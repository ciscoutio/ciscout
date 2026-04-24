# landing-page-stub

## Why

`ciscout.io` needs to resolve to *something* from day one. Two concrete drivers:
1. Week 1 content — LinkedIn + X "I'm building CIScout" post links to the domain. A dead URL kills the launch.
2. Waitlist capture — content-track traffic arrives before product exists; without an email list, that attention vaporizes.

The stub is deliberately minimal. Full landing polish (pricing, testimonials, comparison) is owned by `landing-page-v2` (week 8) and `landing-page-v3` (week 11) per `docs/EXECUTION_PLAN.md`. This change ships the thinnest thing that can capture a real email.

## What

Populate `landing/` with a single static HTML page deployed to Cloudflare Pages on the apex `ciscout.io`.

Page content (subset of `docs/LANDING.md` — hero + problem + waitlist only; do not paraphrase copy):
- Hero: H1 + subheadline + waitlist email input + "See how it works" anchor.
- Problem: one short section (from `docs/LANDING.md` §3).
- How it works: the three-step visual described in `docs/LANDING.md` §4 (text-only for now — no illustrations).
- Waitlist form: inline at bottom of page + duplicated in hero.
- Footer: copyright, link to GitHub org, `mailto:hello@ciscout.io`.

Tech:
- Static HTML, Tailwind via CDN (no build step for v1 stub).
- Vanilla JS for form submission (no framework).
- Cloudflare Pages Function at `landing/functions/api/waitlist.ts` forwards submissions to the chosen email provider.
- Cloudflare Turnstile for spam protection.
- No analytics, no cookies, no tracking — nothing to disclose.

Files:
```
landing/
├── index.html                       (new — hero, problem, how-it-works, waitlist, footer)
├── _headers                         (new — security headers)
├── _redirects                       (new — www → apex if needed)
└── functions/
    └── api/
        └── waitlist.ts              (new — Pages Function)
README.md                            (new "Landing page" section)
```

## Acceptance criteria

- [ ] `ciscout.io` resolves over HTTPS and serves the page.
- [ ] Uncached first paint under 500ms from a European vantage (Cloudflare CDN).
- [ ] Submitting the form with a valid email adds the address to the chosen email provider's list and displays a success state.
- [ ] Invalid email submissions show an inline error with no network call.
- [ ] Turnstile (or honeypot fallback) blocks a programmatic submission without a valid token.
- [ ] Lighthouse scores ≥ 95 on Performance, Accessibility, Best Practices.
- [ ] Mobile viewport (iPhone SE + Pixel 7 widths) renders without horizontal scroll.
- [ ] Copy on the page exactly matches `docs/LANDING.md` sections 2–4 (no paraphrasing).
- [ ] Form provider API key lives in Cloudflare Pages env vars only — not committed to the repo.

## Out of scope

- Pricing, testimonials, FAQ, comparison table — owned by `landing-page-v2` (week 8).
- Final copy polish and conversion optimization — owned by `landing-page-v3` (week 11).
- Blog, docs, or any multi-page navigation — single `index.html` only.
- Analytics (Plausible, Fathom, GA) — separate change if and when we want it.
- A/B testing framework.
- Custom fonts — system font stack is fine.
- React, build tooling, bundlers — static HTML stays static.
- Dashboard or app shell — owned by later changes.
- Localization.

## Dependencies

- Domain `ciscout.io` purchased and in Cloudflare Registrar (Phase 0 item — user action, not a code task).
- Cloudflare account with Pages access.
- `docs/LANDING.md` §2–4 is the source of truth for copy; any edits there cascade here.

## Risks

- **Email provider lock-in.** Migrating a list later is annoying. `@opus` decision task weighs provider trade-offs explicitly.
- **API key leakage.** Client-side submission with a provider API key is a footgun. Architecture decision task forces server-side-only secret handling via Pages Functions.
- **Spam signups.** Without Turnstile, unattended forms fill with junk within days. Turnstile is included from day one.
- **Copy drift from `docs/LANDING.md`.** Re-typing copy risks divergence. Tasks require verbatim copy-paste, and the review task checks for drift.
