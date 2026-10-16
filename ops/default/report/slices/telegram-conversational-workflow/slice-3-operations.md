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

The use-case metadata owner now provides the same immutable request/result
binding, serializes ordinary and bound saves through one exclusion lock and
recovers a two-image metadata/receipt journal. Bounded reads reject symlink
escapes. Its receipts retain relative references.

The central record API and deterministic receiver commands expose bounded
list/show/create/update for native tickets, proposed slice sets and pending
slices, and use cases. Updates require actual observed images/revisions.
Contract files, parent lifecycle and task membership remain owner-checked.
Display updates preserve requirements and omitted context; record commands do
not admit work or change execution/review/integration authority.

The receiver retains its normalized request and frozen target before preparing
serialized owner input. Application intent survives restart; original confirmed
facts, source acknowledgement and result outbox commit together after owner
confirmation. Ordered Unicode parts fit both limits; oversized output retains a
summary/reference. Send uncertainty remains independent of application replay.

## Implementation Notes

- T3.1 native-owner foundation: `f39c12e96c3d2228e8411d954f0a219b86f2d497`;
  case-owner completion: `1b0ec78267b3`.
- T3.2 central record API: `9ebe146da339`; live receiver:
  `5507e7041a3ddcdbd8021e9919c17f6b6ae555fc`.
- T3.3 retained request/outbox foundation:
  `e5de10f0eb5e1a44cd4cc787186ab10a5da60bb3`.

Shared bindings contain no bot, chat, topic, sender or callback metadata.
Native request reservations and original results remain under the existing
workflow namespace. Case receipts remain adjacent to the existing case owner.
Both reserve identity before effects and retain divergent pending images.

Migration 18 adds application requests and the action-binding schema. Requests
bind the configured operator, operation, typed arguments and frozen project/root/
namespace. Prepared input is checksum-bound. SQLite transactions end before
project-owner calls. Reopening after a lost acknowledgement reuses the original
owner input even if `OPS_USER` changes. A replaced registry root blocks
application rather than redirecting it.

Single-use action APIs, callback/text-answer parity and their complete crash
proof remain under T3.3–T3.4. Their schema alone does not establish delivery.

## Contracts Added or Changed

`WorkflowMutationBinding` and the workflow/case mutation result contracts
provide caller-neutral replay. `ReadRecords`, `PrepareRecordMutation` and
`ApplyPreparedRecordMutation` expose native ownership through the central
service. `TelegramOperationRequest` retains transport application state.

The [mutation receipt reference](../../../../../docs/reference/workflow-record-mutations.md)
and [Telegram record reference](../../../../../docs/reference/telegram-record-operations.md)
describe identity, guards, bounds, command syntax and recovery.

## Files of Interest

- `internal/core/workflow_mutation*.go`, `case_mutation*.go`: owner bindings,
  original receipts and journal recovery.
- `internal/app/service_record*.go`, `telegram_record*.go`: shared record API,
  strict deterministic commands and frozen-input dispatch.
- `internal/core/telegram_operation*.go`, `telegram_response_chunks.go`:
  retained application state, target guards and bounded output.
- `internal/core/telegram_operation_*_test.go` and
  `internal/app/telegram_operation_*_test.go`: real owner, reopened SQLite,
  concurrency and HTTP-to-owner crash fixtures.

## Validation

Task checks for T3.1/T3.2 and the retained transport foundation passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramOperation|TelegramReceiver|TelegramConversation|MaintenanceDBStatus|CaseMetadata)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramOperation|TelegramReceiver|TelegramConversation|MaintenanceDBStatus|CaseMetadata)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(TelegramOperation|TelegramReceiver|TelegramConversation)' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^Test(TelegramOperation|TelegramReceiver|TelegramConversation)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Owner fixtures inject loss after metadata, body, coordinator and receipt writes;
reopen project stores; preserve divergent journals; reject changed successful or
failed arguments; and prove concurrent/caller-neutral reuse. Case fixtures add
recovery through ordinary saves, stale hashes and escaped paths.

Transport fixtures reopen real SQLite with application intent, retain one
original result, roll back acknowledgement at outbox capacity, reject changed or
corrupt prepared input, order multi-part sends and exercise astral Unicode
bounds. Real HTTP fixtures create/update/read records, reject stale observations,
recover one ticket after lost owner acknowledgement despite a namespace change,
and block a replacement registered root. Work admission remains absent.

Final exact-HEAD controller validation is pending until the remaining Slice 3
work and canonical PR references are committed. Task fixtures do not establish
live bot or exhaustive scheduled certification.

## Risks and Follow-ups

T3.3 still requires single-use actions, exact action reply associations and
callback/text-answer request parity. T3.4 must complete action and uncertain-send
crash proof. No Slice 3 PR is ready for integration yet.
