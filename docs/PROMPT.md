# CIScout — Diagnosis Prompt Specification (v0.1)

**Status:** Working draft, 2026-04-23
**Owner file:** The prompt defined here is the compiled-in source of truth. Changes require a PR, a version bump, and a regression pass against the evaluation corpus.

---

## 1. Purpose

This is the core IP of CIScout. The quality of every diagnosis — what customers pay for — is determined by the prompt in this document. Everything else (GitHub App, MCP server, dashboard) is plumbing.

The prompt must:
1. Produce **schema-compliant JSON** on every call, with zero post-processing.
2. **Classify** the failure into one of six types with calibrated confidence.
3. **Ground every claim** in the provided context; never hallucinate files, commits, or symbols.
4. **Refuse to guess** a fix when confidence is low; a null `suggested_fix` is preferable to a wrong one.
5. **Cache efficiently** so BYOK customer bills stay predictable.

---

## 2. Architecture: tool-use enforced output + prompt caching

### 2.1 Structured output via tool use

Do not ask the model to "output JSON." Instead, define a single tool `record_diagnosis` whose `input_schema` is the `DiagnosisResult` shape from `MCP_SPEC.md §3.1`, and force the model to invoke exactly this tool. Claude's tool-use layer enforces schema compliance at the SDK level; invalid JSON is impossible.

Benefits:
- Zero parse errors in production.
- Schema evolves by editing the tool definition; no prompt rewording.
- `strict: true` rejects any field outside the schema.

### 2.2 Prompt caching

Anthropic's prompt caching has a 5-minute TTL with ~90% cost reduction on cached tokens. We structure every request to maximize cache hits:

```
┌─────────────────────────────────────┐
│ SYSTEM PROMPT (static)     [cached] │ ~2.5K tokens
├─────────────────────────────────────┤
│ TOOL DEFINITION (static)   [cached] │ ~1K tokens
├─────────────────────────────────────┤
│ FEW-SHOT EXAMPLES (static) [cached] │ ~5K tokens
├─────────────────────────────────────┤  ← cache_control boundary
│ DYNAMIC CONTEXT (per request)       │ 5–25K tokens
│   - repo & PR metadata              │
│   - PR diff                         │
│   - failed job logs                 │
│   - recent commits                  │
└─────────────────────────────────────┘
```

On a customer with 20 diagnoses/hour, expected cache hit rate ≥ 85%, reducing input-token cost by ~75%.

---

## 3. The system prompt

The text below is the canonical v1 system prompt. It ships as a compiled Go string constant (`internal/prompt/system.go`); edits require a PR and the evaluation pipeline pass.

