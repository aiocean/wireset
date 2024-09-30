package realtime

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/models"
	realtimemodel "github.com/aiocean/wireset/feature/realtime/models"
	shopifycommand "github.com/aiocean/wireset/feature/shopifyapp/command"
)

type SendMessageToShopInput struct {
	MyShopifyDomain string
	Topic models.WebsocketTopic
	Payload interface{}
}

func SendMessageToShop(ctx context.Context, commandBus cqrs.CommandBus, input *SendMessageToShopInput) error {
	return commandBus.Send(ctx, &shopifycommand.SendMessage{
		MyShopifyDomain: input.MyShopifyDomain,
		Payload: realtimemodel.WebsocketMessage[any]{
			Topic: input.Topic,
			Payload: input.Payload,
		},
	})
}
