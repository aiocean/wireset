# Creating Commands

Commands represent actions or requests that change the state of your application. This guide will walk you through the process of creating and handling commands using Watermill.

## Defining a Command

1. Create a new Go file in the `model` package to define your command structure.
2. Define a struct that represents your command, including all relevant fields.

Example:

```go
package model

type CreateInstallMetafieldCmd struct {
	ShopID       string `json:"shop_id"`
	MetafieldKey   string `json:"metafield_key"`
	MetafieldValue string `json:"metafield_value"`
}
```

## Creating a Command Handler

1. Create a new file in the `feature/<your-feature>/command` directory.
2. Define a struct that implements the `cqrs.CommandHandler` interface. This interface requires three methods:
   - `HandlerName() string`: Returns the name of the handler.
   - `NewCommand() interface{}`: Returns a new instance of the command struct.
   - `Handle(ctx context.Context, cmd interface{}) error`: Contains the logic to handle the command.

Example:

```go
package command

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type CreateInstallMetafieldHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
}

func (h *CreateInstallMetafieldHandler) HandlerName() string {
	return "CreateInstallMetafieldHandler"
}

func (h *CreateInstallMetafieldHandler) NewCommand() interface{} {
	return &model.CreateInstallMetafieldCmd{}
}

func (h *CreateInstallMetafieldHandler) Handle(ctx context.Context, cmd interface{}) error {
	command, ok := cmd.(*model.CreateInstallMetafieldCmd)
	if !ok {
		return ErrInvalidCommandType
	}
	// Handle the command here
	h.Logger.Info("CreateInstallMetafield command handled", zap.String("shop_id", command.ShopID))
	return nil
}
```

## Registering the Command Handler

In your feature struct, you add the command handler as a field, ensure you have the CommandProcessor in the wireset:

```go
type Feature struct {
    // other fields
    CommandProcessor *cqrs.CommandProcessor
	CreateInstallMetafieldHandler *command.CreateInstallMetafieldHandler
}
```

then declare it as an injectable dependency:

```go
var DefaultWireset = wire.NewSet(
	wire.Struct(new(Feature), "*"),
	wire.Struct(new(command.CreateInstallMetafieldHandler), "*"),
)
```

in the feature's init function, you need to register the command handler to the command processor:

```go
func (f *Feature) Init() error {
	if err := f.CommandProcessor.AddHandlers(f.CreateInstallMetafieldHandler); err != nil {
		return err
	}
	return nil
}
```

finally, need to run the wire command to generate the code:

```bash
wire gen ./cmd/server/...
```

## Dispatching Commands

To dispatch a command:

1. Create a new instance of your command struct.
2. Use `commandBus.Send(ctx, command)` to dispatch the command.

Example:

```go
cmd := &model.CreateInstallMetafieldCmd{
	ShopID:       "shop123",
	MetafieldKey:   "key",
	MetafieldValue: "value",
}
err := commandBus.Send(ctx, cmd)
if err != nil {
	// Handle error
}
```

This guide provides a basic overview of creating and handling commands. For more advanced use cases, refer to the Watermill documentation.
