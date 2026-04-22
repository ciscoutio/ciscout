# CIScout — 12-Week Execution Plan

**Status:** Working draft, 2026-04-23
**Horizon:** 90 days from project start (week 1 = week of 2026-04-27).
**Success target at day 90:** 5 paying customers, $500 MRR, public launch executed, 2× audience growth on LinkedIn.

---

## 1. North Star metric

**Paying customers at day 90.** Not signups, not waitlist, not GitHub stars. Paid customers on Stripe.

Everything else (content, landing polish, MCP marketplace presence) is instrumental to this number.

---

## 2. Phase overview

| Phase | Weeks | Outcome | Gate to next phase |
|---|---|---|---|
| 0 — Setup | Now | Domain, org, repo, landing stub live | All infrastructure decisions locked; PRODUCT/ARCH/MCP/PROMPT docs merged |
| 1 — Foundations | 1–2 | Go monolith + Postgres scaffold; GitHub App registered; CI pipeline green | Can run tests locally; staging deploy works |
| 2 — Core diagnosis loop | 3–5 | End-to-end: webhook → Claude call → PR comment, on test repos | One real diagnosis posted to a friend's PR within target latency |
| 3 — BYOK + MCP | 6–7 | Customer-provided key, MCP server serving diagnoses | Claude Code queries MCP and receives real data |
| 4 — Dashboard + billing | 8 | Minimal dashboard, Stripe live | First paying customer ($29 Solo) |
| 5 — Private beta | 9–10 | 5–10 teams installed, feedback loop running | Evaluation corpus ≥ 80% accuracy; thumbs-up rate ≥ 75% |
| 6 — Public launch | 11–12 | Show HN, LinkedIn launch, docs polished | 5 paying customers, $500 MRR |

---

## 3. Parallel tracks

Two tracks run in parallel every week. Each has its own budget of hours/focus.

- **Build track** — the code, the docs, the infrastructure. ~70% of weekly focus.
- **Content track** — building in public, audience growth, outreach. ~30% of weekly focus.

Neither track pauses for the other. Content without product is noise; product without content is invisible.

---

## 4. Week-by-week plan

Each week lists: build deliverables, content deliverables, and the gate (what must be true by Friday).

### Phase 0 — Setup (this week, already in flight)

**Build:**
- ✅ `PRODUCT.md`, `ARCHITECTURE.md`, `MCP_SPEC.md`, `PROMPT.md`, `WORKFLOW.md` merged to main.
- ⏳ Domain `ciscout.io` purchased (Cloudflare Registrar).
- ⏳ GitHub App "CIScout" registered (reserve the name; no code yet).
- ⏳ Acronis contract review completed.
- ⏳ Trademark check on "CIScout" completed.

**Content:**
- None yet — nothing to announce until the docs stabilize.

**Gate:** All docs live on `main`; domain purchased; GitHub org exists; Acronis contract risk cleared or mitigated.

---

### Week 1 — Foundations: repo scaffold & landing stub

**Build (via OpenSpec changes):**
- `change/repo-scaffolding` — `@sonnet` — Go module init, folder structure (`api/`, `mcp/`, `ui/`, `landing/`, `internal/`), `Makefile`, `.gitignore`, `.env.example`.
- `change/dev-environment` — `@haiku` — Docker Compose for Postgres, goose migrations stub, `main.go` with healthz handler.
- `change/ci-pipeline` — `@haiku` — `.github/workflows/ci.yml` for `go test`, `go vet`, `staticcheck`, build binary.
- `change/landing-page-stub` — `@sonnet` — Static HTML page at `ciscout.io` with headline + waitlist form (Resend or Plunk for email capture). Cloudflare Pages deploy.

**Content:**
- "I'm building CIScout" launch post on LinkedIn + X. One tweet thread, one LinkedIn article (medium length), both linking to ciscout.io. Emphasize the MCP differentiator.
- Pin a profile-level announcement on both platforms.

**Gate:** ciscout.io loads, waitlist captures one email (yours), Go binary runs locally with `make dev`, CI passes on PRs.

---

### Week 2 — Foundations: GitHub App + webhook receiver

**Build:**
- `change/github-app-install-flow` — `@opus` for decisions, `@sonnet` for impl — App registration metadata (permissions in ARCHITECTURE.md §6.1), OAuth flow for installer, `installations` + `repositories` tables + repo layer, install handler.
- `change/webhook-receiver` — `@sonnet` — HMAC verification, delivery deduplication, event dispatch skeleton (no diagnosis yet, just log + ack).
- `change/worker-bootstrap` — `@sonnet` — River queue integration, one no-op job type.

