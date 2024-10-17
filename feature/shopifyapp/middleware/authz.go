package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	LocalKeyMyshopifyDomain = "myshopifyDomain"
	LocalKeyAccessToken     = "accessToken"
	LocalKeyShopID          = "shopID"
	LocalKeySid             = "sid"
)

type ShopifyAuthzMiddleware struct {
	configService   *configsvc.ConfigService
	shopifyConfig   *shopifysvc.Config
	tokenRepository *repository.TokenRepository
	shopRepository  *repository.ShopRepository
	cacheSvc        *cachesvc.CacheService
	logger          *zap.Logger
	shopifySvc      *shopifysvc.ShopifyService
}

func NewAuthzController(
	configSvc *configsvc.ConfigService,
	shopifyConfig *shopifysvc.Config,
	tokenRepository *repository.TokenRepository,
	shopRepository *repository.ShopRepository,
	logger *zap.Logger,
	cacheSvc *cachesvc.CacheService,
	shopifySvc *shopifysvc.ShopifyService,
) *ShopifyAuthzMiddleware {
	localLogger := logger.Named("shopifyAuthzMiddleware")
	controller := &ShopifyAuthzMiddleware{
		logger:          localLogger,
		configService:   configSvc,
		shopifyConfig:   shopifyConfig,
		tokenRepository: tokenRepository,
		shopRepository:  shopRepository,
		shopifySvc:      shopifySvc,
		cacheSvc:        cacheSvc,
	}

	return controller
}

// IsAuthRequired check if the path is required authentication
// TODO by hard code this path, it's become not flexible, hard to maintain. Maybe let's feature to register it into the http api registry is better
func (s *ShopifyAuthzMiddleware) IsAuthRequired(path string) bool {
	if strings.HasPrefix(path, "/auth") {
		return false
	}

	if strings.HasPrefix(path, "/metrics") {
		return false
	}

	if strings.HasPrefix(path, "/app") {
		return false
	}

	if strings.HasPrefix(path, "/charge/confirmation") {
		return false
	}

	return true
}

type AuthData struct {
	AccessToken     string
	MyshopifyDomain string
	ShopID          string
	Iss             string
	Dest            string
	Aud             string
	Sub             string
	Exp             int
	Nbf             int
	Iat             int
	Jti             string
	Sid             string
}

// Handle TODO the token which sent from shopify have expired time, we can use this time to cache the authz result, so that we do not need to query database every time
func (s *ShopifyAuthzMiddleware) Handle(c *fiber.Ctx) error {
	if !s.IsAuthRequired(c.OriginalURL()) {
		return c.Next()
	}

	authHeader := c.Get("authorization")
	if authHeader == "" {
		authHeader = c.Params("authorization")
	}

	if authHeader == "" {
		authHeader = c.Query("authorization")
	}

	if authHeader == "" {
		authHeader = gjson.GetBytes(c.Body(), "authorization").String()
	}

	authentication := strings.TrimPrefix(authHeader, "Bearer ")

	if authentication == "" {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized: missing authentication header",
		})
	}

	var claims model.CustomJwtClaims
	token, err := jwt.ParseWithClaims(authentication, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.shopifyConfig.ClientSecret), nil
	})
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	if !token.Valid {
		return fiber.NewError(http.StatusUnauthorized, "Couldn't handle this token")
	}

	cacheKey := "sessionId:" + claims.Jti + ":" + claims.Dest

	if authDataCache, ok := s.cacheSvc.Get(cacheKey); ok {
		s.logger.Info("get auth data from cache", zap.Any("authData", authDataCache))
		authData := authDataCache.(AuthData)
		setLocal(c, &authData)
		return c.Next()
	}

	authData := AuthData{
		Iss:             claims.Iss,
		Dest:            claims.Dest,
		Aud:             claims.Aud,
		Sub:             claims.Sub,
		Exp:             claims.Exp,
		Nbf:             claims.Nbf,
		Iat:             claims.Iat,
		Jti:             claims.Jti,
		Sid:             claims.Sid,
		MyshopifyDomain: strings.Split(claims.Dest, "/")[2],
	}

	// exchange the session token with access token
	accessTokenResponse, err := shopifysvc.ExchangeAccessToken(authData.MyshopifyDomain, s.shopifyConfig.ClientId, s.shopifyConfig.ClientSecret, authentication)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	authData.AccessToken = accessTokenResponse.AccessToken

	shopifyClient := s.shopifySvc.GetShopifyClient(authData.MyshopifyDomain, authData.AccessToken)
	shop, err := shopifyClient.GetShopDetails()
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	authData.ShopID = shop.ID

	s.logger.Info("set auth data to cache", zap.Any("authData", authData))
	s.cacheSvc.SetWithTTL(cacheKey, authData, 3*time.Minute)

	setLocal(c, &authData)
	return c.Next()
}

func setLocal(c *fiber.Ctx, authData *AuthData) {
	c.Locals(LocalKeyMyshopifyDomain, authData.MyshopifyDomain)
	c.Locals(LocalKeyAccessToken, authData.AccessToken)
	c.Locals(LocalKeyShopID, authData.ShopID)
	c.Locals(LocalKeySid, authData.Sid)
}

func GetMyShopifyDomain(c *fiber.Ctx) (string, bool) {
	myshopifyDomain, ok := c.Locals(LocalKeyMyshopifyDomain).(string)
	return myshopifyDomain, ok
}

func GetAccessToken(c *fiber.Ctx) (string, bool) {
	accessToken, ok := c.Locals(LocalKeyAccessToken).(string)
	return accessToken, ok
}

func GetShopID(c *fiber.Ctx) (string, bool) {
	shopID, ok := c.Locals(LocalKeyShopID).(string)
	return shopID, ok
}

func GetSid(c *fiber.Ctx) (string, bool) {
	sid, ok := c.Locals(LocalKeySid).(string)
	return sid, ok
}
