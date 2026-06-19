# Slice 5: Execution Workspace Results

Status: delivered
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-5-execution-workspace-results`
PR: #332

## Purpose

Make completed execution records and result application understand the
execution workspace that produced the result, so worktree-mode output is read
from the ticket worktree rather than the user's checkout.

## Delivered Behavior

- Execution records now carry durable workspace metadata: workspace mode,
  project checkout path, execution workspace path, base ref, base SHA, and
  ticket branch.
- Execution reports render a dedicated workspace section with that metadata.
- SQLite persistence stores the new execution workspace metadata through a
  forward migration.
- Changed-file discovery can inspect an explicit workspace root while keeping
  discovered paths project-relative.
- Changed-file discovery rejects absolute or parent-relative paths.
- `ticket apply-result` discovers changed files from the execution workspace in
  worktree mode.
- Project-mode result application keeps the previous result-only changed-file
  behavior.
- Evidence readiness and validation blockers still use the same result contract
  and artifact paths when workspace metadata is present.

## Implementation Notes

- The new SQLite migration is additive and keeps legacy rows compatible through
  empty-string defaults.
- Result application only performs host-side changed-file discovery when the
  execution record is in `worktree` mode and has an execution workspace path.
- Host-side discovery uses Git porcelain output from the execution workspace
  and does not inspect the user's checkout in worktree mode.
- Discovered changed-file paths remain relative to the project root contract
  used by tickets and execution results.

## Contracts Added or Changed

- `ExecutionWorkspaceRef` records workspace mode, checkout path, execution
  workspace path, base ref, base SHA, and ticket branch.
- `ExecutionRecord.Workspace` persists execution workspace metadata.
- `ExecutionReport.Workspace` renders execution workspace metadata.
- `DiscoverChangedFilesOptions` accepts separate `ProjectDir` and
  `WorkspaceDir` values.
- `ticket apply-result` passes discovered worktree changed files into
  `NewResultApplication`.

## Files of Interest

- `internal/core/execution.go`
- `internal/core/execution_report.go`
- `internal/core/execution_repository.go`
- `internal/core/changed_file_discovery.go`
- `internal/core/result_application_test.go`
- `internal/app/ticket.go`
- `internal/app/cli_test.go`
- `cmd/chc/assets/migration/sqlite/20260619000011-add_execution_workspace_metadata.sql`

## Validation

- `go test ./internal/core -run 'ExecutionRecord|ExecutionRepository|ExecutionReport|Migrator|Migration'` (pass)
- `go test ./internal/core -run 'ChangedFileDiscoverer|ParseGitStatusChangedFiles'` (pass)
- `go test ./internal/core -run 'TestNewResultApplication|TestChangedFileDiscoverer|TestParseGitStatusChangedFiles' && go test ./internal/app -run 'TestTicketApplyResultDiscoversWorktreeChangedFiles|TestTicketApplyResultAppliesExecutionResult'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Slice 6 still owns ticket branch preparation, host-side push, and PR creation
  from the ticket worktree.
- Review repair and post-merge cleanup still depend on later worktree lifecycle
  slices.
- Worktree mode remains opt-in until the final QA proving slice validates the
  full cycle end to end.
