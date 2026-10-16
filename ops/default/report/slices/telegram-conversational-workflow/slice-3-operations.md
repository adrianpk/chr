# Slice 3: Replay-safe Operations

Status: reviewing

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-3-telegram-operations`

PR: [#392](https://forge.adrianpk.com/adrianpk/chr/pulls/392)

## Purpose

Deliver guarded record operations, shared-owner replay receipts and durable
Telegram requests, actions and outbound results.

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

T3.3 completion: `76c4862df1b7dd68eb366d1395c2e1ec81ea0c69`.
T3.4 acknowledgement proofs: `c0385ee781452a7dfdad5aadeeb971e3b62cb60e`.

A prompt retains one random 128-bit action token, up to eight choices and at most
24 hours of validity. Buttons, exact replies and `/answer` claim the same
immutable guard and first answer request. Current peer/operator, registry root,
conversation revision, expiry and actual owner image are rechecked. The owner
call inherits the action deadline. Only confirmed results consume the action;
stale or denied outcomes supersede it. Both request results, consumption, input
acknowledgement and result outbox commit atomically. Definite send acknowledgement
retains the exact reply binding; stale or pruned action replies cannot become
capture text. Authorization-denied intake supersedes that peer's pending actions.

## Contracts Added or Changed

`WorkflowMutationBinding` and the workflow/case mutation result contracts
provide caller-neutral replay. `ReadRecords`, `PrepareRecordMutation` and
`ApplyPreparedRecordMutation` expose native ownership through the central
service. `TelegramOperationRequest` retains transport application state.
`TelegramActionOffer`, `TelegramActionClaim` and `TelegramActionResult` bind
prompt delivery, the first answer and confirmed owner facts. The existing HTTP
adapter sends optional retained inline choices with plain text.

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

Required final controller checks, in order:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramOperation' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramOperation' -count=1 -timeout=3m`
- `git diff --check`

The controller records these against the exact clean PR HEAD after every PR
reference is committed. The canonical PR validation list and controller receipt
carry that result; task checks below are separate implementation evidence.

Successful T3.3/T3.4 task checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramOperation|TelegramConversation|TelegramReceiver|CaseMetadata|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^TestTelegramOperation(Action|Uncertain)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramOperation(Action|Uncertain)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
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
and block a replacement registered root.

Action fixtures prove one answer across independent SQLite owners, original
binding reuse after reopening, atomic rollback when the outbox is full and
rejection of changed choices, wrong peers, expired prompts, changed operators,
revoked projects, replaced roots, revisions and guards. Real HTTP tests cover
reply/callback/explicit-answer parity, independent case edits, lost owner
acknowledgement, uncertain result delivery and callbacks arriving before prompt
send acknowledgement. Repeated prompts retain their token; consumed reply
copies cannot become capture or authorize a second effect. Work admission
remains absent.

A race-mode prompt fixture initially assumed retry ordering based on execution
speed. It now controls the retained retry timestamp, and focused normal/race
checks pass for the intended interleaving. These deterministic fixtures do not
establish live-bot or exhaustive scheduled certification.

## Risks and Follow-ups

None recorded.
