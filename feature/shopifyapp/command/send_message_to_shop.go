package command

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	realtimecommand "github.com/aiocean/wireset/feature/realtime/command"
	realtimemodel "github.com/aiocean/wireset/feature/realtime/models"
	"github.com/aiocean/wireset/feature/shopifyapp/wsresolver"
	"github.com/aiocean/wireset/shopifysvc"
)

type SendMessage struct {
	MyShopifyDomain string                       `json:"myShopifyDomain"`
	Topic           realtimemodel.WebsocketTopic `json:"topic"`
	Payload         interface{}                  `json:"payload"`
}

type SendMessageHandler struct {
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
	ShopifySvc *shopifysvc.ShopifyService
}

func (h *SendMessageHandler) HandlerName() string {
	return "core.sendMessageToShop"
}

func (h *SendMessageHandler) NewCommand() interface{} {
	return &SendMessage{}
}

func (h *SendMessageHandler) Handle(ctx context.Context, cmdItf interface{}) error {
	cmd := cmdItf.(*SendMessage)

	h.CommandBus.Send(ctx, &realtimecommand.SendWsMessageCmd{
		RoomID:   cmd.MyShopifyDomain,
		Username: wsresolver.DefaultUsername,
		Payload: realtimemodel.WebsocketMessage[any]{
			Topic:   cmd.Topic,
			Payload: cmd.Payload,
		},
	})

	return nil
}
