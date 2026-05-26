# Architecture Boundary Cleanup Slicing Tracker

Status: in_progress

Plan: `.prj/plans/slicing-architecture-boundary-cleanup-plan.md`

## Active Slice

Slicing name: Architecture Boundary Cleanup

Active slice number: Slice 1

Active slice short name: dependency context cleanup

Expected branch name: `feat/slice-1-dependency-context-cleanup`

Expected PR title: `feat(slice-1): dependency context cleanup`

Expected report path:
`.prj/reports/slices/architecture-boundary-cleanup/slice-1-dependency-context-cleanup.md`

Prerequisite PRs or delivered work:

- CLI TUI Surface Parity slicing completed.
- Documentation and parity matrix already describe current CLI and TUI
  operation surfaces.

## Slicing Checklist

- [x] Capture architecture cleanup objective.
- [x] Define complete ordered slice map.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Dependency Context Cleanup | Reviewing | `feat/slice-1-dependency-context-cleanup` | `feat(slice-1): dependency context cleanup` | `.prj/reports/slices/architecture-boundary-cleanup/slice-1-dependency-context-cleanup.md` | PR `#209`; hidden startup asset dependency removed; `make verify` passed. |
| Slice 2: Operation Context Boundary | Planned | `feat/slice-2-operation-context-boundary` | `feat(slice-2): operation context boundary` | `.prj/reports/slices/architecture-boundary-cleanup/slice-2-operation-context-boundary.md` | Move cancellable service work to operation-scoped context. |
| Slice 3: Adapter Service Shape Cleanup | Planned | `feat/slice-3-adapter-service-shape-cleanup` | `feat(slice-3): adapter service shape cleanup` | `.prj/reports/slices/architecture-boundary-cleanup/slice-3-adapter-service-shape-cleanup.md` | Move metadata workflow logic into service methods, split TUI adapters, and rename neutral scope helpers. |

## Slice 1: Dependency Context Cleanup

Candidate tasks:

- [x] T1.1 Replace package-level asset filesystem state with explicit app
  dependencies.
- [x] T1.2 Pass asset dependencies into project context resolution without
  mutable global state.
- [x] T1.3 Preserve CLI and TUI startup behavior with focused tests or existing
  coverage.
- [x] T1.4 Create slice implementation report.
- [x] T1.5 Run validation.

Task commits:

- `32922ed refactor(app): pass embedded assets explicitly`
- `bd7843f refactor(ticket): share ticket index setup`
- `5958ca5 docs(architecture): report dependency context cleanup`
- `2019d0a docs(architecture): record dependency cleanup validation`

Validation:

- `go test ./internal/chcapp`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 2: Operation Context Boundary

Candidate tasks:

- [ ] T2.1 Identify service operations that should receive operation-scoped
  context.
- [ ] T2.2 Update service method signatures and call sites for selected
  operations.
- [ ] T2.3 Remove obsolete `contextcheck` suppressions from CLI flow code.
- [ ] T2.4 Preserve CLI and TUI behavior with focused tests or existing
  coverage.
- [ ] T2.5 Create slice implementation report.
- [ ] T2.6 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 3: Adapter Service Shape Cleanup

Candidate tasks:

- [ ] T3.1 Move metadata project patch/upsert workflow into service methods.
- [ ] T3.2 Move metadata case patch/upsert workflow into service methods.
- [ ] T3.3 Split TUI adapter wrappers into domain-focused files.
- [ ] T3.4 Rename shared scope helpers from quiz-specific naming to neutral
  naming.
- [ ] T3.5 Update documentation or parity matrix paths if implementation paths
  changed.
- [ ] T3.6 Create slice implementation report.
- [ ] T3.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.
