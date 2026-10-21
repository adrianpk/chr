# Telegram Message Presentation Tracker

Status: implementing

Approved: maintainer, 2026-10-21; specification, model, slice map and Looper execution accepted.

Full gate: `make verify`

Spec: [Message presentation](../spec/stable/telegram-message-presentation.md)

Model: [Document delivery](../spec/stable/telegram-message-presentation-model.md)

Plan: [Slice map and validation](../plan/telegram-message-presentation.md)

Inventory: [48-family review](../report/telegram-message-review.md)

Implementation and merge authorization: granted for this delivery set and
Looper-owned Rebase + Fast-forward slice PR merges into dev.
Active slice: Slice 4 of 5, Progress, notifications and failure consistency.

## Canonical Slice Map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Readable text and conversation messages | feat/slice-1-telegram-message-composition | feat(slice-1): compose readable Telegram messages | ops/default/report/slices/telegram-message-presentation/slice-1-message-composition.md |
| Slice 2 | Document excerpts and complete attachments | feat/slice-2-telegram-document-delivery | feat(slice-2): deliver Telegram document previews and attachments | ops/default/report/slices/telegram-message-presentation/slice-2-document-delivery.md |
| Slice 3 | Clear guarded confirmations and controls | feat/slice-3-telegram-action-presentation | feat(slice-3): clarify Telegram confirmations and controls | ops/default/report/slices/telegram-message-presentation/slice-3-action-presentation.md |
| Slice 4 | Progress, notifications and failure consistency | feat/slice-4-telegram-status-presentation | feat(slice-4): unify Telegram progress and notification presentation | ops/default/report/slices/telegram-message-presentation/slice-4-status-presentation.md |
| Slice 5 | Complete coverage and real client acceptance | test/slice-5-telegram-presentation-acceptance | test(slice-5): certify Telegram message presentation | ops/default/report/slices/telegram-message-presentation/slice-5-presentation-acceptance.md |

## Tasks

Slices 1-3 are delivered through PRs #424, #425 and #427. Slice 4 is active;
T4.1-T4.3 are implemented. Slice 5 remains pending.
The plan owns exact branch names, PR titles, report paths and focused commands.

| Task | State | Commit | Required outcome |
| --- | --- | --- | --- |
| T1.1 | delivered | `ffcbc9b4dabc8fbff111e46ef74ad1b55b172b2a` | Shared optional blocks and bounded rich/plain renderer |
| T1.2 | delivered | `c5665e602b5f9b00203e0366216d7a2dbdd7f3ae` | Readable conversation/text families 01-20 |
| T1.3 | delivered | `eb7096813576290130d4f4edd583195a62ea443a` | Renderer, transport and conversation regressions |
| T2.1 | delivered | `0c50fa368801ef245c1e59a28e5386bc7df58f81` | Durable receiver snapshots, migrations and quota |
| T2.2 | delivered | `bb6a5575d69dbaa4d804e82c73adcb6dae561ca2` | Document upload, caption, acknowledgement and fallback |
| T2.3 | delivered | `c48039eb4af550ee760e35af0c0dc8c9de648de6` | Primary artifact/export families 21-24 and 28 |
| T2.4 | delivered | `7e2b7d21e1368ffe1eea6bbf83ad70fe3520f6b6` | Exact bytes, bounds, rollback and recovery evidence |
| T3.1 | delivered | `b7a1174a5d099a663c93ba4d746903725301b264` | Human guarded previews and controls |
| T3.2 | delivered | `94975b70a20b87639e1df5ee1d9ebca4fbb741de` | Original operation, document and action bindings |
| T3.3 | delivered | `099b9a965b0bbe2690ca0191c3d51250a39deadb` | Stale scope, ambiguity and replay regressions |
| T4.1 | implemented | `708afec21584e546624848391a287a59fb5da5d4` | Progress, waits, failures and partial/final results |
| T4.2 | implemented | `04fddeb15cf6184f4c263a55bdb222fcd74a6b1e` | Notification sender and event snapshot integration |
| T4.3 | implemented | `01d9be1341e5c207c1748a22ced4a5aac8eafe1d` | Both-path faults, legacy records and truthful outcomes |
| T5.1 | pending | pending | Complete producer and variant coverage |
| T5.2 | pending | pending | Real mobile/desktop and natural-language journey |
| T5.3 | pending | pending | User documentation, cycle evidence and exact nightly |

## Message-Family Coverage

Numbers retain the source review's identities. Rows 01-20 record Slice 1
implementation and fixture evidence; integration and real client acceptance
remain separate. Later rows remain obligations. Record the actual source owner,
regression and client result as each row is delivered.

