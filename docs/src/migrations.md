# Migrations

Schema changes live in
[`internal/store/postgres/migrations`](https://github.com/picunada/flagcel/tree/main/internal/store/postgres/migrations)
and are managed with [goose](https://github.com/pressly/goose). Migrations are
embedded into the binary at build time.

By default the server applies pending migrations on startup. For production
deployments where you want to run migrations out-of-band, set
`MIGRATE_ON_STARTUP=false` and use the `migrate` subcommand:

```sh
flagcel migrate up       # apply all pending
flagcel migrate down     # roll back the most recent
flagcel migrate status   # show applied / pending
flagcel migrate version  # print current version
```

The same targets are exposed via `make migrate-up`, `make migrate-status`, and
related commands.

To add a new migration, create `internal/store/postgres/migrations/0000N_name.sql`:

```sql
-- +goose Up
ALTER TABLE flags ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE flags DROP COLUMN description;
```

Run the migration status command before and after applying production
migrations so operators can confirm the expected version.
