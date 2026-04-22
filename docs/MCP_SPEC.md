# CIScout — MCP Server Specification (v0.1)

**Status:** Working draft, 2026-04-23
**Endpoint:** `https://mcp.ciscout.io`
**Protocol:** Model Context Protocol (MCP) over streamable HTTP

---

## 1. Purpose

The MCP server exposes CIScout diagnoses to coding agents (Claude Code, Cursor, Codex, and any MCP-compatible client). It is the primary product differentiator — the reason CIScout's output is useful to agents, not just humans.

**Usage pattern:**
1. Developer's CI fails on a PR.
2. CIScout's GitHub App posts a human-facing PR comment (their teammate reads it).
3. The developer opens their coding agent and asks "why did CI fail on my PR?"
4. The agent calls `get_ci_diagnosis(repo, pr_number)` against the CIScout MCP server.
5. The agent receives structured diagnosis JSON including a suggested fix.
6. The agent applies the fix, commits, pushes — CI turns green.

The diagnosis is the upstream signal; the coding agent is the downstream actuator. CIScout does not try to be the agent that fixes the bug.

---

## 2. Tools exposed (v0.1)

Two tools. Every addition requires PRODUCT.md update and architectural review.

### 2.1 `get_ci_diagnosis`

Fetch the most recent diagnosis for a specific pull request.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "repo": {
      "type": "string",
      "pattern": "^[^/]+/[^/]+$",
      "description": "GitHub repository in owner/name format, e.g. 'acme/api-server'."
    },
    "pr_number": {
      "type": "integer",
      "minimum": 1,
      "description": "Pull request number the CI failure occurred on."
    }
  },
  "required": ["repo", "pr_number"]
}
```

**Returns:** `DiagnosisResult` (see §3) for the most recent completed diagnosis on that PR, or a structured 404 if none exists.

### 2.2 `list_diagnoses`

List recent diagnoses across a repo. Useful when the coding agent needs context without knowing a specific PR number.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "repo": {
      "type": "string",
      "pattern": "^[^/]+/[^/]+$"
    },
    "limit": {
      "type": "integer",
      "minimum": 1,
      "maximum": 50,
      "default": 10
    },
    "status": {
      "type": "string",
      "enum": ["complete", "failed", "all"],
      "default": "complete"
    }
  },
  "required": ["repo"]
}
```

**Returns:** `{ "diagnoses": DiagnosisSummary[] }` — most recent first. Each entry contains the summary fields only (see §3.2), not the full result. Callers fetch full detail via `get_ci_diagnosis` when needed.

---

## 3. Data shapes

### 3.1 `DiagnosisResult` — full diagnosis payload

```json
{
  "diagnosis_id": "3f8c1b2a-...-uuid",
  "created_at": "2026-04-23T14:22:08Z",
  "completed_at": "2026-04-23T14:22:19Z",
  "repo": "acme/api-server",
  "pr_number": 847,
  "pr_url": "https://github.com/acme/api-server/pull/847",
  "github_run_id": 12389476123,
  "github_run_url": "https://github.com/acme/api-server/actions/runs/12389476123",
  "status": "complete",
  "failure_type": "flaky_test",
  "confidence": 0.87,
  "summary": "Intermittent race condition in session cleanup; test passes on retry.",
  "reasoning": "The failing test `TestLoginTimeout` relies on a goroutine in `internal/auth/session.go` that cleans up expired sessions without synchronization. Under load (observed twice in the last 30 runs), the cleanup goroutine completes before the test assertion reads session state, causing an intermittent nil-pointer dereference. This is not caused by changes in this PR — the PR diff modifies only `api/handlers/billing.go`.",
  "suspect_files": [
    {
      "path": "internal/auth/session.go",
      "line_range": [142, 168],
      "reason": "Concurrent write to `sessions` map from cleanup goroutine without mutex."
    }
  ],
  "suspect_commits": [],
  "suggested_fix": {
    "file": "internal/auth/session.go",
    "diff": "@@ -140,3 +140,5 @@\n func (s *Store) cleanup() {\n+\ts.mu.Lock()\n+\tdefer s.mu.Unlock()\n \tfor id, sess := range s.sessions {\n \t\tif sess.expired() {",
    "rationale": "Guard the cleanup loop with the store's existing mutex. The mutex is held by `Get` and `Set` but missing from `cleanup`."
  },
  "raw_logs_excerpt": "--- FAIL: TestLoginTimeout (0.12s)\n    session_test.go:58: expected session, got nil\n...",
  "anthropic_usage": {
    "input_tokens": 12450,
    "output_tokens": 890,
    "model": "claude-sonnet-4-6"
  },
  "metadata": {
    "duration_ms": 11240,
    "pr_comment_url": "https://github.com/acme/api-server/pull/847#issuecomment-4421882"
  }
}
```

