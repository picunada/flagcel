package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	v1 "github.com/picunada/flagcel/internal/api/http/v1"
	"github.com/picunada/flagcel/internal/config"
	"github.com/picunada/flagcel/internal/service"
	"github.com/picunada/flagcel/internal/store/postgres"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	logger := cfg.Logger()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect pgx pool", "err", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("ping db", "err", err)
		os.Exit(1)
	}
	store := postgres.NewStore(pool)

	flagSvc := service.NewFlagService(store)
	srv := v1.NewServer(v1.Config{
		Port:            cfg.Port,
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		IdleTimeout:     cfg.HTTP.IdleTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, flagSvc, logger)

	if err := srv.Start(ctx); err != nil {
		slog.Error("http server", "err", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := store.Close(shutdownCtx); err != nil {
		slog.Error("store shutdown", "err", err)
	}

	slog.Info("server stopped")
}
