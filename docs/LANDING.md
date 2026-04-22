# CIScout — Landing Page Copy

**Status:** Working draft, 2026-04-23
**Target audience:** Platform engineers and dev-tools leads at teams of 5–50 engineers using GitHub Actions.
**Target reading level:** Senior engineer. Direct. No hand-waving.
**CTA hierarchy:** Waitlist signup (pre-launch) → Free install (post-launch) → Paid upgrade.

---

## 1. Page structure

```
┌─────────────────────────────────────────────────┐
│ HERO — one-line value prop + CTA                │
├─────────────────────────────────────────────────┤
│ THE PROBLEM — why CI failures waste time        │
├─────────────────────────────────────────────────┤
│ HOW IT WORKS — 3-step visual                    │
├─────────────────────────────────────────────────┤
│ TWO OUTPUTS — human-facing + agent-facing       │
├─────────────────────────────────────────────────┤
│ WHY DIFFERENT — comparison table                │
├─────────────────────────────────────────────────┤
│ BYOK — trust narrative                          │
├─────────────────────────────────────────────────┤
│ PRICING — 4 tiers                               │
├─────────────────────────────────────────────────┤
│ FAQ                                             │
├─────────────────────────────────────────────────┤
│ FOOTER — socials, legal, status                 │
└─────────────────────────────────────────────────┘
```

---

## 2. Hero

**Headline (H1):**
> Your CI failed. Your agent already knows why.

**Subheadline:**
> CIScout analyzes every failed GitHub Actions run and posts the diagnosis as both a PR comment your team can read and an MCP signal your coding agent can act on. Bring your own Claude key.

**Primary CTA button:**
> Join the waitlist (pre-launch) / Install on GitHub (post-launch)

**Secondary link:**
> See how it works ↓

**Visual:**
Split screenshot:
- Left: a GitHub PR comment from CIScout with a clear diagnosis.
- Right: a Claude Code chat where the user asked "why did CI fail?" and the agent cited CIScout's MCP output and proposed a fix.

---

## 3. The problem

**Section heading (H2):**
> CI failures eat hours. Most of them shouldn't.

**Body:**
> A failing CI run on a pull request means someone stops what they're doing, opens the logs, scrolls, squints, and decides: is this a real bug, a flaky test, or did npm blip again? That work repeats on every failure, across every engineer on the team. It doesn't need to.
>
> An AI agent can read the logs, compare against the diff and recent commits, and tell you what happened — often with a concrete fix. That's not future technology. It's available today. The question is what you do with the answer.

**Three supporting bullets:**
> - **Flaky tests drain focus.** A quarter of red CI runs are the same three flakes. You already know this; you still have to check.
> - **Real bugs hide in noise.** When a genuine regression lands, it looks identical to the flakes — until someone burns 20 minutes confirming.
> - **Your agent is already here.** You have Claude Code, Cursor, or Codex open. They could fix this, if they only knew what broke.

---

## 4. How it works

**Section heading (H2):**
> Three steps. Two minutes to install.

**Step 1 — Install the GitHub App**
> Click install, pick your repos, paste your Anthropic API key. CIScout is live.

**Step 2 — Ship code like normal**
> Open a PR. CI runs. CI fails. CIScout reads the logs, compares them to your diff and recent commits, and calls Claude to produce a structured diagnosis.

**Step 3 — Get the diagnosis where you need it**
> A concise PR comment lands for your team. A structured MCP signal is available for your coding agent to use.

**Visual:** animated diagram or simple numbered GIF showing the webhook → Claude → two outputs flow.

---

## 5. Two outputs

**Section heading (H2):**
> For humans. For agents. Same diagnosis.

**Two-column section:**

**Column left — Human-facing:**
> **PR comment on GitHub**
>
> A short, specific diagnosis posted directly to the failing pull request. Classification (real bug, flaky test, infra, dependency, config), suspected files, confidence score, and — when confident — a suggested fix. Your team reads it and moves on.

*[Screenshot of a realistic PR comment.]*

**Column right — Agent-facing:**
> **MCP signal for coding agents**
>
> The same diagnosis, served as a structured tool call via Model Context Protocol. Claude Code, Cursor, Codex, and any MCP-compatible client can query `get_ci_diagnosis` and use the response to generate a fix — without you copy-pasting logs.

