package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/picunada/flagcel/internal/api/http/middleware"
	"github.com/picunada/flagcel/internal/service"
)

type Config struct {
	Port            uint16
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	cfg    Config
	http   *http.Server
	logger *slog.Logger
}

func NewServer(cfg Config, flagSvc *service.FlagService, logger *slog.Logger) *Server {
	handlers := &Handlers{
		Flags: NewFlagsHandler(flagSvc),
	}

	router := NewRouter(handlers)

	chain := middleware.Chain(
		middleware.Logging(logger),
		middleware.Recover(logger),
		middleware.RequestID(logger),
	)

	return &Server{
		cfg:    cfg,
		logger: logger,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      chain(router),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("http server listening", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("http server: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutdown signal received, stopping http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
