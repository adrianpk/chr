# Slice 2: Userspace Resolver

Status: reviewing
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-2-userspace-resolver`
PR: pending

## Purpose

Introduce the project-local active userspace boundary needed by later Board
work without relocating existing ChrysoCoord artifacts.

## Delivered Behavior

- Core code can resolve `.chc/user/` as the active project userspace.
- `CHC_OPERATOR_ID` identifies the active local operator for userspace-bound
  workflows.
- Operator ids are validated before use.
- `chc init` creates `.chc/user/`.
- `EnsureProjectUserspace` can create the userspace directory for repair or
  future targeted setup paths.
- Documentation covers `.envrc` usage and non-direnv shell setup.

## Implementation Notes

- The active userspace path is `.chc/user/`, not
  `.chc/operators/<operator-id>/...`.
- The operator id is still required even though the MVP path is single active
  userspace. This preserves the operator ownership boundary for later
  multi-operator work.
- Existing dumps, quizzes, specs, plans, tickets, runtime evidence, and reports
  are not moved in this slice.

## Contracts Added or Changed

- Added `CHC_OPERATOR_ID`.
- Added `ProjectUserspace`.
- Added `ResolveProjectUserspace`.
- Added `EnsureProjectUserspace`.
- Added `IsValidOperatorID`.
- Added `.chc/user/` to initialized workspace directories.

## Files of Interest

- `internal/chc/project_userspace.go`
- `internal/chc/project_userspace_test.go`
- `internal/chc/workspace_init.go`
- `internal/chc/workspace_init_test.go`
- `docs/reference/project-workspace.md`
- `docs/reference/user-environment.md`
- `docs/how-to/user-guide/index.md`

## Validation

- `go test ./internal/chc -run 'TestResolveProjectUserspace|TestEnsureProjectUserspace|TestIsValidOperatorID|TestInitCreatesProjectWorkspace'`

## Risks and Follow-ups

- No workflow consumes `ProjectUserspace` yet. Later slices should adopt it
  only where userspace ownership is relevant.
- Full multi-operator layout remains deferred.
