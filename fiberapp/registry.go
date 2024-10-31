package fiberapp

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type HttpHandler struct {
	Method   string
	Path     string
	Handlers []fiber.Handler
}

type Registry struct {
	HttpHandlers    map[string]*HttpHandler
	HttpMiddlewares map[string]interface{}
	StaticRoutes    map[string]string // New field for static routes
}

func NewRegistry() *Registry {
	return &Registry{
		HttpHandlers:    map[string]*HttpHandler{},
		HttpMiddlewares: map[string]interface{}{},
		StaticRoutes:    map[string]string{}, 
	}
}

func (r *Registry) AddHttpHandlers(handlers ...*HttpHandler) {
	for _, handler := range handlers {
		r.HttpHandlers[createHandlerID(handler.Method, handler.Path)] = handler
	}
}

func (r *Registry) AddHttpMiddleware(path string, handler interface{}) {
	r.HttpMiddlewares[path] = handler
}

func (r *Registry) GetHttpHandler(method, path string) *HttpHandler {
	id := createHandlerID(method, path)
	return r.HttpHandlers[id]
}

func createHandlerID(method, path string) string {
	return strings.ToLower(method + " " + path)
}

func (r *Registry) RegisterHandlers(app *fiber.App) {
	for _, handler := range r.HttpHandlers {
		app.Add(handler.Method, handler.Path, handler.Handlers...)
	}
}

// New function to register static routes
func (r *Registry) RegisterStaticRoutes(app *fiber.App) {
	for urlPrefix, directory := range r.StaticRoutes {
		app.Static(urlPrefix, directory)
	}
}

func (r *Registry) RegisterMiddlewares(app *fiber.App) {
	for path, middleware := range r.HttpMiddlewares {
		app.Use(path, middleware)
	}
}

// New method to add static routes
func (r *Registry) AddStaticRoute(urlPrefix, directory string) {
	r.StaticRoutes[urlPrefix] = directory
}
