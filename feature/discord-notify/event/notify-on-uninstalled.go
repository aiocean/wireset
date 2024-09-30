package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/discord-notify/config"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type NotifyDiscordOnUninstallHandler struct {
	logger     *zap.Logger
	config     *config.Config
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

// NewNotifyDiscordOnUninstallHandler creates a new NotifyDiscordOnUninstallHandler.
func NewNotifyDiscordOnUninstallHandler(
	logger *zap.Logger,
	config *config.Config,
) *NotifyDiscordOnUninstallHandler {
	handler := &NotifyDiscordOnUninstallHandler{
		logger: logger,
		config: config,
	}

	return handler
}

func (h *NotifyDiscordOnUninstallHandler) HandlerName() string {
	return "NotifyDiscordOnUninstallHandler"
}

func (h *NotifyDiscordOnUninstallHandler) NewEvent() interface{} {
	return &model.ShopUninstalledEvt{}
}

func (h *NotifyDiscordOnUninstallHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

// Handle handles the ShopUninstalledEvt event by sending a notification to a Discord webhook.
// It constructs a Discord message containing the uninstalled shop's domain and sends it
// as a POST request to the configured webhook URL.
//
// If an error occurs during the process, it logs the error and returns it.
func (h *NotifyDiscordOnUninstallHandler) Handle(ctx context.Context, event interface{}) error {
	cmd, ok := event.(*model.ShopUninstalledEvt)
	if !ok {
		return fmt.Errorf("invalid event type: expected *model.ShopUninstalledEvt, got %T", event)
	}

	payload := map[string]string{
		"content": fmt.Sprintf("Shop uninstalled: %s", cmd.MyshopifyDomain),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Error marshaling payload", zap.Error(err))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.config.NewInstallWebhook, bytes.NewReader(jsonPayload))
	if err != nil {
		h.logger.Error("Error creating request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		h.logger.Error("Error sending Discord notification", zap.Error(err))
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(res.Body)
		h.logger.Error("Discord API returned an error",
			zap.Int("status_code", res.StatusCode),
			zap.String("response", string(body)))
		return fmt.Errorf("discord API error: status code %d", res.StatusCode)
	}

	h.logger.Info("Discord notification sent successfully",
		zap.String("shop", cmd.MyshopifyDomain))
	return nil
}
