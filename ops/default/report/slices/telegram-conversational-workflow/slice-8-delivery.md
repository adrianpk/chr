# Slice 8: Live Journey and Closure

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `test/slice-8-telegram-delivery`

PR: pending

## Purpose

Certify the complete Telegram preparation and native development journey,
including bounded live QA, source self-development and exact-candidate closure.

## Delivered Behavior

The deterministic combined journey reaches confirmed native integration after
real fixture HTTP/model clarification, canonical dump/spec/plan generation,
native ticket preparation, SQLite reopen and exact work confirmation. Replaying
the admitted update preserves the completed invocation, attempts and publication.
This fixture uses an actual bare Git target and a simulated forge boundary.
No live journey or Slice 8 acceptance result is claimed.

The journey reproduced a preparation/admission gap: prepared descriptions stayed
in backlog and lacked the contract required by native execution. Preparation now
retains a complete independent-task contract and the project's configured
validation command while preserving backlog status. The shared preview displays
the exact readiness transition, contract, validation arguments and deadline.
The bound confirmation journals readiness and invocation creation together.
Stale input changes neither ticket readiness nor native ownership.

The dedicated private bot passes read-only credential, identity and webhook
checks. Its current project mapping points to the registered source checkout;
live validation must use the designated disposable QA target instead.

## Implementation Notes

PR #396 integration is verified at `c14f1aea640657e27c97bb64e40b46f30f698451`.
Its three tasks are closed with their controller validation receipt. Slice 8
uses its own worktree and leaves the primary checkout on dev. The existing
installed CLI predates the Telegram commands; live validation requires a fixed
receiver executable built from the integrated source.

The existing User Guide, receiver reference and pending changelog entry explain
the prepared-ticket transition. The native owner retains the original backlog
admission image separately from its accepted ready image. Original request-key
replay precedes changed-source guards, and retained continuation includes the
same readiness binding. Preparation remains independent of admission/execution.

## Contracts Added or Changed

- `PrepareIndependentWorkflowTask` seals a validated description without
  mutating its input or promoting backlog work.
- Optional `ready_ticket` binds a single independent ticket's observed revision,
  contract and validation to explicit admission. Existing admissions omit it.
- Shared invocation creation commits the ready image, invocation and its budget
  through the existing allocation/journal owner.

Deterministic fixture results cannot replace real bot/model reception, QA
integration, source self-development or the scheduled final-candidate gate.

## Files of Interest

- `docs/how-to/user-guide/index.md`: existing product User Guide.
- `docs/reference/telegram-receiver.md`: receiver operation and bounds.
- `internal/app/telegram_work_automatic_test.go`: bounded native integration fixture.
- `internal/app/telegram_journey_test.go`: complete asynchronous conversation fixture.
- `internal/core/telegram_journey_test.go`: preparation, guarded admission and replay.
- `internal/core/workflow_ticket_admission.go`: shared atomic readiness transition.
- `internal/app/service_authoring_work_contract.go`: observed description and project gate.

## Validation

Activation documentation passed `git diff --check`. The integrated Slice 7
binary's read-only `telegram status` and `telegram doctor --check-connection`
checks passed. They establish local bot readiness, not live execution or model
connectivity. Slice 8 focused checks and exact-HEAD controller validation remain
pending until the slice's remaining live work and final head are ready.

Task checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramJourney|TelegramAuthoringPrepared|TelegramAuthoringWork|TelegramWorkAdmission|TelegramWorkAutomatic|ServiceWorkflowPreview)' -count=1 -timeout=3m`: core 1.162 seconds; app 34.566 seconds.
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramJourney|TelegramAuthoringPrepared|TelegramAuthoringWork|TelegramWorkAdmission|TelegramWorkAutomatic|ServiceWorkflowPreview)' -count=1 -timeout=3m`: core 2.736 seconds; app 75.807 seconds.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./...`: zero issues.
- `scripts/check-go-func-signatures.sh`, `scripts/check-package-readmes.sh` and `git diff --check`: passed.

These are implementation-task checks. They do not constitute final clean-HEAD
controller validation or the scheduled exhaustive gate.

## Risks and Follow-ups

Prepare an isolated binding to an existing designated QA repository and a fixed
installed receiver. The authorized Telegram user must supply the live request
and any bound clarification/admission messages. Record actual source/base/result
and canonical integration receipt separately from fixtures. Scheduled exhaustive
certification and delivery-set closure remain pending the final integrated
candidate.
