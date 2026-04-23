# ci-pipeline

## Why

Every PR needs automated verification before merge, or regressions land silently and the scaffolding's acceptance criteria decay inside a week. Until CI exists, "make test passes on my machine" is the only signal, and solo development amplifies the blind spot. A minimal GitHub Actions pipeline (test / vet / staticcheck / build) is enough to protect `main` through the 12-week plan.

## What

- Add `.github/workflows/ci.yml` triggered on `pull_request` to `main` and `push` to any branch.
- Single job on `ubuntu-latest`, single Go version (1.26, matching `go.mod`).
- Steps: checkout → setup-go with module cache → `go mod download` → `go vet ./...` → `staticcheck ./...` → `go test ./... -race -count=1` → `go build ./...`.
- Pin `staticcheck` to a specific version so CI behavior is reproducible.
- README gets a CI status badge.
- Branch-protection setup documented as a human action (not automated) in the README — Claude cannot enable it from outside the GitHub UI.

Files touched:
```
/
├── .github/
│   └── workflows/
│       └── ci.yml                   (new)
└── README.md                        (badge + branch-protection note)
```

## Acceptance criteria

- [ ] CI runs on every PR to `main` and every push.
- [ ] A PR with a failing test, a `go vet` violation, or a staticcheck violation is red.
- [ ] Green CI run completes in under 3 minutes with a warm cache on the current scaffold.
- [ ] `go vet`, `staticcheck`, and `go test -race` all gate the pipeline (no warn-only).
- [ ] Module cache keyed on `go.sum` — cold restore stays under 30s.
- [ ] Staticcheck version is pinned in the workflow (not `@latest`).
- [ ] README shows a CI status badge linking to the workflow.

## Out of scope

- Release builds, artifact uploads, deploy hooks — owned by `fly-staging-deploy` (week 4).
- Matrix testing across Go versions — single version is fine for v0.1; revisit when a supported version drifts.
- Coverage reports / Codecov integration — defer until diagnosis pipeline lands and branch coverage matters.
- Lint beyond staticcheck (golangci-lint, errcheck, revive) — add when staticcheck alone misses something real.
- Frontend CI (landing page lint, dashboard type-check) — each frontend change owns its CI.
- DB-backed integration tests — add alongside the first feature that needs them.
- Self-hosted runners — GitHub-hosted is plenty for current volume.

## Dependencies

- `repo-scaffolding` merged (provides go.mod, test suite, `make` targets referenced in the workflow).

## Risks

- **Staticcheck false positives on scaffold code.** Unlikely at the current size; if it happens, fix the code, don't suppress the check.
- **Cache bust churn.** If `go.sum` changes on every branch, the cache degrades to cold restore. Acceptable at current volume; revisit if it becomes painful.
- **Branch protection forgotten.** CI running but not required = no gate. README note + `@opus` review task catch this.
