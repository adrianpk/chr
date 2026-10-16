# Slice 7: Recovery and User Operation

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-7-telegram-recovery`

PR: pending

## Purpose

Recover retained Telegram work and operate the receiver across restart,
shutdown, retry exhaustion and retention pressure.

## Delivered Behavior

T7.1 (`0c2fee3f1ad68bdba578bae9cd75605057d53b05`) gives intake, application calls,
native work and delivery one 30-second shutdown grace. SIGTERM cancels only the
receiver's owned calls. If a call does not quiesce, shutdown reports an explicit
failure and keeps its database and configuration/bot exclusions until that call
returns or the process exits. Cancellation cannot hide a checkpoint failure.

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
T7.2 is ready; T7.3 remains pending in the approved order.

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

Final slice validation remains pending T7.2, T7.3 and the PR-reference commits.
It will use the controller receipt for the exact clean PR HEAD with the frozen
ordered commands.

## Risks and Follow-ups

T7.2 retry/retention and T7.3 operational tooling remain unfinished. The slice
report remains drafting and no PR readiness is claimed. Live journey
certification belongs to the next recorded slice.
