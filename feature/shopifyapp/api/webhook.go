package api

import (
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/repository"
	"github.com/tidwall/gjson"

	"github.com/aiocean/wireset/feature/shopifyapp/event/model"
	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	ShopRepo *repository.ShopRepository
	TokenRepo *repository.TokenRepository
	EventBus *cqrs.EventBus
	FiberApp *fiber.App
}

func (s *WebhookHandler) OnWebhookOccurred(c *fiber.Ctx) error {
	/*
	headers:
	X-Shopify-Topic: `orders/create`
	X-Shopify-Hmac-Sha256: `XWmrwMey6OsLMeiZKwP4FppHH3cmAiiJJAweH5Jo4bM=`
	X-Shopify-Shop-Domain: `{shop}.myshopify.com`
	X-Shopify-API-Version: `2024-10`
	X-Shopify-Webhook-Id: `b54557e4-bdd9-4b37-8a5f-bf7d70bcd043`
	X-Shopify-Triggered-At: `2023-03-29T18:00:27.877041743Z`
	X-Shopify-Event-Id: `98880550-7158-44d4-b7cd-2c97c8a091b5`
	*/

	myshopifyDomain := c.Get("X-Shopify-Shop-Domain")
	topic := c.Get("X-Shopify-Topic")
	gBody := gjson.ParseBytes( c.Body())

	shop, err := s.ShopRepo.GetByDomain(c.UserContext(), myshopifyDomain)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to get shop")
	}

	switch topic {
	case "app/uninstalled":
		uninstalledEvt := &model.ShopUninstalledEvt{
			ShopID:          shop.ID,
			UninstalledAt:   time.Now(),
			Reason:          "Unknown",
			MyshopifyDomain: myshopifyDomain,
		}

		if err := s.EventBus.Publish(c.UserContext(), uninstalledEvt); err != nil {
			return fiber.NewError(http.StatusInternalServerError, "Failed to publish uninstall event")
		}

	case "orders/create":
		token, err := s.TokenRepo.GetToken(c.UserContext(), shop.ID)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "Failed to get token")
		}

		orderCreatedEvt := &model.OrderCreatedEvt{
			ShopID: shop.ID,
			MyshopifyDomain: myshopifyDomain,
			AccessToken:     token.AccessToken,
			Order: model.Order{
				ID: gBody.Get("id").String(),
			},
		}

		if err := s.EventBus.Publish(c.UserContext(), orderCreatedEvt); err != nil {
			return fiber.NewError(http.StatusInternalServerError, "Failed to publish order created event")
		}
	}

	return c.SendStatus(http.StatusOK)
}
