# Slice 1: Dedicated Live Receiver

Status: reviewing

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-1-telegram-receiver`

PR: [#390](https://forge.adrianpk.com/adrianpk/chr/pulls/390)

## Purpose

Provide an authenticated, read-only dedicated bot with retained polling and
response delivery through the existing application service and user database.

## Delivered Behavior

- Private setup reads a masked terminal token or an environment reference,
  preserves differing configuration unless replacement is explicit, and does
  not start work or send messages. Local status preserves `--dev-root`.
- Connection doctor checks bot identity, webhook absence and receiver exclusion.
  Local diagnostics report registered project runtime files and configured
  model credential/executable presence without invoking a model or agent.
- Foreground serving polls bounded message/callback batches, authenticates the
  complete numeric peer before dispatch, and rejects unsupported or forwarded
  content. OS locks exclude another local receiver and setup replacement.
- `/help`, `/projects` and `/status` use a typed read registry. Project reads
  go through the central service and disclose only enabled allowed projects.
  Other input returns available commands or an unavailable-action diagnostic.
- Complete sparse batches and the next cursor commit atomically. Duplicate
  sources retain their receipt; a changed digest rolls back the whole batch.
  Denied bodies are discarded before persistence. Capacity prevents unsafe
  acknowledgement of input or unretained responses.
- Responses enter a durable outbox before sending. Poll and delivery attempts,
  deadlines, uncertainty and exhaustion survive reopening. Configuration
  changes block output rendered under a previous authorization digest.

## Implementation Notes

Primary implementation commit: `99f0868a6b616dd44dab95dc758718b91addd313`.
Follow-up commits record the PR reference, align database-status assertions
and normalize malformed UTF-8 in read output. Final merge evidence must
identify the exact clean PR HEAD after those commits.

Task commits: T1.1 `e38f90f8`, T1.2 `063a2a45`, T1.3 `d32b92a9`,
T1.4 `99f0868a`. Connection doctor followed the transport implementation;
foreground assembly and project readiness completed with T1.4 after storage
existed. The T1.4 commit also applies strict lint formatting to the new files.

Migration `20261016000016` adds receiver, inbox and initial read-response outbox
tables to the existing user database. No transaction spans an HTTP call.
Future migration identities and receiver versions are refused. The existing
notification and saved-update importer retain their separate contracts.

## Contracts Added or Changed

`TelegramReceiverConfig`, `TelegramReceiverAPI`, `TelegramReceiverStore` and
the central service's bounded registry/status reads expose the new capability.
The [receiver reference](../../../../../docs/reference/telegram-receiver.md)
documents private files, strict YAML grammar, exact identity rules, commands,
capacity limits, Unicode limits and retry behavior.

## Files of Interest

- `internal/app/telegram_serve.go`, `telegram_drain.go` and
  `telegram_operations.go`: lifecycle and typed read dispatch.
- `internal/app/service_receiver.go`: shared bounded reads and local readiness.
- `internal/core/telegram_receiver_http.go` and `telegram_receiver_update.go`:
  bounded transport, safe failures and authenticated normalization.
- `internal/core/telegram_receiver_inbox.go`, `telegram_receiver_outbox.go` and
  `telegram_receiver_store.go`: retained transactions and retry reservations.

## Validation

Successful task checks against the implementation worktree:

- `go test ./internal/core ./internal/app -run '^TestTelegramReceiver' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramReceiver(LockProcess|Shutdown|ReadJourney|Batch|Outbox|PollBudget)$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app` — zero issues.

Final merge evidence is recorded through `looper validate` after report and PR
references are committed and pushed. Its receipt supplies the immutable HEAD,
clean worktree and exact ordered command results:

1. `make verify`
2. `go test ./internal/core ./internal/app -run '^Test(TelegramReceiver|TelegramSender|RenderTelegramNotification|DecodeTelegramUpdates|RemoteCommandRecordFromTelegramUpdate)' -count=1 -timeout=3m`
3. `go test -race ./internal/core ./internal/app -run '^Test(TelegramReceiver|TelegramSender|RenderTelegramNotification|DecodeTelegramUpdates|RemoteCommandRecordFromTelegramUpdate)' -count=1 -timeout=3m`
4. `git diff --check`

`TestTelegramReceiverConfig`, `ConfigErrors`, `Token`, `Setup` and `Doctor`
cover TG-02. `Authorization`, `Callback`, `HTTPFailures`, `Webhook`,
`LockProcess`, `Identity` and `Revocation` cover TG-03 and transport ownership.
`Batch`, `IntakeCapacity`, `OutboxCapacity`, `Outbox`, `PollBudget` and
`Migration` cover TG-19, rollback, sparse replay, reopening and retained retries.
`ReadJourney` retains an accepted historical handoff while proving that reads
and unsupported work commands create zero executions. Existing sender/import
regressions cover TG-29. `OutputBounds` and `Shutdown` cover Unicode and owned
lifecycle boundaries.

All evidence here uses deterministic local HTTP and real temporary SQLite
stores. An initial shutdown fixture timed out during server cleanup; explicit
fixture release corrected it before the successful normal and race checks.
`make verify` on initial controller candidate
`cc106fde867cf861b1e19c2627c86003bb5609c3` failed in the two existing
`TestMaintenanceDBStatus` cases that still expected 15 migrations. They now
expect 16 and explicitly check the new receiver migration.
`go test ./internal/app -run '^TestMaintenanceDBStatus' -count=1 -timeout=3m`
passed after that correction. The final controller receipt supersedes the
failed candidate's evidence.
The malformed UTF-8 regression reproduced a slice-bounds panic in read-output
truncation. Normalizing invalid byte sequences before applying Unicode limits
corrected it; `go test ./internal/core -run '^TestTelegramReceiverOutputBounds$' -count=1 -timeout=3m` passed afterward.
No live bot/model demonstration or scheduled exhaustive certification is claimed.

## Risks and Follow-ups

An uncertain send can repeat a read notice. Automatic restart preserves blocked
retry state. Explicit local retry cycles, long-outage cursor recovery and
retirement/compaction belong to Slice 7. Conversational and mutating operations
remain assigned to their approved successor slices. Canonical integration is
pending the external controller; no successor is activated by this worker.