**Content:**
- Technical post: "Why we're building CIScout as a Go monolith in 2026" — reference architecture, emphasize opinionated tech choices. LinkedIn primary.
- Reply-guy mode: comment on 10 platform-engineering / CI-related LinkedIn posts per week with substantive technical contributions. Slow but compounds.

**Gate:** GitHub Actions on a test repo fires a webhook, CIScout receives + verifies + acks it, and enqueues a (stub) job that runs.

---

### Week 3 — Core diagnosis loop: context gathering

**Build:**
- `change/github-context-fetcher` — `@sonnet` — Installation token cache, GitHub REST calls for logs, diff, commit history. Truncation per PROMPT.md §6.
- `change/diagnosis-job` — `@sonnet` — River job that takes a repo+PR+run and assembles the `<ci_failure>` XML block.
- `change/anthropic-client` — `@sonnet` — HTTP client with tool-use support, retry/backoff, prompt caching headers. Hardcoded API key for now (BYOK is week 6).

**Content:**
- Weekly progress post on LinkedIn. Short: "Week 3 of CIScout — here's what shipped, here's what's next."
- Start a blog on `ciscout.io/blog` (or hashnode/dev.to crossposted). First post: "Structured diagnosis: why AI for CI needs to talk to other AIs."

**Gate:** Given a real failing CI run on a test repo, the system produces a full `<ci_failure>` input block and successfully calls Anthropic with tool use, receiving a structured `record_diagnosis` call back.

---

### Week 4 — Core diagnosis loop: PR comment output

**Build:**
- `change/pr-comment-rendering` — `@sonnet` — Markdown template from the structured diagnosis. Edge cases: long reasoning, nil suggested_fix, multiple suspect files.
- `change/pr-comment-poster` — `@sonnet` — Post via GitHub API, store comment_id, handle re-run (update existing vs new comment — decide: new comment per run).
- `change/feedback-reactions` — `@haiku` — Listen for 👍/👎 reactions on CIScout comments, store in `diagnosis_feedback`.
- `change/fly-staging-deploy` — `@haiku` — `fly.toml`, deploy staging environment.

**Content:**
- **Major post:** "First real CI failure diagnosed by CIScout" — screenshots of the PR comment, commentary on what worked, what surprised you. LinkedIn + X + blog.
- Reach out to 5 platform-eng leads in network with a demo GIF + "would love your feedback, not selling." Warm intros only; no cold outreach yet.

**Gate:** End-to-end on three friends' repos. Each has installed the GitHub App. Each has had CIScout post at least one diagnosis comment within 90 seconds of a CI failure. At least one thumbs-up reaction logged.

---

### Week 5 — Diagnosis quality & evaluation corpus

**Build:**
- `change/evaluation-corpus-v1` — `@sonnet` — Build 20 of the 30 target corpus cases. Fixtures + expected.json + Go test runner.
- `change/prompt-iteration-v1.1` — `@opus` — Based on first 50 real diagnoses + corpus results, revise the system prompt. MINOR version bump.
- `change/post-hoc-validators` — `@sonnet` — Implement the hallucination catchers from PROMPT.md §13.
- `change/opus-escalation` — `@sonnet` — Wire the model-escalation path.

**Content:**
- Weekly progress post.
- Blog: "What we learned from diagnosing 50 CI failures" — concrete pattern observations. This is the kind of content that ranks on Hacker News.
- LinkedIn: "Here's how I'm using Claude Opus for planning and Sonnet for building" — meta post on the AI-assisted workflow. Your model-tagging convention is itself a content moment.

**Gate:** Evaluation corpus (20 cases) passes ≥ 75%. At least one prompt revision has shipped via the tagged workflow.

---

### Week 6 — BYOK + MCP server

**Build:**
- `change/byok-key-storage` — `@opus` for encryption scheme review, `@sonnet` for impl — AES-256-GCM envelope encryption, KEK in Fly secrets, rotation scaffolding, `anthropic_keys` table + repo.
- `change/byok-onboarding-ui` — `@sonnet` — Dashboard form for pasting key, last-4 display, revoke flow.
- `change/mcp-server-core` — `@opus` for transport decision, `@sonnet` for impl — streamable HTTP endpoint, auth via `cis_live_*` bearer tokens, `get_ci_diagnosis` and `list_diagnoses` tools.
- `change/mcp-config-dashboard-section` — `@sonnet` — Settings page with copy-paste Claude Code config block.

