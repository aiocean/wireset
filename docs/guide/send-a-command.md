# Send a command

This document describes how to public a command in the system.

## Define command struct

A command is a data structure that represents an action to be performed by the system. It should be a serializable struct.

The name convention of a command struct is `[Verb][Noun]Cmd`. it mean do something to a noun, it is a action.

```go
package model

type CreateInstallMetafieldCmd struct {
	ShopId string `json:"shop_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}
```

## Add CommandBus to the struct of feature

You can declare the `CommandBus` in the struct of your feature, and it will be injected by Wire, dont worry about how to initialize it.

```go
type MyFeature struct {
	CommandBus *cqrs.CommandBus
}
```

## Publish the command

You can publish the command using the `Send` method of the `CommandBus`:

```go
err := f.CommandBus.Send(ctx, &model.CreateInstallMetafieldCmd{
	ShopId: "shop-id",
	Key:    "key",
	Value:  "value",
})
if err != nil {
	// Handle the error
}
```

Notice that a command can only handled by one handler, and execute exactly once.