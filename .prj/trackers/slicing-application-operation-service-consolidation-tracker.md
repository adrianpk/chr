# Application Operation Service Consolidation Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`

## Slicing Checklist

- [x] Capture operation service consolidation direction.
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
| Slice 1: Operation Service Boundary | Delivered | `feat/slice-1-operation-service-boundary` | `feat(slice-1): operation service boundary` | `.prj/reports/slices/application-operation-service-consolidation/slice-1-operation-service-boundary.md` | PR #156. |
| Slice 2: Workspace And Metadata Operations | Planned | `feat/slice-2-workspace-metadata-operations` | `feat(slice-2): workspace metadata operations` | `.prj/reports/slices/application-operation-service-consolidation/slice-2-workspace-metadata-operations.md` | Move refresh, Project, and Case operations. |
| Slice 3: Artifact And Ticket Operations | Planned | `feat/slice-3-artifact-ticket-operations` | `feat(slice-3): artifact ticket operations` | `.prj/reports/slices/application-operation-service-consolidation/slice-3-artifact-ticket-operations.md` | Move artifact creation and ticket mutations. |
| Slice 4: Work Control Operations | Planned | `feat/slice-4-work-control-operations` | `feat(slice-4): work control operations` | `.prj/reports/slices/application-operation-service-consolidation/slice-4-work-control-operations.md` | Move execution and daemon controls. |
| Slice 5: Adapter Interface Cleanup | Planned | `feat/slice-5-adapter-interface-cleanup` | `feat(slice-5): adapter interface cleanup` | `.prj/reports/slices/application-operation-service-consolidation/slice-5-adapter-interface-cleanup.md` | Reduce adapter-specific operation wrappers. |
| Slice 6: CLI Convergence Pass | Planned | `feat/slice-6-cli-convergence-pass` | `feat(slice-6): cli convergence pass` | `.prj/reports/slices/application-operation-service-consolidation/slice-6-cli-convergence-pass.md` | Reuse service from overlapping CLI commands. |
| Slice 7: Architecture Documentation | Planned | `docs/slice-7-architecture-documentation` | `feat(slice-7): architecture documentation` | `.prj/reports/slices/application-operation-service-consolidation/slice-7-architecture-documentation.md` | Document operation boundary. |

## Slice 1: Operation Service Boundary

Candidate tasks:

- [x] Add central operation service type.
- [x] Centralize shared service dependencies.
- [x] Add adapter-neutral request/result types for the first migrated boundary.
- [x] Delegate at least one low-risk TUI executor path through the service.
- [x] Keep TUI options and behavior stable.
- [x] Add focused tests where practical.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `aa65793` adds service type and dependency boundary.
- T1.2: `aa65793` adds workspace refresh service method.
- T1.3: `aa65793` delegates the TUI workspace refresh wrapper to the service.
- T1.4: `aa65793` adds tests, report, and validation.

## Slice 2: Workspace And Metadata Operations

Candidate tasks:

- [ ] Move workspace graph refresh into service methods.
- [ ] Move Project metadata save into service methods.
- [ ] Move Case metadata save into service methods.
- [ ] Map service results to TUI results outside the service.
- [ ] Add focused tests for refresh and metadata save operations.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1: pending.
- T2.2: pending.
- T2.3: pending.
- T2.4: pending.

## Slice 3: Artifact And Ticket Operations

Candidate tasks:

- [ ] Move default ticket creation into service methods.
- [ ] Move authoring session creation into service methods.
- [ ] Move selected-ticket comment addition into service methods.
- [ ] Move selected-ticket status updates into service methods.
- [ ] Centralize selected ticket lookup and ticket reload paths.
- [ ] Preserve unsupported action behavior.
- [ ] Add focused tests for artifact and ticket operations.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending.
- T3.2: pending.
- T3.3: pending.
- T3.4: pending.

## Slice 4: Work Control Operations

Candidate tasks:

- [ ] Move selected-ticket execution into service methods.
- [ ] Move one-ticket ready-loop execution into service methods.
- [ ] Move daemon stop requests into service methods.
- [ ] Centralize observability reload after work actions.
- [ ] Preserve execution failure handling.
- [ ] Add focused tests for work-control operation outcomes where practical.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending.
- T4.2: pending.
- T4.3: pending.
- T4.4: pending.

## Slice 5: Adapter Interface Cleanup

Candidate tasks:

- [ ] Reduce TUI executor wrappers to thin adapter mappings.
- [ ] Remove obsolete TUI-named operation structs from `internal/chcapp`.
- [ ] Keep adapter-neutral operation types in `internal/chcapp`.
- [ ] Ensure operation service files do not import `internal/chctui`.
- [ ] Organize service methods across capability files without separate entity services.
- [ ] Add boundary tests or checks where practical.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending.
- T5.2: pending.
- T5.3: pending.
- T5.4: pending.

## Slice 6: CLI Convergence Pass

Candidate tasks:

- [ ] Identify overlapping CLI operations.
- [ ] Migrate lowest-risk overlapping CLI commands to service calls.
- [ ] Preserve current CLI output contracts.
- [ ] Avoid broad CLI rewrites.
- [ ] Add focused CLI tests for migrated commands.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending.
- T6.2: pending.
- T6.3: pending.
- T6.4: pending.

## Slice 7: Architecture Documentation

Candidate tasks:

- [ ] Update workspace architecture documentation with the service boundary.
- [ ] Document that adapters do not implement use cases.
- [ ] Document that one central service may be split across files.
- [ ] Document when narrow interfaces are appropriate.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending.
- T7.2: pending.
- T7.3: pending.

## Open Decisions

- Whether the central type is exported as `Service` immediately or remains an
  unexported `service` until another package needs it directly.
- Whether service request/result types should live in one file or next to their
  capability methods.
- How much CLI convergence belongs in this slicing versus a follow-up tuning
  slice.

## Validation Log

- 2026-05-23: `go test ./internal/chcapp ./internal/chctui` passed for Slice 1.
- 2026-05-23: `make verify` passed for Slice 1.
