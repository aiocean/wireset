# Creating Events

Events are a crucial part of the realtime system. They represent significant occurrences or state changes in your application. This guide will walk you through the process of creating and handling events using Watermill.

## Defining an Event

1. Create a new Go file in the `model` package to define your event structure.
2. Define a struct that represents your event, including all relevant fields.

Example:

```go
package model

import "time"

type ShopInstalledEvt struct {
	ShopID    string
	Timestamp time.Time
}
```

## Creating an Event Handler

1. Create a new file in the `feature/<your-feature>/event` directory.
2. Define a struct that implements the `cqrs.EventHandler` interface. This interface requires three methods:
   - `HandlerName() string`: Returns the name of the handler.
   - `NewEvent() interface{}`: Returns a new instance of the event struct.
   - `Handle(ctx context.Context, event interface{}) error`: Contains the logic to handle the event.

Example:

```go
package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type ShopInstalledHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
}

func (h *ShopInstalledHandler) HandlerName() string {
	return "ShopInstalledHandler"
}

func (h *ShopInstalledHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *ShopInstalledHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*model.ShopInstalledEvt)
	// Handle the event here
	h.Logger.Info("Shop installed", zap.String("shop_id", evt.ShopID))
	return nil
}
```

## Registering the Event Handler

In your feature struct, you add the event handler as a field, ensure you have the EventProcessor in the wireset:

```go
type Feature struct {
    // other fields
    EventProcessor *cqrs.EventProcessor
	ShopInstalledHandler *event.ShopInstalledHandler
}
```

then declare it as an injectable dependency:

```go
var DefaultWireset = wire.NewSet(
	wire.Struct(new(Feature), "*"),
	wire.Struct(new(event.ShopInstalledHandler), "*"),
)
```

in the feature's init function, you need to register the event handler to the event processor:

```go
func (f *Feature) Init() error {
	if err := f.EventProcessor.AddHandlers(f.ShopInstalledHandler); err != nil {
		return err
	}
	return nil
}
```

finally, need to run the wire command to generate the code:

```bash
wire gen ./cmd/server/...
```

## Publishing Events

To publish an event when a significant action occurs in your application:

1. Create a new instance of your event struct.
2. Use `eventBus.Publish(ctx, event)` to publish the event.

Example:

```go
evt := &model.ShopInstalledEvt{
	ShopID:    "shop123",
	Timestamp: time.Now(),
}
err := eventBus.Publish(ctx, evt)
if err != nil {
	// Handle error
}
```

This guide provides a basic overview of creating and handling events. For more advanced use cases, refer to the Watermill documentation.
