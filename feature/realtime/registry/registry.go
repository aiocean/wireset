package registry

import (
	"errors"
	"sync"

	"github.com/aiocean/wireset/feature/realtime/models"
	"github.com/gofiber/contrib/websocket"
	"github.com/hashicorp/go-multierror"
	"github.com/tidwall/gjson"
)

type HandlerRegistry struct {
	handlers map[string][]HandlerFunc
	mu       sync.RWMutex
}

func NewWsHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string][]HandlerFunc),
	}
}

type HandlerFunc func(conn *websocket.Conn, payload *gjson.Result) error

type WebsocketHandler struct {
	Topic   models.WebsocketTopic
	Handler HandlerFunc
}

func (r *HandlerRegistry) AddWebsocketHandler(handlers ...*WebsocketHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, h := range handlers {
		r.handlers[h.Topic.String()] = append(r.handlers[h.Topic.String()], h.Handler)
	}
}

var ErrHandlerNotFound = errors.New("handler not found")

func (r *HandlerRegistry) Handle(topic string, conn *websocket.Conn, payload *gjson.Result) error {
	r.mu.RLock()
	handlers, ok := r.handlers[topic]
	r.mu.RUnlock()
	if !ok {
		return ErrHandlerNotFound
	}
	var wg sync.WaitGroup
	var result *multierror.Error
	var mu sync.Mutex
	for _, handler := range handlers {
		wg.Add(1)
		go func(h HandlerFunc) {
			defer wg.Done()
			if err := h(conn, payload); err != nil {
				mu.Lock()
				result = multierror.Append(result, err)
				mu.Unlock()
			}
		}(handler)
	}
	wg.Wait()
	return result.ErrorOrNil()
}