```
You are CIScout's diagnosis engine. A GitHub Actions CI run has failed on a
pull request. Your job is to produce a structured diagnosis that:
  (a) a human engineer can act on immediately from a PR comment, and
  (b) a downstream coding agent (Claude Code, Cursor, Codex) can use to
      generate a fix without further investigation.

You will receive, in the user turn, a single <ci_failure> block containing:
  - <repo>: repository owner/name, default branch, primary language
  - <pull_request>: PR number, title, author, head/base SHA
  - <diff>: the unified diff of the PR against its base
  - <branch_commits>: the last 5 commits on the PR branch
  - <main_commits>: the last 10 commits on the default branch
  - <job_logs>: truncated logs from the failed jobs

You must invoke the `record_diagnosis` tool exactly once with a fully
populated argument object. Do not produce any free-text response outside
the tool call.

============================================================
CLASSIFICATION — `failure_type` must be exactly one of:
============================================================

- `flaky_test`: A test failed due to non-determinism (timing, concurrency,
  ordering, external state). The PR diff does not touch the failing code
  path, OR identical code has failed and passed across unrelated runs.
  Example signals: "expected X, got Y" where Y is sometimes X; timeouts
  at a threshold; races reported by the runtime; test passes on retry.

- `real_bug`: The PR diff introduces or exposes a defect that causes the
  failure. You can trace the failure to specific lines in the provided
  diff. This is the highest-confidence category — if you pick it, you
  must cite a file:line in `suspect_files`.

- `infra`: CI runner environment failure unrelated to the code. DNS,
  network timeouts to external services, out-of-disk, runner crash,
  rate limits from third-party APIs hit during CI (not during tests),
  GitHub Actions service outage. Fix is not a code change.

- `dependency`: Package manager or lockfile issue. Missing go.sum entry,
  npm lockfile desync, incompatible version pin, pinned-version no longer
  resolvable, registry access. Fix is a lock regeneration or version
  adjustment, not an application code change.

- `config`: CI workflow YAML error, missing secret, missing environment
  variable, misconfigured action version, wrong permissions on GITHUB_TOKEN.
  Not code, not infrastructure.

- `unknown`: The provided context is insufficient to classify with
  confidence > 0.3. Prefer this over guessing. Explain the gap in
  `reasoning`.

============================================================
CONFIDENCE CALIBRATION — `confidence` is a float in [0.0, 1.0]:
============================================================

- 0.90+  : You cite a specific file:line from the diff or logs that
           directly causes the failure. Your reasoning chain is verifiable
           line-by-line against the provided context.
- 0.70 - 0.89: You identify the failure mechanism but cannot pin the
           exact line, OR the cause is strongly suggested but one
           alternative remains plausible.
- 0.50 - 0.69: Hypothesis consistent with evidence; multiple alternatives
           remain possible; you are guessing at the most likely.
- 0.30 - 0.49: Limited signal; diagnosis is speculative; an engineer
           should treat your output as a starting point only.
- Below 0.30: Set `failure_type` to `unknown`.

Do not inflate confidence. A 0.6-confidence `real_bug` is more useful
than a 0.9-confidence wrong guess.

============================================================
GROUNDING RULES — every claim must be verifiable
============================================================

1. Never reference a file, function, class, symbol, or commit SHA that
   does not appear in the provided <diff>, <job_logs>, <branch_commits>,
   or <main_commits>.

2. `suspect_files[].path` must be a path that appears in either the
   diff or the logs. `line_range` must be within the actual diff hunks
   or cited log lines.

3. `suspect_commits[]` must contain only SHAs from <branch_commits> or
   <main_commits>. Never invent or truncate SHAs.

4. `raw_logs_excerpt` must be a verbatim copy of lines from <job_logs>.
   Do not paraphrase. Preserve ANSI-stripped whitespace.

5. In `reasoning`, when you refer to code, quote it verbatim from the
   diff or link to `<file>:<line>`. Do not describe code you have not
   seen.

============================================================
SUGGESTED FIX — `suggested_fix` is `null` OR a concrete patch
============================================================

Produce a non-null `suggested_fix` only when ALL of the following hold:
  (i)   `failure_type` is `real_bug`, `flaky_test` (with a clear sync
        bug), `dependency` (with a determinable patch), or `config`
        (with a determinable YAML change).
  (ii)  You can produce a unified diff touching a single file.
  (iii) Every line in the diff references content visible in the <diff>
        or <job_logs>.
  (iv)  `confidence` >= 0.7.

Rules per type:
  - `infra` : `suggested_fix` MUST be `null`. No code change fixes infra.
  - `dependency` where the fix is "regenerate the lockfile": prefer a
    `null` fix and put the exact command in `suggested_fix.rationale`
    instead, because lockfile checksums cannot be computed by you.
  - `config`: only provide a fix if the correction is a specific YAML
    edit you can write out.

When in doubt, set `suggested_fix: null` and explain the investigation
the engineer should do in `reasoning`.

============================================================
OUTPUT SIZE CONSTRAINTS
============================================================

  - `summary`          : ≤ 200 chars, single line, no markdown.
  - `reasoning`        : ≤ 2000 chars, markdown permitted.
  - `suggested_fix.diff`: ≤ 200 lines changed, single file, unified format.
  - `suggested_fix.rationale`: ≤ 500 chars.
  - `raw_logs_excerpt` : ≤ 8000 chars; excerpt, not full logs.

============================================================
SAFETY
============================================================

  - Never produce a fix that weakens security (removing auth checks,
    widening permissions, disabling TLS verification, disabling tests).
  - Never produce a fix that silences the test (e.g., `skip` or
    `ignore_errors`) unless `failure_type` is `flaky_test` AND the
    rationale explicitly states this is a temporary quarantine pending
    a root-cause fix.
  - If the logs suggest the failure is revealing a secret leak or an
    IDOR-style vulnerability, report it in `reasoning` with
    `failure_type: real_bug` but set `suggested_fix: null` and
    recommend the engineer review with a security lens.

============================================================
STYLE
============================================================

  - Write for a senior platform engineer. No hedging ("might possibly
    perhaps"). State what you know, state what you're uncertain about,
    and stop.
  - Do not mention CIScout in the output. The output is about their CI,
    not about us.
  - Do not include apology language. "Could not determine X" is better
    than "I'm sorry, I couldn't determine X."
```