*[Snippet of Claude Code responding to "why did CI fail?" with a cited diagnosis + fix.]*

**Closing line:**
> Most AI CI tools end at the PR comment. CIScout starts there.

---

## 6. Why CIScout is different

**Section heading (H2):**
> Built for the agent-first workflow.

**Comparison table:**

| | CIScout | Typical AI PR-review tool | Generic agent platform |
|---|---|---|---|
| Focus | CI failure diagnosis, specifically | Code review subjectivity | Build any agent for any workflow |
| Where you see the output | PR comment **+ your coding agent** | PR comment only | Slack / Discord |
| Works with Claude Code / Cursor / Codex via MCP | **Yes, native** | No | Limited |
| Bring your own Anthropic key | **Yes, required** | No (token markup) | Mixed |
| Token cost on your invoice | Your Anthropic bill (see exact tokens) | Bundled, opaque | Bundled, opaque |
| Ships in 12 weeks | v0.1 | — | — |
| Quota-based pricing | No — per-repo flat | Variable | Task-based |

**Positioning line below the table:**
> CIScout does one job exceptionally. Everything we ship is in service of making CI failure diagnosis useful to both humans and agents. No platform to configure. No skills marketplace to browse. One install, one key, immediate value.

---

## 7. BYOK — the trust narrative

**Section heading (H2):**
> Your Claude. Your data. We stay out of it.

**Body:**
> CIScout never stores your code. When CI fails, we fetch your logs and diff from GitHub, send them through Anthropic's API using **your** Claude API key, and receive the diagnosis. Your key is encrypted at rest with AES-256-GCM and decrypted only for the duration of a single API call.
>
> This means:
>
> - Every token billed is visible in your own Anthropic console. No markup, no surprise.
> - Your company's existing Anthropic DPA and compliance posture covers the LLM layer.
> - If we disappeared tomorrow, your code never went anywhere you don't already trust.

**Trust signals row:**
> 🔒 AES-256-GCM envelope encryption · 🔑 Keys decrypted only for single API call · 👁️ Every token visible in your Anthropic console · 🛑 Source code never written to disk

---

## 8. Pricing

**Section heading (H2):**
> Per repository, not per analysis. Predictable.

**Four-tier grid:**

**Free**
> **$0** / month
> - 1 public repo
> - Unlimited diagnoses
> - PR comment output
> - Community support
>
> *[Get started]*

**Solo**
> **$29** / month
> - 1 private repo
> - Unlimited diagnoses
> - PR comment output
> - Email support
>
> *[Start 14-day trial]*

**Team** (recommended)
> **$149** / month
> - Up to 20 private repos
> - Unlimited diagnoses
> - PR comment + Slack/Discord output
> - **MCP server access**
> - Priority email support
>
> *[Start 14-day trial]*

**Org**
> **$499** / month
> - Unlimited repos
> - Unlimited diagnoses
> - All outputs
> - MCP server access
> - Audit log
> - Dedicated founder Slack
>
> *[Contact us]*

**Note below pricing:**
> All tiers: bring your own Anthropic API key. Your tokens, your bill. We charge only for the CIScout platform.

---

## 9. FAQ

**Q: Does CIScout work with GitHub Enterprise Server?**
> Not yet in v0.1. We support github.com only. If you're on GHES and would pay for support, email us.

**Q: Does CIScout work with GitLab / Bitbucket / Jenkins / CircleCI?**
> Not in v0.1. We're GitHub Actions only. This is deliberate — we'd rather do one thing excellently than four things mediocrely. If your team would pay for other CI systems, tell us which ones.

**Q: What models does CIScout use?**
> Claude Sonnet 4.6 by default. We escalate to Claude Opus 4.7 on low-confidence cases. You bring your own Anthropic API key — every diagnosis is billed directly against your Anthropic account.

**Q: How long does a diagnosis take?**
> Target is ≤ 90 seconds from CI failure to PR comment. Most take 10–20 seconds.

**Q: What if the diagnosis is wrong?**
> Give it a thumbs-down on the PR comment. We review every thumbs-down, iterate on the prompt, and ship improvements weekly. Our goal is ≥ 90% accuracy within 60 days of your install.

