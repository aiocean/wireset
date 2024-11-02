package command

import (
	"context"

	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/shopifysvc"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type InstallWebhookHandler struct {
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
	ShopifySvc *shopifysvc.ShopifyService
}

// NewInstallWebhookHandler creates a new InstallWebhookHandler.
func NewInstallWebhookHandler(shopifySvc *shopifysvc.ShopifyService) *InstallWebhookHandler {
	return &InstallWebhookHandler{
		ShopifySvc: shopifySvc,
	}
}

func (h *InstallWebhookHandler) HandlerName() string {
	return "core.InstallWebhookCmd"
}

func (h *InstallWebhookHandler) NewCommand() interface{} {
	return &model.InstallWebhookCmd{}
}

func (h *InstallWebhookHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.EventBus = eventBus
	h.CommandBus = commandBus
}

func (h *InstallWebhookHandler) Handle(ctx context.Context, cmdItf interface{}) error {
	// TODO: migrated to shopify managed webhooks
	return nil
}