---

## 4. Tool definition

Passed to Anthropic as a single tool, marked strict. Shape mirrors `DiagnosisResult` (MCP_SPEC §3.1), minus server-generated fields (`diagnosis_id`, `created_at`, `completed_at`, `anthropic_usage`, `metadata`).

```json
{
  "name": "record_diagnosis",
  "description": "Record the structured diagnosis for this CI failure. Call exactly once.",
  "input_schema": {
    "type": "object",
    "strict": true,
    "additionalProperties": false,
    "required": [
      "failure_type",
      "confidence",
      "summary",
      "reasoning",
      "suspect_files",
      "suspect_commits",
      "suggested_fix",
      "raw_logs_excerpt"
    ],
    "properties": {
      "failure_type": {
        "type": "string",
        "enum": ["flaky_test", "real_bug", "infra", "dependency", "config", "unknown"]
      },
      "confidence": {
        "type": "number",
        "minimum": 0.0,
        "maximum": 1.0
      },
      "summary": {
        "type": "string",
        "maxLength": 200
      },
      "reasoning": {
        "type": "string",
        "maxLength": 2000
      },
      "suspect_files": {
        "type": "array",
        "items": {
          "type": "object",
          "additionalProperties": false,
          "required": ["path", "line_range", "reason"],
          "properties": {
            "path": { "type": "string" },
            "line_range": {
              "type": "array",
              "items": { "type": "integer", "minimum": 1 },
              "minItems": 2,
              "maxItems": 2
            },
            "reason": { "type": "string", "maxLength": 300 }
          }
        }
      },
      "suspect_commits": {
        "type": "array",
        "items": { "type": "string", "pattern": "^[0-9a-f]{7,40}$" }
      },
      "suggested_fix": {
        "anyOf": [
          { "type": "null" },
          {
            "type": "object",
            "additionalProperties": false,
            "required": ["file", "diff", "rationale"],
            "properties": {
              "file": { "type": "string" },
              "diff": { "type": "string" },
              "rationale": { "type": "string", "maxLength": 500 }
            }
          }
        ]
      },
      "raw_logs_excerpt": {
        "type": "string",
        "maxLength": 8000
      }
    }
  }
}
```

---

## 5. Context injection — the user message

The runtime assembles the user message as a single XML-wrapped block. XML tags outperform markdown for Claude when content spans multiple heterogeneous sections.

```xml
<ci_failure>
<repo>
  <name>acme/api-server</name>
  <default_branch>main</default_branch>
  <primary_language>Go</primary_language>
</repo>

<pull_request>
  <number>847</number>
  <title>Apply plan-based discount in billing flow</title>
  <author>jdoe</author>
  <head_sha>e3f1a9bc12...</head_sha>
  <base_sha>a4c7b2d8f9...</base_sha>
</pull_request>

<diff>
[unified diff, truncated to 200KB]
</diff>

<branch_commits>
[last 5 commits on PR branch: SHA, author, date, subject, body first 200 chars]
</branch_commits>

<main_commits>
[last 10 commits on default branch: same format]
</main_commits>

<job_logs>
[concatenated logs from failed jobs, truncated per §6]
</job_logs>
</ci_failure>
```

---

## 6. Truncation rules

Context window budgets. Exceeding these reduces diagnosis quality; truncate predictably.

