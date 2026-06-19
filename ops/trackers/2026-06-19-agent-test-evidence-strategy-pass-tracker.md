# Agent Test/Evidence Strategy Pass Tracker

Status: active

Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`

Spec: `ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md`

Quiz: `ops/conv/quiz/2026-06-19-agent-test-evidence-strategy-pass-quiz.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | evidence strategy config | reviewing | `feat/slice-1-evidence-strategy-config` | pending | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-1-evidence-strategy-config.md` |
| Slice 2 | evidence checklist writer | planned | `feat/slice-2-evidence-checklist-writer` | pending | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-2-evidence-checklist-writer.md` |
| Slice 3 | worker strategy context | planned | `feat/slice-3-worker-strategy-context` | pending | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-3-worker-strategy-context.md` |
| Slice 4 | review evidence gate | planned | `feat/slice-4-review-evidence-gate` | pending | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-4-review-evidence-gate.md` |
| Slice 5 | gherkin test policy | planned | `feat/slice-5-gherkin-test-policy` | pending | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-5-gherkin-test-policy.md` |
| Slice 6 | evidence visibility docs | planned | `feat/slice-6-evidence-visibility-docs` | pending | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-6-evidence-visibility-docs.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add evidence strategy config structs, defaults, and parser support under `.chc/user/config.yml`. | done | `44ea5a0` | Defaults preserve current behavior until later slices enforce strategy. |
| T1.2 | Add enum validation for strategy mode, checklist policy, Gherkin policy, strict gating, user edits, and report copy behavior. | done | `b880c70` | Invalid enum values fail clearly. |
| T1.3 | Add core strategy status and diagnostic value types used by later slices. | done | `d655841` | No behavior gate yet. |
| T1.4 | Add config tests for omitted defaults, explicit overrides, and invalid values. | done | `ed8d9b4` | Covers defaults, overrides, and invalid enum values. |
| T1.5 | Document the config contract in reference docs. | done | `9ead9ca` | Sane defaults plus override behavior. |
| T1.6 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add parser coverage for strategy block anchors, required labels, checked state, status values, and malformed blocks. | pending | pending | Foundation for all later enforcement. |
| T2.2 | Add a renderer for the v1 strategy checklist block. | pending | pending | Keep Markdown human-readable. |
| T2.3 | Add a safe ticket Markdown updater that inserts or updates only the evidence strategy block. | pending | pending | Preserve user ticket content. |
| T2.4 | Wire a core/app service path for evidence checklist write operations. | pending | pending | May also unblock existing checklist capability surfaces. |
| T2.5 | Add CLI-level tests or service tests that prove user content is preserved. | pending | pending | Include frontmatter, comments, blocker, and body sections. |
| T2.6 | Add the slice implementation report and validation evidence. | pending | pending | Report required before PR. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Extend context materialization with an Evidence Strategy section. | pending | pending | Inline worker pass, no second container. |
| T3.2 | Include current evidence config and parsed checklist state in the worker context. | pending | pending | Existing user edits are strong input. |
| T3.3 | Update the worker result contract to require evidence summary metadata. | pending | pending | Keep result envelope machine-readable. |
| T3.4 | Copy strategy and evidence summary into runtime reports when enabled. | pending | pending | Reports are trace evidence, not primary state. |
| T3.5 | Add context and result-contract tests for docs-only and behavior-change tickets. | pending | pending | Proves no-test and required-evidence paths. |
| T3.6 | Add the slice implementation report and validation evidence. | pending | pending | Report required before PR. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add an evidence readiness evaluator over parsed ticket strategy state, execution result, validation result, and review refs. | pending | pending | Produces classified blockers. |
| T4.2 | Integrate the evaluator into ticket result application. | pending | pending | Worker completion cannot bypass evidence. |
| T4.3 | Integrate strict evidence gating into ticket status transitions to `ready_for_review`. | pending | pending | Manual status changes should respect strict mode. |
| T4.4 | Add tests for passed, missing, malformed, blocked, validation-failed, and optional-strategy paths. | pending | pending | Include disabled/optional config. |
| T4.5 | Update reference docs for review-transition evidence requirements. | pending | pending | Keep review workflow docs current. |
| T4.6 | Add the slice implementation report and validation evidence. | pending | pending | Report required before PR. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add project-level active Gherkin detection based on existing acceptance feature configuration/artifacts. | pending | pending | Must be explicit and testable. |
| T5.2 | Add strategy guidance rendering for change kind, risk, test timing, and evidence matrix. | pending | pending | Evidence-first, not universal TDD. |
| T5.3 | Add required Gherkin checks for active/applicable behavior tickets. | pending | pending | If active and applicable, `.feature` must exist or update. |
| T5.4 | Add tests for docs-only no-test strategy, user-visible Gherkin-required strategy, bugfix regression guidance, and parser table/fuzz guidance. | pending | pending | Use table-driven tests where possible. |
| T5.5 | Update acceptance/Gherkin docs to explain the evidence strategy interaction. | pending | pending | Tie into existing acceptance workflow. |
| T5.6 | Add the slice implementation report and validation evidence. | pending | pending | Report required before PR. |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Add evidence status to indexed ticket detail or a derived Board read model. | pending | pending | Avoid making runtime logs the only visibility path. |
| T6.2 | Render evidence status in Board detail and relevant Dash messages. | pending | pending | Concise status only. |
| T6.3 | Update ticket show/list/board CLI output where appropriate. | pending | pending | Keep output readable. |
| T6.4 | Update user guide, command reference, runtime config, ticket execution, ticket review workflow, and testing doctrine docs. | pending | pending | Document defaults and overrides. |
| T6.5 | Add or update QA fixture coverage for no-test docs work, behavior work with Gherkin, and blocked validation. | pending | pending | Keep scenarios resettable. |
| T6.6 | Add the slice implementation report and validation evidence. | pending | pending | Report required before PR. |

## Validation Log

- Slice 1: `go test ./internal/core -run 'WorkspaceConfig|EvidenceStrategy|Validate.*Config'` (pass)
- Slice 1: `golangci-lint run --timeout=5m ./internal/core` (pass)
- Slice 1: `make verify` (pass)

## Notes

- The plan intentionally starts with config and checklist infrastructure because
  the current TUI exposes checklist capabilities but the safe checklist writer
  is not implemented.
- Strategy execution stays inline in the development worker by default.
- Separate-pass strategy remains future-compatible but not implemented in this
  slicing.
- The ticket checklist is the primary operational source; runtime reports are
  secondary trace evidence.
