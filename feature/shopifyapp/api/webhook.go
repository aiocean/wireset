package api

import (
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	ShopRepo *repository.ShopRepository
	EventBus *cqrs.EventBus
	FiberApp *fiber.App
}

func (s *WebhookHandler) Uninstalled(c *fiber.Ctx) error {
	shop := c.Query("shop")

	// Fetch additional data from the request or database
	shopDetails, err := s.ShopRepo.GetByDomain(c.UserContext(), shop)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to fetch shop ID")
	}

	uninstalledEvt := &model.ShopUninstalledEvt{
		MyshopifyDomain: shop,
		ShopID:          shopDetails.ID,
		UninstalledAt:   time.Now(),
		Reason:          c.Query("reason", "Unknown"), // Assuming reason is provided in query params, default to "Unknown"
	}

	if err := s.EventBus.Publish(c.UserContext(), uninstalledEvt); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to publish uninstall event")
	}

	return c.SendStatus(http.StatusOK)
}
