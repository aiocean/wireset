package event

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	eventmodel "github.com/aiocean/wireset/feature/shopifyapp/event/model"
	"github.com/aiocean/wireset/shopifysvc"
	"go.uber.org/zap"
)

type SubscriptionCheckHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
	ShopifySvc *shopifysvc.ShopifyService
}

func (h *SubscriptionCheckHandler) HandlerName() string {
	return "check-subscription-handler"
}

func (h *SubscriptionCheckHandler) NewEvent() interface{} {
	return &eventmodel.ShopLoggedInEvt{}
}

func (h *SubscriptionCheckHandler) Handle(ctx context.Context, event interface{}) error {
	evt, ok := event.(*eventmodel.ShopLoggedInEvt)
	if !ok {
		return fmt.Errorf("invalid event type: expected *eventmodel.ShopLoggedInEvt, got %T", event)
	}

	shopifyClient := h.ShopifySvc.GetShopifyClient(evt.MyshopifyDomain, evt.AccessToken)

	activeSubscription, err := shopifyClient.GetActiveSubscriptions()
	if err != nil {
		if errors.Is(err, shopifysvc.ErrorSubscriptionNotFound) {
			h.Logger.Info("No active subscription found",
				zap.String("myshopifyDomain", evt.MyshopifyDomain),
				zap.String("shopID", evt.ShopID))
			return h.publishNoSubscriptionEvent(ctx, evt)
		}

		// Check for API errors
		var graphQLErr *shopifysvc.GraphQLError
		if errors.As(err, &graphQLErr) {
			h.Logger.Error("GraphQL API error occurred",
				zap.String("myshopifyDomain", evt.MyshopifyDomain),
				zap.String("shopID", evt.ShopID),
				zap.Error(graphQLErr))
			// Handle GraphQL error specifically if needed
			return err
		}

		h.Logger.Error("Failed to get active subscription",
			zap.String("myshopifyDomain", evt.MyshopifyDomain),
			zap.String("shopID", evt.ShopID),
			zap.Error(err))
		return err
	}

	h.Logger.Info("Active subscription found",
		zap.String("myshopifyDomain", evt.MyshopifyDomain),
		zap.String("shopID", evt.ShopID),
		zap.Any("subscription", activeSubscription))
	return nil
}

func (h *SubscriptionCheckHandler) publishNoSubscriptionEvent(ctx context.Context, evt *eventmodel.ShopLoggedInEvt) error {
	cmd := &eventmodel.ShopWithoutSubscriptionFoundEvt{
		MyshopifyDomain: evt.MyshopifyDomain,
		AccessToken:     evt.AccessToken,
		ShopID:          evt.ShopID,
	}

	if err := h.EventBus.Publish(ctx, cmd); err != nil {
		h.Logger.Error("Failed to publish ShopWithoutSubscriptionFoundEvt",
			zap.String("myshopifyDomain", evt.MyshopifyDomain),
			zap.String("shopID", evt.ShopID),
			zap.Error(err))
		return err
	}

	return nil
}