| Family | Required presentation/outcome | Owning slice | Evidence |
| --- | --- | --- | --- |
| 01 | Readable help; advanced material conditional | Slice 1 | Changed: bounded basic help and explicit complete groups; `app/telegram_help.go`; `TestTelegramPresentationConversationHelp`, `TestTelegramHelpHTTP`. |
| 02 | Relevant command/unsupported-action guidance | Slice 1 | Changed: concise unsupported-command guidance; keep unavailable callback wording; `app/telegram_operations.go`; `TestTelegramConversationUnknown`, `TestTelegramHelpHTTP`. |
| 03 | Human project list and empty state | Slice 1 | Changed: human project list and explicit local setup for empty access; `app/telegram_operations.go`; `TestTelegramReceiverReadJourney`, `TestTelegramPresentationConversationReads`. |
| 04 | Selected-project acknowledgement | Slice 1 | Keep: selected registered name, omit an unnecessary next step; `core/telegram_conversation.go`; `TestTelegramConversationProjects`. |
| 05 | Safe project/reply/root ambiguity and recovery | Slice 1 | Changed: safe project/root/target recovery; `core/telegram_conversation.go`; `TestTelegramConversationBinding`, `TestTelegramConversationProjectNames`, `TestTelegramAuthoringIntentRootChange`. |
| 06 | Receiver health and next responsible party | Slice 1 | Changed: health and relevant retry/error guidance; omit bot/cursor metadata; `app/telegram_operations.go`; `TestTelegramPresentationConversationReads`, `TestTelegramHelpHTTP`. |
| 07 | Request acknowledged; extra capture details optional | Slice 1 | Changed: acknowledge the retained request and optional extra details; `core/telegram_conversation.go`; `TestTelegramPresentationConversationJourney`. |
| 08 | Concrete start/capacity limits | Slice 1 | Changed: explicit topic/allocation bounds and session selection; `core/telegram_conversation.go`; `TestTelegramConversationIsolation`, `TestTelegramConversationBotQuota`. |
| 09 | Human session topics, states and targeting choices | Slice 1 | Changed: human topics, project names, states and complete exact target rows; `core/telegram_conversation_presentation.go`; `TestTelegramPresentationConversationJourney`. |
| 10 | Clear focus effect or actual unavailable state | Slice 1 | Changed: focus effect or actual noncollecting state; `core/telegram_conversation_target.go`; `TestTelegramConversationFocus`, `TestTelegramPresentationConversationJourney`. |
| 11 | Stale/conflicting prompt and preserved work | Slice 1 | Changed: stale/conflicting reply guidance preserves saved context; `core/telegram_conversation_target.go`; `TestTelegramConversationReplies`, `TestTelegramConversationReplyGuards`, `TestTelegramConversationReplyPeers`. |
| 12 | Details saved; one useful next action | Slice 1 | Changed: concise acknowledgement, accounting only near bounds; `core/telegram_conversation_presentation.go`; `TestTelegramPresentationConversationJourney`, `TestTelegramConversationCapture`. |
| 13 | Readable retained capture and inspection | Slice 1 | Changed: human capture view and literal Markdown entity; `core/telegram_conversation_presentation.go`; `TestTelegramPresentationConversationJourney`, `TestTelegramPresentationFormatHTTP`. |
| 14 | Capture/finished-state errors with unchanged context | Slice 1 | Changed: relevant capture limit and preserved-context guidance; `core/telegram_conversation_capture.go`; `TestTelegramConversationCapture`, `TestTelegramConversationCompletion`. |
| 15 | Saved request and truthful generation state | Slice 1 | Changed: saved capture versus confirmed canonical dump generation; `app/telegram_authoring_apply.go`; `TestTelegramAuthoringBridgeJourney`, `TestTelegramConversationCompletion`. |
| 16 | Explicit unstarted-capture discard | Slice 1 | Changed: explicit discard of unstarted capture; `core/telegram_conversation_capture.go`; `TestTelegramPresentationConversationJourney`, `TestTelegramConversationCompletion`. |
| 17 | Relevant syntax/example only | Slice 1 | Changed: only relevant command syntax and a start example; `app/telegram_conversation.go`, `app/telegram_authoring_command.go`; `TestTelegramConversationCommands`. |
| 18 | One concrete natural-language target question | Slice 1 | Changed: concrete operation/target question; keep retained target guards; `app/telegram_intent_dispatch.go`, `app/telegram_intent_guard.go`; `TestTelegramAuthoringIntentInvalid`, `TestTelegramAuthoringIntentRootChange`. |
| 19 | Interpretation failure/retry exhaustion guidance | Slice 1 | Changed: interpretation unavailable/exhausted guidance preserves prior conversations; `app/telegram_intent_dispatch.go`; `TestTelegramAuthoringIntentProviderFailure`. |
| 20 | Actual clarification question and clear answer route | Slice 1 | Changed: actual question first with original reply route; `app/telegram_authoring_apply.go`; `TestTelegramPresentationConversationQuestion`, `TestTelegramHelpQuestion`. |
| 21 | Primary generated artifact excerpt and full file | Slice 2 | Changed: one primary complete dump/spec/plan with direct literal excerpt; `app/telegram_document_results.go`; `TestTelegramPresentationDocumentAuthoring`. |
| 22 | Authoring wait/failure from real owner state | Slice 2 | Keep: existing durable provider wait/retry ownership and context-preserving failure; omit an unscheduled retry promise; `app/telegram_authoring_apply.go`; `TestTelegramAuthoringExplainFailure`, `TestTelegramAuthoringBridgeLostAck`. |
| 23 | Saved result after conversation advances | Slice 2 | Changed: attach the saved result while preserving the advanced conversation; `app/telegram_authoring_apply.go`; `TestTelegramPresentationDocumentAdvancedConversation`. |
| 24 | Bounded explanation, with full content when needed | Slice 2 | Changed: bounded explanation preview and complete identified Markdown export with canonical sources; `app/telegram_explanation_dispatch.go`; `TestTelegramAuthoringExplainRouting`. |
| 25 | Short ticket scope and bound complete document | Slice 3 | Changed: readable preparation scope and complete identified confirmation export; `app/telegram_prepared_work_result.go`, `telegram_confirmation_document.go`; `TestTelegramAuthoringPreparedAction`, `TestTelegramPresentationConfirmationBinding`. |
| 26 | Prepared backlog result and correct next step | Slice 3 | Changed: backlog result and exact next work preview; replay shows the original result without raw JSON; `app/telegram_prepared_work_result.go`; `TestTelegramAuthoringPreparedIntent`. |
| 27 | Explicit missing intent/context for preparation | Slice 3 | Keep: actual missing-instruction/intent diagnostics preserve retained context and never infer authority; `core/telegram_conversation_instruction.go`, `app/telegram_authoring_apply.go`; existing prepared-work intent/failure regressions. |
| 28 | Readable/paginated records and identified full export | Slice 2 | Changed: human record rows and explicit complete records-page JSON export; preserve pagination; `app/telegram_record_dispatch.go`; `TestTelegramOperationLiveRecords`. |
| 29 | Readable proposed record change and consequence | Slice 3 | Changed: human record title/state and canonical reference in results; retain existing direct typed metadata-write ownership and legacy confirmation semantics; `app/telegram_record_apply.go`, `telegram_action_record.go`; `TestTelegramOperationLiveRecords`, `TestTelegramOperationActionReplay`. |
| 30 | Actionable malformed/stale record failure | Slice 3 | Changed: malformed JSON gives a concrete correction and unchanged-record outcome; retain stale revision/source and unsupported-lifecycle guards; `app/telegram_record_dispatch.go`; existing record action/live regressions. |
| 31 | Exact work selection, policy, validation and scope | Slice 3 | Changed: human selection, target, policy, limits and validation; full selected source/contracts attached; incomplete scope refuses a prompt; `app/telegram_work_dispatch.go`, `service_workflow_selection.go`; `TestTelegramPresentationActionWork`, `TestTelegramPresentationConfirmationIncomplete`. |
| 32 | Admission/current-owner/capacity state | Slice 3 | Changed: admission is distinct from execution/integration; owner and capacity failures are separate; `app/telegram_work_action.go`, `telegram_work_failure.go`; `TestTelegramWorkExtents`, `TestTelegramWorkLostAck`. |
| 33 | Real progress, review and partial/final result | Slice 4 | Changed: human work/project view, actual counts, wait/stop/review and positive native-receipt distinction; `app/telegram_work_progress.go`; `TestTelegramPresentationProgressStates`, existing native evidence/automatic integration fixtures. |
| 34 | Durable progress/recovery without a duplicate action | Slice 4 | Changed: frozen bounded presentation under the original monotonic revision; same-operation recovery and original retry allowance; `core/telegram_work_projection.go`; `TestTelegramPresentationProgressRetention`, `TestTelegramWorkProjection`, `TestTelegramWorkDeliveryFailure`. |
| 35 | Exact control scope, effect and bounds | Slice 3 | Changed: exact control effect, bound scope and remaining allowance with complete confirmation export; `app/telegram_action_presentation.go`, `telegram_control_prepare.go`; `TestTelegramPresentationControlPreview`, existing lifecycle/decision regressions. |
| 36 | Confirmed/replayed/stale control outcome | Slice 3 | Changed: actual control result and retained original receipt; no extra mutation on replay; `app/telegram_control_action.go`; existing lifecycle acknowledgement and stale-receipt regressions. |
| 37 | Decision question and supported bound choices | Slice 3 | Changed: one actual question first, bounded choices and exact answer-preview command; complete questions/options/candidate bindings exported; `app/telegram_control_status.go`; `TestTelegramPresentationControlQuestions`. |
| 38 | Readable harness selections and actual readiness | Slice 3 | Changed: concise role defaults and local readiness through the shared renderer; retain remote-authentication distinction and admitted bindings; `app/telegram_harness_dispatch.go`; existing harness visibility/preparation regressions. |
| 39 | Explicit harness-change consequence | Slice 3 | Changed: new-operations-only preview, original save/cancel guards and retained result replay; `app/telegram_harness_prepare.go`, `telegram_harness_action.go`; `TestTelegramHarnessSelectionConfirm`, `TestTelegramHarnessSelectionSendRetry`. |
| 40 | Original receipt, expired action and safe retry | Slice 3 | Changed: original saved result on repeat, current-preview recovery on unavailable actions; `app/telegram_action_dispatch.go`, `telegram_control_diagnostic.go`; `TestTelegramPresentationActionWork`, existing stale action/receipt regressions. |
| 41 | Human button labels; original tokens/choice keys | Slice 3 | Changed: human labels with original token/choice keys; labels, choices and complete image are sealed together; `core/telegram_action_presentation.go`; `TestTelegramPresentationActionLabels`, `TestTelegramPresentationConfirmationGuards`. |
| 42 | Queue rejection versus waiting and actual acceptance | Slice 4 | Changed: full queue means the new request was not accepted; retained work and reserved controls stay available; `app/telegram_drain.go`, `core/telegram_intake_reserve.go`; `TestTelegramControlReserve`. |
| 43 | Uncertain send and bounded plain fallback | Slice 4 | Changed: lost-ack repeat warning; retain original content, identities and bounded formatting fallback; `app/telegram_drain.go`; existing work/action/harness lost-ack and help-format HTTP regressions. |
| 44 | Bounded text, Unicode and full-document route | Slice 4 | Changed: one bounded new plain view with narrowing/local inspection route, without automatic content chunks; keep stored legacy multipart ordering and formatted help; `core/telegram_response_chunks.go`; Unicode/legacy-part/help regressions. |
| 45 | Unavailable service and preserved state | Slice 4 | Changed: failed read identifies the unavailable operation, preserved state and local doctor command; existing preparation/record outcomes retain their owners; `app/telegram_drain.go`; `TestTelegramPresentationFailureRead` injects an unavailable read and verifies unchanged saved state and bounded recovery. |
| 46 | Silent unauthorized rejection; local outage diagnosis | Slice 4 | Keep: unauthorized/revoked input and delivery remain silent; poll/config failures use local-only classified diagnostics; `app/telegram_failure.go`, `telegram_serve.go`, `telegram_drain.go`; `TestTelegramHelpUnauthorized`, `TestTelegramPresentationDeliveryRevocation`, `TestTelegramRecoveryFailures`, `TestTelegramRecoveryDoctor` and `TestTelegramRecoveryPollEpoch` passed normally and with race instrumentation. |
| 47 | Concise workflow notifications and full details | Slice 4 | Changed: human project/work activity and receipt distinction, one complete frozen event JSON export; `app/telegram_notification_presentation.go`, existing notification repository and sender; `TestTelegramPresentationNotificationDispatch`, snapshot/HTTP fault tests. |
| 48 | Clearly labelled notification test | Slice 4 | Changed: short human project name and explicit test purpose, omit routine directory and attachment; `app/telegram_notification_presentation.go`, `notify.go`; `TestNotifyTelegramTest`. |

