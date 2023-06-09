package handler

import (
	"firebase.google.com/go/auth"
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHandler struct {
	ShopRepo       *repository.ShopRepository
	ShopifyService *shopifysvc.ShopifyService
	ConfigSvc      *configsvc.ConfigService
	ShopifyConfig  *shopifysvc.Config
	ShopifyApp     *goshopify.App
	TokenRepo      *repository.TokenRepository
	PubsubSvc      *pubsub.Pubsub
	LogSvc         *zap.Logger
	CacheSvc       *cachesvc.CacheService
	FireAuth       *auth.Client
}

func (s *AuthHandler) Register(fiberApp *fiber.App) {
	authGroup := fiberApp.Group("/auth")
	{
		authGroup.Get("shopify/login-callback", s.loginCallback)
		authGroup.Get("shopify/checkin", s.checkin)
	}
}