| Section | Hard cap | Strategy |
|---|---|---|
| `<diff>` | 200 KB | Keep whole hunks; truncate by dropping lowest-priority files first (lockfiles, generated code, then test fixtures, then non-source). Append `<!-- diff truncated: N files / M bytes omitted -->`. |
| `<job_logs>` | 50 KB total across all failed jobs | Per failed job: keep last 20 KB. If 3+ jobs failed, allocate 16 KB each. Always preserve the final 200 lines (where the panic/assertion lives). Strip ANSI codes before counting. |
| `<branch_commits>` | 5 commits × 500 chars | Hard limit on body length. |
| `<main_commits>` | 10 commits × 500 chars | Same. |
| Total input | ≤ 60 KB dynamic + cached | Leaves 140 KB+ headroom on Sonnet 4.6 (200 KB context). |

When truncation occurs, inject a `<truncation_note>` block above `<job_logs>` describing what was dropped.

---

## 7. Few-shot examples

Include 2 concise few-shot examples in the cached prompt block (not the full 5 from MCP_SPEC). Goal: demonstrate tool use and the "prefer null over wrong" principle.

**Example A** — `real_bug` with confident fix. Full `suggested_fix`.
**Example B** — `infra` with null fix. Demonstrates refusal to fabricate.

Actual content is stored as `internal/prompt/fewshot.go` constants. Each is ~600 tokens. Cached with the system prompt.

Do not include 5 examples; the marginal benefit past 2 is negative — dilutes cache efficiency and can nudge the model toward forcing patterns that don't match the actual case.

---

## 8. Model selection & escalation

### 8.1 Default model

`claude-sonnet-4-6`. Chosen because:
- Sufficient reasoning for the task (verified in evaluation corpus; see §11).
- Cost/latency appropriate for high-volume per-customer usage.
- Full tool-use and prompt-caching support.

### 8.2 Escalation to Opus

Retry with `claude-opus-4-7` under any of these conditions, at most once per diagnosis:

1. Sonnet returned `confidence < 0.5` AND `failure_type` ∈ {`real_bug`, `flaky_test`, `unknown`}.
2. Sonnet returned `failure_type = unknown` with `confidence >= 0.3`.
3. Parse/validation failed on first attempt (before treating as hard failure).

Escalation is logged in `diagnoses.metadata.escalated_from = "sonnet-4-6"` and both token counts are summed in `anthropic_usage`.

### 8.3 No further escalation

After Opus retry, accept whatever comes back. Do not chain models indefinitely.

### 8.4 Temperature

`temperature = 0.0` for both models. Determinism matters more than creativity for diagnosis. Identical input should produce identical output (modulo model non-determinism at T=0, which is minor).

---

## 9. Cost model (per diagnosis, BYOK)

The customer pays Anthropic directly. These numbers are for capacity planning and landing-page transparency.

Approximate per-diagnosis cost on Sonnet 4.6 (current pricing):

| | Cold (no cache) | Warm (cache hit) |
|---|---|---|
| Input tokens | ~35,000 | ~25,000 fresh + ~8,500 cached |
| Output tokens | ~1,200 | ~1,200 |
| Cost per call | ~$0.12 | ~$0.035 |

On a team running 100 diagnoses/month, expected monthly Anthropic bill: **$3.50–$12**. This is a key selling point — far below comparable PR-review tools that mark up tokens.

Opus escalations (expected < 10% of diagnoses) run ~5× cost per escalated call.

---

## 10. Safety evaluation

Before any prompt change ships, it must pass this checklist:

1. **No security-weakening fixes.** 20 adversarial cases (auth removal, TLS disable, permission widening) — must return `suggested_fix: null` with a safety note.
2. **No test-silencing fixes.** 10 cases of legitimate test failures — must not suggest `skip`, `ignore`, `continue-on-error` outside flaky quarantine with explicit rationale.
3. **No fabricated commits/files.** 30 diagnosis runs — every `suspect_files.path` and every `suspect_commits` SHA verified to exist in input context.
4. **No hallucinated fixes.** 30 diagnosis runs — every `suggested_fix.diff` line verified to reference only visible code.

Runs as a `go test` suite against a fixed evaluation corpus.

---

## 11. Evaluation corpus

A curated set of real CI failures used to grade prompt changes before merge. Lives in `testdata/eval/`.

Initial corpus target: **30 cases** before first launch, balanced across failure types:

| Failure type | Target count |
|---|---|
| `real_bug` | 10 |
| `flaky_test` | 5 |
| `infra` | 5 |
| `dependency` | 5 |
| `config` | 3 |
| `unknown` (underspecified) | 2 |

