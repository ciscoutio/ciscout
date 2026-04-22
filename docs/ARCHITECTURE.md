# CIScout — Architecture (v0.1)

**Status:** Working draft, 2026-04-23
**Scope:** Layer 1 only (CI failure diagnosis). Layers 2 and 3 are explicitly out of scope for this document.

---

## 1. Guiding principles

1. **Monolith first.** One Go binary, one Postgres, one deploy. Split only when scale or team demands it.
2. **Boring technology.** Nothing that requires explanation during a 2 AM debugging session.
3. **BYOK is structural, not a feature.** Customer's Anthropic key flows through our system to Anthropic and never touches our cost line.
4. **Synchronous acknowledgement, asynchronous work.** Webhooks return 200 in <500ms. All real work happens off the request path.
5. **Opinionated defaults, no configuration surface.** One trigger, one output format, one notification channel per install. Platform primitives are out of scope.

---

## 2. System overview

```
                                 ┌─────────────────────────┐
                                 │     GitHub.com          │
                                 │  (Actions + PR + API)   │
                                 └────────┬────────────────┘
                                          │ webhook
                                          ▼
┌──────────────────┐     ┌─────────────────────────────────────┐
│  Cloudflare      │     │  CIScout API (Go monolith, Fly.io)  │
│  Pages           │     │  ┌───────────────────────────────┐  │
│  - Landing       │     │  │ /webhooks/github (HMAC verify)│  │
│  - Dashboard SPA │◄───►│  │ /api/* (REST for dashboard)   │  │
│                  │     │  │ /mcp/* (MCP protocol endpoint)│  │
└──────────────────┘     │  └──────────────┬────────────────┘  │
                         │                 │                   │
                         │  ┌──────────────▼────────────────┐  │
                         │  │   Background worker (River)   │  │
                         │  │   - diagnosis jobs            │  │
                         │  │   - notification fanout       │  │
                         │  └───┬──────────┬──────────┬─────┘  │
                         └──────┼──────────┼──────────┼────────┘
                                ▼          ▼          ▼
                         ┌──────────┐ ┌────────┐ ┌─────────┐
                         │ Anthropic│ │ GitHub │ │Slack/   │
                         │   API    │ │  API   │ │Discord  │
                         │ (BYOK)   │ │        │ │webhooks │
                         └──────────┘ └────────┘ └─────────┘
                                │
                                ▼
                         ┌─────────────────┐
                         │ Postgres        │
                         │ (Fly.io managed)│
                         │ - all app data  │
                         │ - queue (River) │
                         └─────────────────┘
```

**Deploy units:**
- One Go binary running on Fly.io (API + worker + MCP in same process, different goroutines).
- One Postgres instance on Fly.io managed Postgres.
- One static site on Cloudflare Pages (landing + dashboard SPA).

No Redis, no separate queue service, no microservices. All added complexity requires explicit justification and PRODUCT.md update.

---

## 3. Primary data flow — CI failure diagnosis

Happy path, end-to-end, target latency ≤ 90 seconds from webhook to PR comment:

1. GitHub Actions workflow fails on a pull request.
2. GitHub sends `workflow_run` (or `check_run`) webhook with `conclusion = failure` to `POST /webhooks/github`.
3. API server:
   - Verifies HMAC-SHA256 signature using per-installation webhook secret.
   - Looks up `github_installations.id` by `installation_id` from payload.
   - Looks up `repositories.id` by `repo_id` from payload; bails if repo not enrolled.
   - Inserts a `diagnoses` row with `status = pending`.
   - Enqueues a River job `DiagnoseCIFailure{diagnosis_id, installation_id, repo_id, pr_number, run_id}`.
   - Returns `200 OK`. End of request path.
