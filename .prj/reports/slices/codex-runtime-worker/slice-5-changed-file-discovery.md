# Slice 5: Changed File Discovery

Status: reviewing
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/slice-5-changed-file-discovery`
PR: pending

## Purpose

Discover workspace file changes caused by the worker run.

## Delivered Behavior

The core package now exposes a changed-file discoverer that uses git status to
find workspace-relative changed paths.

The discoverer:

- runs `git status --porcelain=v1 --untracked-files=all`;
- scopes discovery to the project directory;
- returns clean worktrees as an empty list;
- ignores operational `.chc/runtime/` artifacts;
- does not stage, commit, revert, clean, or otherwise mutate files.

## Implementation Notes

The discoverer returns `ExecutionResultChangedFile` values with only `Path`
populated. Later orchestration can combine these paths with worker-provided
summaries.

## Contracts Added or Changed

New core types:

- `ChangedFileDiscoverer`
- `DiscoverChangedFilesOptions`

New function:

- `NewChangedFileDiscoverer`

## Files of Interest

- `internal/chc/changed_file_discovery.go`
- `internal/chc/changed_file_discovery_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pending

## Risks and Follow-ups

Discovery currently depends on the project directory being a git worktree. Later
orchestration should classify non-git discovery failures clearly.