### 3.2 `DiagnosisSummary` — compact payload for list views

```json
{
  "diagnosis_id": "3f8c1b2a-...-uuid",
  "created_at": "2026-04-23T14:22:08Z",
  "pr_number": 847,
  "pr_url": "https://github.com/acme/api-server/pull/847",
  "status": "complete",
  "failure_type": "flaky_test",
  "confidence": 0.87,
  "summary": "Intermittent race condition in session cleanup; test passes on retry."
}
```

### 3.3 Field contracts

| Field | Type | Notes |
|---|---|---|
| `diagnosis_id` | UUID string | Stable identifier; use for deduplication. |
| `status` | enum | `complete` / `failed`. Callers should only expect full fields when `complete`. |
| `failure_type` | enum | `flaky_test`, `real_bug`, `infra`, `dependency`, `config`, `unknown`. |
| `confidence` | float 0.0–1.0 | Model-reported; treat < 0.5 as "hint, not conclusion." |
| `summary` | string | One-line, < 200 chars. Safe to display inline. |
| `reasoning` | string | Markdown. May contain multi-paragraph analysis. Show on request, not always. |
| `suspect_files` | array | May be empty. `line_range` is inclusive `[start, end]`. |
| `suspect_commits` | array of SHAs | May be empty. Populated when a specific commit introduces the failure. |
| `suggested_fix` | object or null | Null when confidence is low or the fix requires human judgment. |
| `suggested_fix.diff` | string | Unified diff format. Single-file only in v0.1. |
| `raw_logs_excerpt` | string | ≤ 8 KB. The specific failing lines, not full job logs. |
| `anthropic_usage` | object | For BYOK cost visibility; not normally shown to end users. |

---

## 4. Five canonical examples

One example per `failure_type` to pin the shape.

### 4.1 `flaky_test`

See §3.1 — the running example above.

### 4.2 `real_bug` (introduced by this PR)

```json
{
  "diagnosis_id": "7b2d3e4f-...",
  "status": "complete",
  "failure_type": "real_bug",
  "confidence": 0.94,
  "summary": "PR introduces nil dereference in `calculateDiscount` when `customer.Plan` is unset.",
  "reasoning": "The diff in `billing/discount.go:34` calls `customer.Plan.Tier` unconditionally. Earlier in the function, `customer.Plan` is loaded via `loadPlan(id)` which returns nil for customers without an active subscription. The test `TestFreeCustomerCheckout` exercises this code path and fails with `SIGSEGV` at the dereference.",
  "suspect_files": [
    {
      "path": "billing/discount.go",
      "line_range": [34, 34],
      "reason": "Unchecked nil pointer dereference — `customer.Plan.Tier` when `Plan` is nil for free-tier customers."
    }
  ],
  "suspect_commits": ["e3f1a9b"],
  "suggested_fix": {
    "file": "billing/discount.go",
    "diff": "@@ -32,3 +32,6 @@\n func calculateDiscount(customer *Customer) decimal.Decimal {\n+\tif customer.Plan == nil {\n+\t\treturn decimal.Zero\n+\t}\n \tswitch customer.Plan.Tier {",
    "rationale": "Guard against nil Plan. Free-tier customers have no subscription and should receive no discount (zero)."
  },
  "raw_logs_excerpt": "=== RUN   TestFreeCustomerCheckout\n    panic: runtime error: invalid memory address or nil pointer dereference\n    billing/discount.go:34 +0x28\n...",
  "anthropic_usage": { "input_tokens": 14220, "output_tokens": 1103, "model": "claude-sonnet-4-6" }
}
```

### 4.3 `infra`

```json
{
  "diagnosis_id": "a1b2c3d4-...",
  "status": "complete",
  "failure_type": "infra",
  "confidence": 0.82,
  "summary": "GitHub Actions runner could not reach npm registry; not a code issue.",
  "reasoning": "The `npm install` step failed with `ETIMEDOUT` connecting to `registry.npmjs.org`. No changes to `package.json`, `package-lock.json`, or CI config in this PR. Prior 20 runs on this repo succeeded with identical dependencies. Most likely transient npm or GitHub runner networking issue.",
  "suspect_files": [],
  "suspect_commits": [],
  "suggested_fix": null,
  "raw_logs_excerpt": "npm ERR! network request to https://registry.npmjs.org/react failed, reason: ETIMEDOUT\nnpm ERR! A complete log of this run can be found in: /home/runner/.npm/_logs/...",
  "anthropic_usage": { "input_tokens": 8900, "output_tokens": 412, "model": "claude-sonnet-4-6" }
}
```