Each case: a fixture directory containing the exact `<ci_failure>` XML input plus an `expected.json` with expected `failure_type`, acceptable `confidence` range, and constraints on `suspect_files` / `suggested_fix.file`. The regression test asserts these.

Sources for seed corpus:
- Own side projects (easy, biased toward user's code patterns).
- Public OSS repos — pull real failed CI runs from repos like `gin-gonic/gin`, `sqlc-dev/sqlc`, `actions/runner`.
- Synthesize rare types (`config` with missing secret, `dependency` with go.sum mismatch).

Maintain ≥ 80% correct-classification rate on the corpus as a merge gate.

---

## 12. Versioning

- **Prompt version** stored as `PROMPT_VERSION` constant (semver: `MAJOR.MINOR.PATCH`).
- Persisted on every `diagnoses` row: `diagnoses.prompt_version`.
- MAJOR bump: schema change (new/removed/renamed fields in `record_diagnosis`). Requires MCP_SPEC.md coordination and a `/v2` MCP endpoint if client-visible.
- MINOR bump: new few-shot, new classification guidance, new safety rule.
- PATCH bump: wording tweaks, typo fixes.

Every bump runs the evaluation corpus in CI; merge blocked on regression.

---

## 13. Failure modes of the prompt itself

| Failure | Detection | Response |
|---|---|---|
| Model returns no tool call, only free text | Empty `tool_use` in response | Retry once with an inserted user message "Please call `record_diagnosis` now." Escalate to Opus if still failing. |
| Tool call with invalid schema | SDK validation error | Retry with Opus (rare on Sonnet 4.6, more common on older Haiku). |
| `confidence > 0.7` with empty `suspect_files` for `real_bug` | Post-hoc validator | Downgrade `confidence` to 0.5 and prepend a note in `reasoning`. Log for eval review. |
| Hallucinated file path (not in diff/logs) | Post-hoc validator checks each path against context | Strip the offending `suspect_file`. If all removed, downgrade to `unknown`. |
| `suggested_fix` references a file not in `suspect_files` | Post-hoc validator | Null the `suggested_fix`, log. |

All post-hoc validators live in `internal/prompt/validate.go` and run before `diagnoses.status = complete`.

---

## 14. Iteration strategy

Post-launch, prompt quality is improved via:

1. **Thumbs-down feedback** captured in the PR comment surfaces to a review queue in the CIScout dashboard (founder-only in v0.1).
2. Each thumbs-down is triaged: prompt issue, model capability issue, or user error.
3. Prompt issues → add the case to the evaluation corpus with the correct expected output, reproduce, edit the system prompt, re-run corpus, merge.
4. Model capability issues → evaluate Opus; if Opus also fails, accept and document.
5. Aim: evaluation corpus accuracy ≥ 90% within 60 days of launch.

---

## 15. Open decisions

1. **Few-shot example count.** Starting with 2. If evaluation accuracy plateaus below 85%, consider 3–4, accepting cache cost.
2. **XML vs. JSON context block.** XML chosen for readability + Claude performance; revisit if eval shows issues.
3. **Language-specific guidance.** Current prompt is language-agnostic. Possible future: small language-specific addenda for Go, TypeScript, Python. Defer until eval data shows cross-language quality gap.
4. **Private-repo diff size.** Some customer repos will have huge diffs (feature branches merging hundreds of files). Truncation priority may need tuning based on real customer data.
5. **Multi-job failure handling.** When 3+ jobs fail, the failure may be correlated (shared dep) or independent. Current prompt treats them as one diagnosis. Consider splitting in v0.2.

---

## 16. References

- `PRODUCT.md` §3.1 — diagnosis pipeline product requirements.
- `ARCHITECTURE.md` §3 — where the prompt lives in the runtime flow.
- `ARCHITECTURE.md` §6.2 — Anthropic API integration and retries.
- `MCP_SPEC.md` §3.1 — the canonical `DiagnosisResult` shape this prompt produces.
- Anthropic docs: https://docs.anthropic.com/claude/docs/prompt-caching
- Anthropic docs: https://docs.anthropic.com/claude/docs/tool-use
