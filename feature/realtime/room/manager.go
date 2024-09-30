package room

import (
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Manager struct {
	Logger *zap.Logger
	mu     sync.RWMutex
	Rooms  map[string]*Room
}

func NewRoomManager(
	logger *zap.Logger,
) (*Manager, error) {
	return &Manager{
		Logger: logger,
		mu:     sync.RWMutex{},
		Rooms:  make(map[string]*Room),
	}, nil
}

// IsRoomExists checks if a room exists.
func (h *Manager) IsRoomExists(roomName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.Rooms[roomName]
	return ok
}

// AddNewRoom adds a new room to the api.
func (h *Manager) AddNewRoom(roomName string) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Rooms[roomName] = NewRoom(roomName)

	return h.Rooms[roomName], nil
}

var ErrRoomNotFound = errors.New("room not found")

// GetRoom returns a room by its name.
func (h *Manager) GetRoom(roomName string) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if foundRoom, found := h.Rooms[roomName]; found {
		return foundRoom, nil
	}

	return nil, ErrRoomNotFound
}

// DeleteRoom deletes a room by its name.
func (h *Manager) DeleteRoom(roomName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.Rooms, roomName)
	return nil
}
