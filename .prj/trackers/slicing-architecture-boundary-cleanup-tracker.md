# Architecture Boundary Cleanup Slicing Tracker

Status: in_progress

Plan: `.prj/plans/slicing-architecture-boundary-cleanup-plan.md`

## Active Slice

Slicing name: Architecture Boundary Cleanup

Active slice number: Slice 3

Active slice short name: adapter service shape cleanup

Expected branch name: `feat/slice-3-adapter-service-shape-cleanup`

Expected PR title: `feat(slice-3): adapter service shape cleanup`

Expected report path:
`.prj/reports/slices/architecture-boundary-cleanup/slice-3-adapter-service-shape-cleanup.md`

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
| Slice 1: Dependency Context Cleanup | Delivered | `feat/slice-1-dependency-context-cleanup` | `feat(slice-1): dependency context cleanup` | `.prj/reports/slices/architecture-boundary-cleanup/slice-1-dependency-context-cleanup.md` | PR `#209`; hidden startup asset dependency removed; `make verify` passed. |
| Slice 2: Operation Context Boundary | Delivered | `feat/slice-2-operation-context-boundary` | `feat(slice-2): operation context boundary` | `.prj/reports/slices/architecture-boundary-cleanup/slice-2-operation-context-boundary.md` | PR `#210`; generation operations now receive operation-scoped context; `make verify` passed. |
| Slice 3: Adapter Service Shape Cleanup | Reviewing | `feat/slice-3-adapter-service-shape-cleanup` | `feat(slice-3): adapter service shape cleanup` | `.prj/reports/slices/architecture-boundary-cleanup/slice-3-adapter-service-shape-cleanup.md` | PR `#211`; metadata patching moved to service methods, TUI adapters split, and shared scope helpers renamed; `make verify` passed. |

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
- `4bd7889 docs(architecture): link dependency cleanup pr`

Validation:

- `go test ./internal/chcapp`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 2: Operation Context Boundary

Candidate tasks:

- [x] T2.1 Identify service operations that should receive operation-scoped
  context.
- [x] T2.2 Update service method signatures and call sites for selected
  operations.
- [x] T2.3 Remove obsolete `contextcheck` suppressions from CLI flow code.
- [x] T2.4 Preserve CLI and TUI behavior with focused tests or existing
  coverage.
- [x] T2.5 Create slice implementation report.
- [x] T2.6 Run validation.

Task commits:

- `138c8c4 refactor(generation): pass operation contexts explicitly`
- `db6d4e8 docs(architecture): report operation context boundary`
- `be2916b docs(architecture): link operation context pr`

Validation:

- `go test ./internal/chcapp`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 3: Adapter Service Shape Cleanup

Candidate tasks:

- [x] T3.1 Move metadata project patch/upsert workflow into service methods.
- [x] T3.2 Move metadata case patch/upsert workflow into service methods.
- [x] T3.3 Split TUI adapter wrappers into domain-focused files.
- [x] T3.4 Rename shared scope helpers from quiz-specific naming to neutral
  naming.
- [x] T3.5 Update documentation or parity matrix paths if implementation paths
  changed.
- [x] T3.6 Create slice implementation report.
- [x] T3.7 Run validation.

Task commits:

- `ed7138b refactor(metadata): move cli patching into service`
- `4efdbd5 refactor(cli): rename shared scope helpers`
- `916c5bd refactor(tui): split adapter domains`
- `f9069be docs(architecture): update adapter cleanup finding`
- `7b9d0ae docs(architecture): report adapter service cleanup`

Validation:

- `go test ./internal/chcapp`: pass.
- `go test ./...`: pass.
- `make verify`: pass.
