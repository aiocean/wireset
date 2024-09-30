package api

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/registry"
	"github.com/aiocean/wireset/feature/realtime/resolver"
	"github.com/aiocean/wireset/feature/realtime/room"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const roomIDKey = "roomID"
const usernameKey = "username"

const errorKey = "error"

type WebsocketHandler struct {
	RoomManager      *room.Manager
	Logger           *zap.Logger
	IdentityResolver resolver.IdentityResolver
	Registry         *registry.HandlerRegistry
	EventBus         *cqrs.EventBus
}

func NewWebsocketHandler(
	logger *zap.Logger,
	roomManager *room.Manager,
	identityResolver resolver.IdentityResolver,
	registry *registry.HandlerRegistry,
	eventBus *cqrs.EventBus,
) *WebsocketHandler {
	return &WebsocketHandler{
		RoomManager:      roomManager,
		Logger:           logger.Named("websocket"),
		IdentityResolver: identityResolver,
		Registry:         registry,
		EventBus:         eventBus,
	}
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
// require "room" and "username" query params.
// we need to extend this function to allow another query params.
func (h *WebsocketHandler) Upgrade(ctx *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(ctx) {
		return fiber.ErrUpgradeRequired
	}

	identity, err := h.IdentityResolver.Resolve(ctx)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	ctx.Locals(roomIDKey, identity.Room)
	ctx.Locals(usernameKey, identity.Username)

	currentRoom, err := h.RoomManager.GetRoom(identity.Room)
	switch {
	case err != nil && !errors.Is(err, room.ErrRoomNotFound):
		return err
	case currentRoom == nil:
		return ctx.Next()
	case currentRoom.IsMemberExists(identity.Username):
		return errors.New("username already exists")
	}

	return ctx.Next()
}

// OnDisconnect is called when a client disconnects from the server.
func (h *WebsocketHandler) OnDisconnect(ctx *websocket.Conn, room *room.Room, username string) {
	defer func() {
		if room.IsEmpty() {
			if err := h.RoomManager.DeleteRoom(room.ID); err != nil {
				h.Logger.Error("delete room", zap.String(roomIDKey, room.ID))
			}
			h.Logger.Info("Room Deleted")
		}
	}()

	if err := room.DeleteMember(username); err != nil {
		h.Logger.Error("delete member", zap.String(usernameKey, username), zap.String(roomIDKey, room.ID))
		return
	}

	h.Logger.Info("Member Left", zap.String(usernameKey, username), zap.String(roomIDKey, room.ID))
}
func (h *WebsocketHandler) handleError(conn *websocket.Conn, logger *zap.Logger, err error, message string) {
	logger.Error(message, zap.Error(err))
	conn.WriteJSON(map[string]string{
		errorKey: message + ": " + err.Error(),
	})
}

// getTopic returns the topic from the message.
func getTopic(buf []byte) (string, error) {
	topic := gjson.GetBytes(buf, "topic").String()
	if topic == "" {
		return "", errors.New("command is empty")
	}

	return topic, nil
}
