# Slice 2: Agent Work Service Facade

Status: reviewing

## Purpose

Add a service-level agent-work facade that names the product intent without
duplicating ticket execution or loop execution internals.

## Delivered

- Added `AgentWorkResult` for service callers that need execution, loop, queue,
  and eligible-count details.
- Added `StartAgentWorkTicket` as the single-ticket agent-work facade over the
  existing ticket execution path.
- Added `StartReadyAgentWorkLoop` for the existing one-ticket ready-loop TUI
  behavior.
- Added `StartAllReadyAgentWork` for run-all ready tickets using the existing
  loop runner without a max-iteration cap.
- Added `PreviewAllReadyAgentWork` for queue preview without execution.
- Kept existing `StartSelectedTicket` and `StartReadyTicketLoop` methods as
  compatibility wrappers.
- Added service tests proving single-ticket work creates execution records
  without loop records, run-all creates loop and execution records, and preview
  does not execute.

## Validation

- `go test ./internal/chcapp`
- `make verify`
- `git diff --check`
