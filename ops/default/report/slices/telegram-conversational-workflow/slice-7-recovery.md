# Slice 7: Recovery and User Operation

Status: delivered

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-7-telegram-recovery`

PR: [#396](https://forge.adrianpk.com/adrianpk/chr/pulls/396)

## Purpose

Recover retained Telegram work and operate the receiver across restart,
shutdown, retry exhaustion and retention pressure.

## Delivered Behavior

T7.1 (`0c2fee3f1ad68bdba578bae9cd75605057d53b05`) gives intake, application calls,
native work and delivery one 30-second shutdown grace. SIGTERM cancels only the
receiver's owned calls. If a call does not quiesce, shutdown reports an explicit
failure and keeps its database and configuration/bot exclusions until that call
returns or the process exits. Cancellation cannot hide a checkpoint failure; `8fd3713efaf1` also preserves it
through application completion and canceled SQL transaction cleanup.

Receiver startup recovers an accepted authoring request from its canonical
receipt, preserving a paused invocation and a human review wait in other
conversations. Actual dead native and model processes are fenced through the
existing shared owners without resetting attempts or overriding operator stops.
Another adapter's live worker and validation retain their original ownership,
policy and capacity reservations through receiver startup and shutdown.

A fixture receiver receives real SIGTERM after its installed executable is
atomically replaced. Its mapped running image remains fixed, retained
conversation state survives, and normal shutdown releases receiver exclusion.

## Implementation Notes

The receiver starts four bounded hosts under one cancellation owner. The
shutdown result retains a quiescence signal; resource release waits for that
signal after a deadline failure. A buffered completion channel lets late hosts
finish without a second consumer or premature database/lock closure.

Startup uses pending source/request identities, native receipts and OS-owned
model exclusions already delivered by earlier slices. There is no timestamp
based transport-claim takeover and no reset of native authority or budgets.
T7.2 (`b2647940ccdd`) retains poll, delivery and model attempts across reopened
stores. An explicit offline transport retry records the local actor/time and a
new cycle without resetting lifetime counters, source interpretation attempts
or domain work. A poll after more than 24 hours opens one recovery epoch with
no offset, preserving the retained inbox and reporting possible remote loss.

Seven-day terminal compaction clears only resolved bodies and delivered replies;
canonical request receipts, focused conversations, unresolved prompts and unsent
results remain. Exact retired IDs merge only adjacent observed identities, with
at most 1,024 intervals. Intake observes the 64-MiB logical payload bound,
10,000 receipt-only sources plus 1,024 pending sources, and 1,024 unresolved
outbound entries. Capacity pressure retains existing authority and refuses new
input without acknowledging it.

T7.3 (`5546d2ab4b7726951ba885c8932ca88bdc3ea1b7`) installs an optional private
user service for the current executable without enabling or starting it. The
unit preserves user environment paths and reads private token/model assignments.
Correction-required configuration, credentials, schema, poll exhaustion and
ownership failures use exit 78; recoverable process/transport failures remain
restartable. Explicit retry acquires receiver exclusion before opening or
migrating its database and derives the actor from the OS user.

Read-only status reports installed build, cursor/gap/retry metadata and bounded
queue details without exposing message bodies. Doctor distinguishes changed bot
identity, webhooks and local exclusion, reports missing/disabled/unready projects
and the selected providers' local prerequisites, and labels model connectivity
and remote polling ownership unchecked. A real getUpdates conflict is retained
and blocks subsequent polls without takeover. Installed help and the receiver
reference document service operation, retry selection, private configuration
and retention bounds.

## Contracts Added or Changed

A shutdown that cannot prove quiescence is a failure, with local exclusions
retained for the remaining calls. The running executable is never replaced by
source changes; installing a candidate remains an explicit operation.

## Files of Interest

- `internal/app/telegram_lifecycle.go`: bounded host shutdown and resource lifetime.
- `internal/app/telegram_serve.go`: receiver assembly and ownership.
- `internal/app/telegram_recovery_startup_test.go`: lost-ack startup with pause/review waits.
- `internal/app/telegram_recovery_signal_test.go`: real SIGTERM and fixed executable identity.
- `internal/app/telegram_recovery_validation_test.go`: independent live host validation.
- `internal/core/telegram_recovery_owner_test.go`: real model-owner death and retained attempts.
- `internal/core/telegram_retention.go`: bounded terminal compaction and recovery epochs.
- `internal/core/telegram_transport_retry.go`: exclusion before migration and attributed cycles.
- `internal/app/telegram_service.go`: optional inactive user-service installation.
- `internal/app/telegram_status.go`: sanitized bounded retained-state diagnostics.
- `docs/reference/telegram-receiver.md`: private setup, service operation and recovery.

## Validation

T7.1 focused normal lifecycle checks passed in core 0.065 seconds and app
10.913 seconds. The expanded race selection passed in core 5.967 seconds and
app 68.276 seconds. After the final wrapped-error guard, its focused race check
passed in 2.067 seconds; the affected normal receiver/model/worker selection
passed in 2.167 seconds. These are task-level checks, not final slice evidence.

```sh
go test ./internal/core ./internal/app -run '^TestTelegramRecovery(ModelOwner|DeadWork|Grace|HostFailure|Startup|OwnedStop|Signal|Validation)$' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^TestTelegram(Recovery(ModelOwner|DeadWork|Grace|HostFailure|Startup|OwnedStop|Signal|Validation)|ReceiverShutdown|WorkRecovery|WorkResponsive|ControlModelLoad)$' -count=1 -timeout=3m
go test -race ./internal/app -run '^TestTelegramRecovery(Grace|HostFailure)$' -count=1 -timeout=3m
go test ./internal/app -run '^TestTelegram(Recovery(Grace|HostFailure)|ReceiverShutdown|WorkResponsive|ControlModelLoad)$' -count=1 -timeout=3m
```

Strict focused lint, function signatures, package README checks and whitespace
validation passed. Initial strict lint found helper complexity and an unchecked
ready-pipe read; both were corrected before the successful runs.

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app
scripts/check-go-func-signatures.sh
scripts/check-package-readmes.sh
git diff --check
```

