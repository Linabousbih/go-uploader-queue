package apiserver

import (
	"async/config"
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type ApiServer struct {
	Config *config.Config
	logger *slog.Logger
}

func New(config *config.Config, logger *slog.Logger) *ApiServer {

	return &ApiServer{Config: config}
}

func (s *ApiServer) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {

	}
}

func (s *ApiServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", s.ping)
	server := &http.Server{
		Addr:    net.JoinHostPort(s.Config.ApiServerHost, s.Config.ApiServerPort),
		Handler: mux,
	}

	go func() {
		s.logger.Info("starting server at", "port", s.Config.ApiServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("api server failed to listen and serve", "error", err)
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("api server faield to shutdown", "error", err)
		}
	}()

	wg.Wait()
	return nil
}
