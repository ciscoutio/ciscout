# Tasks — dev-environment

## Budget
- **Opus:** 3 tasks — Postgres version pin, readyz DB-check strategy, final review
- **Sonnet:** 4 tasks — config extension, db package, main.go wiring, readyz change
- **Haiku:** 3 tasks — docker-compose, baseline migration, Makefile + .env.example

Total: 10 tasks. Expected effort: 3–4 hours.

## Tasks

### Decisions first

- [x] `@opus` Pin Postgres major version. **Decision: Postgres 16.** Fly.io Managed Postgres docs (https://fly.io/docs/mpg/) reference only the Postgres 16 distribution as of 2026-04-24, so matching the major version eliminates SQL / extension-behavior drift between dev and prod. Use the `postgres:16` tag in `docker-compose.yml` (major-pin only — security patches flow through `docker compose pull`; minor-version pinning is overkill for a local dev DB). Restate the version + this rationale in the `docker-compose.yml` commit message.

- [x] `@opus` Decide `/readyz` DB-check strategy. **Decision: (a) ping per request, `SELECT 1` with a 1s context timeout.** Fly checks `/readyz` every ~10s, so one ping per cycle is effectively free; detection of a bad DB happens within one check cycle so Fly can cycle the instance quickly; the 1s context timeout caps handler latency when the DB is hung, keeping responses inside Fly's check window. Option (b) background probe is a premature optimization on a `SELECT 1` and adds goroutine lifecycle + cache-staleness concerns; (c) startup-only ping defeats the purpose of a readiness probe. Revisit only if `/readyz` latency shows up in a flame graph.

### Scaffold

- [x] `@haiku` Write `docker-compose.yml` at repo root: one `postgres:<pinned-version>` service, named volume for data, healthcheck (`pg_isready`), port 5432 exposed. User/password/db all `ciscout`. Include a comment linking to the Fly version pin decision.

- [x] `@haiku` Create `db/migrations/00001_baseline.sql` with `-- +goose Up` / `-- +goose Down` sections, each containing only a comment. Purpose: exercise the migration pipeline end-to-end before any real schema lands.

- [x] `@sonnet` Extend `internal/config/config.go` — add `DatabaseURL string` with tag `env:"DATABASE_URL,required"`. Extend `internal/config/config_test.go` with subtests for present + missing cases.

- [x] `@sonnet` Write `internal/db/db.go` — `New(ctx context.Context, url string) (*sqlx.DB, error)` that opens a pooled connection using `sqlx.ConnectContext`, sets `MaxOpenConns=25`, `MaxIdleConns=5`, `ConnMaxLifetime=30*time.Minute`, and pings with a 5s context timeout before returning. Add `internal/db/db_test.go` that skips if `DATABASE_URL` is unset (keeps CI green without a DB).

- [x] `@sonnet` Wire DB into `cmd/ciscout/main.go`: after `config.Load()`, construct `*sqlx.DB` via `db.New(ctx, cfg.DatabaseURL)`, defer `db.Close()`, pass into `httpserver.New(cfg, db)`. Startup failure logs and exits non-zero.

- [x] `@sonnet` Update `internal/httpserver/server.go` — `New` signature accepts `*sqlx.DB`; `/readyz` handler pings the DB with a 1s context timeout and returns 503 with JSON `{"status":"unavailable"}` on failure. Update `server_test.go` — `/healthz` test unchanged; `/readyz` test covers both DB-ok and DB-down paths using a mock or a closed `*sqlx.DB`.

- [x] `@haiku` Extend `Makefile` with targets: `db-up` (`docker compose up -d postgres`), `db-down` (`docker compose down`), `db-reset` (`docker compose down -v && $(MAKE) db-up db-migrate`), `db-migrate` (install goose if missing, then `goose -dir db/migrations postgres "$$DATABASE_URL" up`), `db-migrate-new` (takes `NAME=foo`, runs `goose -dir db/migrations create $$NAME sql`). Update `.env.example` with `DATABASE_URL=postgres://ciscout:ciscout@localhost:5432/ciscout?sslmode=disable`.

### Review

- [x] `@opus` End-to-end smoke test. Passed 2026-04-25. Two deviations fixed inline: (1) `command -v goose` failed because `~/go/bin` is not in make's PATH — fixed by resolving via `go env GOPATH`; (2) `db-reset` raced between `docker compose up` and `goose` — fixed by switching `db-up` to `docker compose up -d --wait postgres` which blocks until the healthcheck passes. All acceptance criteria met: `make db-up db-migrate` succeeds; server starts and `/readyz` returns 200 with DB up, 503 within 1s of DB going down; `make db-reset` restores 200; `make test` green (DB-backed tests run when `DATABASE_URL` is set, skip cleanly without it).

### Post-smoke-test addition

- [x] `@sonnet` Add Postgres service to CI. Extend `.github/workflows/ci.yml` with a `services: postgres:16` block (same credentials as docker-compose: user/password/db all `ciscout`). Set `DATABASE_URL` in the `go test` step env. This makes `TestNew` and `TestReadyz/db_available` run for real in CI instead of skipping, and exercises the full startup path including migration-compatible connectivity.

## Escalations

Record here if any task was escalated mid-execution. Format:
`<task line> — escalated from @X to @Y because <reason>`
