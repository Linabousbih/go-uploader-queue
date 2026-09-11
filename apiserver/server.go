package apiserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"async/config"
)

type ApiServer struct {
	config  *config.Config
	logger  *slog.Logger
	handler http.Handler
}

func New(config *config.Config, logger *slog.Logger, handler http.Handler) *ApiServer {
	return &ApiServer{config: config, logger: logger, handler: handler}
}

func (s *ApiServer) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    net.JoinHostPort(s.config.ApiServerHost, s.config.ApiServerPort),
		Handler: s.handler,
	}

	go func() {
		s.logger.Info("starting server at", "port", s.config.ApiServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("api server failed to listen and serve", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("api server faield to shutdown", "error", err)
		}
	}()

	return nil
}
