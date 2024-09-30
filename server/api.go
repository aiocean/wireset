package server

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ApiServer struct {
	MsgRouter           *message.Router
	ConfigSvc           *configsvc.ConfigService
	LogSvc              *zap.Logger
	FiberSvc            *fiber.App
	HttpHandlerRegistry *fiberapp.Registry
	Features            []Feature
}

var DefaultWireset = wire.NewSet(
	wire.Struct(new(ApiServer), "*"),
	wire.Bind(new(Server), new(*ApiServer)),
)

func (s *ApiServer) Start(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	g, ctx := errgroup.WithContext(ctx)

	if err := s.initFeatures(); err != nil {
		errChan <- err
		close(errChan)
		return errChan
	}

	g.Go(func() error { return s.runMessageRouter(ctx) })
	g.Go(s.runHTTPServer)

	go func() {
		errChan <- g.Wait()
		close(errChan)
	}()

	return errChan
}

func (s *ApiServer) initFeatures() error {
	for _, f := range s.Features {
		s.LogSvc.Info("Initializing feature", zap.String("feature", f.Name()))
		if err := f.Init(); err != nil {
			return fmt.Errorf("failed to init feature %s: %w", f.Name(), err)
		}
	}
	return nil
}

func (s *ApiServer) runMessageRouter(ctx context.Context) error {
	s.LogSvc.Info("Starting message router")
	if err := s.MsgRouter.Run(ctx); err != nil {
		return fmt.Errorf("message router error: %w", err)
	}
	s.LogSvc.Info("Message router stopped")
	return nil
}

func (s *ApiServer) runHTTPServer() error {
	port := s.ConfigSvc.Port
	s.LogSvc.Info("Starting HTTP server", zap.String("port", port))

	s.HttpHandlerRegistry.RegisterMiddlewares(s.FiberSvc)
	s.HttpHandlerRegistry.RegisterHandlers(s.FiberSvc)

	if err := s.FiberSvc.Listen(":" + port); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	s.LogSvc.Info("HTTP server stopped")
	return nil
}

func (s *ApiServer) Shutdown(ctx context.Context) error {
	s.LogSvc.Info("Shutting down server")
	if err := s.FiberSvc.Shutdown(); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}
	s.MsgRouter.Close()
	return nil
}