T7.2 focused normal checks passed in core 1.752 seconds and app 0.121 seconds.
The expanded race selection passed in core 136.988 seconds and app 60.113
seconds. Fixtures reopen exhausted poll/delivery/model budgets; exercise actual
64-MiB payload and receipt/pending/outbox boundaries; retain sparse IDs and
unresolved authority; and prove additive migration rollback, unknown-schema
refusal and recovery polls without destructive offsets. Strict focused lint
reported zero issues; repository style and whitespace checks passed.

```sh
go test ./internal/core ./internal/app -run '^Test(TelegramRecovery(Retry|Models|Sparse|Intervals|Compaction|Quota|Payload|PollEpoch|Schema|Malformed|Pending)|MaintenanceDBStatusReports)' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^Test(TelegramRecovery(Retry|Models|Sparse|Intervals|Compaction|Quota|Payload|PollEpoch|Schema|Malformed|Pending|OwnedStop|Validation|Startup)|TelegramReceiver(Batch|IntakeCapacity|OutboxCapacity|PollBudget|Migration|Identity)|MaintenanceDBStatusReports)' -count=1 -timeout=3m
```

T7.2 activity receipt: `01M53CYHZ9NHH8F69ZJ9AJ2V48`.

T7.3 focused normal checks passed in core 0.434 seconds and app 0.526 seconds;
the expanded operator race selection passed in core 6.731 seconds and app 4.606
seconds. Fixtures verify the real CLI/store boundary, private token replacement
preserving model assignments, malformed input refusal, inactive service
installation/replacement, systemd unit syntax, HTTP identity/webhook/remote
conflicts and live local exclusion. Strict lint reported zero issues after
splitting helpers to meet repository complexity bounds; style and whitespace
checks passed.

```sh
go test ./internal/core ./internal/app -run '^Test(TelegramRecovery(Environment|Offline|Operator|Service|Help|Doctor|Failures|Providers|Remote|Retry)|TelegramReceiver(Setup|Doctor|Config|Token|Identity))' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^Test(TelegramRecovery(Environment|Offline|Operator|Service|Help|Doctor|Failures|Providers|Remote|Retry|ConfigExit)|TelegramReceiver(Setup|Doctor|Config|Token|Identity))' -count=1 -timeout=3m
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./cmd/chr
```

T7.3 activity receipt: `01M53DJPBSCRP7DH6KQ237TEK2`.

The additional reply regression first failed during receiver shutdown with
`sql: transaction has already been committed or rolled back`. Its focused
reproduction (`go test -race ./internal/app -run '^TestTelegramControlReplyArrival$' -count=3 -timeout=3m`)
also failed. Real canceled-transaction action/commit fixtures reproduced the
cause before the correction. Only a bare automatic-rollback error under this
transaction owner's canceled context becomes cancellation; uncanceled errors
and independent joined/wrapped checkpoint failures remain visible. Application
completion now removes expected deferral causes individually rather than
suppressing an unrelated joined error.

The corrected normal cancellation checks passed in core 0.267 seconds and app
0.266 seconds. The combined cancellation, Unicode splitting, peer/topic/replaced
reply and uncertain-prompt race checks passed in core 18.333 seconds and app
7.785 seconds. Strict lint and repository style/whitespace checks passed.
This completes deterministic TG-25 coverage with the earlier conversation and
operation fixtures. Correction activity receipt: `01M53DZ8V64JXF4GEEWBN74T81`.

```sh
go test ./internal/core ./internal/app -run '^TestTelegramRecovery(CanceledTx|ApplicationError|HostFailure|Grace)$' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^TestTelegram(Recovery(CanceledTx|ApplicationError|HostFailure|Grace)|OperationUnicodeParts|ConversationReply(Peers|Retention|Guards)|OperationUncertainPrompt|ControlReply(Binding|Exhaustion|Priority|Arrival))$' -count=1 -timeout=3m
```

Controller receipt `JR526N5RLBBVIBT62W5YOZMNB6` passed all four ordered checks
at clean HEAD `c14f1aea640657e27c97bb64e40b46f30f698451`. The canonical
forge verifies PR #396 merged that same commit into dev at 2026-10-16T22:44:10Z.
Acta merge event: `01M53E66GA9C9VH1CKRKYRDM5B`. Normal core/app checks passed
in 2.719/21.135 seconds; race checks passed in 56.358/57.946 seconds.
The frozen ordered sequence was:

```sh
make verify
go test ./internal/core ./internal/app -run '^TestTelegramRecovery' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^TestTelegramRecovery' -count=1 -timeout=3m
git diff --check
```

Verified integration, audited 2026-10-19: [PR #396](https://forge.adrianpk.com/adrianpk/chr/pulls/396)
merged into `dev` at `c14f1aea640657e27c97bb64e40b46f30f698451` on
2026-10-16T22:44:10Z. Controller validation `JR526N5RLBBVIBT62W5YOZMNB6`
records the four commands above in order on that exact clean HEAD; all passed.
The focused normal run passed in core 2.719 seconds and app 21.135 seconds;
the focused race run passed in core 56.358 seconds and app 57.946 seconds.

## Risks and Follow-ups

Deterministic fixture evidence does not establish a live bot/model journey or
scheduled exhaustive certification. Both remain assigned to the next recorded
slice. No owned fixture service remains installed or active outside temporary
test environments.
