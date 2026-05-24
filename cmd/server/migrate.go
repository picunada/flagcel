package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/picunada/flagcel/internal/config"
	"github.com/picunada/flagcel/internal/store/postgres/migrations"
)

func runMigrate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: flagcel migrate <up|down|status|version>")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect pgx pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	cmd := args[0]
	switch cmd {
	case "up":
		err = migrations.Up(ctx, db)
	case "down":
		err = migrations.Down(ctx, db)
	case "status":
		err = migrations.Status(ctx, db)
	case "version":
		err = migrations.Version(ctx, db)
	default:
		fmt.Fprintf(os.Stderr, "unknown migrate command: %s\n", cmd)
		os.Exit(2)
	}
	if err != nil {
		slog.Error("migrate", "cmd", cmd, "err", err)
		os.Exit(1)
	}
}
