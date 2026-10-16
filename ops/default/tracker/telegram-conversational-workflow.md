# Telegram Conversation Tracker

Status: Approved

Date: 2026-10-16

Approved: maintainer, 2026-10-16; specification scope and delivery design accepted through the request for spec, plan and tracker.

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../plan/telegram-conversational-workflow.md)

Spec: [Telegram Conversation and Workflow Control](../spec/stable/telegram-conversational-workflow.md)

Model: [Telegram Conversation and Workflow Model](../spec/stable/telegram-conversational-workflow-model.md)

Implementation authorization: Granted by the maintainer on 2026-10-16 for this exact set, canonical slice PRs and Looper-owned Rebase + Fast-forward merges into dev.

Active slice: Slice 2, Conversation routing; implementation in progress

Next slice: Slice 3, Replay-safe operations; after verified Slice 2 integration

Progress: 1 of 8 slices integrated; 2 of 29 tasks validated and 4 done; 4 of 32 acceptance criteria have deterministic fixture evidence.

Full gate: `make verify`

Exhaustive certification: pending the scheduled nightly for the final integrated candidate.

## State and Evidence Rules

Task states: `pending`, `ready`, `in_progress`, `validated`, `blocked`, `done`.
Slice 1 tasks are done after verified PR #390 integration. Slice 2 is active;
later tasks remain pending until their prerequisites integrate. A validated task becomes
done after verified slice integration; a successful test or worker exit alone
does not close it.

Use one Conventional Commit per task when practical and one PR per slice.
The canonical identities below match the plan. Do not renumber slices or create
one PR per task. Looper starts only after execution authority is granted;
read-only status must not start or resume it.

Each task's scope and validation are defined in the matching plan section.
Record the implementation commit and report evidence here as work proceeds.
The required order is `make verify`, focused checks, applicable focused race
checks, then `git diff --check`, all for the exact candidate HEAD.

The final scheduled nightly remains separate from Looper's fast local gate.
After all slices integrate, use `awaiting_nightly` until the recorded exact
candidate is green. Preserve active validation ownership and frozen run
configuration. Never launch an interactive full suite to fill that evidence gap.

## Canonical Slice Map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Dedicated live receiver | `feat/slice-1-telegram-receiver` | `feat(slice-1): connect a dedicated Telegram receiver` | `ops/default/report/slices/telegram-conversational-workflow/slice-1-receiver.md` |
| Slice 2 | Conversation routing | `feat/slice-2-telegram-conversations` | `feat(slice-2): retain project-bound Telegram conversations` | `ops/default/report/slices/telegram-conversational-workflow/slice-2-conversations.md` |
| Slice 3 | Replay-safe operations | `feat/slice-3-telegram-operations` | `feat(slice-3): apply replay-safe Telegram operations` | `ops/default/report/slices/telegram-conversational-workflow/slice-3-operations.md` |
| Slice 4 | Conversational authoring | `feat/slice-4-telegram-authoring` | `feat(slice-4): prepare work through Telegram conversations` | `ops/default/report/slices/telegram-conversational-workflow/slice-4-authoring.md` |
| Slice 5 | Native work continuation | `feat/slice-5-telegram-work` | `feat(slice-5): advance native work from Telegram` | `ops/default/report/slices/telegram-conversational-workflow/slice-5-work.md` |
| Slice 6 | Guarded controls and capacity | `feat/slice-6-telegram-controls` | `feat(slice-6): control shared work through Telegram` | `ops/default/report/slices/telegram-conversational-workflow/slice-6-controls.md` |
| Slice 7 | Recovery and user operation | `feat/slice-7-telegram-recovery` | `feat(slice-7): recover and operate the Telegram service` | `ops/default/report/slices/telegram-conversational-workflow/slice-7-recovery.md` |
| Slice 8 | Live journey and closure | `test/slice-8-telegram-delivery` | `test(slice-8): certify the Telegram development journey` | `ops/default/report/slices/telegram-conversational-workflow/slice-8-delivery.md` |

## Slice 1: Dedicated Live Receiver

State: delivered

