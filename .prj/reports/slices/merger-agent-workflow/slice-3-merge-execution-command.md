# Slice 3: Merge Execution Command

Status: delivered
Slicing: Merger Agent Workflow
Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
Branch: `feat/slice-3-merge-execution-command`
PR: `#285`

## Purpose

Add the execution path that lets the merger role merge a reviewed ticket pull
request after objective preflight checks.

## Delivered Behavior

- Added `RunTicketMergeCommand` as the service boundary for reviewed ticket
  merge execution.
- Added `chc ticket merge <ticket-key>` with `--dry-run`, `--strategy`, and
  `--base` options.
- Added Forgejo-compatible merge execution through `tea api`.
- Added GitHub merge execution through `gh pr merge`.
- Successful merges update the ticket to `done` and record merge metadata.
- Blocked merge attempts record compact local blocker metadata without adding a
  new board state.
- Dry-run mode reports preflight and intended action without mutating remote or
  ticket state.

## Implementation Notes

The command reuses Slice 2 preflight before any provider merge call. Provider
merge execution stays behind the existing remote review adapter, which keeps CLI
integration details out of ticket state logic.

The command acquires the `merge` operation lock for real merges so development,
review, and merge operations remain serial in one workspace. Dry-runs skip the
lock and state writes.

Blocked provider failures and blocked preflight attempts keep the ticket in
`ready_for_merge` with `merge_state: "blocked"` and a compact `merge_blocker`
value. Successful merges clear blocker metadata and set `merged_at`.

## Contracts Added or Changed

- Added `MergeRequest` and `MergeResult` for provider merge execution.
- Extended `RemoteReviewAdapter` with `Merge`.
- Added `TicketMergeCommandOptions` and `TicketMergeCommandResult`.
- Added CLI output under `ticket_merge: result`.
- Added `AgentOperationMerge` for serial operation locking.

## Files of Interest

- `internal/chc/ticket_merge_command.go`
- `internal/chc/ticket_merge_command_test.go`
- `internal/chc/review_remote_adapter.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/cli_test.go`

## Validation

- `scripts/check-go-func-signatures.sh`
- `golangci-lint run --timeout=5m ./internal/chc/... ./internal/chcapp/...`
- `go test ./internal/chc ./internal/chcapp`
- `git diff --check`
- `make verify`

## Risks and Follow-ups

- Provider merge response SHA extraction is intentionally narrow and records an
  empty merge commit when the provider response does not include `sha`.
- Slice 4 must document the command, metadata, and role handoff behavior.
