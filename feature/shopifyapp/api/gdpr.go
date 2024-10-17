package api

/*

all 3 endpoints are required for the GDPR compliance. All shopify app must implement these endpoints.
for current implementation, we are just returning 200 status code.
*/

import "github.com/gofiber/fiber/v2"

type GdprHandler struct {
}

func (g *GdprHandler) CustomerDataRequest(ctx *fiber.Ctx) error {
	
	return ctx.SendStatus(200)
}

func (g *GdprHandler) CustomerRedact(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func (g *GdprHandler) ShopRedact(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}
