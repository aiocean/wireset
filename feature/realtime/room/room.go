package room

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/gofiber/contrib/websocket"
)

type Room struct {
	ID          string
	membersLock sync.RWMutex
	Members     map[string]*Member
}

func NewRoom(id string) *Room {
	return &Room{
		ID:          id,
		membersLock: sync.RWMutex{},
		Members:     make(map[string]*Member),
	}
}

// IsMemberExists checks if a member exists.
func (r *Room) IsMemberExists(username string) bool {
	r.membersLock.RLock()
	defer r.membersLock.RUnlock()
	isMemberExist := r.Members[username] != nil

	return isMemberExist
}

// AddMember adds a new member to the room.
// It acquires a write lock on the room's mutex to ensure thread safety.
func (r *Room) AddMember(username string, conn *websocket.Conn) error {
	r.membersLock.Lock()
	defer r.membersLock.Unlock()

	if r.Members[username] != nil {
		return errors.New("member already exists")
	}

	r.Members[username] = NewMember(username, conn)
	return nil
}

// DeleteMember deletes a member from the room.
// It acquires a write lock on the room's mutex to ensure thread safety.
func (r *Room) DeleteMember(username string) error {
	r.membersLock.Lock()
	defer r.membersLock.Unlock()
	if _, exists := r.Members[username]; !exists {
		return ErrMemberNotFound
	}
	delete(r.Members, username)
	return nil
}

// IsEmpty checks if the room is empty.
// It acquires a read lock on the room's mutex to ensure thread safety.
func (r *Room) IsEmpty() bool {
	r.membersLock.RLock()
	defer r.membersLock.RUnlock()
	return len(r.Members) == 0
}

var ErrMemberNotFound = errors.New("member not found")

func (r *Room) SendMessageTo(username string, message interface{}) error {
	r.membersLock.RLock()
	defer r.membersLock.RUnlock()
	if member := r.Members[username]; member != nil {
		return member.Send(message)
	}

	return ErrMemberNotFound
}

// SendSystemMessage sends a system message to a member.
func (r *Room) SendSystemMessage(username string, messageType string, message interface{}) error {
	return r.SendMessageTo(username, &Message{
		Sender:    "system",
		Recipient: username,
		Type:      messageType,
		Message:   message,
	})
}

// SendSystemError sends a system error message to a member.
func (r *Room) SendSystemError(username string, message interface{}) error {
	return r.SendSystemMessage(username, "error", message)
}

func (r *Room) BroadcastMessage(message interface{}) []error {
	r.membersLock.RLock()
	defer r.membersLock.RUnlock()
	var errors []error
	for _, member := range r.Members {
		if err := member.Send(message); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
