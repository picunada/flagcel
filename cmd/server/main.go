package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/api/http/debug"
	v1 "github.com/picunada/flagcel/internal/api/http/v1"
	"github.com/picunada/flagcel/internal/config"
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

	celEnv, err := evalcore.NewCELEnv()
	if err != nil {
		logger.Error("init cel env", "err", err)
		os.Exit(1)
	}
	eng := evalcore.NewEngine(celEnv)

	evalSvc := service.NewEvalService(store, eng)
	flagSvc := service.NewFlagService(store, evalSvc.InvalidateFlag)
	ruleSvc := service.NewRuleService(store, evalSvc.InvalidateFlag)
	ctxSvc := service.NewContextService(store, evalSvc.InvalidateContext)
	authSvc, err := service.NewAuthService(ctx, service.AuthConfig{
		OIDCIssuerURL:     cfg.Auth.OIDCIssuerURL,
		OIDCClientID:      cfg.Auth.OIDCClientID,
		OIDCSecret:        cfg.Auth.OIDCClientSecret,
		OIDCRedirectURL:   cfg.Auth.OIDCRedirectURL,
		AdminEmails:       cfg.Auth.AdminEmails,
		BootstrapEmail:    cfg.Auth.BootstrapEmail,
		BootstrapPassword: cfg.Auth.BootstrapPassword,
		BootstrapName:     cfg.Auth.BootstrapName,
		SessionSecret:     cfg.Auth.SessionSecret,
		CookieSecure:      cfg.Auth.CookieSecure,
		SessionTTL:        cfg.Auth.SessionTTL,
	}, store)
	if err != nil {
		logger.Error("init auth", "err", err)
		os.Exit(1)
	}
	srv := v1.NewServer(v1.Config{
		Port:            cfg.Port,
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		IdleTimeout:     cfg.HTTP.IdleTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, flagSvc, ruleSvc, ctxSvc, evalSvc, authSvc, logger)

	if cfg.DebugAddr != "" {
		dbg := debug.NewServer(cfg.DebugAddr, logger)
		go func() {
			if err := dbg.Start(ctx, cfg.HTTP.ShutdownTimeout); err != nil {
				logger.Error("debug server", "err", err)
			}
		}()
	}

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
