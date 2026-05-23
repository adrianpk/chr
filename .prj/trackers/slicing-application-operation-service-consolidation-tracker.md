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
| Slice 2: Workspace And Metadata Operations | Delivered | `feat/slice-2-workspace-metadata-operations` | `feat(slice-2): workspace metadata operations` | `.prj/reports/slices/application-operation-service-consolidation/slice-2-workspace-metadata-operations.md` | PR #157. |
| Slice 3: Artifact And Ticket Operations | Delivered | `feat/slice-3-artifact-ticket-operations` | `feat(slice-3): artifact ticket operations` | `.prj/reports/slices/application-operation-service-consolidation/slice-3-artifact-ticket-operations.md` | PR #158. |
| Slice 4: Work Control Operations | Delivered | `feat/slice-4-work-control-operations` | `feat(slice-4): work control operations` | `.prj/reports/slices/application-operation-service-consolidation/slice-4-work-control-operations.md` | PR #159. |
| Slice 5: Adapter Interface Cleanup | Delivered | `feat/slice-5-adapter-interface-cleanup` | `feat(slice-5): adapter interface cleanup` | `.prj/reports/slices/application-operation-service-consolidation/slice-5-adapter-interface-cleanup.md` | PR #160. |
| Slice 6: CLI Convergence Pass | Delivered | `feat/slice-6-cli-convergence-pass` | `feat(slice-6): cli convergence pass` | `.prj/reports/slices/application-operation-service-consolidation/slice-6-cli-convergence-pass.md` | PR #161. |
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

- [x] Move workspace graph refresh into service methods.
- [x] Move Project metadata save into service methods.
- [x] Move Case metadata save into service methods.
- [x] Map service results to TUI results outside the service.
- [x] Add focused tests for refresh and metadata save operations.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `9695605` adds metadata service operations.
- T2.2: `9695605` delegates TUI metadata wrappers to the service.
- T2.3: `9695605` adds service metadata tests.
- T2.4: `9695605` adds report and validation.

## Slice 3: Artifact And Ticket Operations

Candidate tasks:

- [x] Move default ticket creation into service methods.
- [x] Move authoring session creation into service methods.
- [x] Move selected-ticket comment addition into service methods.
- [x] Move selected-ticket status updates into service methods.
- [x] Centralize selected ticket lookup and ticket reload paths.
- [x] Preserve unsupported action behavior.
- [x] Add focused tests for artifact and ticket operations.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `54e6536` adds artifact operation service methods.
- T3.2: `54e6536` adds ticket operation service methods.
- T3.3: `54e6536` delegates TUI wrappers to the service and adds focused tests.
- T3.4: `54e6536` adds report and validation.

## Slice 4: Work Control Operations

Candidate tasks:

- [x] Move selected-ticket execution into service methods.
- [x] Move one-ticket ready-loop execution into service methods.
- [x] Move daemon stop requests into service methods.
- [x] Centralize observability reload after work actions.
- [x] Preserve execution failure handling.
- [x] Add focused tests for work-control operation outcomes where practical.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `adc2a05` adds selected-ticket execution service operation.
- T4.2: `adc2a05` adds ready-loop and daemon stop service operations.
- T4.3: `adc2a05` delegates TUI wrapper to the service and adds focused tests.
- T4.4: `adc2a05` adds report and validation.

## Slice 5: Adapter Interface Cleanup

Candidate tasks:

- [x] Reduce TUI executor wrappers to thin adapter mappings.
- [x] Remove obsolete TUI-named operation structs from `internal/chcapp`.
- [x] Keep adapter-neutral operation types in `internal/chcapp`.
- [x] Ensure operation service files do not import `internal/chctui`.
- [x] Organize service methods across capability files without separate entity services.
- [x] Add boundary tests or checks where practical.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `aaa261a` isolates TUI adapter wrappers.
- T5.2: `aaa261a` adds service import boundary test.
- T5.3: `aaa261a` adds report and validation.
- T5.4: `aaa261a` verifies final adapter cleanup.

## Slice 6: CLI Convergence Pass

Candidate tasks:

- [x] Identify overlapping CLI operations.
- [x] Migrate lowest-risk overlapping CLI commands to service calls.
- [x] Preserve current CLI output contracts.
- [x] Avoid broad CLI rewrites.
- [x] Add focused CLI tests for migrated commands.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `d00e592` adds generic ticket service operations.
- T6.2: `d00e592` routes ticket CLI commands through the service.
- T6.3: `d00e592` records focused CLI validation.
- T6.4: `d00e592` adds report and validation.

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
- 2026-05-23: `go test ./internal/chcapp ./internal/chctui` passed for Slice 2.
- 2026-05-23: `make verify` passed for Slice 2.
- 2026-05-23: `go test ./internal/chcapp ./internal/chctui` passed for Slice 3.
- 2026-05-23: `make verify` passed for Slice 3.
- 2026-05-23: `go test ./internal/chcapp ./internal/chctui` passed for Slice 4.
- 2026-05-23: `make verify` passed for Slice 4.
- 2026-05-23: `go test ./internal/chcapp ./internal/chctui` passed for Slice 5.
- 2026-05-23: `make verify` passed for Slice 5.
- 2026-05-23: `go test ./internal/chcapp -run 'TestTicket(New|Comment|Status)'` passed for Slice 6.
- 2026-05-23: `go test ./internal/chcapp ./internal/chctui` passed for Slice 6.
- 2026-05-23: `make verify` passed for Slice 6.
