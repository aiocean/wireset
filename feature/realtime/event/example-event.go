package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type ExampleHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
}

// HandlerName returns the name of the handler. The name should be unique.
func (h *ExampleHandler) HandlerName() string {

	return "EventExampleHandler"
}

// NewEvent returns a new event instance. the instance is used to determine which event the handler is listening to.
func (h *ExampleHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

// Handle is the method that will be called when an event is received.
func (h *ExampleHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopInstalledEvt)
	_ = cmd
	return nil
}