### 4.4 `dependency`

```json
{
  "diagnosis_id": "c9e8d7f6-...",
  "status": "complete",
  "failure_type": "dependency",
  "confidence": 0.91,
  "summary": "`go.sum` mismatch — `github.com/go-chi/chi/v5` was upgraded in `go.mod` but `go.sum` not regenerated.",
  "reasoning": "The PR upgrades `github.com/go-chi/chi/v5` from v5.0.10 to v5.1.0 in `go.mod`, but `go.sum` still contains only the v5.0.10 checksum. The CI `go build` step fails with `missing go.sum entry for module providing package github.com/go-chi/chi/v5 (imported by cmd/server)`. Standard fix: run `go mod tidy` locally and commit.",
  "suspect_files": [
    {
      "path": "go.sum",
      "line_range": [1, 1],
      "reason": "Missing entries for upgraded chi/v5 v5.1.0."
    }
  ],
  "suspect_commits": ["d4e5f6a"],
  "suggested_fix": {
    "file": "go.sum",
    "diff": "# Run locally:\n#   go mod tidy\n#   git add go.sum && git commit --amend --no-edit && git push --force-with-lease\n",
    "rationale": "`go.sum` must be regenerated after any go.mod dependency change. CIScout cannot produce the exact checksums without running `go mod download`; the command above handles it."
  },
  "raw_logs_excerpt": "cmd/server/main.go:12:2: missing go.sum entry for module providing package github.com/go-chi/chi/v5\nrun: 'go mod tidy' to add missing entries\n",
  "anthropic_usage": { "input_tokens": 11200, "output_tokens": 678, "model": "claude-sonnet-4-6" }
}
```

### 4.5 `config`

```json
{
  "diagnosis_id": "f1e2d3c4-...",
  "status": "complete",
  "failure_type": "config",
  "confidence": 0.88,
  "summary": "Workflow YAML references secret `DOCKER_HUB_TOKEN` which is not configured on this repository.",
  "reasoning": "The `publish.yml` workflow step `Login to Docker Hub` uses `${{ secrets.DOCKER_HUB_TOKEN }}`, but the secret does not exist on this repo or the organization. The step fails with `Error: Username and password required`. This is a configuration problem, not code; the secret must be added via Settings → Secrets and variables → Actions.",
  "suspect_files": [
    {
      "path": ".github/workflows/publish.yml",
      "line_range": [22, 26],
      "reason": "Step references undefined secret `DOCKER_HUB_TOKEN`."
    }
  ],
  "suspect_commits": [],
  "suggested_fix": null,
  "raw_logs_excerpt": "Run docker/login-action@v3\nError: Username and password required\n",
  "anthropic_usage": { "input_tokens": 7850, "output_tokens": 521, "model": "claude-sonnet-4-6" }
}
```

---

## 5. Transport & protocol

### 5.1 Transport

Streamable HTTP (MCP spec 2025-03-26 or later). Endpoint: `POST https://mcp.ciscout.io/v1`.

Single endpoint, content-type `application/json`, Server-Sent Events for streaming responses. TLS only. No HTTP fallback.

### 5.2 Authentication

Bearer token in `Authorization` header:

