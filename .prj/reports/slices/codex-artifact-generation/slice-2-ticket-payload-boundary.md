# Slice 2: Ticket Payload Boundary

Status: reviewing
Slicing: Codex Artifact Generation
Plan: `.prj/plans/slicing-codex-artifact-generation-plan.md`
Tracker: `.prj/trackers/slicing-codex-artifact-generation-tracker.md`
Branch: `feat/slice-2-ticket-payload-boundary`
PR: pending

## Purpose

Harden the boundary between generated plan tracking payloads and repository
ticket creation so generated data is validated before any ticket file is
created.

## Delivered Behavior

- Added model-content validation that rejects generated ticket `key` and `path`
  values before materialization.
- Added generated ticket validation for required `summary` content.
- Added a service-level preflight that validates the complete ticket payload
  batch before creating the first ticket.
- Added coverage proving invalid later payloads do not leave partial ticket
  files behind.
- Added coverage proving Go owns generated ticket keys and paths.

## Implementation Notes

The Codex and model adapters still synthesize ticket content only. They do not
own ticket keys or ticket paths. The service validates all generated ticket
payloads before calling `CreateTicket`, and `CreateTicket` remains responsible
for key reservation, path derivation, rendering, file creation, and index
refresh.

The model-content validation now rejects non-empty `GeneratedPlanTrackingTicket`
`Key` and `Path` fields. Those fields are populated only after successful Go
materialization.

## Contracts Added or Changed

- `ValidateGeneratedPlanTrackingModelContent` now rejects generated ticket keys
  and paths before materialization.
- Plan tracking ticket payload creation now preflights the full batch before
  writing any ticket files.

## Files of Interest

- `internal/chr/plan_tracking.go`
- `internal/chrapp/service_plan.go`
- `internal/chrapp/service_plan_test.go`

## Validation

- `gofmt -w internal/chr/plan_tracking.go internal/chrapp/service_plan.go internal/chrapp/service_plan_test.go`
- `go test ./internal/chr ./internal/chrapp`
- `make verify`

## Risks and Follow-ups

- Slice 3 should decide the default routing policy for dump, spec, plan, and
  tracking generation.