4. Background worker picks up the job:
   - Mints a GitHub App installation access token (1-hour lifetime, cached until 5 min before expiry).
   - Fetches, in parallel:
     - Failed job logs (truncated to last 50 KB per job, max 3 failed jobs).
     - PR diff (unified format, truncated to 200 KB).
     - Last 5 commits on the PR branch.
     - Last 10 commits on `main` (or default branch).
   - Loads the customer's encrypted Anthropic API key from `anthropic_keys`, decrypts using KEK from Fly secrets.
   - Constructs the prompt (template in `PROMPT.md`, shipped as a compiled Go string constant).
   - Calls Anthropic API (Sonnet 4.6 default, Opus fallback for low-confidence cases — decision deferred until `PROMPT.md` is written).
   - Parses the JSON response. On parse failure, retries once with a stricter "return JSON only" reminder. On second failure, marks diagnosis `failed` and posts a generic "diagnosis unavailable" PR comment.
   - Stores the structured diagnosis JSON in `diagnoses.result_json`.
   - Renders the human-facing markdown comment from the JSON using a template.
   - Posts the PR comment via GitHub API (`POST /repos/{owner}/{repo}/issues/{pr}/comments`). Stores the returned comment ID for future edits/updates.
   - Enqueues follow-up notification jobs: `NotifySlack`, `NotifyDiscord` (one or both, if configured).
   - Marks `diagnoses.status = complete`.
5. Downstream: the MCP server serves this diagnosis when queried by `get_ci_diagnosis(pr_number)` against the org's API token.

**Idempotency:** GitHub retries webhooks on non-2xx. Every webhook delivery carries a `X-GitHub-Delivery` UUID; store it in a `webhook_deliveries` table with a unique constraint. Drop duplicates at the door before enqueuing.

---

## 4. Components

### 4.1 API server (`api/`)

Single Go binary. Responsibilities:

- HTTP handlers under four route trees:
  - `/webhooks/*` — GitHub webhooks (no auth; HMAC verified per payload).
  - `/api/*` — Dashboard JSON API (auth via session cookie).
  - `/mcp/*` — MCP protocol endpoint (auth via API token bearer).
  - `/healthz`, `/readyz` — liveness/readiness probes.
- Runs the River worker in the same process on dedicated goroutines.
- Graceful shutdown: stop accepting new HTTP, drain worker, close DB.

**Stack:**
- Router: `chi` (lightweight, idiomatic) or `gin`. Default to `chi` — less magic.
- DB: `sqlx` + raw SQL. No ORM.
- Migrations: `goose` or `golang-migrate`. Default `goose`.
- Queue: `riverqueue/river` — Postgres-backed, supports retries, scheduled jobs, unique jobs.
- Logging: `log/slog` (stdlib). JSON output in prod.
- HTTP client: stdlib `net/http` with sensible timeouts (5 s connect, 30 s total for GitHub; 60 s for Anthropic).

### 4.2 Background worker

Same binary, same process. River jobs executed on worker pool. Job types for v0.1:

| Job | Trigger | Max attempts | Backoff |
|---|---|---|---|
| `DiagnoseCIFailure` | Webhook | 3 | Exponential, 30 s base |
| `NotifySlack` | Post-diagnosis | 5 | Exponential, 10 s base |
| `NotifyDiscord` | Post-diagnosis | 5 | Exponential, 10 s base |
| `RefreshInstallationToken` | Scheduled, every 45 min | 3 | Linear, 60 s |

No cross-job dependencies beyond the trivial "diagnosis must complete before notifications are queued." Notifications are best-effort; their failure does not fail the diagnosis.

### 4.3 MCP server (`mcp/` — served from same binary)

Exposes one tool: `get_ci_diagnosis(pr_number: int, repo: string)`.

- Protocol: MCP over HTTP (SSE or streamable HTTP — confirm against current MCP spec at implementation time).
- Auth: `Authorization: Bearer <api_token>` where token is a 32-byte random value minted in dashboard, stored hashed (SHA-256) in `mcp_api_tokens`.
- Resolution:
  - Token → `organization_id`.
  - `(organization_id, repo, pr_number)` → most recent `diagnoses` row.
  - Returns `result_json` or 404 if no diagnosis exists for that PR.
- Rate limit: 60 requests/minute per token.

**Language choice:** Go, co-located in the API binary. Reasons: simpler deploy, shared DB access, no extra service. If MCP Go SDK maturity becomes a blocker, break out as a separate service running Python SDK — defer that decision until implementation.

