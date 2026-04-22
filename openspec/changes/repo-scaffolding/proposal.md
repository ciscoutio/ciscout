# Repo scaffolding

## Why

The repo currently holds only docs. Week 1 of EXECUTION_PLAN.md requires a runnable Go monolith skeleton so subsequent changes (dev environment, GitHub webhook receiver, etc.) have somewhere to land. A proper scaffold sets tech-stack conventions in code — router choice, logger choice, config pattern — before the first feature is written, avoiding drift later.

## What

Stand up the minimum set of files needed for:
- A compilable Go binary at `cmd/ciscout` that serves `/healthz` and `/readyz`.
- The monorepo folder layout described in `docs/ARCHITECTURE.md` and `docs/WORKFLOW.md` §2.
- Developer ergonomics: `Makefile` for common commands, `.gitignore`, `.env.example`, `README.md`.
- Licensing chosen and committed (one `@opus` decision task covers this).
- Pre-commit hooks wired (format, vet, secret-scan).

The binary does no real work yet — no Postgres, no GitHub App, no queue. Those land in follow-up changes (`dev-environment`, `github-app-install-flow`, `worker-bootstrap`). This change is strictly the scaffold.

Files and folders created:
```
/
├── README.md
├── LICENSE
├── Makefile
├── .gitignore
├── .env.example
├── .pre-commit-config.yaml
├── go.mod
├── go.sum
├── cmd/
│   └── ciscout/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   └── httpserver/
│       └── server.go
├── api/                 # (empty for now; populated by future changes)
├── mcp/                 # (empty for now)
├── ui/                  # (empty for now)
├── landing/             # (empty for now)
└── openspec/
    ├── changes/         # this folder
    └── archive/         # (empty)
```

Tech stack locked by this change (consistent with `docs/ARCHITECTURE.md` §4.1):
- Go 1.23+
- Router: `github.com/go-chi/chi/v5`
- Logger: stdlib `log/slog` with JSON handler
- Config: env-driven via `github.com/caarlos0/env/v10` (or equivalent small lib; decision in tasks)

## Acceptance criteria

- [ ] `go build ./cmd/ciscout` succeeds on a fresh clone.
- [ ] `./ciscout` starts a server bound to `:8080` (configurable via `PORT` env).
- [ ] `curl localhost:8080/healthz` returns `200 OK`.
- [ ] `curl localhost:8080/readyz` returns `200 OK`.
- [ ] `make help` lists all available targets.
- [ ] `make fmt`, `make vet`, `make test` run clean on the scaffold.
- [ ] `.env.example` documents every env var the binary reads.
- [ ] `README.md` explains: what CIScout is (one line pointing to `docs/PRODUCT.md`), how to run it locally, where to find docs.
- [ ] `LICENSE` file committed; choice documented in the license task.
- [ ] `.pre-commit-config.yaml` present with `gofmt`, `go vet`, and `gitleaks`.
- [ ] No secrets in the repo. `gitleaks` run on the scaffold passes.

## Out of scope

- Postgres, migrations, data model — covered by `change/dev-environment`.
- CI workflow (GitHub Actions) — covered by `change/ci-pipeline`.
- GitHub App registration and webhook handling — covered by `change/github-app-install-flow` and `change/webhook-receiver`.
- Frontend (React dashboard, landing page) — covered by `change/landing-page-stub` and later UI changes.
- Any business logic (diagnosis, MCP, BYOK).
- Observability beyond basic logger init — metrics/tracing land later.
- Deployment config (`fly.toml`) — covered by `change/fly-staging-deploy` in week 4.

## Dependencies

None. This is the first code change.

## Risks

- **License choice locks public posture.** Choosing MIT vs AGPL vs BSL has downstream effects on commercial clone risk and enterprise adoption friction. Handled via a dedicated `@opus` decision task with explicit rationale recorded.
- **Env lib choice is sticky.** Swapping config libraries later is annoying once handlers read from them. The `@opus` decision task pins the choice with reasoning.
- **Over-scaffolding.** Temptation to add middleware, metrics, tracing stubs "while we're here." Don't. Every added dependency is a commitment. The acceptance criteria are deliberately minimal.
