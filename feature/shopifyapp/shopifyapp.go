package shopifyapp

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/registry"
	"github.com/aiocean/wireset/feature/shopifyapp/api"
	"github.com/aiocean/wireset/feature/shopifyapp/command"
	"github.com/aiocean/wireset/feature/shopifyapp/event"
	"github.com/aiocean/wireset/feature/shopifyapp/middleware"
	"github.com/aiocean/wireset/feature/shopifyapp/plan"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/poolsvc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureCore), "*"),

	command.NewSetShopStateHandler,

	wire.Struct(new(plan.Registry), "*"),
	plan.NewPlanRepository,

	wire.Struct(new(command.SendMessageHandler), "*"),

	wire.Struct(new(event.CreateUserHandler), "*"),
	wire.Struct(new(event.WelcomeHandler), "*"),
	wire.Struct(new(event.OnUserConnectedHandler), "*"),
	wire.Struct(new(event.OnCheckedInHandler), "*"),

	middleware.NewAuthzController,

	wire.Struct(new(api.AuthHandler), "*"),
	wire.Struct(new(api.WebhookHandler), "*"),
	poolsvc.DefaultWireset,
)

type FeatureCore struct {
	// Command Handlers
	SetShopStateCmdHandler *command.SetShopStateHandler
	SendMessageCmdHandler  *command.SendMessageHandler

	// Event Handlers
	ShopInstalledEvtHandler *event.CreateUserHandler
	WelcomeEvtHandler       *event.WelcomeHandler
	OnUserConnectedHandler  *event.OnUserConnectedHandler
	OnCheckedInHandler      *event.OnCheckedInHandler

	// Middleware
	AuthzMiddleware *middleware.ShopifyAuthzMiddleware

	// API Handlers
	AuthHandler    *api.AuthHandler
	WebhookHandler *api.WebhookHandler

	// Processors
	EventProcessor   *cqrs.EventProcessor
	CommandProcessor *cqrs.CommandProcessor

	// Registries
	HttpRegistry *fiberapp.Registry
	WsRegistry   *registry.HandlerRegistry
}

func (f *FeatureCore) Name() string {
	return "shopifyapp"
}

func (f *FeatureCore) Init() error {

	// Register command handlers
	if err := f.CommandProcessor.AddHandlers(
		f.SetShopStateCmdHandler,
		f.SendMessageCmdHandler,
	); err != nil {
		return err
	}

	// Register event handlers
	if err := f.EventProcessor.AddHandlers(
		f.ShopInstalledEvtHandler,
		f.WelcomeEvtHandler,
		f.OnUserConnectedHandler,
		f.OnCheckedInHandler,
	); err != nil {
		return err
	}

	f.HttpRegistry.AddHttpMiddleware("/", f.AuthzMiddleware.Handle)

	f.HttpRegistry.AddHttpHandlers(
		// in the future, we do not have login callback
		&fiberapp.HttpHandler{
			Method:   fiber.MethodGet,
			Path:     "/auth/shopify/login-callback",
			Handlers: []fiber.Handler{f.AuthHandler.LoginCallback},
		},
		&fiberapp.HttpHandler{
			Method: fiber.MethodGet,
			Path:   "/auth/shopify/checkin",
			Handlers: []fiber.Handler{
				f.AuthHandler.Checkin,
			},
		},
		&fiberapp.HttpHandler{
			Method:   fiber.MethodPost,
			Path:     "/webhooks",
			Handlers: []fiber.Handler{f.WebhookHandler.OnWebhookOccurred},
		},
	)

	return nil
}