Full schema, request/response examples, and edge cases live in `MCP_SPEC.md` (not this doc).

### 4.4 Dashboard (`ui/`)

React + Tailwind SPA. Deployed to Cloudflare Pages.

Routes:
- `/` — marketing landing (may be a separate Astro/static site — decide at implementation; not architecturally significant).
- `/app` — authenticated dashboard.
  - `/app/install` — GitHub App install entry + repo selection.
  - `/app/setup` — Anthropic key paste + Slack/Discord channel config.
  - `/app/diagnoses` — history list.
  - `/app/diagnoses/:id` — detail view.
  - `/app/settings` — key rotation, MCP token management, billing.

Auth: session cookie set by API after GitHub OAuth login. Cookie is `HttpOnly`, `Secure`, `SameSite=Lax`.

### 4.5 Landing page

Static HTML (Astro or plain HTML — decide later). Cloudflare Pages. Waitlist form posts to `POST /api/waitlist` (unauthenticated, rate-limited by IP).

---

## 5. Data model (v0.1)

All tables include `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`, `created_at`, `updated_at`.

### 5.1 Identity

**`users`**
- `email` (unique, citext)
- `github_id` (unique, nullable until linked)
- `github_login`
- `name`

**`organizations`**
- `slug` (unique)
- `name`
- `owner_user_id` → users

**`organization_members`**
- `organization_id` → organizations
- `user_id` → users
- `role` enum(`owner`, `member`)
- unique(`organization_id`, `user_id`)

### 5.2 GitHub integration

**`github_installations`**
- `organization_id` → organizations
- `github_installation_id` bigint, unique
- `github_account_login`
- `github_account_type` enum(`User`, `Organization`)
- `webhook_secret` (encrypted; unique per install)
- `suspended_at` nullable

**`repositories`**
- `github_installation_id` → github_installations
- `github_repo_id` bigint, unique
- `owner` text
- `name` text
- `default_branch` text
- `enrolled` bool (whether CIScout is active for this repo)

**`webhook_deliveries`**
- `delivery_id` text, unique (from `X-GitHub-Delivery`)
- `event_type`, `received_at`, `processed` bool

### 5.3 Diagnoses

**`diagnoses`**
- `organization_id` → organizations
- `repository_id` → repositories
- `pr_number` int
- `github_run_id` bigint
- `github_comment_id` bigint nullable (after comment posted)
- `status` enum(`pending`, `complete`, `failed`)
- `failure_type` enum(`flaky_test`, `real_bug`, `infra`, `dependency`, `config`, `unknown`) nullable until complete
- `confidence` numeric(3,2) nullable
- `result_json` jsonb nullable (structured diagnosis; source of truth for MCP)
- `anthropic_input_tokens`, `anthropic_output_tokens` int nullable
- `error_message` text nullable
- `duration_ms` int nullable
- index on (`organization_id`, `created_at desc`)
- index on (`repository_id`, `pr_number`)

**`diagnosis_feedback`**
- `diagnosis_id` → diagnoses
- `rating` enum(`thumbs_up`, `thumbs_down`)
- `comment` text nullable
- `user_id` → users nullable (may be anonymous from PR reaction)

### 5.4 BYOK & tokens

**`anthropic_keys`**
- `organization_id` → organizations, unique
- `encrypted_key` bytea — AES-256-GCM ciphertext
- `nonce` bytea — GCM nonce
- `key_version` int — for KEK rotation
- `last_four` char(4) — for UI display
- `status` enum(`active`, `invalid`, `rotated`)
- `last_used_at` nullable

**`mcp_api_tokens`**
- `organization_id` → organizations
- `token_hash` bytea — SHA-256 of the actual token
- `name` text — human label ("Claude Code prod")
- `last_four` char(4)
- `last_used_at` nullable
- `revoked_at` nullable

### 5.5 Notifications

**`notification_channels`**
- `organization_id` → organizations
- `kind` enum(`slack`, `discord`)
- `webhook_url` (encrypted)
- `enabled` bool

### 5.6 Billing

