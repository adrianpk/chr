# Slice 3: Replay-safe Operations

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-3-telegram-operations`

PR: pending

## Purpose

Provide shared-owner mutation receipts and retained transport requests before
exposing canonical record mutations through the live receiver.

## Delivered Behavior

The shared workflow record owner now accepts caller-neutral opaque request and
operator bindings. It normalizes and binds the complete scoped arguments before
semantic preparation, retaining identity even after a failed attempt. Changed
arguments conflict. Existing record lifecycle and observed-image/revision checks
remain authoritative.

Native ticket and plan mutations reserve their artifact identities under the
existing allocator. Their record images and confirmed result receipt share the
existing workflow journal and repository exclusion. Lost acknowledgement
recovers the same journal and result; a divergent image leaves the pending
journal intact. Concurrent independent owners applying one request produce one
artifact. Replay returns the original result after later independent edits.

Live record dispatch, other record-owner bindings, durable SQLite application
requests, single-use actions and complete outbound chunk delivery remain in
progress under T3.2–T3.4.

## Implementation Notes

T3.1 workflow-owner foundation commit: `f39c12e96c3d2228e8411d954f0a219b86f2d497`. Other owner bindings
complete with the recorded T3.2 scope before T3.1 is closed.

The shared binding contains no bot, chat, topic, sender or callback metadata.
`run/operation-bindings/` retains accepted identity reservations;
`run/operation-receipts/` retains checksum-bound original results. Limits are
64 KiB normalized arguments and 4,096 reservations per namespace. The existing
1-MiB workflow journal bound also applies. Pending allocation ownership is
released only when its journal completes or no pending journal exists.

The central operation service resolves the requested namespace explicitly and
calls the same domain owner. It adds no model, agent, forge or transport call.

## Contracts Added or Changed

`WorkflowMutationBinding`, `WorkflowMutationRequest` and
`WorkflowMutationResult` add the shared native record contract.
`ApplyWorkflowRecordMutation` exposes it through the central service.
The [record mutation reference](../../../../../docs/reference/workflow-record-mutations.md)
describes replay, guards and capacity.

## Files of Interest

- `internal/core/workflow_mutation.go` and `workflow_mutation_binding.go`:
  normalized bindings, original receipts and existing-owner journal integration.
- `internal/app/service_workflow_mutation.go`: caller-neutral service boundary.
- `internal/core/telegram_operation_mutation_test.go` and
  `internal/app/telegram_operation_service_test.go`: real owner reopening,
  fault injection, concurrency and cross-adapter reuse.

## Validation

T3.1 foundation task checks passed:

- `go test ./internal/core ./internal/app -run '^TestTelegramOperation' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramOperation' -count=1 -timeout=3m`

The fixtures inject loss after metadata, body, coordinator and receipt writes;
reopen the real project store; preserve divergent journal images; reject changed
arguments after both success and failed preparation; and apply one request
through independent owners. They also check argument/identity bounds, receipt
tampering and central-service reuse. Focused strict lint reported zero issues:

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `git diff --check`

The final ordered controller checks remain pending until every Slice 3 task and
canonical PR reference is committed. These task checks establish local owner
behavior, not live bot or exhaustive scheduled certification.

## Risks and Follow-ups

T3.2 must complete record surfaces and other shared-owner bindings. T3.3 must
complete retained application requests, single-use actions and outbound delivery.
T3.4 must connect the owner proofs to SQLite and send-acknowledgement crash
boundaries. No Slice 3 PR is ready for integration yet.
