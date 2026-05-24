package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var FS embed.FS

func init() {
	goose.SetBaseFS(FS)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("goose set dialect: %w", err))
	}
}

func Up(ctx context.Context, db *sql.DB) error {
	return goose.UpContext(ctx, db, ".")
}

func Down(ctx context.Context, db *sql.DB) error {
	return goose.DownContext(ctx, db, ".")
}

func Status(ctx context.Context, db *sql.DB) error {
	return goose.StatusContext(ctx, db, ".")
}

func Version(ctx context.Context, db *sql.DB) error {
	return goose.VersionContext(ctx, db, ".")
}
