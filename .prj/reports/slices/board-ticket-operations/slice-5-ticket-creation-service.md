# Slice 5: Ticket Creation Service

Status: delivered
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-5-ticket-creation-service`
PR: #217

## Purpose

Harden chr-created tickets so manual and generated creation paths produce
Board-ready metadata and immediately refresh the SQLite ticket index.

## Delivered Behavior

- `CreateTicketRequest` and the CLI `ticket new` command accept `synopsis`.
- chr-created tickets require non-empty title and synopsis before key
  allocation.
- Tickets created without an explicit status default to `backlog`.
- Ticket key allocation uses the SQLite `ticket_key_sequences` table with the
  filesystem scan as a seed.
- Existing filesystem tickets can advance the SQLite sequence if the filesystem
  seed is higher than the stored next number.
- Created tickets are reconciled into the ticket index immediately after the
  Markdown file is written.
- TUI-created tickets use a default synopsis and show it in the confirmation
  details.
- Plan tracking ticket creation passes generated ticket summaries as ticket
  synopsis and uses `ready_for_dev`.

## Implementation Notes

- The SQLite key sequence is scoped by project ID and ticket key prefix.
- The sequence path still checks existing Markdown ticket files before reserving
  a number so external files do not force an immediate duplicate.
- The service validates title and synopsis before reserving a key to avoid
  consuming sequence numbers for invalid requests.
- This slice does not add generated source metadata; Slice 6 owns source refs
  and slice metadata for generated plan tracking tickets.

## Contracts Added or Changed

- `CreateTicketRequest.Synopsis`
- `CreateTicketOptions.Synopsis`
- `WithTicketRepositoryNextKeyNumber`
- `NextTicketKeyNumber`
- `TicketKeySequenceRepository`
- `ReserveNextTicketKeyOptions`

## Files of Interest

- `internal/chr/ticket_repository.go`
- `internal/chr/ticket_key_sequence.go`
- `internal/chrapp/service_artifact_ticket.go`
- `internal/chrapp/ticket.go`
- `internal/chrapp/service_plan.go`
- `internal/chrtui/artifact_actions.go`

## Validation

- `go test ./internal/chr ./internal/chrapp ./internal/chrtui`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- CLI documentation is still updated in the final Board docs validation slice.
- Plan tracking tickets still need source metadata and slice metadata in Slice
  6.
