# Tasks — ci-pipeline

## Budget
- **Opus:** 2 tasks — staticcheck gating decision, end-to-end verification
- **Sonnet:** 1 task — workflow authoring (tool pinning is the tricky part)
- **Haiku:** 2 tasks — README badge, branch-protection note

Total: 5 tasks. Expected effort: 1–2 hours.

## Tasks

### Decisions first

- [ ] `@opus` Decide staticcheck gating: hard-fail vs warn-only at v0.1. Trade-off: hard-fail = clean history but an unrelated stylistic nit can block a PR; warn-only = easier merges but rules decay. Default bias: hard-fail — scaffold is small, cleaning up now is cheaper than later. Pin the staticcheck version while deciding (so the rule set is reproducible).

### Scaffold

- [ ] `@sonnet` Write `.github/workflows/ci.yml`. Triggers: `pull_request` on `main`, `push` on `**`. Single job `test` on `ubuntu-latest`. Steps:
  1. `actions/checkout@v4`
  2. `actions/setup-go@v5` with `go-version: '1.26'` and `cache: true` (built-in module cache via `go.sum`)
  3. `go mod download`
  4. `go vet ./...`
  5. Install staticcheck at the pinned version via `go install honnef.co/go/tools/cmd/staticcheck@v0.7.0`, then `staticcheck ./...`
  6. `go test ./... -race -count=1`
  7. `go build ./...`
  Pin the staticcheck version (no `@latest`). Use `GOTOOLCHAIN=local` to prevent silent toolchain upgrades.

### Review

- [ ] `@opus` Verify on a throwaway PR. Push three PRs in sequence (or force-push through states): (1) a failing test → red, (2) a `go vet` violation → red, (3) a staticcheck violation → red. Then a clean commit → green. Confirm warm-cache run is under 3 minutes. Document timings + any surprises in commit.

- [ ] `@haiku` Add a CI status badge to the top of `README.md` (one line, directly under the title). Format: `![CI](https://github.com/ciscoutio/ciscout/actions/workflows/ci.yml/badge.svg)`.

- [ ] `@haiku` Append a "Branch protection" subsection to `README.md` (after "Pre-commit hooks"): human-only steps to enable required status checks on `main`, require branches up-to-date, require linear history. Max 6 lines.

## Escalations

Record here if any task was escalated mid-execution. Format:
`<task line> — escalated from @X to @Y because <reason>`
