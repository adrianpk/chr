# Slice 3: Worker Strategy Context

Status: reviewing
Slicing: Agent Test/Evidence Strategy Pass
Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`
Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`
Branch: `feat/slice-3-worker-strategy-context`
PR: pending

## Purpose

Make development worker runs aware of the evidence strategy contract before
they edit project files, and make completed worker runs emit trace evidence
that can be read from runtime reports.

This slice keeps the strategy pass inline in the existing worker execution.
It does not launch a second planning container and it does not yet enforce the
ready-for-review gate; that enforcement is reserved for a later slice.

## Delivered Behavior

- Codex worker context now includes an `Evidence Strategy` section.
- Worker context instructs the worker to classify change kind and risk, decide
  Gherkin/test/validation obligations, update the ticket checklist, and
  reconcile produced evidence before finishing.
- Worker context includes effective evidence config values.
- Worker context includes parsed ticket checklist state when present.
- Malformed checklist state is surfaced to the worker instead of silently
  ignored.
- Worker result JSON now requires `evidence_summary.strategy`,
  `evidence_summary.produced`, and `evidence_summary.validation`.
- Worker result JSON can optionally include evidence deviations and remaining
  risk.
- Runtime execution reports copy the worker evidence summary and ticket
  checklist state when `execution_report_copy` is enabled.
- Tests cover docs-only and behavior-change evidence strategies.

## Implementation Notes

- The ticket checklist remains the operational source of truth.
- Runtime execution reports are secondary trace evidence for inspection and
  debugging.
- Missing ticket checklist state is represented as `missing`; malformed state
  is represented as `malformed` with the parse error.
- The result contract is intentionally machine-readable so later gate slices
  can classify missing or incomplete evidence.
- During validation, the local signature lint flagged a helper name that
  forced `gofumpt` to split a function signature. The helper/type were renamed
  to keep the signature on one line under the repository lint rule.

## Contracts Added or Changed

- Added `WorkerEvidenceContext`.
- Added evidence config/checklist rendering to `RenderCodexWorkerContext`.
- Added `ExecutionResultEvidenceSummary`.
- Added required `evidence_summary` validation to `ValidateExecutionResult`.
- Added report-level evidence copy fields to `ExecutionReport`.
- Added `IncludeEvidence`, `EvidenceChecklist`, and
  `EvidenceChecklistError` to `ExecutionReportOptions`.

## Files of Interest

- `internal/core/codex_worker_context.go`
- `internal/core/codex_worker_context_test.go`
- `internal/core/execution_result.go`
- `internal/core/execution_result_test.go`
- `internal/core/execution_report.go`
- `internal/core/execution_report_test.go`
- `internal/app/ticket.go`
- `internal/app/cli_test.go`

## Validation

- `go test ./internal/core -run 'ExecutionResult|CodexWorkerContext|ExecutionRunner|ExecutionReport|ReviewHandoff|ResultApplication|PreviousAttempt|GateEvidenceClassifier'` (pass)
- `go test ./internal/app -run 'Ticket(Start|ApplyResult|Loop)|ApplyResult|TicketStart'` (pass)
- `go test ./internal/core -run 'ExecutionReport|ExecutionResult'` (pass)
- `go test ./internal/app -run 'TicketApplyResult'` (pass)
- `go test ./internal/core -run 'CodexWorkerContext|ExecutionResult'` (pass)
- `go test ./internal/core ./internal/app` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Ready-for-review enforcement is not implemented in this slice.
- Gherkin applicability is copied from the checklist state; deeper inference
  and policy checks are delivered in later slices.
- Execution reports intentionally duplicate evidence for traceability but do
  not replace the ticket checklist as primary state.