```
Authorization: Bearer cis_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Token format: `cis_` prefix, followed by environment marker (`live` / `test`), followed by 32-char base32 random. Tokens are minted in the CIScout dashboard and displayed once at creation (plaintext); only SHA-256 hash is stored (see ARCHITECTURE.md §5.4).

Token scope: organization-level. Resolves to exactly one `organization_id`. All queries scoped to that org's repos; cross-org access returns 404, not 403, to avoid leaking repo existence.

### 5.3 Versioning

- `/v1` path prefix on the endpoint.
- Breaking changes → `/v2`. `/v1` kept live for at least 12 months after deprecation announced.
- Additive changes (new fields on existing responses) are not breaking; clients must ignore unknown fields.

---

## 6. Error model

All errors returned as MCP `ToolError` objects with stable `code` identifiers:

| Code | HTTP | Meaning | Client action |
|---|---|---|---|
| `auth_missing` | 401 | No bearer token provided | Prompt user to configure token |
| `auth_invalid` | 401 | Token doesn't match any active org | Prompt user to regenerate token |
| `auth_revoked` | 401 | Token was revoked in dashboard | Prompt user to mint a new one |
| `repo_not_enrolled` | 404 | Repo not installed in CIScout | Show install CTA |
| `diagnosis_not_found` | 404 | No diagnosis exists for this PR | Inform user CI hasn't failed on this PR (yet) |
| `rate_limited` | 429 | Token exceeded rate limit | Back off per `Retry-After` |
| `invalid_argument` | 400 | Schema validation failed | Fix caller logic |
| `internal_error` | 500 | Unexpected server failure | Retry once with backoff; surface to user on second failure |

Error response shape:
```json
{
  "error": {
    "code": "diagnosis_not_found",
    "message": "No completed diagnosis found for acme/api-server PR #847.",
    "details": { "repo": "acme/api-server", "pr_number": 847 }
  }
}
```

---

## 7. Rate limits

- **60 requests per minute per token.** Exceeded → `429 rate_limited`.
- **Concurrency cap:** 5 in-flight requests per token.
- **Response headers** on every response:
  - `X-RateLimit-Limit: 60`
  - `X-RateLimit-Remaining: <n>`
  - `X-RateLimit-Reset: <unix_seconds>`
- Rate limits apply per token, not per org. An org can mint multiple tokens if separate agents need isolated quotas.

---

## 8. Client configuration

### 8.1 Claude Code

`~/.config/claude-code/mcp.json` (or wherever Claude Code stores MCP config):

```json
{
  "mcpServers": {
    "ciscout": {
      "url": "https://mcp.ciscout.io/v1",
      "headers": {
        "Authorization": "Bearer cis_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      }
    }
  }
}
```

Copy-pasteable block is surfaced in the CIScout dashboard (Settings → MCP → "Configure Claude Code") with the user's token pre-filled.

### 8.2 Cursor

Cursor settings → MCP section:
- Name: `ciscout`
- Type: `HTTP`
- URL: `https://mcp.ciscout.io/v1`
- Header: `Authorization: Bearer cis_live_...`

### 8.3 Generic MCP client

Any MCP-spec-compliant client can connect. Required capabilities: streamable HTTP transport, bearer auth in request headers.

---

## 9. Security considerations

- **Tokens are secrets.** Never commit to repos. CIScout will not scan for leaked tokens in v0.1, but will rotate on user request.
- **Diagnosis payloads can contain code snippets and log lines** from the customer's repository. Treat the MCP response as sensitive as the customer's source code.
- **Cross-org isolation** is enforced by token → `organization_id` binding in query filters, tested in CI.
- **Request logging** strips the `Authorization` header and never logs response bodies for MCP endpoints.
- **Client-side responsibility:** coding agents consuming the response should honor the customer's data residency expectations. CIScout cannot control what the client does with the fetched JSON.

---

## 10. Out of scope for v0.1

- WebSocket transport (only streamable HTTP).
- Subscription-style tools (e.g., "notify me when diagnosis completes"). Clients poll via `get_ci_diagnosis` or respond to the PR comment directly.
- Write tools (e.g., `submit_feedback`, `trigger_reanalysis`). Possible in v0.2.
- OAuth / dynamic token exchange. Bearer tokens only for v0.1.
- Per-repo tokens. Org-level scope only.
- Multi-tenant queries (e.g., fetch diagnoses across orgs). Not a valid use case.
- Non-GitHub repo support in the tool schema (the `repo` field is a GitHub `owner/name` pair only).

---

## 11. Open decisions

1. **Token format variant.** `cis_live_*` vs `cis_prod_*` — pick one before shipping.
2. **Rate-limit bucket window.** Fixed 60/min vs token-bucket with burst. Start fixed for simplicity.
3. **Streaming vs. single response.** Current responses fit in a single HTTP response easily. Streaming capability reserved for future tools that return partial progress (e.g., `watch_diagnosis`).
4. **Client ID / user-agent normalization.** Whether to require clients to send a `User-Agent` identifying which agent called us. Useful for metrics; not required in v0.1.

---

## 12. References

- MCP specification: https://spec.modelcontextprotocol.io/
- ARCHITECTURE.md §4.3 — server-side implementation notes.
- PROMPT.md (next) — the Claude prompt producing the `DiagnosisResult` payload.
- PRODUCT.md §3.3 — product-level definition of the MCP surface.
