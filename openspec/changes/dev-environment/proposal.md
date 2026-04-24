# dev-environment

## Why

CIScout's v0.1 data model (installations, diagnoses, BYOK keys, River queue state — see `docs/ARCHITECTURE.md` §5) all lives in Postgres. Before any feature lands, local development needs a reproducible Postgres + migration workflow. Without this, every subsequent change blocks on "how do I run the DB?" The scaffolding binary currently connects to nothing; this change wires in the first external dependency.

## What

Stand up the minimum DB ergonomics needed for later feature work:

- `docker-compose.yml` at repo root running a pinned Postgres version matching Fly.io managed Postgres.
- `goose` adopted for migrations. First migration is an intentional no-op baseline that proves the pipeline end-to-end.
- `internal/db/db.go` package: opens a pooled `*sqlx.DB`, runs a startup ping.
- `DatabaseURL` added to `internal/config/config.go` (required, no default).
- `/readyz` changed from stub to actual DB ping so Fly.io health checks and local verification reflect reality.
- `Makefile` targets for db lifecycle (`db-up`, `db-down`, `db-reset`, `db-migrate`, `db-migrate-new`).
- `.env.example` updated with `DATABASE_URL`.

Files and folders touched:
```
/
├── docker-compose.yml              (new)
├── db/
│   └── migrations/
│       └── 00001_baseline.sql       (new, no-op)
├── internal/
│   ├── config/config.go             (extend)
│   ├── config/config_test.go        (extend)
│   ├── db/db.go                     (new)
│   ├── db/db_test.go                (new)
│   └── httpserver/server.go         (wire /readyz DB ping)
├── cmd/ciscout/main.go              (construct + inject *sqlx.DB)
├── Makefile                         (extend)
└── .env.example                     (extend)
```

## Acceptance criteria

- [ ] `make db-up` starts Postgres, reachable on `localhost:5432`.
- [ ] `make db-migrate` applies the baseline migration and exits 0.
- [ ] `./bin/ciscout` connects to the DB on startup; startup fails loudly (non-zero exit, logged error) if DB is unreachable.
- [ ] `curl localhost:8080/readyz` returns 200 when DB is up, 503 within 10s of DB going down.
- [ ] `curl localhost:8080/healthz` still returns 200 regardless of DB state (liveness, not readiness).
- [ ] `make db-reset` tears down the volume and recreates a clean DB.
- [ ] `make test` still passes; unit tests do not require a running DB.
- [ ] No application tables created — baseline migration is a comment-only no-op.

## Out of scope

- Application tables (installations, diagnoses, anthropic_keys, etc.) — each feature change owns its migration.
- Connection retry/backoff strategy beyond `sql.DB` pool defaults.
- Dockerfile / container image for the app itself — deploy-time concern owned by `fly-staging-deploy` (week 4).
- Test infrastructure for DB-backed tests (testcontainers, fixtures, transactional test helpers) — land when the first DB-backed feature does.
- Migration rollback playbook — nothing to roll back yet.
- Separate read replica / pooling via pgbouncer — premature at v0.1.

## Dependencies

- `repo-scaffolding` merged (provides config package, server, Makefile to extend).

## Risks

- **Postgres version drift between dev and Fly.io prod.** Pinning via `@opus` decision task with explicit version recorded.
- **`/readyz` hot path querying DB.** Fly.io health checks every ~10s; a ping per check is cheap but worth a conscious decision. `@opus` task picks the strategy.
- **Goose tool install friction.** If contributors see "goose not found" on first `make db-migrate`, adoption suffers. Makefile targets should self-install on first run or fail with a clear message.
