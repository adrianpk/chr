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
Separate live reception and canonical capture evidence is recorded below;
the complete live journeys remain pending.

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
The categorized-help image was built from
`7f0e493431227e752fe7ce5c89f6b2950bc237ef`, with Go 1.27.1 and
`vcs.modified=false`; its SHA-256 was
`422039e26460491c0a34124a097fb4b3cd2bf5e1d2623fcf81e72e4216e3c7d9`.
The maintainer then requested reinstalling the cleaner registered-project list.
Current dev is synchronized into this candidate before rebuilding the isolated
receiver, preserving the pending admission fixes. The private fixture descriptor
owns the current executable source and digest and retains retired image history.
No development conversation was active when this upgrade was prepared.

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

The acceptance traceability audit at
`0274135cc98b4640624ab6fc634cb1b6bc69d6d5` found exactly one tracker entry
for each of TG-01 through TG-32. All report links, section anchors and referenced
commits resolve. The canonical forge confirms the planned title, branch, dev
base and recorded merge commit for PRs #390 through #396; all seven merge
commits are ancestors of the audited candidate. Their delivered reports contain
the configured command lists, and every focused selection has matching tests in
the candidate source. This checks traceability without rerunning earlier suites.
The matrix retains 28 fixture-backed criteria, live reception for TG-01 and
pending TG-27, TG-28 and TG-32.

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

The maintainer-requested project-list reinstall composition passed
`go test ./internal/app -run '^TestTelegramReceiverReadJourney$' -count=1 -timeout=3m`,
`make verify` and `git diff --check` before its merge commit. These checks cover
the composed receiver implementation, not the pending live development journeys.

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
the configured user. Neither was a development-journey result. This initial
activation preceded the live development request recorded below.

The owned receiver stopped at the end of its 300-second window. The timeout
wrapper returned its expected status 124; the receiver process exited, both
configuration/bot locks were released and retained SQLite `quick_check` passed.
The executable digest was unchanged. At that checkpoint the retained fixture
had one incoming update, one acknowledged response, zero development
conversations and zero operation requests. Acta setup/activation note:
`01M53G6BMNSN6ERW3Y64QXNQ1T`. Subsequent observations reuse this same private
development root and both targets.

After the receiver image update, a genuine incoming `/help` received a
categorized reply accepted by Telegram in Markdown mode on its first delivery
attempt, without plain fallback. This verifies the deployed help path. It does
not establish a development conversation, native invocation or integration;
the two complete live development journeys remain pending.

Before the first development request, another retained snapshot contained six
processed messages and six delivered application replies, with zero development
conversations and zero operation requests. The configured user also received an
informational notice with the two QA start commands. That notice is recorded
privately and is not a development request or live journey result.

## Live QA Progress

TG-01 is validated by a genuine `/develop` request in `telegram-journey-qa` and
its delivered project-bound acknowledgement. The request starts a small README
change that appends a verification sentence while preserving existing content.
Conversation `tgc-38356dda2b04b98de3d419985fa97481` retains the `develop` intent,
project `project-0b671359780bf6ba` and namespace `telegram-qa`. Its revision-one
reply is delivered on the first attempt and has a retained Telegram
acknowledgement. Private evidence is retained as `live-reception-qa.json` beside
the fixture descriptor; no private peer identifiers are copied here.

The receiver is fixed at source
`14c563d40f6a3bc0c41f7e3a8b4fc52e1c2cb166`, with SHA-256
`419cf6654b5c9a3f40cb69ab42056ebb1dd00d944e0ced4fc6043276ad16a8a6`.
Its digest matches the private descriptor. This build remains fixed while the
candidate documentation changes. The existing user service is managed outside
this worker and remains active. Its retained polling health is `ready`, with no
last error. The remaining live steps depend on the configured user's
conversation input.

The subsequent genuine `/done` produced confirmed request
`tgr-15e80739483b1a83389faf08385db9f0` for `authoring.continue`. Its canonical
session is `.chr/authoring/sessions/author-b041d04024ce5905d0675851/session.md`;
the dump is
`ops/telegram-qa/dump/00000-author-b041d04024ce5905d0675851-dump/artifact.md`.
The shared owner's retained receipt, session and dump exist on disk; their
bounded metadata and dump digest are retained privately in `live-qa-dump.json`.
The bot's prepared-authoring response is delivered. This capture operation has
zero model attempts, so it does not establish the live provider path.