**Q: Can CIScout fix the bug automatically?**
> No, and that's intentional. We produce the diagnosis + a suggested fix; your team or your coding agent (Claude Code, Cursor) decides whether to apply it. We are not a merge-queue agent and we won't push commits to your repo.

**Q: Is my code sent to CIScout's servers?**
> Logs and diff pass through our server transiently to construct the Anthropic API call. We don't persist source code to disk. The diagnosis output (structured JSON) is stored so your agent can query it later via MCP.

**Q: Can I self-host CIScout?**
> Not in v0.1. Self-hosting is on the roadmap for enterprise customers. If this is a deal-breaker, email us and we'll expedite.

**Q: What about compliance — SOC 2, ISO 27001?**
> We're pre-SOC-2 in v0.1. We're transparent about this. Our BYOK architecture means the LLM layer is governed by your existing Anthropic compliance relationship, not ours. If your company requires SOC 2 from CIScout itself, reach out and let's discuss timing.

**Q: Who built CIScout?**
> Solo founder, ex–platform engineer. Building in public — follow along on LinkedIn and X.

---

## 10. Footer

**Three columns:**

**Product**
- Install
- Pricing
- Documentation
- Changelog

**Company**
- About
- Blog
- Contact

**Legal / Status**
- Privacy
- Terms
- Security
- Status page

**Bottom row:**
> © 2026 CIScout. Built in public. · [LinkedIn] · [X] · [GitHub]

---

## 11. SEO / metadata

- **Title tag:** CIScout — AI diagnosis for GitHub Actions failures. For humans and agents.
- **Meta description:** Structured CI failure diagnosis via GitHub App + MCP. Works with Claude Code, Cursor, Codex. Bring your own Anthropic key. Install in two minutes.
- **Canonical:** https://ciscout.io/
- **Open Graph image:** split-screen hero screenshot.
- **Schema.org:** SoftwareApplication with pricing info for Google rich results.

---

## 12. Design notes

- **Typography:** monospace accent for code snippets (e.g. Berkeley Mono or IBM Plex Mono); sans-serif for body (Inter).
- **Color palette:** two colors maximum plus neutrals. Suggested accent: a CI-inspired green or a single brand color TBD. No gradients. No glass morphism.
- **No stock photos.** Use real product screenshots and functional diagrams only.
- **No hero animation** unless it is genuinely explanatory (e.g., the "how it works" numbered GIF).
- **Dark mode default.** Platform engineers use dark editors. Light mode available via toggle.
- **Page weight budget:** < 500 KB total on first paint, no third-party JS beyond Plausible analytics.

---

## 13. Copy rules (style guide)

- Write for a senior engineer who has seen dozens of AI tool landing pages. Assume skepticism.
- No "revolutionize." No "next-gen." No "transform your workflow."
- Use the second person (you) when addressing the reader.
- Don't claim things that aren't true yet. If something ships in v0.2, it is not on the landing page.
- Every claim is verifiable. If we say "90-second diagnosis," we hit it. If we say "BYOK," we mean it.
- One idea per sentence. Short paragraphs. Scannable.

---

## 14. Post-launch iteration

The landing page is a living surface. Revisit copy when:

- Customers repeat the same misunderstanding in sales conversations (add to FAQ).
- A competitor ships a feature worth differentiating from.
- A metric-driven A/B test has run on a specific section and finished.
- Pricing changes.

Do not tinker weekly. Edit with intent.

---

## 15. Assets needed

Before v2 of the landing page ships (week 8):

- 3 product screenshots (install flow, PR comment, dashboard).
- 1 "how it works" animated diagram or looping GIF.
- 1 Claude Code MCP interaction demo (screenshot or GIF).
- 1 founder headshot (for About page and social OG images).
- 3 customer testimonials (collected during private beta, weeks 9–10).

---

## 16. Open decisions

1. **Accent color and typography final pick.** Defer to design sprint in week 8.
2. **Blog integration.** `ciscout.io/blog` vs dev.to / Hashnode as primary. Default to `/blog` on the main site; cross-post elsewhere.
3. **Annual pricing discount.** Offer `-20%` on annual for all tiers? Decide at week 8.
4. **Referral program.** Maybe in Phase 6. Not v0.1.