**Content:**
- **Key post:** "CIScout now speaks MCP — your Claude Code can fix your own CI failures." Demo video/GIF. This is the positioning moment.
- Cross-post to r/ClaudeAI, r/LocalLLaMA (if community fits), Anthropic Discord.

**Gate:** A friend configures their Claude Code to point at your MCP endpoint, queries a real diagnosis, and uses the suggested_fix to ship a real fix to a real PR. Document the session.

---

### Week 7 — Dashboard polish + Slack/Discord

**Build:**
- `change/dashboard-diagnosis-history` — `@sonnet` — List view, detail view, thumbs feedback UI.
- `change/slack-discord-notifier` — `@sonnet` — Worker jobs for each; one channel per install.
- `change/dashboard-settings` — `@haiku` — Rotate key, manage repos, MCP tokens.
- `change/evaluation-corpus-v2` — `@haiku` — Complete cases 21–30.

**Content:**
- Weekly progress post.
- Blog: "The MCP server is what makes AI CI tools useful" — positioning essay, not a build post.

**Gate:** Evaluation corpus (30 cases) ≥ 80% accuracy. Slack notification tested end-to-end by at least one friend.

---

### Week 8 — Billing + public signups

**Build:**
- `change/stripe-integration` — `@sonnet` — Checkout sessions, webhooks, subscription lifecycle, tier enforcement.
- `change/public-signup-flow` — `@sonnet` — Remove waitlist gate; GitHub OAuth login → install GitHub App → setup wizard.
- `change/tier-enforcement` — `@sonnet` — Free/Solo/Team/Org checks at enroll and diagnose time.
- `change/landing-page-v2` — `@sonnet` — Real copy from LANDING.md, pricing, feature comparison, testimonials (if any).

**Content:**
- **Launch moment #1 (soft):** "CIScout is now open for signups" on LinkedIn + X + dev.to. Not a full public launch — that's week 12.
- Direct outreach to the 5 friends who tested: convert to paid. Even $29 each = $145 MRR floor.
- Ask each paying friend for permission to quote them in future content.

**Gate:** Stripe live, at least one paying customer on a Solo or Team tier.

---

### Week 9 — Private beta cohort

**Build:**
- Prompt + product iteration based on beta feedback. No major new features.
- `change/observability-metrics` — `@haiku` — Prometheus metrics per ARCHITECTURE.md §9.2.
- `change/alerting-minimal` — `@haiku` — Diagnosis failure rate alert via Grafana → Slack DM.

**Content:**
- Cold(ish) outreach to 30 platform-engineering leads on LinkedIn. Template personalized per lead, offering free private beta access in exchange for a 20-minute feedback call.
- Start a weekly "CIScout Digest" email: one diagnosis pattern observation per week, surface via the waitlist → paid list.

**Gate:** 10 organizations have tried CIScout this week (paying or beta). 3+ have given substantive feedback. Conversion target: 2 of the 10 → paid.

---

### Week 10 — Case study generation & content asset buildup

**Build:**
- `change/prompt-iteration-v1.2` — `@opus` — Second major prompt revision based on beta data. Corpus accuracy target ≥ 85%.
- `change/dashboard-share-diagnosis` — `@sonnet` — Public-by-link diagnosis share page (opt-in, stripped of customer-identifying info). Good for content marketing.

**Content:**
- **Case study blog posts (3):** "How CIScout diagnosed a multi-day flaky test at [customer X]" format. Customer quotes, before/after, workflow integration. These are your most valuable long-term assets.
- Start appearing on podcasts: email 10 platform-engineering / dev-tools podcasts pitching an episode on "AI in CI/CD — what works, what doesn't."

**Gate:** 3 case study posts drafted (at least 1 published). 3+ paying customers. Podcast outreach sent to 10 shows.

---

### Week 11 — Launch prep

**Build:**
- `change/docs-site` — `@sonnet` — Public documentation at `docs.ciscout.io` (or `/docs` on the main site): getting started, GitHub App install, BYOK setup, MCP configuration, FAQ.
- `change/landing-page-v3` — `@sonnet` — Final copy polish. Add testimonials, pricing confidence, comparison to CodeRabbit/Greptile/Nairi.
- `change/load-test` — `@haiku` — Basic load test: 100 concurrent webhooks, verify no crashes, measure p95 latency.