**`subscriptions`**
- `organization_id` → organizations, unique
- `stripe_customer_id`, `stripe_subscription_id`
- `tier` enum(`free`, `solo`, `team`, `org`)
- `status` enum(`active`, `past_due`, `canceled`, `trialing`)
- `current_period_end` timestamp

### 5.7 Queue

**`river_jobs`** — managed by River library, not hand-rolled.

---

## 6. Integrations

### 6.1 GitHub App

- App registered at `github.com/apps/ciscout`.
- Permissions requested:
  - Repository: **Checks** (read), **Contents** (read), **Issues** (write), **Metadata** (read), **Pull requests** (write).
  - No organization permissions.
- Webhook events subscribed:
  - `workflow_run` (primary trigger).
  - `check_run` (fallback; some workflows use Checks API directly).
  - `installation`, `installation_repositories` (install lifecycle).
  - `repository` (rename/delete).
- Installation access tokens cached in-memory per worker, refreshed proactively 5 min before expiry.
- Secondary rate limits: respect `Retry-After` on 429. Log and alert if any org hits 1000 requests/hour.

### 6.2 Anthropic API (BYOK)

- Model: `claude-sonnet-4-6` default. Confidence < 0.6 → retry with `claude-opus-4-7` once.
- Prompt caching enabled for the system prompt and few-shot examples (static portion — high cache hit rate).
- Timeout: 60 s. One retry on network or 5xx errors.
- Customer's key is loaded per request, never kept in memory beyond the request lifetime.
- Errors:
  - 401 from Anthropic → mark key `invalid`, notify user via dashboard + email, skip subsequent diagnoses for that org until key is rotated.
  - 429 → exponential backoff, max 3 retries.
  - 5xx → one retry, then fail.

### 6.3 Slack / Discord

