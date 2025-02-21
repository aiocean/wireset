package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/feature/shopifyapp/models"
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

	// Cache configuration
	defaultCacheTTL = 3 * time.Minute
	cacheKeyPrefix  = "sessionId"
)

var (
	// ErrMissingToken represents a missing authentication token error
	ErrMissingToken = errors.New("missing authentication token")
	// ErrInvalidToken represents an invalid authentication token error
	ErrInvalidToken = errors.New("invalid authentication token")
)

// Config represents the middleware configuration
type Config struct {
	PublicPaths []string
	CacheTTL    time.Duration
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		PublicPaths: []string{
			"/auth",
			"/metrics",
			"/app",
			"/webhooks",
		},
		CacheTTL: defaultCacheTTL,
	}
}

type ShopifyAuthzMiddleware struct {
	configService   *configsvc.ConfigService
	shopifyConfig   *shopifysvc.Config
	tokenRepository *repository.TokenRepository
	shopRepository  *repository.ShopRepository
	cacheSvc        *cachesvc.CacheService
	logger          *zap.Logger
	shopifySvc      *shopifysvc.ShopifyService
	config          Config
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
		config:          DefaultConfig(),
	}

	return controller
}

// IsAuthRequired check if the path requires authentication
func (s *ShopifyAuthzMiddleware) IsAuthRequired(path string) bool {
	for _, publicPath := range s.config.PublicPaths {
		if strings.HasPrefix(path, publicPath) {
			return false
		}
	}
	return true
}

// Handle processes the authentication middleware
func (s *ShopifyAuthzMiddleware) Handle(c *fiber.Ctx) error {
	if !s.IsAuthRequired(c.OriginalURL()) {
		return c.Next()
	}

	token, err := s.extractToken(c)
	if err != nil {
		return s.unauthorizedResponse(c, err)
	}

	claims, err := s.parseToken(token)
	if err != nil {
		return s.unauthorizedResponse(c, err)
	}

	cacheKey := s.getCacheKey(claims)
	if authData, ok := s.getCachedAuthData(cacheKey); ok {
		setLocal(c, authData)
		return c.Next()
	}

	authData, err := s.buildAuthData(claims, token)
	if err != nil {
		return s.unauthorizedResponse(c, err)
	}

	s.cacheAuthData(cacheKey, authData)
	setLocal(c, authData)
	return c.Next()
}

// extractToken extracts the token from various sources
func (s *ShopifyAuthzMiddleware) extractToken(c *fiber.Ctx) (string, error) {
	sources := []func() string{
		func() string { return c.Get("authorization") },
		func() string { return c.Params("authorization") },
		func() string { return c.Query("authorization") },
		func() string { return gjson.GetBytes(c.Body(), "authorization").String() },
	}

	for _, source := range sources {
		if token := strings.TrimPrefix(source(), "Bearer "); token != "" {
			return token, nil
		}
	}

	return "", ErrMissingToken
}

// parseToken parses and validates the JWT token
func (s *ShopifyAuthzMiddleware) parseToken(tokenString string) (*model.CustomJwtClaims, error) {
	var claims model.CustomJwtClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// TODO: Add signing method validation if needed
		return []byte(s.shopifyConfig.ClientSecret), nil
	})

	if err != nil {
		s.logger.Error("failed to parse token", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// getCacheKey generates a cache key for the auth data
func (s *ShopifyAuthzMiddleware) getCacheKey(claims *model.CustomJwtClaims) string {
	return fmt.Sprintf("%s:%s:%s", cacheKeyPrefix, claims.Jti, claims.Dest)
}

// getCachedAuthData retrieves auth data from cache
func (s *ShopifyAuthzMiddleware) getCachedAuthData(cacheKey string) (*models.AuthData, bool) {
	if authDataCache, ok := s.cacheSvc.Get(cacheKey); ok {
		return authDataCache.(*models.AuthData), true
	}
	return nil, false
}

// buildAuthData creates AuthData from claims and external services
func (s *ShopifyAuthzMiddleware) buildAuthData(claims *model.CustomJwtClaims, token string) (*models.AuthData, error) {
	authData := &models.AuthData{
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

	if err := s.enrichAuthData(authData, token); err != nil {
		return nil, err
	}

	return authData, nil
}

// enrichAuthData enriches auth data with external service data
func (s *ShopifyAuthzMiddleware) enrichAuthData(authData *models.AuthData, token string) error {
	accessTokenResponse, err := shopifysvc.ExchangeAccessToken(
		authData.MyshopifyDomain,
		s.shopifyConfig.ClientId,
		s.shopifyConfig.ClientSecret,
		token,
	)
	if err != nil {
		s.logger.Error("failed to exchange access token", zap.Error(err))
		return fmt.Errorf("failed to exchange token: %w", err)
	}

	authData.AccessToken = accessTokenResponse.AccessToken

	shopifyClient := s.shopifySvc.GetShopifyClient(authData.MyshopifyDomain, authData.AccessToken)
	shop, err := shopifyClient.GetShopDetails()
	if err != nil {
		s.logger.Error("failed to get shop details", zap.Error(err))
		return fmt.Errorf("failed to get shop details: %w", err)
	}

	authData.ShopID = shop.ID
	return nil
}

// cacheAuthData stores auth data in cache
func (s *ShopifyAuthzMiddleware) cacheAuthData(cacheKey string, authData *models.AuthData) {
	s.logger.Debug("caching auth data", zap.String("cacheKey", cacheKey))
	s.cacheSvc.SetWithTTL(cacheKey, *authData, s.config.CacheTTL)
}

// unauthorizedResponse returns a standardized unauthorized response
func (s *ShopifyAuthzMiddleware) unauthorizedResponse(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
		Message: fmt.Sprintf("Unauthorized: %v", err),
	})
}

func setLocal(c *fiber.Ctx, authData *models.AuthData) {
	c.Locals(LocalKeyMyshopifyDomain, authData.MyshopifyDomain)
	c.Locals(LocalKeyAccessToken, authData.AccessToken)
	c.Locals(LocalKeyShopID, authData.ShopID)
	c.Locals(LocalKeySid, authData.Sid)
}

// Helper functions to get values from context
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
