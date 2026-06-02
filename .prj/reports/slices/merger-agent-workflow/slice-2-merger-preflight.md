# Slice 2: Merger Preflight

Status: reviewing
Slicing: Merger Agent Workflow
Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
Branch: `feat/slice-2-merger-preflight`
PR: `#284`

## Purpose

Add an objective merger preflight that can decide whether a reviewed ticket pull
request is integration-ready before any provider merge call exists.

## Delivered Behavior

- Added `RunTicketMergePreflight` as a pure service boundary under
  `internal/chc`.
- Added preflight result and blocker models with machine-readable blocker codes
  and blocker kind classification.
- Reused the Slice 1 merge handoff contract to require `ready_for_merge`,
  passed review metadata, branch metadata, and a pull request review reference.
- Added remote pull request metadata checks for open state, expected base,
  expected head branch, and explicit non-mergeability.
- Added serial repair blocking so pending review-repair tickets stop merge
  preflight.
- Added merge strategy validation for `merge`, `rebase`, and `squash`.

## Implementation Notes

The preflight reads remote pull request snapshots through a narrow interface.
It does not call provider merge APIs and does not write ticket state.

Blockers are classified as:

- `operational`: missing handoff, unreadable provider, closed PR, wrong base,
  or invalid merge strategy.
- `repair`: head mismatch, not mergeable, or pending serial review repair.

The remote metadata parser supports the JSON shapes already read by the remote
review adapter for Forgejo-compatible providers and GitHub.

## Contracts Added or Changed

- Added ticket merge preflight constants for blocker codes, blocker kinds, and
  merge strategies.
- Added `TicketMergePreflightRemoteReader`.
- Added `TicketMergePreflightOptions`, `TicketMergePreflightResult`, and
  `TicketMergePreflightBlocker`.

## Files of Interest

- `internal/chc/ticket_merge_preflight.go`
- `internal/chc/ticket_merge_preflight_test.go`

## Validation

- `scripts/check-go-func-signatures.sh`
- `go test ./internal/chc`
- `git diff --check`

## Risks and Follow-ups

- Provider mergeability semantics are still intentionally normalized only to
  the explicit values needed for this preflight.
- Slice 3 must connect this preflight to merge execution and ticket updates.
