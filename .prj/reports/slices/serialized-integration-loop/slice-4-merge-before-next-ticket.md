# Slice 4: Merge Before Next Ticket

Status: ready for review
Slicing: Serialized Integration Loop
Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`
Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`
Branch: `feat/slice-4-merge-before-next-ticket`
PR: pending

## Purpose

Require the integration loop to merge the current ticket before selecting later
normal work.

## Delivered Behavior

- `agentic_review` mode now runs merge preflight and merge execution after a
  clean review pass.
- `trusted_auto_merge` mode now writes trusted pass review metadata and still
  runs merge before continuing.
- Merge success marks the ticket `done` and allows the next eligible ticket to
  be selected.
- Merge preflight blocks and provider merge failures leave the ticket
  `ready_for_merge`, record blocked merge metadata, and stop the loop.
- `work run-all` execution output now prints merge result rows.
- Existing merge command behavior remains the merge implementation path, so
  provider handling and worktree cleanup policy stay centralized.

## Implementation Notes

- The app loop delegates merge to `core.RunTicketMergeCommand`; it does not
  duplicate provider, preflight, or cleanup logic.
- The loop carries merge results in `AgentWorkResult` and
  `ticketLoopExecutionResult`.
- The post-development cycle is split into review/trusted handoff and merge
  helpers to keep the runner readable and lint-compliant.
- Worktree-mode merge reads ticket state from the execution workspace while
  project-mode merge uses the normal project ticket list.

## Contracts Added or Changed

- `core.ApplyTrustedAutoMergeReview` writes pass review handoff metadata for
  trusted auto-merge mode.
- `StartAgentWorkLoop` now continues past a ticket only after merge success
  when the cycle policy requires merge.
- `work run-all` output includes `merge: <ticket> <state> <result> <commit>`
  lines for completed merge attempts.

## Files of Interest

- `internal/core/trusted_auto_merge.go`
- `internal/core/trusted_auto_merge_test.go`
- `internal/app/service_work.go`
- `internal/app/service_work_test.go`
- `internal/app/work.go`
- `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`

## Validation

- `go test ./internal/core -run 'TrustedAutoMerge|TicketMergeCommand|TicketMergePreflight|IntegrationCycle'` (pass)
- `go test ./internal/app -run 'StartAllReadyAgentWork|WorkRunAll'` (pass)
- `go test ./internal/app` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Repair after `needs_changes` is still deferred to Slice 5; a review block
  still stops the loop instead of repairing in-place.
- Dash and richer event/report visualization remain deferred to Slice 6.
