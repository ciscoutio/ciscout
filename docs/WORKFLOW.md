# CIScout — Working Agreement

**Status:** Working draft, 2026-04-23
**Applies to:** All non-trivial work on the CIScout codebase.

---

## 1. Purpose

This document codifies how work happens on CIScout — specifically:
- When to use OpenSpec versus direct commit.
- How to decompose work into tasks.
- How to assign AI model to each task.
- How changes move from idea to shipped code.

It exists because solo AI-assisted development is high-leverage but easy to lose discipline on. A short written agreement is cheaper than rework.

---

## 2. Document layers

Three layers of artifact, each with a distinct purpose:

```
docs/                Long-lived specs that describe the product
  PRODUCT.md         What CIScout is, roadmap, out-of-scope
  ARCHITECTURE.md    How the system is built
  MCP_SPEC.md        External contract for MCP clients
  PROMPT.md          The diagnosis prompt IP
  WORKFLOW.md        This file
  EXECUTION_PLAN.md  12-week strategic calendar
  LANDING.md         Landing page source text

openspec/            Discrete units of change
  changes/           One folder per in-flight or archived change
    <slug>/
      proposal.md    Why, what, acceptance criteria
      tasks.md       Decomposed tasks, each tagged with a model
      design.md      (optional) Non-obvious implementation notes
  specs/             Machine-readable capability specs (lazy — only when needed)
  archive/           Completed changes, moved here after merge

api/, ui/, mcp/      The actual code
```

`docs/` changes slowly and describes the destination.
`openspec/changes/` changes per-feature and describes the journey.
Code is the destination made real.

---

## 3. When OpenSpec, when direct commit

### Use an OpenSpec change when:

- Estimated effort ≥ 1 day of focused work.
- The change touches a new subsystem or integrates with an external service.
- There's ambiguity worth resolving before writing code (data model, API shape, trade-offs).
- The change would benefit from a written plan that outlives the branch.
- You want me (Claude) to help decompose and review before implementation.

### Direct commit, no OpenSpec, when:

- Typo fix, doc tweak, dependency bump, formatting.
- A bug fix that is genuinely one-line.
- Experimental spike you will throw away or replace.
- Tests for existing code with no new design decisions.

When in doubt, prefer OpenSpec. The overhead of a lightweight change folder (proposal + tasks) is ~15 minutes and prevents scope drift.

---

## 4. Change lifecycle

```
 propose  →  review  →  apply  →  archive
   ↓         ↓         ↓          ↓
 draft     revise    branch     merge to main
 proposal  proposal  execute    move folder to
 + tasks   if needed tasks      openspec/archive/
```

### 4.1 Propose

Create `openspec/changes/<slug>/proposal.md`. Use kebab-case slugs rooted in the feature name (e.g. `github-app-webhook`, `byok-encryption`, `mcp-server-core`).

Minimum content:

```markdown
# <Change title>

## Why
<1–3 sentences. What problem does this solve? What's the trigger?>

## What
<Concrete description of what changes. Reference files/components that will
be touched. Reference sections in docs/ that this change implements.>

## Acceptance criteria
- [ ] Specific, testable outcome 1
- [ ] Specific, testable outcome 2
- [ ] ...

## Out of scope
<What this change deliberately does NOT do. Forces boundary discipline.>

## Dependencies
<Other changes that must land first, or external blockers.>
```

Create `openspec/changes/<slug>/tasks.md` with the decomposition (see §5).

### 4.2 Review

Proposal is the contract. You (the human) read it and raise objections. I (Claude) can be asked to critique by running `openspec-explore` or directly reading the proposal and pushing back.

Signoff on the proposal = scope is frozen for this change. Additions require a proposal edit with explicit note.

### 4.3 Apply

Create a git branch named `change/<slug>`. Execute tasks.md top to bottom, checking each off as completed. Each task produces one or more commits on the branch.

### 4.4 Archive

On merge to main, move the change folder from `openspec/changes/<slug>/` to `openspec/archive/<YYYY-MM-DD>-<slug>/`. Keep the proposal, tasks, and any design notes. Future readers can trace why something was built.

---

## 5. Task decomposition

### 5.1 Task format

In `tasks.md`:

```markdown
# Tasks — <change slug>

## Budget
Opus: ~3 tasks    (planning, reviews, architecturally tricky decisions)
Sonnet: ~8 tasks  (most implementation)
Haiku: ~5 tasks   (boilerplate, generated code, mechanical tests)

## Tasks

- [ ] `@opus` Decide table column for <ambiguous concern>
- [ ] `@sonnet` Implement `<Component>` with <specific scope>
- [ ] `@haiku` Generate goose migration for <tables>
- [ ] `@sonnet` Wire handler to service layer
- [ ] `@haiku` Table-driven tests for <component>
- [ ] `@opus` Final review before PR
```

