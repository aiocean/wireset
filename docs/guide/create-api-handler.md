# Create API Handler

This guide will walk you through the process of creating an HTTP API handler, registering it, and integrating it into your feature struct.

## Defining an API Handler

1. Create a new Go file in your feature's directory (e.g., `feature/<your-feature>/api/`).
2. Define a struct for your handler. This struct will typically hold dependencies like loggers or service clients.
3. Define a handler function within your struct. This function should accept a `*fiber.Ctx` argument and return an `error`.

Example:

**feature/example/api/handler.go

```go
package api

import (
	"github.com/aiocean/wireset/fiberapp"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type ExampleHandler struct {
	Logger *zap.Logger
}

func (h *ExampleHandler) GetExample(c *fiber.Ctx) error {
	exampleID := c.Params("id")
	// Implement your handler logic here
	h.Logger.Info("Fetching example", zap.String("id", exampleID))
	return c.SendString("Example: " + exampleID)
}
```

## Registering the API Handler

1. In your feature's main file (e.g., `feature/<your-feature>/feature.go`), obtain an instance of `*fiberapp.Registry` (usually injected via Wire).
2. Create a new `*fiberapp.HttpHandler` instance, specifying the HTTP method, path, and a slice of handler functions.
3. Use the `AddHttpHandlers` method of the `*fiberapp.Registry` to register your handler.

Example:

**/feature/example/feature.go**

```go
package example

import (
	"github.com/aiocean/wireset/fiberapp"
	"github.com/google/wire"
)

type Feature struct {
	// ... other fields
	HttpRegistry *fiberapp.Registry
	API      *api.ExampleHandler
}

var DefaultWireset = wire.NewSet(
	// ... other dependencies
	wire.Struct(new(Feature), "*"),
	wire.Struct(new(api.ExampleHandler), "*"),
)

func (f *Feature) Init() error {
	// ... other initialization logic

	f.HttpRegistry.AddHttpHandlers(
		&fiberapp.HttpHandler{
			Method:   fiber.MethodGet,
			Path:     "/examples/:id",
			Handlers: []fiber.Handler{f.API.GetExample},
		},
	)
	return nil
}
```

## Summary

By following these steps, you can create well-structured API handlers, register them with the central registry, and manage them within your feature's lifecycle. This approach promotes code organization and maintainability as your application grows.