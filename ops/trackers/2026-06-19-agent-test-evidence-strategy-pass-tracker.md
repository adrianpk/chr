# Agent Test/Evidence Strategy Pass Tracker

Status: delivered

Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`

Spec: `ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md`

Quiz: `ops/conv/quiz/2026-06-19-agent-test-evidence-strategy-pass-quiz.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | evidence strategy config | delivered | `feat/slice-1-evidence-strategy-config` | `#320` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-1-evidence-strategy-config.md` |
| Slice 2 | evidence checklist writer | delivered | `feat/slice-2-evidence-checklist-writer` | `#321` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-2-evidence-checklist-writer.md` |
| Slice 3 | worker strategy context | delivered | `feat/slice-3-worker-strategy-context` | `#322` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-3-worker-strategy-context.md` |
| Slice 4 | review evidence gate | delivered | `feat/slice-4-review-evidence-gate` | `#323` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-4-review-evidence-gate.md` |
| Slice 5 | gherkin test policy | delivered | `feat/slice-5-gherkin-test-policy` | `#324` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-5-gherkin-test-policy.md` |
| Slice 6 | evidence visibility docs | delivered | `feat/slice-6-evidence-visibility-docs` | `#325` | `.prj/reports/slices/agent-test-evidence-strategy-pass/slice-6-evidence-visibility-docs.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add evidence strategy config structs, defaults, and parser support under `.chr/user/config.yml`. | done | `44ea5a0` | Defaults preserve current behavior until later slices enforce strategy. |
| T1.2 | Add enum validation for strategy mode, checklist policy, Gherkin policy, strict gating, user edits, and report copy behavior. | done | `b880c70` | Invalid enum values fail clearly. |
| T1.3 | Add core strategy status and diagnostic value types used by later slices. | done | `d655841` | No behavior gate yet. |
| T1.4 | Add config tests for omitted defaults, explicit overrides, and invalid values. | done | `ed8d9b4` | Covers defaults, overrides, and invalid enum values. |
| T1.5 | Document the config contract in reference docs. | done | `9ead9ca` | Sane defaults plus override behavior. |
| T1.6 | Add the slice implementation report and validation evidence. | done | `141225c` | Report created with reviewing status. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add parser coverage for strategy block anchors, required labels, checked state, status values, and malformed blocks. | done | `6b1ce94` | Parser rejects malformed blocks instead of overwriting them. |
| T2.2 | Add a renderer for the v1 strategy checklist block. | done | `c58c28d` | Renderer keeps canonical order and readable Markdown. |
| T2.3 | Add a safe ticket Markdown updater that inserts or updates only the evidence strategy block. | done | `d2ca2f1` | Updater preserves valid non-updated user-authored items. |
| T2.4 | Wire a core/app service path for evidence checklist write operations. | done | `2826a3c` | Service writes by ticket key and reloads ticket state. |
| T2.5 | Add CLI-level tests or service tests that prove user content is preserved. | done | `50403f6` | Service test covers body section preservation. |
| T2.6 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Extend context materialization with an Evidence Strategy section. | done | `47b243c` | Inline worker pass, no second container. |
| T3.2 | Include current evidence config and parsed checklist state in the worker context. | done | `c76a8f5` | Existing user edits are strong input. |
| T3.3 | Update the worker result contract to require evidence summary metadata. | done | `38526a4` | Result envelope now requires evidence summary metadata. |
| T3.4 | Copy strategy and evidence summary into runtime reports when enabled. | done | `69a5ef5` | Reports copy strategy trace evidence when configured. |
| T3.5 | Add context and result-contract tests for docs-only and behavior-change tickets. | done | `cd8f8de` | Proves no-test and required-evidence paths. |
| T3.6 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add an evidence readiness evaluator over parsed ticket strategy state, execution result, validation result, and review refs. | done | `d915612` | Produces classified blockers. |
| T4.2 | Integrate the evaluator into ticket result application. | done | `f917d38` | Worker completion cannot bypass evidence. |
| T4.3 | Integrate strict evidence gating into ticket status transitions to `ready_for_review`. | done | `f917d38` | Manual status changes respect strict mode. |
| T4.4 | Add tests for passed, missing, malformed, blocked, validation-failed, and optional-strategy paths. | done | `f917d38` | Includes disabled/optional config and loop fixture coverage. |
| T4.5 | Update reference docs for review-transition evidence requirements. | done | `30054e0` | Review workflow docs describe the gate and opt-outs. |
| T4.6 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add project-level active Gherkin detection based on existing acceptance feature configuration/artifacts. | done | `271d865` | Detects `.chr/user/acceptance/*.feature` under active policy. |
| T5.2 | Add strategy guidance rendering for change kind, risk, test timing, and evidence matrix. | done | `45febeb`, `10d4f26` | Worker context renders active Gherkin, suggested strategy, and decision matrix. |
| T5.3 | Add required Gherkin checks for active/applicable behavior tickets. | done | `45febeb`, `e75f453` | Active applicable Gherkin rejects `not_applicable` at review gate. |
| T5.4 | Add tests for docs-only no-test strategy, user-visible Gherkin-required strategy, bugfix regression guidance, and parser table/fuzz guidance. | done | `271d865`, `45febeb` | Core policy, worker context, readiness, and apply-result coverage. |
| T5.5 | Update acceptance/Gherkin docs to explain the evidence strategy interaction. | done | `b37d681` | Acceptance feature docs explain active/applicable Gherkin. |
| T5.6 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Add evidence status to indexed ticket detail or a derived Board read model. | done | `4c2e18a` | Core derives concise evidence summaries from ticket Markdown. |
| T6.2 | Render evidence status in Board detail and relevant Dash messages. | done | `b1a6775` | Board detail and Dash work-action details show concise evidence references. |
| T6.3 | Update ticket show/list/board CLI output where appropriate. | done | `3163031` | CLI show/list/board print evidence status without changing normal columns. |
| T6.4 | Update user guide, command reference, runtime config, ticket execution, ticket review workflow, and testing doctrine docs. | done | `93630e6` | Docs cover defaults, visibility, review gating, and evidence references. |
| T6.5 | Add or update QA fixture coverage for no-test docs work, behavior work with Gherkin, and blocked validation. | done | `bf44f46` | Tic-tac-toe reference includes resettable evidence examples and active Gherkin. |
| T6.6 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Validation Log