### 5.2 Task writing rules

- **Each task is independently executable.** A task should be shippable in one focused work session without waiting on a sibling task (other than dependencies called out in order).
- **State the output, not the process.** "Implement the HMAC verifier" — not "think about how to verify HMAC and then write it."
- **Reference spec sections.** "Per ARCHITECTURE.md §6.1" is worth a paragraph of re-explanation.
- **Order matters.** Tasks run top-to-bottom. Pull dependencies to the top.
- **No open-ended research tasks.** If research is needed, make it a specific question with a deadline ("Confirm MCP streamable-HTTP transport vs SSE by reading spec + one OSS example; 1 hour cap").

### 5.3 Model tagging rules

Inline `@opus`, `@sonnet`, or `@haiku` tag at the start of each task. The tag is the recommended model; the human or Claude running the task can escalate but not silently downgrade without updating the task.

Default mapping (see §6 for the full rubric):

| Work character | Tag |
|---|---|
| Planning, architecture decisions, reviews, ambiguity resolution | `@opus` |
| Novel logic, integration code, prompt engineering, debugging | `@sonnet` |
| Boilerplate, scaffolds, mechanical refactors, straightforward tests | `@haiku` |

If a task would need two models (e.g., "design then implement"), split it into two tasks.

---

## 6. Model-to-task rubric

The explicit rubric. When in doubt, consult this.

| Task type | Model | Rationale |
|---|---|---|
| Write or refine an OpenSpec proposal | `@opus` | Planning quality compounds |
| Decompose a proposal into tasks | `@opus` | Same |
| Architecture trade-off decisions | `@opus` | Same |
| Code review before merge | `@opus` | Fresh deep reasoning catches most issues |
| Prompt engineering (changes to PROMPT.md) | `@opus` | Core IP; worth the spend |
| Complex implementation with novel logic (diagnosis pipeline orchestration, BYOK envelope encryption, MCP server core) | `@sonnet` | Best balance; handles nuance without overspending |
| Standard feature implementation (HTTP handler, service, repo) | `@sonnet` | Default |
| Debugging a non-trivial failure | `@sonnet` | Sonnet is competent; Opus only if Sonnet stalls |
| Writing or extending the evaluation corpus | `@sonnet` | Requires judgment about test cases |
| SQL migration from an already-designed schema | `@haiku` | Mechanical translation |
| Scaffolding code (main.go, config loader, logger init) | `@haiku` | Pattern-match work |
| Refactors that rename or reorganize without logic change | `@haiku` | Same |
| Table-driven tests for existing code | `@haiku` or `@sonnet` | Haiku if assertions are obvious; Sonnet if edge cases require thought |
| Documentation updates for already-decided content | `@haiku` | Transcription |

**Escalation rule:** if a `@haiku` task proves non-trivial, escalate to `@sonnet` and update the task tag. If a `@sonnet` task needs multiple revision rounds, escalate to `@opus`. Log the escalation in a task checkbox note.

**Downgrade rule:** Do not downgrade silently. If `@opus` tagged work turns out to be trivial, complete at `@opus` (cost is small) and adjust the rubric if a pattern emerges.

---

## 7. Git conventions

- **Branch names:** `change/<slug>` for OpenSpec changes; `fix/<short>` or `chore/<short>` for direct commits.
- **Commit messages:** Imperative mood; first line ≤ 72 chars; body explains why.
- **Include** `Co-Authored-By: Claude <model>` on any commit where Claude drafted substantive content.
- **One change folder, one PR.** Don't bundle unrelated changes.
- **Squash-merge to main by default.** Keep `main` history linear.

---

## 8. How Claude is used

Claude (this assistant, any model version) is a collaborator, not a replacement for decision-making. Applied conventions:

- **Claude can propose changes** by writing `openspec/changes/<slug>/proposal.md` drafts on request. The human decides what ships.
- **Claude executes tasks** per their tag. The human or Claude-running-the-task can escalate models mid-task.
- **Claude reviews** on request (`@opus` for critical reviews, `@sonnet` for routine).
- **Claude pushes back.** If a proposal violates PRODUCT.md scope, has an architecture smell, or risks scope drift, Claude says so. This is non-optional.

---

## 9. Not in scope for this workflow

- Formal RFC process with stakeholder circulation. (Solo project.)
- Design reviews requiring more than one reviewer. (Solo project.)
- Linting of `proposal.md` / `tasks.md` format. If the sections exist, it's fine.
- `openspec/specs/` population — only fill when a machine-readable capability spec has a concrete consumer. Do not pre-populate.

---

## 10. Change to this document

Changes to `WORKFLOW.md` itself are an OpenSpec change (slug: `workflow-<what>`). This document is load-bearing; edits deserve a proposal.
