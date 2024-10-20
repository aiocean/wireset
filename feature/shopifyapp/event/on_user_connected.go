package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/command"
	realtimemodel "github.com/aiocean/wireset/feature/realtime/models"
	"github.com/aiocean/wireset/feature/shopifyapp/models"
	"github.com/aiocean/wireset/feature/shopifyapp/plan"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/pkg/errors"
)

type OnUserConnectedHandler struct {
	CommandBus *cqrs.CommandBus
	TokenRepo  *repository.TokenRepository
	ShopifySvc *shopifysvc.ShopifyService
	ShopRepo   *repository.ShopRepository
	PlanRepo   *plan.PlanRepository
}

func (h *OnUserConnectedHandler) HandlerName() string {
	return "OnUserConnectedHandler"
}

func (h *OnUserConnectedHandler) NewEvent() interface{} {
	return &realtimemodel.UserJoinedEvt{}
}

const FreePlanName = "Free"
const FreePlanID = "-1"

func (h *OnUserConnectedHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*realtimemodel.UserJoinedEvt)
	shopifyDomain := evt.RoomID

	shop, err := h.ShopRepo.GetByDomain(ctx, shopifyDomain)
	if err != nil {
		return err
	}

	accessToken, err := h.TokenRepo.GetToken(ctx, shop.ID)
	if err != nil {
		return err
	}

	shopifyClient := h.ShopifySvc.GetShopifyClient(shopifyDomain, accessToken.AccessToken)

	activeSubscription, err := shopifyClient.GetActiveSubscriptions()
	if err != nil {
		if errors.Is(err, shopifysvc.ErrorSubscriptionNotFound) {
			// it is not an error, just means there is no active subscription
			return h.CommandBus.Send(ctx, &command.SendWsMessageCmd{
				RoomID:   evt.RoomID,
				Username: evt.UserName,
				Payload: realtimemodel.WebsocketMessage[models.SetActivateSubscriptionPayload]{
					Topic: models.TopicSetActivateSubscription,
					Payload: models.SetActivateSubscriptionPayload{
						ID:     FreePlanID,
						Status: models.SubscriptionStatusActive,
						Name:      FreePlanName,
					},
				},
			})
		}
		return err
	}

	plan, err := h.PlanRepo.GetPlanByName(activeSubscription.Name)
	if err != nil {
		return err
	}

	return h.CommandBus.Send(ctx, &command.SendWsMessageCmd{
		RoomID:   evt.RoomID,
		Username: evt.UserName,
		Payload: realtimemodel.WebsocketMessage[models.SetActivateSubscriptionPayload]{
			Topic: models.TopicSetActivateSubscription,
			Payload: models.SetActivateSubscriptionPayload{
				ID:     activeSubscription.ID,
				Status: activeSubscription.Status,
				TrialDays: activeSubscription.TrialDays,
				Name:      activeSubscription.Name,
				Plan:      plan,
			},
		},
	})
}
