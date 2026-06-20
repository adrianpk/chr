# Slice 3: High-Level Agentic Review

Status: reviewing
Slicing: Serialized Integration Loop
Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`
Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`
Branch: `feat/slice-3-high-level-agentic-review`
PR: pending

## Purpose

Add the high-level reviewer phase for `agentic_review` mode so a selected
ticket can move from development result application to local agentic review
before the loop allows merge or repair handling.

## Delivered Behavior

- Agentic review now has a structured context artifact, prompt contract, JSON
  output schema, parser, and validator.
- Review artifacts are persisted under `.chc/user/runtime/reviews/<ticket>/<execution>/`.
- Review decisions map to ticket review metadata:
  `pass`, `needs_changes`, `insufficient_evidence`, and
  `needs_human_decision`.
- Clean pass requires existing branch and pull request review metadata before
  the ticket can move to `ready_for_merge`.
- Non-blocking review notes become local ticket comments, limited by
  `review.non_blocking_comment_limit`.
- `work run-all` executes the agentic review phase when `review.mode` is
  `agentic_review`, then stops the loop before later tickets because merge is
  still required by the cycle policy.

## Implementation Notes

- The reviewer prompt explicitly blocks only objective correctness, scope,
  evidence, contract, safety, or mergeability issues.
- Cosmetic style, optional refactors, naming preferences, and non-critical
  cleanup are non-blocking when configured validation passed.
- In project workspace mode, review ticket mutations are reconciled into the
  ticket index after comments are written.
- In worktree workspace mode, review reads and writes the execution worktree
  ticket and keeps the project checkout index untouched.
- Merge orchestration is intentionally deferred to Slice 4.

## Contracts Added or Changed

- `core.AgenticReviewRunner` is the app-level review execution interface.
- `core.CodexAgenticReviewRunner` runs the local Codex reviewer and writes
  context, result, and summary artifacts.
- `core.ApplyAgenticReviewResult` converts structured review output into
  existing ticket review metadata.
- `StartAgentWorkLoop` now invokes agentic review after successful development
  result application when the selected review workflow mode is
  `agentic_review`.

## Files of Interest

- `internal/core/agentic_review.go`
- `internal/core/agentic_review_test.go`
- `internal/app/service_work.go`
- `internal/app/service_work_test.go`
- `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`

## Validation

- `go test ./internal/core -run AgenticReview` (pass)
- `go test ./internal/app -run 'AgenticReview|StartAllReadyAgentWork|WorkRunAll'` (pass)
- `make lint-strict` (pass)
- `go test ./internal/app ./internal/core` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Merge is not executed in this slice. A clean agentic pass moves the ticket to
  `ready_for_merge`, then the loop stops until Slice 4 wires merge execution.
- Repair is not retried in this slice. `needs_changes` returns the ticket to
  `ready_for_dev`; finite repair handling is planned for Slice 5.