## Acceptance Evidence

| Criterion | Owner tasks | State |
| --- | --- | --- |
| TMP-01 | T1.1, T4.1 | Shared bounded renderer and progress fixture evidence verified; real client acceptance pending |
| TMP-02 | T4.2, T4.3 | Both delivery owners verified through real HTTP and SQLite; real client acceptance pending |
| TMP-03 | T1.1, T1.3, T4.3 | Long Unicode, original complete exports and bounded fallback fixtures verified |
| TMP-04 | T2.3 | Slice 2 fixtures verified and integrated; client acceptance pending |
| TMP-05 | T2.1, T2.2, T2.4 | Slice 2 fixtures verified and integrated; client acceptance pending |
| TMP-06 | T1.2, T3.1 | Slice 3 fixtures verified and integrated; client acceptance pending |
| TMP-07 | T3.2, T3.3 | Guard, original-result and exact-scope fixtures verified and integrated |
| TMP-08 | T2.4, T3.3, T4.3 | Slice 2 integrated; Slice 3 guard/recovery fixtures verified; Slice 4 both-owner recovery fixtures verified |
| TMP-09 | T2.4, T3.3, T4.3 | Slice 2 integrated; Slice 3 fail-closed scope/quota fixtures verified; Slice 4 shared quota/target/failure fixtures verified |
| TMP-10 | T2.1, T2.4, T3.3, T4.2 | Receiver/action and physical pre-28 notification migrations verified; NULL/version-zero legacy payload semantics preserved |
| TMP-11 | T5.1 | pending |
| TMP-12 | T5.2, T5.3 | pending |

