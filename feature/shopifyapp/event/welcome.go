package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type WelcomeHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
}

func (h *WelcomeHandler) HandlerName() string {
	return "send-welcome-email-on-shop-installed"
}

func (h *WelcomeHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *WelcomeHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.EventBus = eventBus
	h.CommandBus = commandBus
}

func (h *WelcomeHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopInstalledEvt)
	_ = cmd
	// TODO send email
	return nil
}
