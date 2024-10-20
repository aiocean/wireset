# Realtime Wireset

The Realtime Wireset provides a pre-configured setup for real-time communication in your applications using WebSockets. It's designed to work with the Go Fiber framework and implements a room-based messaging system.

## Features

- WebSocket-based real-time communication
- Room management for organizing users
- Event-driven architecture using CQRS (Command Query Responsibility Segregation)
- Scalable design (with known scaling issues to be addressed)

## Add the feature to your FeatureList

**examples/api/cmd/server/main.go**

```go
import "github.com/aiocean/wireset/feature/realtime"

func FeatureList(
	realtime *realtime.FeatureRealtime,
) []server.Feature {
	return []server.Feature{
		realtime,
	}
}
```

## Add provider to injector

**examples/api/cmd/server/wire.go**

```go

func InitializeServer() (server.Server, func(), error) {
	wire.Build(
		FeatureList,
		realtime.DefaultWireset,
	)
}
```

## Handle websocket topic

3. Handle messages:

Add WsHandlerRegistry to your feature struct

```go
import "github.com/aiocean/wireset/feature/realtime/registry"

type YourFeature struct {
	WsRegistry *registry.HandlerRegistry
}
```

Implement your handler `<your_feature>/ws/example_ws_handler.go`

```go
type ExampleWsHandler struct {
	Logger *zap.Logger
}

func (h *ExampleWsHandler) Handle(conn *websocket.Conn, payload *gjson.Result) error {
	return conn.WriteJSON(models.WebsocketMessage[model.DashboardReport]{
		Topic:   model.TopicSetDashboardReport,
		Payload: report,
	})
}
```

Add the handler to feature struct, wireset:

```go
type FeatureReport struct {
	WsRegistry *registry.HandlerRegistry
	ExampleWsHandler *ExampleWsHandler
}


var DefaultWireset = wire.NewSet(
	wire.Struct(new(YourFeature), "*"),
	wire.Struct(new(ws.ExampleWsHandler), "*"),
)
```

Add to WsRegistry in Init() function:

```go
func (f *FeatureReport) Init() error {
    f.WsRegistry.AddWebsocketHandler(&registry.WebsocketHandler{
        Topic:   models.WebsocketTopic("your_topic"),
        Handler: yourHandlerFunc,
    })
}
```

4. Send messages programmatically:

   Use the `SendWsMessageHandler` to send messages to specific users:

   ```go
   cmd := &command.SendWsMessageCmd{
       RoomID:   "room_id",
       Username: "username",
       Payload:  yourPayload,
   }
   err := sendWsMessageHandler.Handle(context.Background(), cmd)
   ```