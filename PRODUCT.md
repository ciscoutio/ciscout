# CIScout — Product Definition

**Status:** Working draft, 2026-04-22
**Domain:** ciscout.io
**GitHub org:** github.com/ciscoutio

---

## 1. Product statement (one-liner)

> **Structured CI/CD diagnosis for humans and AI agents. When CI fails, get a PR comment for your team and an MCP signal that Claude Code / Cursor / Codex can act on.**

---

## 2. Three-layer roadmap

```
LAYER 3 (year 2):  CI Health Advisory · CD failure diagnosis · Incident response
                   ↑ Earned by customer pull + accumulated trust
LAYER 2 (month 4–6): CI insights dashboard (observed patterns across diagnoses)
                   ↑ Earned by Layer 1 at volume
LAYER 1 (v0.1, weeks 1–12): CI failure diagnosis (PR comment + MCP + Slack)
                   ↑ Ship this now
```

### Layer 1 — CI failure diagnosis (active)
GitHub App that analyzes failed CI runs, posts PR comments with diagnosis, and exposes the same diagnosis to coding agents via MCP.

### Layer 2 — CI insights (deferred, v0.2)
Monthly per-repo reports: flaky tests, slowest jobs, failure-type breakdown. Emerges from data already collected in Layer 1.

### Layer 3 — Advisory + CD + incident (deferred, v1.0)
Actual recommendations for CI restructuring, deployment failure diagnosis, production incident response. Enters Nairi AI SRE / Incident.io territory — requires trust and data earned in prior layers.

### Why this sequencing
- A v0.1 promising CI+CD+advisory ships mediocre all three and wins none.
- Advisory as a lead feature is a consulting pitch, not a tool pitch — nobody pays $79/mo to be told they're doing CI wrong. Advisory as data-backed observation is defensible.
- CD diagnosis requires 5x more integrations and carries prod-stakes trust — wrong starting point solo.
- Each layer compounds: Layer 1 data makes Layer 2 possible; Layer 2 patterns make Layer 3 credible.

---

## 3. Layer 1 — Feature list (v0.1 scope)

### 3.1 Core: diagnosis pipeline
- **Trigger:** GitHub `check_run` / `workflow_run` webhook with conclusion `failure`.
- **Context gathering:** fetch failed job logs, PR diff, last 5 branch commits, last 10 main commits.
- **LLM call:** single Claude API call (model configurable; default Sonnet 4.6) with opinionated system prompt. BYOK — customer's API key.
- **Output:** structured JSON:
  ```
  {
    failure_type: "flaky_test" | "real_bug" | "infra" | "dependency" | "config",
    confidence: 0.0–1.0,
    summary: "one-line human diagnosis",
    suspect_files: [{path, line_range}],
    suspect_commits: [sha],
    reasoning: "markdown explanation",
    suggested_fix: { file, diff, rationale } | null,
    raw_logs_excerpt: "..."
  }
  ```

### 3.2 Output surface: GitHub PR comment
- Rendered from structured JSON to markdown.
- Posted to the PR whose CI failed.
- Contains: summary, failure type, suspect files as GitHub line links, suggested fix as code block, confidence indicator.
- Includes thumbs up/down feedback buttons (via GitHub comment reactions or an inline form).

### 3.3 Output surface: MCP server
- MCP server exposing one tool: `get_ci_diagnosis(pr_number: int)`.
- Returns the structured JSON above.
- Authentication: customer's CIScout API token (generated in dashboard).
- Deployment: hosted MCP endpoint; customer configures their Claude Code / Cursor / Codex to point at it.
- **This is the positioning differentiator. It must be in v0.1, not deferred.**

### 3.4 Output surface: Slack / Discord notification
- One channel per install, configured at install time.
- Posts a short summary with a link back to the PR comment.
- Opinionated: no per-repo routing rules, no filters. If granularity needed later, add in v0.2.

### 3.5 BYOK (Bring Your Own Key)
- Customer provides Anthropic API key during onboarding.
- Encrypted at rest (AES-256, key stored in secrets manager).
- Key never logged, never shown in dashboard after save.
- Marketing angle: "we never store your code — diagnosis runs against your Claude key, your data path."