**Content:**
- Draft launch assets: Show HN post, r/devops post, LinkedIn launch post, X thread, Product Hunt listing.
- Pre-line up 10 people to upvote/comment on launch day (friends, beta customers, LinkedIn network).
- Draft the "launch day" blog post.

**Gate:** All launch assets drafted and reviewed. Docs site live. Load test passes at target scale.

---

### Week 12 — Public launch

**Build:**
- Bug fixes only. No new features this week.
- Launch-day monitoring watch.

**Content:**
- **Launch day (Tuesday recommended for Show HN):**
  - Show HN post at 6am PT.
  - LinkedIn + X launch posts same time.
  - r/devops, r/programming, r/SaaS posts later in the day.
  - Email waitlist with "we're live" + discount code for first 20 signups.
  - Product Hunt listing.
- Every day this week: respond to every comment, DM, email. Convert interest to customers.

**Gate — day 90 targets:**
- 5 paying customers.
- $500 MRR.
- Public launch executed across all planned channels.
- 2× LinkedIn follower count vs. week 1.
- Evaluation corpus accuracy ≥ 88%.

---

## 5. Risks and mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Acronis contract restricts moonlighting | Medium | Kills the project or forces LLC | Week 1: finish contract review, form LLC if needed |
| Diagnosis quality plateau below 75% | Medium | Slows customer acquisition | Weekly prompt iteration via corpus + thumbs-down review |
| Anthropic rate limits per customer key | Low | Affects customer experience, not ours | BYOK = customer's problem; surface clearly in UI |
| GitHub App secondary rate limits | Low | Delays diagnosis delivery | Per-installation token caching; fallback to polling unused events |
| Solo burnout | Medium | Ships slip, content dries up | Hard weekly boundary; Saturdays off; parallel-track rule prevents all-or-nothing |
| Name collision / trademark issue | Low | Rebrand cost | Week 1 trademark check; if blocked, fall back to next candidate from PRODUCT naming shortlist |
| Early customer churn | Medium | MRR stalls | Weekly check-in with every paying customer; fix their problem within 24 hours |
| Competitor (CodeRabbit, Nairi, Greptile) ships CI-diagnosis feature | High over 12 weeks | Erodes differentiation | MCP angle is the durable moat; if lost, accept and pivot positioning to "best GHA CI diagnosis experience" |

---

## 6. Decision points

Moments where the plan may pivot based on data:

- **End of week 5:** If corpus < 65% or no organic interest, revise prompt strategy and possibly narrow target customer further.
- **End of week 8:** If no paying customers by end of week 8, investigate root cause before pushing public launch. Public launch without paying customers signals the product isn't ready.
- **End of week 10:** If 3 paying customers feels like a hard ceiling (not driven by distribution), the wedge may be wrong. Revisit narrowing further or adjusting price.
- **End of week 12:** Full retrospective. Set next 90-day plan based on what's working vs. what's noise.

---

## 7. What's not in this plan

- **Fundraising.** Not a VC-scale business. Bootstrap-only until $100k ARR or a clear reason to reconsider.
- **Hiring.** Solo for the 12-week horizon. Revisit after day 90 if MRR trajectory warrants.
- **Enterprise features.** SOC 2, SSO, on-prem, DPAs — out of scope until a specific named customer requires them to close.
- **Layer 2 features.** Insights dashboard, flaky test tracker — per PRODUCT.md §2, these ship month 4–6, not in this 90-day horizon.
- **International expansion.** Default to English-speaking customers (US, UK, EU, AU). No localization yet.

---

## 8. Tracking

- Weekly status snapshot posted to `openspec/archive/<YYYY-MM-DD>-weekly-status.md` (or a rolling `STATUS.md` on `main`).
- One OpenSpec change per major deliverable, archived on merge.
- Paying customer count + MRR updated weekly in `STATUS.md`.
- Retrospective at end of each phase (phases 1, 2, 3, 4, 5, 6). Short: what worked, what didn't, what to adjust.

---

## 9. Open items that block the plan

- Domain purchase.
- Acronis contract review.
- Trademark check.
- GitHub App reservation on github.com/apps.
- Anthropic API key for staging (your own key, used until BYOK ships in week 6).
- First 3 friends' repos to test on (reach out this week).

All six are actions for you, not for Claude. None should block past week 1 start.