Prerequisite: approved committed planning artifacts; execution authorized on 2026-10-16.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T1.1 | done | `e38f90f8`, `063a2a45`; readiness completed with T1.4 | Private setup/configuration, secret handling and readiness commands. |
| T1.2 | done | `063a2a45`; foreground service wiring completed with T1.4 | Bounded HTTP receiver, exact identity authorization and ownership checks. |
| T1.3 | done | `d32b92a9`; authorization-bound delivery completed with T1.4 | Additive storage, atomic inbox/cursor, initial outbox and retained retries. |
| T1.4 | done | `99f0868a` | Read-only registry/help/project/status operations and existing adapter compatibility. |

Validation: plan section 6; `TestTelegramReceiver` and the named existing
notification/import regressions, including focused race mode.

Primary implementation commit: `99f0868a6b616dd44dab95dc758718b91addd313`. Canonical PR: [#390](https://forge.adrianpk.com/adrianpk/chr/pulls/390) on `feat/slice-1-telegram-receiver`. Verified merge: `3e2eb8825b4bb4c4b86585e5e2f862c4202f51b2`, 2026-10-16. Evidence: [Slice 1 report](../report/slices/telegram-conversational-workflow/slice-1-receiver.md); controller receipt for that exact clean HEAD records all four ordered checks passed.

## Slice 2: Conversation Routing

State: in_progress

Prerequisite: verified Slice 1 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T2.1 | validated | `8e7faf9bf62c` | Frozen project/namespace bindings, peer selection and independent conversations. |
| T2.2 | validated | task commit follows this update | Exact replies/questions, duplicate-safe capture and safe completion/cancellation. |
| T2.3 | in_progress | pending | Deterministic sessions/focus commands and bounded conversation views. |

Validation: plan section 7; `TestTelegramConversation`, including focused race mode.

Task validation: `go test ./internal/core ./internal/app -run '^Test(TelegramConversation|MaintenanceDBStatus)' -count=1 -timeout=3m` passed for T2.1; strict focused lint reported zero issues.

T2.2 validation: `go test ./internal/core -run '^TestTelegramConversation' -count=1 -timeout=3m` passed; strict focused lint reported zero issues.

Candidate / PR / verified merge / evidence: pending.

## Slice 3: Replay-safe Operations

State: pending

Prerequisite: verified Slice 2 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T3.1 | pending | pending | Caller-neutral request binding, owner journals and durable mutation receipts. |
| T3.2 | pending | pending | Guarded ticket/slice/slice-set/use-case record operations through shared owners. |
| T3.3 | pending | pending | Durable outbox, reply/action guards and bounded Unicode delivery. |
| T3.4 | pending | pending | Real-store crash/replay proof and changed-argument rejection. |

Validation: plan section 8; `TestTelegramOperation`, including focused race mode.

Candidate / PR / verified merge / evidence: pending.

## Slice 4: Conversational Authoring

State: pending

Prerequisite: verified Slice 3 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T4.1 | pending | pending | Bounded typed natural-language intent and deterministic command parity. |
| T4.2 | pending | pending | Real configured-provider authoring with exact question/turn binding. |
| T4.3 | pending | pending | Explain/prepare/develop intent, retained scope and bounded authority. |
| T4.4 | pending | pending | Canonical artifact generation/refinement, provenance and replay checks. |

Validation: plan section 9; `TestTelegramAuthoring`, including focused race mode.

Candidate / PR / verified merge / evidence: pending.

## Slice 5: Native Work Continuation

State: pending

Prerequisite: verified Slice 4 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T5.1 | pending | pending | Exact native previews/admission for all five extents and RequestKey replay. |
| T5.2 | pending | pending | Bounded asynchronous continuation under existing workflow ownership. |
| T5.3 | pending | pending | Unattended policy-driven progress, review and integration. |
| T5.4 | pending | pending | Evidence-bound results and shared CLI/TUI/Web invocation references. |

Validation: plan section 10; `TestTelegramWork`, including focused race mode.

Candidate / PR / verified merge / evidence: pending.

## Slice 6: Guarded Controls and Capacity

State: pending

Prerequisite: verified Slice 5 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T6.1 | pending | pending | Shared pause/cancel/resume/continue/reconcile controls and actor mapping. |
| T6.2 | pending | pending | Exact-candidate review, grants and decision answers with single-use receipts. |
| T6.3 | pending | pending | Responsive prioritized controls, stale-action handling and callback/text parity. |
| T6.4 | pending | pending | Shared serial/parallel capacity and cross-adapter ownership checks. |

Validation: plan section 11; `TestTelegramControl`, including focused race mode.

Candidate / PR / verified merge / evidence: pending.

## Slice 7: Recovery and User Operation

State: pending

Prerequisite: verified Slice 6 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T7.1 | pending | pending | Startup reconciliation, owned shutdown and fixed-build self-development behavior. |
| T7.2 | pending | pending | Persistent retry cycles, cursor epochs, exact retirement and bounded retention. |
| T7.3 | pending | pending | Optional user service, local retry/status/doctor/help and operational diagnostics. |

Validation: plan section 12; `TestTelegramRecovery`, including focused race mode.

Candidate / PR / verified merge / evidence: pending.

## Slice 8: Live Journey and Closure

State: pending

Prerequisite: verified Slice 7 integration plus private live readiness and
recorded bot/model/service/QA activation authority from plan section 2.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T8.1 | pending | pending | Authorized real bot/model request-to-integration journey and exact evidence. |
| T8.2 | pending | pending | Chryso source development through a fixed installed receiver in isolated QA. |
| T8.3 | pending | pending | User documentation, changelog, complete evidence audit and nightly-bound closure. |

Validation: plan section 13; `TestTelegramJourney`, including applicable focused
race mode, separate authorized live proof and final exact-candidate nightly.

Candidate / PR / verified merge / evidence: pending.

## Acceptance Evidence

The plan's section 14 maps every criterion to its primary tasks. Add immutable
report/test/live references in the Evidence column as they are verified.

| Criterion | State | Evidence |
| --- | --- | --- |
| TG-01 | pending | pending |
| TG-02 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-03 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-04 | pending | pending |
| TG-05 | pending | pending |
| TG-06 | pending | pending |
| TG-07 | pending | pending |
| TG-08 | pending | pending |
| TG-09 | pending | pending |
| TG-10 | pending | pending |
| TG-11 | pending | pending |
| TG-12 | pending | pending |
| TG-13 | pending | pending |
| TG-14 | pending | pending |
| TG-15 | pending | pending |
| TG-16 | pending | pending |
| TG-17 | pending | pending |
| TG-18 | pending | pending |
| TG-19 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-20 | pending | pending |
| TG-21 | pending | pending |
| TG-22 | pending | pending |
| TG-23 | pending | pending |
| TG-24 | pending | pending |
| TG-25 | pending | pending |
| TG-26 | pending | pending |
| TG-27 | pending | pending |
| TG-28 | pending | pending |
| TG-29 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-30 | pending | pending |
| TG-31 | pending | pending |
| TG-32 | pending | pending |

## External Prerequisite Register

| Requirement | State | Evidence or action |
| --- | --- | --- |
| Looper execution authority | granted | Maintainer instructed this exact prepared set to proceed on 2026-10-16. |
| Dedicated bot credentials and exact identity bindings | pending private setup | Supply locally through masked setup after implementation; never in versioned artifacts. |
| Live service/messages/model/QA authority | granted for bounded plan validation | Use the dedicated bot, configured model and existing designated QA repositories; private readiness remains required. |
| Final scheduled nightly on integrated candidate | pending | Use existing Forgejo schedule; do not dispatch or run the long gate interactively. |

## Final Candidate and Closure

Integrated candidate: pending.

Local controller gate result: pending.

Scheduled nightly run / exact tested HEAD / result: pending.

Live journey build / request / invocation / integration receipt: pending.

Cycle report: expected at `ops/default/report/telegram-conversational-workflow-cycle.md`; not yet created.

- [ ] All 29 task commits and focused checks recorded.
- [ ] All eight canonical PRs integrated and every report delivered.
- [ ] All 32 acceptance criteria evidenced, including authorized live journeys.
- [ ] Exact integrated candidate passed the scheduled exhaustive nightly.
- [ ] Cycle report and closure metadata finalized without changing tested code.

Current state: Slice 1 delivered through PR #390 at `3e2eb8825b4bb4c4b86585e5e2f862c4202f51b2`; Slice 2 is active in its dedicated worktree.