### 3.6 Dashboard (minimal)
- **Install flow:** GitHub App install → select repos → paste Anthropic key → configure Slack/Discord channel.
- **Diagnosis history:** list of recent diagnoses with PR link, failure type, confidence, thumbs feedback.
- **Settings:** rotate API key, manage repo list, MCP endpoint config with copy-paste instructions, billing.
- **No:** fleet UI, agent builder, skill marketplace, prompt editor, vault manager — those are Nairi territory and out of scope.

### 3.7 Billing
- Stripe integration.
- Tiers:
  - **Free:** 1 public repo, no private repos, unlimited diagnoses.
  - **Solo ($29/mo):** 1 private repo.
  - **Team ($149/mo):** up to 20 private repos, Slack/Discord output enabled.
  - **Org ($499/mo):** unlimited repos, MCP access, priority support.
- BYOK = no LLM cost on our P&L; margins clean.

### 3.8 Feedback loop (for product improvement)
- Thumbs up/down on each diagnosis.
- Optional free-text feedback field.
- Stored per customer (not shared); used to tune prompts and identify failure-type gaps.

---

## 4. Layer 1 — Explicitly out of scope

These are NOT v0.1 features. Every expansion here kills the ship date.

- ❌ CD / deployment failure diagnosis
- ❌ Production incident response
- ❌ CI structure advisory / recommendations
- ❌ Flaky test tracking dashboard (Layer 2)
- ❌ Multiple LLM support (Claude only in v0.1)
- ❌ Self-hosted daemon
- ❌ Multi-agent / fleet UI
- ❌ Configurable skills / MCPs / rules (user-facing)
- ❌ Prompt editor
- ❌ Vault manager
- ❌ GitLab / Bitbucket support (GitHub only)
- ❌ Jenkins / CircleCI / Buildkite / Travis support (GitHub Actions only)
  - Rule for reconsidering: 5+ paying customers explicitly ask AND 2+ commit to a higher tier to fund the integration AND product-market fit is proven on GHA first.
- ❌ PR review (subjective code quality review — that's CodeRabbit's job)
- ❌ Email notifications (Slack/Discord only)
- ❌ SSO / SAML (Org tier only, v0.2+)

---

## 5. Tech stack (proposed, for v0.1)

- **Backend:** Go + Gin.
- **Data:** Postgres (installs, repos, diagnoses, users, api_keys).
- **Queue:** Redis or Postgres-backed (avoid new infra). Diagnosis jobs processed async off webhook.
- **Deploy (backend):** Fly.io.
- **Deploy (landing + dashboard):** Cloudflare Pages.
- **GitHub:** GitHub App (webhooks + App token auth).
- **LLM:** Anthropic API directly (customer BYOK).
- **MCP:** MCP server — language chosen during architecture phase.
- **Dashboard:** React + Tailwind.
- **Billing:** Stripe.
- **Domain registrar:** Cloudflare Registrar.

---

## 6. Infrastructure decisions (locked)

| Decision | Choice |
|---|---|
| Product name | CIScout |
| Domain | ciscout.io |
| GitHub org | github.com/ciscoutio |
| Landing page hosting | Cloudflare Pages |
| Backend hosting | Fly.io |
| Domain registrar | Cloudflare Registrar |
| Code language (backend) | Go |

---

## 7. Open decisions

1. **MCP server language** — Go (matches backend) vs. TypeScript/Python (official SDKs more mature). Decide during architecture phase.
2. **Pricing validation** — $29 / $149 / $499 is a guess. Validate with 5 early customers before locking.
3. **Trademark check for "CIScout"** — verify no conflict via USPTO/EUIPO before significant marketing spend.
4. **Acronis employment contract review** — check IP assignment / moonlighting / disclosure clauses before committing branded product code under personal name.

---

## 8. Next artifacts to draft

In this order — each feeds the next:

1. `ARCHITECTURE.md` — end-to-end technical spec: webhook flow, data model, BYOK storage, MCP server, deployment topology.
2. `MCP_SPEC.md` — exact MCP tool schema + 3–5 example JSON outputs for different failure types.
3. `PROMPT.md` — system prompt driving diagnosis quality, with few-shot examples. Core IP.
4. `EXECUTION_PLAN.md` — 12-week breakdown with weekly deliverables and phase gates.
5. `LANDING.md` — landing page copy leading with dual-output positioning.
