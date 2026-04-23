# CIScout

CIScout is an AI-powered CI failure diagnosis product for GitHub Actions. When a CI workflow fails, CIScout posts a structured diagnosis to the PR as both a human-readable comment and an MCP signal for AI agents like Claude Code.

See [`docs/PRODUCT.md`](docs/PRODUCT.md) for the full product scope.

## Quick Start

```bash
git clone https://github.com/ciscoutio/ciscout.git
cd ciscout
cp .env.example .env.local
make dev
```

Server listens on `localhost:8080` by default. Hit `http://localhost:8080/healthz` to verify.

### Pre-commit hooks

```bash
pipx install pre-commit
pre-commit install
pre-commit run --all-files
```

## Project Structure

```
cmd/ciscout/       Entry point
internal/config/   Config loading
internal/httpserver/ HTTP server
api/, mcp/, ui/, landing/ — Future features
```

## Docs

- [`PRODUCT.md`](docs/PRODUCT.md) — Product scope
- [`ARCHITECTURE.md`](docs/ARCHITECTURE.md) — Technical design
- [`WORKFLOW.md`](docs/WORKFLOW.md) — Change process
- [`MCP_SPEC.md`](docs/MCP_SPEC.md) — MCP contract
- [`PROMPT.md`](docs/PROMPT.md) — Diagnosis prompt
- [`EXECUTION_PLAN.md`](docs/EXECUTION_PLAN.md) — 12-week plan
- [`LANDING.md`](docs/LANDING.md) — Marketing copy

## License

[MIT License](LICENSE)