At this checkpoint the conversation is `prepared`, with no native invocation or
confirmed integration. Native work readiness still reports `ready` in worktree
mode after the generated operational files appear in the user checkout; the
readiness report is retained privately as `doctor-after-live-dump.txt`. No
generated files were removed and no policy was weakened. Live model
clarification, work preparation and bound admission remain. TG-27, TG-28 and
TG-32 remain pending. The 28 earlier criteria retain their integrated fixture
references.

## Live Execution Checkpoint: 2026-10-21

Genuine user messages completed model clarification, spec generation and native
ticket preparation for `TGQA-1`. The first admitted invocation,
`00002-workflow-invocation`, never started a worker because its frozen target
omitted the coordinator marker created by authoring. The maintainer authorized
pausing it. The designated QA target was repaired at
`03c16df809f841924f1c4b3cd458644bf77b1381` by committing only that marker, and
the obsolete invocation was cancelled with zero attempts. Its evidence remains.

A fresh user-confirmed preview admitted `00003-workflow-invocation` on that
repaired target. One real Codex worker appended the exact requested paragraph
while preserving the original README bytes, but left its edit uncommitted.
The worker exited zero; native execution correctly paused with
`task_validation_failed` and `workflow lane must be clean`. No host validation
or integration was confirmed. Independent inspection of the retained lane
verified the exact append and the configured README/whitespace checks.

The executor input omitted an explicit commit instruction. New task inputs now
freeze scoped edit, validation and commit instructions alongside the task
contract. Publication and workflow state remain host-owned. The existing real
Git regression now uses a worker that commits only when instructed: it failed
with the same dirty-lane error before the correction and passes afterwards.
Dirty-worker rejection is retained, and aggregate validation receives no task
commit instruction. Focused task, aggregate, manual, certified-driver and
Telegram journey checks passed:

`go test ./internal/core ./internal/app -run '^Test(WorkflowTask(Exact|Failures|Independent|Cumulative)|WorkflowSliceAggregate|WorkflowManualExact|NativeAgent(Settlement|Capability)|TelegramJourney)' -count=1 -timeout=3m`

Core/app completed in 33.640/11.684 seconds. `make verify` also passed: strict
lint reported zero issues, short core/app tests completed in 17.107/28.524
seconds, all eight CI script regressions passed and the application built.
Local document links and `git diff --check` passed. These development checks do
not replace the clean-HEAD controller receipt or scheduled exhaustive gate and
do not prove that the real model follows the corrected instructions. The failed live
attempt and its dirty lane remain intact. The running receiver still uses
`14c563d40f6a`; activation of a corrected build and a successful real retry are
pending. The user-guided slash-command sequence is diagnostic progress. A
complete conversation from an ordinary natural-language request through
confirmed integration remains a separate acceptance requirement.

## Risks and Follow-ups

T8.1 has genuine live reception, clarification, canonical authoring, prepared
work and exact admission evidence. Native integration remains blocked on
activating the executor handoff correction and proving a successful live run.
T8.2 still needs the separately bound source-development request and complete
journey. Fixture checks, synthetic inbound updates and direct application calls
cannot supply those results. The report remains drafting and the canonical PR
remains pending while these requirements are incomplete.

Retain the existing QA attempts and use a new authorized attempt after the
receiver correction. The source QA journey requires its own separately bound
request and fixed receiver image. Record actual
source/base/result and canonical integration receipts separately from fixtures.
Scheduled exhaustive certification and delivery-set closure remain pending the
final integrated candidate.

## Continuation Audit: 2026-10-20

The slice branch now composes current `dev` through `6037036b2fba`, including
the saved-spec approval correction and nightly fixture repairs. Plan/tracker
conflicts were reconciled by preserving genuine live capture, integrated
Slice 7 receipts, pending task states and the retained controller blocker.
The receiver service is active and ready; its fixed executable still matches
the private descriptor digest. No service was restarted or upgraded.

The previous controller validation applies only to `4e4fbf8fadef`. New
exact-HEAD composition checks belong to the existing controller checks ledger;
this report does not infer their outcome before that receipt exists. The next
genuine input is the existing QA conversation's clarification request; no
synthetic update or local direct call substitutes for it. The second source-QA
journey and exact-candidate scheduled exhaustive gate also remain pending.

Pre-commit composition checks passed: `make verify` (strict lint zero issues,
short core/app 17.991/29.684 seconds, eight CI script regressions and build),
operational file links, conflict-marker checks and `git diff --check`. These
working-tree checks do not replace the ordered clean-HEAD controller receipt.