- Incoming webhook URLs configured by customer in dashboard (we don't operate our own Slack app in v0.1).
- Simple `POST` with JSON payload. 5-second timeout. 3 retries with exponential backoff. Non-delivery does not fail the diagnosis.

### 6.4 Stripe

- Checkout session hosted by Stripe.
- Webhook endpoint `POST /webhooks/stripe` — verify signature, update `subscriptions` row.
- Events handled: `customer.subscription.created`, `customer.subscription.updated`, `customer.subscription.deleted`, `invoice.payment_failed`.
- Tier enforcement done at repo-enrollment time and at diagnosis-trigger time. Over-quota = diagnosis skipped + user notified.

---

## 7. BYOK — key management & encryption

The most security-sensitive piece. Choices here set the "we never see your code" marketing story.

### 7.1 Encryption scheme

- **AES-256-GCM** for envelope encryption of the Anthropic key.
- **KEK (Key Encryption Key)**: 32 random bytes stored as Fly.io secret `CISCOUT_KEK`. One per environment (dev, staging, prod).
- **Encryption flow on save:**
  1. Generate random 12-byte nonce.
  2. `ciphertext = AES-256-GCM(KEK, nonce, plaintext_key, aad = organization_id)`.
  3. Store `(ciphertext, nonce, key_version = 1, last_four = plaintext[-4:])`.
  4. Zero plaintext buffer, return.
- **Decryption flow on use:**
  1. Load row by `organization_id`.
  2. Decrypt with current KEK. If current KEK fails, try previous KEK versions (rotation support).
  3. Return plaintext only to the Anthropic HTTP call; don't log, don't pass between goroutines.

### 7.2 Rotation

- KEK rotation: deploy a new KEK as `CISCOUT_KEK_V2`; background job re-encrypts every row with the new KEK and increments `key_version`; old KEK retained until every row is migrated, then removed.
- Customer key rotation: UI lets customer paste new key; old encrypted value overwritten.

### 7.3 Access control

- Keys never returned from any API endpoint. Dashboard shows only `••••••••{last_four}`.
- Request log scrubbing: ensure no HTTP middleware logs Anthropic request headers. Enforce via test.
- Database backups encrypted at rest (Fly.io default).

### 7.4 Trust narrative

Landing page can credibly say: *"We never store your code. CIScout encrypts your Anthropic key with AES-256-GCM, decrypts it only for the duration of a single Claude API call, and sends your logs and diff directly through Anthropic under your own key — you see every token billed in your Anthropic console."*

---

## 8. Security model

### 8.1 Authentication

- **Dashboard:** GitHub OAuth only (no password). Session cookie on success.
- **MCP:** API token (bearer). Tokens minted in dashboard, hashed with SHA-256 before storage.
- **Webhooks:** HMAC-SHA256 on every GitHub webhook; signed Stripe webhooks for billing.

### 8.2 Authorization

- Every DB query filtered by `organization_id` in WHERE clause.
- Middleware extracts `organization_id` from session or token; handlers never trust the client's supplied org ID.
- Multi-tenant boundary enforcement tested explicitly in CI.

### 8.3 Transport

- HTTPS only. `Strict-Transport-Security` header. No HTTP fallback.
- MCP endpoint TLS-terminated at Fly proxy.

### 8.4 Secret handling

- All secrets in Fly.io env (`KEK`, `GITHUB_APP_PRIVATE_KEY`, `STRIPE_SECRET`, `SESSION_KEY`, webhook secrets).
- No secrets in Git. Pre-commit hook scans for common patterns (gitleaks or similar).
- Development: `.env.local` git-ignored; template in `.env.example`.

### 8.5 Data protection

- Code/logs from customer repos flow through our server transiently. Never written to disk (in-memory only during diagnosis processing). `result_json` stores the diagnosis output, not the raw logs.
- Anthropic key, webhook secrets, Slack/Discord URLs: encrypted at rest (AES-256-GCM).
- Postgres backups encrypted at rest (Fly managed).

---

## 9. Observability

### 9.1 Logging

- Structured JSON via `log/slog`.
- Fields on every log line: `trace_id`, `org_id` (where applicable), `event`, `duration_ms`.
- **Never log**: Anthropic API keys, webhook secrets, full log payloads from customer CI, full diff content, PII.
- Log levels: `debug`, `info`, `warn`, `error`. Prod defaults to `info`.

### 9.2 Metrics

Prometheus-style, scraped by Fly metrics or exported to a small Grafana Cloud tier (free tier is sufficient for v0.1).

Core metrics:
- `ciscout_diagnosis_total{status, failure_type, org_id}` — counter.
- `ciscout_diagnosis_duration_seconds{status}` — histogram.
- `ciscout_anthropic_tokens_total{org_id, direction}` — counter (for BYOK cost visibility in dashboard).
- `ciscout_github_api_rate_limit_remaining{installation_id}` — gauge.
- `ciscout_webhook_received_total{event_type, outcome}` — counter.
- `ciscout_mcp_request_total{tool, status}` — counter.

### 9.3 Tracing

Skip OpenTelemetry in v0.1. Revisit in Layer 2.

### 9.4 Alerting

Minimal, via Fly Grafana → PagerDuty-lite or Slack DM to founder:
- Diagnosis failure rate > 10% over 15 min.
- Any 5xx rate > 2% sustained 5 min.
- Postgres connection pool saturated > 80%.
- Worker queue depth > 100.

---

## 10. Deployment topology

### 10.1 Environments

- **prod** — `ciscout.io`, `api.ciscout.io` (Fly), Postgres prod instance.
- **staging** — `staging.ciscout.io`, separate Fly app + Postgres. Used for pre-release testing on own repos.
- **dev** — local (Docker Compose for Postgres + the Go binary).

### 10.2 Infrastructure as code

- Fly.io config: `fly.toml` per environment, checked into repo.
- Cloudflare: config via dashboard initially; consider Terraform later if count of resources grows.
- No Kubernetes, no container orchestration, no service mesh.

### 10.3 CI/CD

- GitHub Actions on `push` to `main`:
  - `go test ./...`
  - `go vet ./...`, `staticcheck`.
  - Build binary → `flyctl deploy` to staging.
  - Manual promotion to prod via `flyctl deploy` with `--app ciscout-prod` label.
- Cloudflare Pages auto-deploys `ui/` and `landing/` on merge to `main`.
- Database migrations run on deploy startup via `goose up`.

---

## 11. Failure modes & recovery

| Failure | Detection | Behavior |
|---|---|---|
| GitHub webhook duplicated | `webhook_deliveries` unique constraint | Silent drop |
| Webhook HMAC invalid | Signature check fails | 401, log warning, no state change |
| Diagnosis job crash mid-processing | River retry | Re-attempt up to 3 times with backoff |
| Anthropic 401 (bad key) | Response parse | Mark key `invalid`, skip future diagnoses for org, notify user |
| Anthropic 429 | Response status | Retry with exponential backoff; after 3 fails, fail the diagnosis |
| Anthropic 5xx | Response status | One retry; if still failing, mark diagnosis `failed` with error |
| Anthropic JSON parse error | JSON unmarshal | One retry with "JSON only" nudge; else fallback to plain-text comment |
| GitHub API rate limit | 403 with `X-RateLimit-Remaining: 0` | Sleep until reset; log; alert if sustained |
| GitHub comment post fails | Non-2xx | Retry up to 3 times; mark diagnosis `complete` but flag `comment_post_failed` |
| Slack webhook down | 5xx / timeout | Retry up to 5 times; silently drop after; diagnosis unaffected |
| Postgres down | Connection failure | Return 503 for webhooks (GitHub will retry); readiness probe fails |
| KEK unavailable at boot | Startup fails | Refuse to start; CrashLoopBackoff; alert |

---

## 12. Scale considerations

Back-of-envelope sanity check for when things break. Not optimizing, just knowing:

- **100 paying customers, 50 diagnoses/day each = 5000/day ≈ 0.06/s.** Single `shared-cpu-2x` Fly machine handles this with headroom. Postgres `shared-cpu-1x` is fine.
- **1000 customers = 0.6/s.** Need one more worker machine, Postgres `dedicated-cpu-2x`. Still a monolith.
- **10,000 customers = 6/s.** Split worker from API. Move queue to dedicated Postgres or switch to Redis. Still no microservices.
- **GitHub API rate limit**: 5000 req/hour per installation. Each diagnosis = ~5 API calls. 1000 diagnoses/hour/install before limit. Not a concern at target scale.
- **Anthropic rate limits**: customer's problem (BYOK). They hit their own limits, not ours.
- **Postgres row counts at 1000 customers × 1000 diagnoses × 12 months = 12M rows.** Index-friendly; well within Postgres comfort zone. Archive old diagnoses to cold storage at 24 months.

---

## 13. Explicitly out of scope for v0.1 architecture

- Multi-region deployment.
- Read replicas.
- Redis or any non-Postgres queue.
- Kubernetes / container orchestration.
- Service mesh.
- Distributed tracing.
- Real-time streaming (WebSocket, SSE) in dashboard — polling is fine.
- Federated auth (SAML, OIDC) beyond GitHub OAuth.
- On-prem / self-hosted option.
- GDPR Data Processing Addendum flows — address when first EU enterprise customer asks.
- SOC 2 controls — v0.2 topic at earliest.

---

## 14. Open decisions (to resolve during implementation)

1. **Router choice:** `chi` vs `gin`. Default `chi` unless a concrete reason emerges.
2. **Migration tool:** `goose` vs `golang-migrate`. Default `goose`.
3. **Landing page framework:** Plain HTML + Tailwind, Astro, or a Next.js static export. Default plain HTML + Tailwind for v0 landing; revisit after week 2.
4. **MCP protocol transport:** SSE vs streamable HTTP. Confirm against current MCP spec at implementation time.
5. **Model fallback policy:** when to escalate Sonnet → Opus. Defer to `PROMPT.md`.
6. **Dashboard auth session store:** encrypted cookie vs server-side session table. Default encrypted cookie (stateless, simpler).

---

## 15. References to sibling documents

- `PRODUCT.md` — product scope, roadmap, out-of-scope features.
- `MCP_SPEC.md` (next) — MCP tool schema, request/response examples.
- `PROMPT.md` (after MCP_SPEC) — system prompt, few-shot examples, output schema.
- `EXECUTION_PLAN.md` — 12-week build plan with gates.
- `LANDING.md` — landing page copy.
