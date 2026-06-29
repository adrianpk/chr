# Slice 6: Generated Ticket Synopsis

Status: delivered
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-6-generated-ticket-synopsis`
PR: #218

## Purpose

Ensure plan tracking materialization creates self-describing Board tickets with
explicit synopsis and source metadata.

## Delivered Behavior

- Plan tracking model responses now require a ticket-level `synopsis`.
- Generated plan tracking ticket structs carry the synopsis through validation
  and materialization.
- Materialized tickets persist the generated synopsis in ticket frontmatter.
- Generated tickets include `source_kind: "plan_tracking"`.
- Generated tickets include `source_ref` pointing to the source plan path.
- Generated tickets include the slice identifier in ticket frontmatter.
- Generated plan tracking tickets continue to include plan and tracking
  references in the Markdown body.
- Service tests assert that reloaded indexed tickets include synopsis and source
  metadata.

## Implementation Notes

- `summary` remains the longer ticket body summary used in the Markdown
  description.
- `synopsis` is the short Board-facing sentence used for list/detail surfaces.
- `source_ref` points to the source plan because the tracking file is a
  materialized execution index and is already referenced in the body.
- This slice does not change Board rendering. Later Board query and UI slices
  consume the enriched metadata.

## Contracts Added or Changed

- `QuizModelPlanTrackingTicketPayload.Synopsis`
- `GeneratedPlanTrackingTicket.Synopsis`
- Plan tracking JSON schema now requires `synopsis` per ticket.
- `CreateTicketRequest` and `CreateTicketOptions` carry source metadata.

## Files of Interest

- `internal/chr/quiz_model_payload.go`
- `internal/chr/quiz_openai_model.go`
- `internal/chr/plan_tracking.go`
- `internal/chr/ticket_repository.go`
- `internal/chrapp/service_artifact_ticket.go`
- `internal/chrapp/service_plan.go`

## Validation

- `go test ./internal/chr ./internal/chrapp ./internal/chrtui`
- `go test ./internal/chrapp -run TestServiceMaterializePlanTrackingCreatesTrackingAndTickets`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- Existing generated plan tracking outputs created before this contract change
  do not include ticket synopsis in their original model payloads.
- Board query and TUI rendering still need later slices to use the generated
  synopsis and source metadata directly.
