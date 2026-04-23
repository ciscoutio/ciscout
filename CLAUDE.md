# CLAUDE.md

This file is loaded by Claude Code when working in the CIScout repository.

## What CIScout is

CIScout is an AI-powered CI failure diagnosis product for teams using GitHub Actions. When CI fails on a PR, CIScout posts a structured diagnosis as both a PR comment (human-facing) and an MCP signal consumable by Claude Code / Cursor / Codex (agent-facing). Customers bring their own Anthropic API key (BYOK).

Full product definition lives in `docs/PRODUCT.md`. Do not assume details — read the doc.

## How we work

This project uses a lightweight OpenSpec workflow with model-tagged task decomposition (`@opus` / `@sonnet` / `@haiku`). Before proposing or executing any significant code change, read `docs/WORKFLOW.md`.

Primary authoritative docs (in priority order):

- `docs/PRODUCT.md` — product scope, three-layer roadmap, explicit out-of-scope items
- `docs/WORKFLOW.md` — how changes flow from proposal to merge, model-tagging rubric
- `docs/ARCHITECTURE.md` — technical architecture for v0.1 (components, data model, integrations, BYOK scheme)
- `docs/MCP_SPEC.md` — external MCP server contract (tool schemas, client configuration)
- `docs/PROMPT.md` — diagnosis prompt specification (core IP)
- `docs/EXECUTION_PLAN.md` — 12-week strategic calendar with weekly deliverables and phase gates
- `docs/LANDING.md` — landing page source copy

In-flight OpenSpec changes live in `openspec/changes/<slug>/`. Archived ones are in `openspec/archive/`.

## Tech stack (locked for v0.1)

- Go 1.26+ monolith deployed on Fly.io
- Postgres (Fly.io managed) as the only datastore; no Redis
- `chi` router, `sqlx`, River queue, goose migrations, stdlib `log/slog`
- React + Tailwind dashboard on Cloudflare Pages
- Anthropic API via customer's own key (BYOK — AES-256-GCM envelope encryption)

## Common commands

To be populated after `openspec/changes/repo-scaffolding/` merges. That change adds the `Makefile` with `make dev`, `make test`, `make build`, `make lint`, `make fmt`, `make tidy`.

## Working agreements

- **Default model assignments:** `@sonnet` for implementation, `@opus` for planning / architecture / review, `@haiku` for boilerplate. See `docs/WORKFLOW.md` §6 for the full rubric.
- **Scope discipline:** resist expanding v0.1 beyond `docs/PRODUCT.md §4` (explicit out-of-scope). Solo-founder scope creep is the top project risk — flagged in `docs/EXECUTION_PLAN.md §5`.
- **BYOK is structural, not a feature.** The customer's Anthropic key never becomes our P&L. This affects encryption, UI, and marketing narrative.
- **MCP output is the positioning differentiator.** Do not treat it as "just a feature." See `docs/MCP_SPEC.md`.
- **Tool-use for diagnosis output.** Never prompt Claude to "return JSON" — always use the `record_diagnosis` tool defined in `docs/PROMPT.md §4`.

## What Claude should push back on

- Any proposal to add CD (deployment) diagnosis, CI advisory features, multi-LLM support, self-hosted daemons, agent-platform primitives, or non-GitHub-Actions CI systems to v0.1.
- Any plan that bundles unrelated changes into one OpenSpec proposal.
- Any tempted cleanup / refactoring outside the scope of a specific tasks.md entry.

## History

CIScout was renamed from the earlier `Groven` project on 2026-04-22. The old Groven codebase lives at `~/dev/groven/` and is retained for historical reference only — it is not being developed and does not share code with CIScout.
