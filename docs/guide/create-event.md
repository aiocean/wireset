# Creating Events

Events are a crucial part of the realtime system. They represent significant occurrences or state changes in your application. This guide will walk you through the process of creating and handling events.

## Defining an Event

1. Create a new Go file in the `model` package to define your event structure.
2. Define a struct that represents your event, including all relevant fields.

Example:

```go
package model

type ShopInstalledEvt struct {
    ShopID    string
    Timestamp time.Time
}
```

## Creating an Event Handler

1. Create a new file in the `feature/realtime/event` directory.
2. Define a struct that implements the event handler interface.

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
    return nil
}
```

## Registering the Event Handler

To ensure your event handler is called when the corresponding event occurs:

1. In your application's initialization code, create an instance of your handler.
2. Register the handler with the event bus.

Example:

```go
shopInstalledHandler := &event.ShopInstalledHandler{
    Logger:     logger,
    EventBus:   eventBus,
    CommandBus: commandBus,
}
eventBus.AddHandler(shopInstalledHandler)
```

## Publishing Events

To publish an event when a significant action occurs in your application:

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