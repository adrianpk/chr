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

The dedicated private bot passes credential, identity and webhook checks. An
isolated user environment now permits only two registered disposable QA clones,
with separate target refs in the existing designated QA repository.

## Implementation Notes

PR #396 integration is verified at `c14f1aea640657e27c97bb64e40b46f30f698451`.
Its three tasks are closed with their controller validation receipt. Slice 8
uses its own worktree and leaves the primary checkout on dev. The existing
installed global CLI predates the Telegram commands. The isolated receiver is
fixed at implementation commit `b72a6955cf7c038142c23b7142215b5ed83d58ec`, built
with Go 1.27.1 and `vcs.modified=false`. Its SHA-256 is
`c888ebbf6481d7bd6c7a44fbb09d7960eb1c3a93a0fc1ce80624e3343c4984a5`.

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
connectivity. The configured final Slice 8 command list and exact-HEAD controller
validation remain pending until the remaining live work and final head are ready.

Task checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramJourney|TelegramAuthoringPrepared|TelegramAuthoringWork|TelegramWorkAdmission|TelegramWorkAutomatic|ServiceWorkflowPreview)' -count=1 -timeout=3m`: core 1.162 seconds; app 34.566 seconds.
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramJourney|TelegramAuthoringPrepared|TelegramAuthoringWork|TelegramWorkAdmission|TelegramWorkAutomatic|ServiceWorkflowPreview)' -count=1 -timeout=3m`: core 2.736 seconds; app 75.807 seconds.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./...`: zero issues.
- `scripts/check-go-func-signatures.sh`, `scripts/check-package-readmes.sh` and `git diff --check`: passed.

These are implementation-task checks. They do not constitute final clean-HEAD
controller validation or the scheduled exhaustive gate.

## Live QA Setup

Owned playground:
`/home/adrian/Projects/playground/chr/20261017-0110-telegram-journey-qa`.
Its `fixture.json` records the executable, project registrations and exact
bootstrap targets; `README.md` contains the bounded foreground command. Private
configuration and retained data use its isolated `receiver/` development root.
The original user configuration and primary dev checkout are preserved.

| Registered project | Project identity | Isolated target | Bootstrap base |
| --- | --- | --- | --- |
| `telegram-journey-qa` | `project-0b671359780bf6ba` | `chr/qa/telegram-journey-20261017-0110-journey` | `96a3addef94da397ea4ddcc3b3ceeb19a09c0a8d` |
| `telegram-source-qa` | `project-51b74a16735269d1` | `chr/qa/telegram-journey-20261017-0110-source` | `ac5f3f5087572a11b72a328c08d678727abc2a49` |

Both targets belong to `adrianpk/chr-integration-qa` on the canonical Forgejo.
They are setup branches, not completed bot-development results. The source QA
bootstrap changes only its fixture policy; the executable remains fixed. The
clone-local HTTPS credential helper references existing host access. Each
project permits one worker, one worker attempt, one integration attempt and
300-second worker/validation bounds, with native publication explicitly enabled.
The ordinary project checks its README and Git whitespace; the source project
uses `make verify`.

Both `work doctor` checks report ready, and both effective configurations and
installed command help were inspected using the fixed executable. The isolated
`telegram doctor --check-connection` succeeds. Its legacy runtime-plan indicator
remains false without a concrete execution workspace; native host-work readiness
is reported separately by `work doctor`.

A bounded receiver activation acknowledged an existing incoming message and its
real outgoing response. A separate informational readiness notice was sent to
the configured user. Neither is a development-journey result. No authoring
conversation, model request, native invocation or confirmed integration has yet
been produced by live user development input. TG-01, TG-27, TG-28 and TG-32 remain
pending; the other 28 criteria retain their prior integrated fixture references.

The owned receiver stopped at the end of its 300-second window. The timeout
wrapper returned its expected status 124; the receiver process exited, both
configuration/bot locks were released and retained SQLite `quick_check` passed.
The executable digest is unchanged. The retained fixture has one incoming update,
one acknowledged response, zero development conversations and zero operation
requests. Acta setup/activation note: `01M53G6BMNSN6ERW3Y64QXNQ1T`. Reuse this
same private development root and both targets when actual user messages arrive.

## Risks and Follow-ups

The isolated bindings and fixed installed receiver are prepared. The authorized
Telegram user must supply the live request
and any bound clarification/admission messages. Record actual source/base/result
and canonical integration receipt separately from fixtures. Scheduled exhaustive
certification and delivery-set closure remain pending the final integrated
candidate.