- Slice 1: `go test ./internal/core -run 'WorkspaceConfig|EvidenceStrategy|Validate.*Config'` (pass)
- Slice 1: `golangci-lint run --timeout=5m ./internal/core` (pass)
- Slice 1: `make verify` (pass)
- Slice 1: PR `#320` merged to `dev` on 2026-06-19.
- Slice 2: `go test ./internal/core ./internal/app -run 'EvidenceStrategyChecklist|ServiceUpsertTicketEvidenceStrategyChecklist'` (pass)
- Slice 2: `golangci-lint run --timeout=5m ./internal/core ./internal/app` (pass)
- Slice 2: `make verify` (pass)
- Slice 2: PR `#321` merged to `dev` on 2026-06-19.
- Slice 3: `go test ./internal/core -run 'ExecutionResult|CodexWorkerContext|ExecutionRunner|ExecutionReport|ReviewHandoff|ResultApplication|PreviousAttempt|GateEvidenceClassifier'` (pass)
- Slice 3: `go test ./internal/app -run 'Ticket(Start|ApplyResult|Loop)|ApplyResult|TicketStart'` (pass)
- Slice 3: `go test ./internal/core -run 'ExecutionReport|ExecutionResult'` (pass)
- Slice 3: `go test ./internal/app -run 'TicketApplyResult'` (pass)
- Slice 3: `go test ./internal/core -run 'CodexWorkerContext|ExecutionResult'` (pass)
- Slice 3: `go test ./internal/core ./internal/app` (pass)
- Slice 3: `make lint-strict` (pass)
- Slice 3: `make verify` (pass)
- Slice 3: PR `#322` merged to `dev` on 2026-06-19.
- Slice 4: `go test ./internal/core -run 'ResultApplication|EvidenceReadiness|TicketStatusUpdater'` (pass)
- Slice 4: `go test ./internal/app -run 'TestTicket(ApplyResult|ReviewRef|Status)'` (pass)
- Slice 4: `go test ./internal/app -run 'TestTicketLoopExecutesEligibleTickets|TestTicketLoopHonorsMaxIterations|TestWorkRunAllExecutesReadyTickets|TestTicketLoopDispatchesConfiguredFinishedNotification|TestServiceStartAgentWorkTicketReusesTicketExecution|TestServiceStartAllReadyAgentWork'` (pass)
- Slice 4: `make lint-strict` (pass)
- Slice 4: `make verify` (pass)
- Slice 4: PR `#323` merged to `dev` on 2026-06-19.
- Slice 5: `go test ./internal/core -run 'Evidence.*Policy|EvidenceReadiness|CodexWorkerContext'` (pass)
- Slice 5: `go test ./internal/app -run 'TestTicketApplyResult'` (pass)
- Slice 5: `make lint-strict` (pass)
- Slice 5: `make verify` (pass)
- Slice 5: PR `#324` merged to `dev` on 2026-06-19.
- Slice 6: `go test ./internal/core -run 'SummarizeEvidenceStrategyChecklist|NewEvidenceStrategySummary'` (pass)
- Slice 6: `go test ./internal/tui -run 'TestModelRendersTicketEvidenceDetail|TestModelRendersWorkActionEventStream'` (pass)
- Slice 6: `go test ./internal/app -run 'TestTicketShowPrintsEvidenceStrategySummary|TestTicketListPrintsIndexedTickets|TestTicketBoardPrintsBoardColumns'` (pass)
- Slice 6: `make qa-smoke-tictactoe` (pass)
- Slice 6: `make lint-strict` (pass)
- Slice 6: `make verify` (pass)
- Slice 6: PR `#325` merged to `dev` on 2026-06-19.

## Notes

- The plan intentionally starts with config and checklist infrastructure because
  the current TUI exposes checklist capabilities but the safe checklist writer
  is not implemented.
- Strategy execution stays inline in the development worker by default.
- Separate-pass strategy remains future-compatible but not implemented in this
  slicing.
- The ticket checklist is the primary operational source; runtime reports are
  secondary trace evidence.
