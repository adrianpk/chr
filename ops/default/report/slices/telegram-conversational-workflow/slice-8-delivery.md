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
uses its own worktree and leaves the primary checkout on dev. The isolated
receiver is pinned separately from the installed developer CLI. Before any live
development conversation began, current dev was merged into the candidate so
the receiver includes categorized help alongside the pending admission fixes.
The fixed source is `7f0e493431227e752fe7ce5c89f6b2950bc237ef`, built with
Go 1.27.1 and `vcs.modified=false`. Its SHA-256 is
`422039e26460491c0a34124a097fb4b3cd2bf5e1d2623fcf81e72e4216e3c7d9`.
The previous fixed image at `b72a6955cf7c` is retained privately. Both source and
digest are recorded in the private fixture descriptor; neither image was
replaced during an active source-development journey.

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
connectivity. The controller owns execution and exact-HEAD evidence for this
required ordered command list:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramJourney' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramJourney' -count=1 -timeout=3m`
- `git diff --check`

Read the controller receipt for results bound to the current candidate; a head
change requires new evidence. Passing this list does not replace live journey
proof or the scheduled exhaustive gate.

Task checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramJourney|TelegramAuthoringPrepared|TelegramAuthoringWork|TelegramWorkAdmission|TelegramWorkAutomatic|ServiceWorkflowPreview)' -count=1 -timeout=3m`: core 1.162 seconds; app 34.566 seconds.
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramJourney|TelegramAuthoringPrepared|TelegramAuthoringWork|TelegramWorkAdmission|TelegramWorkAutomatic|ServiceWorkflowPreview)' -count=1 -timeout=3m`: core 2.736 seconds; app 75.807 seconds.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./...`: zero issues.
- `scripts/check-go-func-signatures.sh`, `scripts/check-package-readmes.sh` and `git diff --check`: passed.

These are implementation-task checks. They do not constitute final clean-HEAD
controller validation or the scheduled exhaustive gate.

The four ordered controller checks passed on source candidate
`7f0e493431227e752fe7ce5c89f6b2950bc237ef` in 141 seconds, with validation
receipt `UEKXWTRHXFC3WG7QWL34QC55MN`. The composed help, journey, authoring,
admission and native-work regressions also passed before that commit. This
report-only update follows those checks; the final PR head still needs its own
exact-HEAD controller receipt.

## Live QA Setup

The private user-state descriptor `chr/qa/telegram-conversational-workflow.json`
locates the owned playground. Its `fixture.json` records the executable,
project registrations and exact bootstrap targets; `README.md` contains the bounded foreground command. Private
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

After the receiver image update, a genuine incoming `/help` received a
categorized reply accepted by Telegram in Markdown mode on its first delivery
attempt, without plain fallback. This verifies the deployed help path. It does
not establish a development conversation, native invocation or integration;
the two complete live development journeys remain pending.

The updated image matches the private fixture's recorded digest. The retained
store now has six processed messages and six delivered application replies,
with zero development conversations and zero operation requests. The configured
user also received an informational notice with the two QA start commands.
That notice is recorded privately and is not a development request or live
journey result. The same receiver, registrations and targets remain in use.

## Risks and Follow-ups

T8.1 and T8.2 are blocked on genuine inbound Telegram user messages. Repeated
bounded receiver checks retained zero development conversations and zero
operation requests. The default store also contains no development conversation.
No authenticated Telegram user client has been supplied; the bot credential
cannot originate authenticated user updates. Resume with requests, clarification
and exact admission from the configured user, or an existing authenticated user
client. Fixture checks, synthetic inbound updates and direct application calls
cannot supply this live evidence. The report remains drafting and the canonical
PR remains pending while these requirements are incomplete.

The isolated bindings and fixed installed receiver are prepared. The authorized
Telegram user must supply the live request
and any bound clarification/admission messages. Record actual source/base/result
and canonical integration receipt separately from fixtures. Scheduled exhaustive
certification and delivery-set closure remain pending the final integrated
candidate.
