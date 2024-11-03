package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/tidwall/gjson"

	"github.com/aiocean/wireset/feature/shopifyapp/event/model"
	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	ShopRepo  *repository.ShopRepository
	TokenRepo *repository.TokenRepository
	EventBus  *cqrs.EventBus
	FiberApp  *fiber.App
	ShopifyApp *goshopify.App
}

func (s *WebhookHandler) VerifyWebhook(c *fiber.Ctx) bool {
	shopifySha256 := c.Get("X-Shopify-Hmac-Sha256")
	actualMac := []byte(shopifySha256)

	mac := hmac.New(sha256.New, []byte(s.ShopifyApp.ApiSecret))
	mac.Write(c.Body())
	macSum := mac.Sum(nil)
	expectedMac := []byte(base64.StdEncoding.EncodeToString(macSum))

	return hmac.Equal(actualMac, expectedMac)
}

func (s *WebhookHandler) handleAppUninstalled(c *fiber.Ctx, shop *shopifysvc.Shop, myshopifyDomain string) error {
	uninstalledEvt := &model.ShopUninstalledEvt{
		ShopID:          shop.ID,
		UninstalledAt:   time.Now(),
		Reason:          "Unknown",
		MyshopifyDomain: myshopifyDomain,
	}

	if err := s.EventBus.Publish(c.UserContext(), uninstalledEvt); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to publish uninstall event")
	}
	return nil
}

func (s *WebhookHandler) handleOrderCreated(c *fiber.Ctx, shop *shopifysvc.Shop, myshopifyDomain string, gBody gjson.Result) error {
	token, err := s.TokenRepo.GetToken(c.UserContext(), shop.ID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to get token")
	}

	orderCreatedEvt := &model.OrderCreatedEvt{
		ShopID:          shop.ID,
		MyshopifyDomain: myshopifyDomain,
		AccessToken:     token.AccessToken,
		Order: model.Order{
			ID: gBody.Get("id").String(),
		},
	}

	if err := s.EventBus.Publish(c.UserContext(), orderCreatedEvt); err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to publish order created event")
	}
	return nil
}

func (s *WebhookHandler) OnWebhookOccurred(c *fiber.Ctx) error {
	if !s.VerifyWebhook(c) {
		return fiber.NewError(http.StatusUnauthorized, "Invalid HMAC")
	}

	myshopifyDomain := c.Get("X-Shopify-Shop-Domain")
	topic := c.Get("X-Shopify-Topic")
	gBody := gjson.ParseBytes(c.Body())

	shop, err := s.ShopRepo.GetByDomain(c.UserContext(), myshopifyDomain)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Failed to get shop")
	}

	switch topic {
	case "app/uninstalled":
		if err := s.handleAppUninstalled(c, shop, myshopifyDomain); err != nil {
			return err
		}
	case "orders/create":
		if err := s.handleOrderCreated(c, shop, myshopifyDomain, gBody); err != nil {
			return err
		}
	}

	return c.SendStatus(http.StatusOK)
}
