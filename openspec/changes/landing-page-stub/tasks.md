# Tasks — landing-page-stub

## Budget
- **Opus:** 3 tasks — email provider decision, form-submit architecture decision, final review
- **Sonnet:** 3 tasks — HTML page, Pages Function, client-side submit handler
- **Haiku:** 2 tasks — `_headers` / `_redirects`, README section

Total: 9 tasks (8 code + 1 human-action deploy step noted inline). Expected effort: 3–5 hours.

## Tasks

### Decisions first

- [ ] `@opus` Decide email provider: Resend vs Plunk vs ConvertKit vs Buttondown. Constraints: developer-first API, free tier covers pre-launch (<500 subscribers), good deliverability, GDPR-friendly (EU audience in scope per `docs/EXECUTION_PLAN.md`). Default bias: Resend (best transactional DX) — but Resend broadcast tooling is thin; may need Buttondown/ConvertKit later for the "CIScout Digest" email mentioned in week 9. Record choice + the digest migration plan in commit.

- [ ] `@opus` Decide form submission architecture. Options:
  - (a) Cloudflare Pages Function → provider API (secret in CF env)
  - (b) Direct client-side POST to provider (only viable if provider supports public form tokens)
  - (c) POST to the Go backend
  Default bias: (a). Backend isn't publicly reachable yet, (b) risks key leakage, (a) keeps secrets server-side with zero backend coupling. Record in commit.

### Scaffold

- [ ] `@sonnet` Create `landing/index.html`. Semantic HTML5, Tailwind CDN in `<head>`, sections: `<header>` + hero (H1, subhead, inline email input, "See how it works" anchor), problem, how-it-works (text-only three-step), waitlist form (repeat of hero CTA), footer. Copy lifted verbatim from `docs/LANDING.md` §2 (hero), §3 (problem), §4 (how it works). OpenGraph + Twitter Card `<meta>` tags with placeholder OG image URL (to be replaced when we have an image). `<html lang="en">`.

- [ ] `@sonnet` Write `landing/functions/api/waitlist.ts`. Cloudflare Pages Function accepting `POST` with JSON `{email, turnstileToken}`. Steps:
  1. Verify Turnstile server-side (`https://challenges.cloudflare.com/turnstile/v0/siteverify`) with secret from `env.TURNSTILE_SECRET`.
  2. Validate email shape (simple regex).
  3. POST to the chosen email provider's API using key from `env.EMAIL_PROVIDER_KEY`.
  4. Return JSON `{ok: true}` on success, 400 with `{error: <message>}` on validation failure, 502 on provider failure.
  No external dependencies beyond `fetch`. TypeScript, one file.

- [ ] `@sonnet` Wire the client-side submit handler in `index.html` (inline `<script>`, vanilla JS). On submit: client-side email regex, disable the button with a spinner, `fetch('/api/waitlist', {...})`, show success or inline error. Include the Turnstile widget (invisible variant) and pass its token with the request.

### Boilerplate

- [ ] `@haiku` Write `landing/_headers` with security headers: `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy: camera=(), microphone=(), geolocation=()`, `X-Frame-Options: DENY`. Scope to all paths.

- [ ] `@haiku` Write `landing/_redirects` with `https://www.ciscout.io/* https://ciscout.io/:splat 301`. (Remove if Cloudflare handles apex/www at the DNS level — the deploy task confirms.)

- [ ] `@haiku` Add a "Landing page" section to `README.md`. Max 12 lines. Content: how to preview locally (`npx wrangler pages dev landing/`), env vars needed (`TURNSTILE_SECRET`, `EMAIL_PROVIDER_KEY`), one-line pointer to the Cloudflare Pages deploy steps below.

### Deploy (human action — documented, not automated)

- [ ] `@sonnet` Document Cloudflare Pages deploy steps in `README.md` under the "Landing page" section. Steps: connect repo in CF dashboard → set build output directory to `landing/` → attach custom domain `ciscout.io` → add environment variables (`TURNSTILE_SECRET`, `EMAIL_PROVIDER_KEY`) → verify Turnstile site key is baked into the HTML. Claude cannot click these buttons; the task output is documentation, not deploy execution.

### Review

- [ ] `@opus` End-to-end check. Submit a real email to the live form, confirm it lands in the provider's list. Run Lighthouse (Performance/Accessibility/Best Practices ≥ 95). Spot-check on one iOS and one Android device. Diff the on-page copy against `docs/LANDING.md` §2–4 — must be verbatim. Document findings in commit.

## Escalations

Record here if any task was escalated mid-execution. Format:
`<task line> — escalated from @X to @Y because <reason>`