## Completion State

- Planning artifacts: accepted; local links, whitespace, family/task coverage and acceptance mappings checked.
- Specification/model promotion and implementation approval: completed on 2026-10-21.
- Slice 1: delivered through [#424](https://forge.adrianpk.com/adrianpk/chr/pulls/424) at `a2bb0974eac21b1fe33cac78e21f4c1845871e96`.
- Slice 2: delivered through [#425](https://forge.adrianpk.com/adrianpk/chr/pulls/425) at `a5c6a7e068a8e5ec547d13d73d39afe1f0477335`.
- Slice 3: delivered through [PR #427](https://forge.adrianpk.com/adrianpk/chr/pulls/427) at `a4d0a865e8bb32a3caac65395a28839b573ed369`.
- Slice 4: active; T4.1-T4.3 implemented; [PR #428](https://forge.adrianpk.com/adrianpk/chr/pulls/428) targets `dev`; report reviewing.
- Slice 5: pending.
- Mobile/desktop and user-operated journey evidence: pending.
- Exact final candidate and scheduled nightly: pending.

The existing conversational workflow and its active QA evidence retain their
own controller, tracker and invocation states. This set uses a separate Looper
configuration. Its Slice 3 prerequisite is resolved by the canonical repair
integration; real client and journey evidence remain required.

## Slice 1 Validation

Focused presentation, existing conversation/help/authoring regressions and strict
package lint passed. [Slice 1 report](../report/slices/telegram-message-presentation/slice-1-message-composition.md)
records commands and coverage. Looper evidence `UYX4BPRVESSKGMZMYMQQG72NMA`
records all five ordered checks passed on the exact merged candidate.

## Slice 2 Validation

[Slice 2 report](../report/slices/telegram-message-presentation/slice-2-document-delivery.md)
records immutable snapshots, upload/producer behavior, quota admission,
retention, migration and fault fixtures. Looper evidence
`NQQQ4DVYVPESUBCN7TURIYUTUC` records five ordered checks passed on the exact
merged candidate `a5c6a7e068a8e5ec547d13d73d39afe1f0477335`.

## Slice 3 Validation

[Slice 3 report](../report/slices/telegram-message-presentation/slice-3-action-presentation.md)
records complete scope exports, action sealing, original-result recovery and
normal/race HTTP/SQLite fixtures. Looper evidence `RUKJK7R5AERILO54DESZZNK6UW`
records all four ordered checks passed on exact merged candidate
`a4d0a865e8bb32a3caac65395a28839b573ed369`. Canonical PR #427 and Acta merge
event `01M5FRH8SFP1F2QX0PDWA4AJ4B` agree.

## Slice 4 Activation

The dedicated Slice 4 branch starts from exact canonical `dev`
`a4d0a865e8bb32a3caac65395a28839b573ed369`. T4.1-T4.3 cover progress,
receiver failures and the existing notification sender/repository. The
[Slice 4 report](../report/slices/telegram-message-presentation/slice-4-status-presentation.md)
is reviewing. All three tasks are committed and focused normal/race checks passed.
Canonical [PR #428](https://forge.adrianpk.com/adrianpk/chr/pulls/428) targets `dev`.
The report records the final ordered command list. Looper retains clean-HEAD
validation after the PR-reference follow-up commit; the PR validation block
records those successful commands. Development checks above do not replace that
candidate-bound evidence.
