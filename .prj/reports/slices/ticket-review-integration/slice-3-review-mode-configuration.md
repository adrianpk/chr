# Slice 3: Review Mode Configuration

Status: delivered
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-3-review-mode-configuration`
PR: #275

## Purpose

Make ticket review behavior explicit and configurable before adding provider
diagnostics or remote review execution.

## Delivered Behavior

- Workspace config supports `review.mode` with `human_review`,
  `agentic_review`, and `trusted_auto_merge`.
- Workspace config supports `review.non_blocking_comment_limit`.
- New workspaces include explicit review defaults in `.chr/default/config.yml`.
- `chr ticket review-mode status` prints the active mode, validation status,
  mode implications, clean-pass remote comment policy, and non-blocking comment
  limit.
- `chr ticket review-mode status --dry-run` prints the same read-only plan
  without performing review work.

## Implementation Notes

- `human_review` is the conservative default.
- Invalid review modes remain loadable so status output can explain the
  configuration error instead of failing before diagnostics are printed.
- This slice only configures review policy. Provider access checks and actual
  review execution are left to later slices.

## Contracts Added or Changed

- `.chr/default/config.yml` may include:
  `review.mode` and `review.non_blocking_comment_limit`.
- Valid review modes are `human_review`, `agentic_review`, and
  `trusted_auto_merge`.
- Clean agentic review passes are represented as no remote comment ceremony.
- Non-blocking remote comments are capped by
  `review.non_blocking_comment_limit`.

## Files of Interest

- `internal/chr/config.go`
- `internal/chr/review_workflow_config.go`
- `internal/chrapp/ticket.go`
- `cmd/chr/assets/workspace/profiles/standard/config.yml`
- `docs/reference/review-workflow-config.md`

## Validation

- `go test ./internal/chr -run 'Config|ReviewWorkflow'`
- `go test ./internal/chrapp -run 'TicketReviewMode|Workspace|Init'`
- `go test ./internal/chr ./internal/chrapp`
- `make verify`

## Risks and Follow-ups

- Provider diagnostics must define what remote access evidence is required for
  `agentic_review`.
- Review execution must apply `trusted_auto_merge` and `human_review` without
  adding remote comment noise.
