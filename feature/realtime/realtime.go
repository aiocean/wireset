// Known issues:
// - Service này có thể scale lên nhiều pod, nếu user connect ws vào pod A, nhưng nếu pod B nhận được command SendMessage, pod B sẽ thực hiện send message về cho user bằng ws, nhưng pod B không có kết nối. sẽ dẫn tới lỗi.

package realtime

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/api"
	"github.com/aiocean/wireset/feature/realtime/command"
	"github.com/aiocean/wireset/feature/realtime/models"
	"github.com/aiocean/wireset/feature/realtime/registry"
	"github.com/aiocean/wireset/feature/realtime/room"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureRealtime), "*"),
	api.NewWebsocketHandler,
	room.NewRoomManager,
	wire.Struct(new(command.SendWsMessageHandler), "*"),
	registry.NewWsHandlerRegistry,
)

type FeatureRealtime struct {
	HttpRegistry     *fiberapp.Registry
	WebsocketHandler *api.WebsocketHandler

	CommandProcessor *cqrs.CommandProcessor
	EventProcessor   *cqrs.EventProcessor

	EventBus *cqrs.EventBus

	SendWsMessageHandler *command.SendWsMessageHandler
}

func NewFeatureRealtime(
	httpRegistry *fiberapp.Registry,
	websocketHandler *api.WebsocketHandler,
	commandProcessor *cqrs.CommandProcessor,
	eventProcessor *cqrs.EventProcessor,
	eventBus *cqrs.EventBus,
	sendWsMessageHandler *command.SendWsMessageHandler,
) *FeatureRealtime {
	return &FeatureRealtime{
		HttpRegistry:         httpRegistry,
		WebsocketHandler:     websocketHandler,
		CommandProcessor:     commandProcessor,
		EventProcessor:       eventProcessor,
		EventBus:             eventBus,
		SendWsMessageHandler: sendWsMessageHandler,
	}
}

func (f *FeatureRealtime) Name() string {
	return "realtime-feature"
}

func (f *FeatureRealtime) Init() error {
	if err := f.CommandProcessor.AddHandlers(f.SendWsMessageHandler); err != nil {
		return errors.Wrap(err, "failed to add command handler")
	}

	f.HttpRegistry.AddHttpMiddleware(models.WebsocketEndpoint, f.WebsocketHandler.Upgrade)
	f.HttpRegistry.AddHttpHandlers(
		&fiberapp.HttpHandler{
			Method: fiber.MethodGet,
			Path:   models.WebsocketEndpoint,
			Handlers: []fiber.Handler{
				websocket.New(f.WebsocketHandler.Handle),
			},
		},
	)

	return nil
}
