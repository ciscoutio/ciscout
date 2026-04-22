# Tasks — repo-scaffolding

## Budget
- **Opus:** 3 tasks — license choice, env-lib choice, final review
- **Sonnet:** 4 tasks — core implementation
- **Haiku:** 4 tasks — boilerplate files and configuration

Total: 11 tasks. Expected effort: 4–6 hours of focused work.

## Tasks

### Decisions first (unblock implementation)

- [ ] `@opus` Decide license: MIT vs AGPL-3.0 vs BSL 1.1. Record decision + rationale in a short note at the top of `LICENSE` (one-sentence header comment if using MIT/AGPL text verbatim; rationale in the commit message). Default bias: MIT for v0.1 simplicity; revisit pre-public-launch if commercial-clone risk becomes credible.

- [ ] `@opus` Decide config library: `github.com/caarlos0/env/v10` vs `github.com/sethvargo/go-envconfig` vs hand-rolled `os.Getenv` wrappers. Picking one means every future handler reads config the same way. Constraints: minimal deps, struct-tag driven, supports defaults and required validation.

### Scaffold

- [ ] `@sonnet` Initialize Go module: `go mod init github.com/ciscoutio/ciscout`. Add `github.com/go-chi/chi/v5` and the env library chosen above. Run `go mod tidy`. Commit `go.mod` and `go.sum`.

- [ ] `@sonnet` Create folder structure per proposal §What. Add a `.keep` file to each empty folder (`api/`, `mcp/`, `ui/`, `landing/`) so git tracks them. Do not add `.keep` to `openspec/` subdirs (they'll fill up quickly).

- [ ] `@sonnet` Write `internal/config/config.go` — `Config` struct with fields: `Port int` (default 8080), `LogLevel string` (default "info"), `Environment string` (required: `dev`/`staging`/`prod`). Load via chosen env lib. Unit test with a subtest for defaults and a subtest for env overrides.

- [ ] `@sonnet` Write `internal/httpserver/server.go` — `New(cfg *config.Config) *http.Server` that builds a chi router with `/healthz` and `/readyz` handlers, applies a request-logging middleware using `slog`, and returns a configured `*http.Server` with sensible timeouts (read: 10s, write: 30s, idle: 120s). Include a basic test that hits `/healthz` via `httptest`.

- [ ] `@sonnet` Write `cmd/ciscout/main.go` — load config, construct a JSON `slog` logger, construct the server via `httpserver.New`, start listening, handle graceful shutdown on `SIGINT`/`SIGTERM` with a 10-second drain. No goroutines doing real work yet.

### Boilerplate

- [ ] `@haiku` Write `.gitignore`: Go build artifacts, IDE files (`.idea/`, `.vscode/`), env files (`.env`, `.env.local`), macOS/Linux/Windows junk, `node_modules/`, `dist/`, `coverage.out`, test binaries.

- [ ] `@haiku` Write `.env.example` with every env var the config reads, each commented with purpose and example value. Include a warning header: "# Copy to .env.local. Never commit .env.local."

- [ ] `@haiku` Write `Makefile` with targets: `help` (self-documenting via comments), `dev` (run with live-reload via `air` if available, else plain `go run`), `build` (compile to `bin/ciscout`), `test` (`go test ./... -race -count=1`), `fmt` (`gofmt -w .`), `vet` (`go vet ./...`), `tidy` (`go mod tidy`), `lint` (if staticcheck is installed, run it; else no-op with a print), `clean`.

- [ ] `@haiku` Write `README.md` — max 40 lines. Sections: What CIScout is (2 sentences, link to `docs/PRODUCT.md`), How to run locally (3-step: clone, `cp .env.example .env.local`, `make dev`), Project structure (brief tree), Where to find docs (list of `docs/*.md` files with one-line description each), License (link to `LICENSE`). No marketing copy. No screenshots.

- [ ] `@haiku` Write `.pre-commit-config.yaml` — hooks: `gofmt`, `go-vet`, `gitleaks` (use `zricethezav/gitleaks` hook). Document install steps in README.

### Review

- [ ] `@opus` Review the full scaffold end-to-end before merge. Specific checks:
  - `make build && ./bin/ciscout` works with `.env.example` copied to `.env.local`.
  - `curl localhost:8080/healthz` returns 200 with a JSON body.
  - Graceful shutdown drains cleanly under `kill -SIGTERM`.
  - No dead-code, no speculative abstractions, no middleware or utility files beyond those in the proposal.
  - `.env.example` is consistent with what `config.go` reads — no drift.
  - License choice is executed correctly (SPDX header where needed, correct year).
  - Pre-commit hooks actually run and catch a staged secret test case.

## Escalations

Record here if any task was escalated mid-execution. Format:
`<task line> — escalated from @X to @Y because <reason>`
