package api

import (
	"github.com/gofiber/fiber/v2"
)

// LoginCallback is a api to handle login callback from shopify
// basically, with the new shopify authh flow, we don't need to do anything here, just redirect back to the app then FE will call checkin api.
func (s *AuthHandler) LoginCallback(ctx *fiber.Ctx) error {
	shopName := ctx.Query("shop")
	redirectUrl := "https://" + shopName + "/admin/apps/" + s.ShopifyConfig.ClientId

	return ctx.Redirect(redirectUrl)
}
