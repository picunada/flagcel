package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	v1 "github.com/picunada/flagcel/internal/api/http/v1"
	"github.com/picunada/flagcel/internal/config"
	"github.com/picunada/flagcel/internal/engine"
	"github.com/picunada/flagcel/internal/service"
	"github.com/picunada/flagcel/internal/store/postgres"
	"github.com/picunada/flagcel/internal/store/postgres/migrations"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrate(os.Args[2:])
		return
	}
	runServer()
}

func runServer() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	logger := cfg.Logger()
	logger.Info("config loaded", "port", cfg.Port, "log_level", cfg.LogLevel)

	logger.Info("connecting to database")
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect pgx pool", "err", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("ping db", "err", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	if cfg.MigrateOnStartup {
		logger.Info("applying migrations")
		db := stdlib.OpenDBFromPool(pool)
		if err := migrations.Up(ctx, db); err != nil {
			logger.Error("migrate up", "err", err)
			os.Exit(1)
		}
		_ = db.Close()
		logger.Info("migrations applied")
	}

	store := postgres.NewStore(pool)

	celEnv, err := engine.NewCELEnv()
	if err != nil {
		logger.Error("init cel env", "err", err)
		os.Exit(1)
	}
	eng := engine.NewEngine(celEnv)

	flagSvc := service.NewFlagService(store)
	ruleSvc := service.NewRuleService(store)
	ctxSvc := service.NewContextService(store)
	evalSvc := service.NewEvalService(store, eng)
	srv := v1.NewServer(v1.Config{
		Port:            cfg.Port,
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		IdleTimeout:     cfg.HTTP.IdleTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, flagSvc, ruleSvc, ctxSvc, evalSvc, logger)

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
