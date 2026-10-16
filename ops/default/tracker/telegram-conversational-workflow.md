# Telegram Conversation Tracker

Status: Approved

Date: 2026-10-16

Approved: maintainer, 2026-10-16; specification scope and delivery design accepted through the request for spec, plan and tracker.

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../plan/telegram-conversational-workflow.md)

Spec: [Telegram Conversation and Workflow Control](../spec/stable/telegram-conversational-workflow.md)

Model: [Telegram Conversation and Workflow Model](../spec/stable/telegram-conversational-workflow-model.md)

Implementation authorization: Granted by the maintainer on 2026-10-16 for this exact set, canonical slice PRs and Looper-owned Rebase + Fast-forward merges into dev.

Active slice: Slice 6, Guarded controls and capacity; T6.1 through T6.3 validated, T6.4 ready

Next slice: Slice 7, Recovery and user operation; after verified Slice 6 integration

Progress: 5 of 8 slices integrated; 3 of 29 tasks validated and 19 done; 18 of 32 acceptance criteria have deterministic fixture evidence.

Full gate: `make verify`

Exhaustive certification: pending the scheduled nightly for the final integrated candidate.

## State and Evidence Rules

Task states: `pending`, `ready`, `in_progress`, `validated`, `blocked`, `done`.
Slice 1 through Slice 5 tasks are done after verified PR #390/#391/#392/#393/#394 integration. Slice 6 is active;
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

State: delivered

Prerequisite: verified Slice 1 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T2.1 | done | `8e7faf9bf62c` | Frozen project/namespace bindings, peer selection and independent conversations. |
| T2.2 | done | `9946757b0528` | Exact replies/questions, duplicate-safe capture and safe completion/cancellation. |
| T2.3 | done | `36f31feaa0c0` | Deterministic sessions/focus commands and bounded conversation views. |

Validation: plan section 7; `TestTelegramConversation`, including focused race mode.

Task validation: `go test ./internal/core ./internal/app -run '^Test(TelegramConversation|MaintenanceDBStatus)' -count=1 -timeout=3m` passed for T2.1; strict focused lint reported zero issues.

T2.2 validation: `go test ./internal/core -run '^TestTelegramConversation' -count=1 -timeout=3m` passed; strict focused lint reported zero issues.

T2.3 validation: focused `TestTelegramConversation` normal and race commands passed in both core and app; strict focused lint reported zero issues. The [Slice 2 report](../report/slices/telegram-conversational-workflow/slice-2-conversations.md) records deterministic evidence and the required exact-HEAD controller checks.

Primary implementation commit: `36f31feaa0c0c0d5b52e9d80071322a8574744b5`. Canonical PR: [#391](https://forge.adrianpk.com/adrianpk/chr/pulls/391) on `feat/slice-2-telegram-conversations`. Verified merge: `f8f33d844c083997ffb0e479e9c22068e5c30e03`, 2026-10-16; all four ordered controller checks passed on that clean HEAD. Evidence: [Slice 2 report](../report/slices/telegram-conversational-workflow/slice-2-conversations.md); final readiness is bound to the exact clean PR HEAD and ordered checks through controller validation.

## Slice 3: Replay-safe Operations

State: delivered

Prerequisite: verified Slice 2 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T3.1 | done | `f39c12e96c3d2228e8411d954f0a219b86f2d497`, `1b0ec78267b3` | Caller-neutral request binding, owner journals and durable mutation receipts. |
| T3.2 | done | `9ebe146da339`, `5507e7041a3ddcdbd8021e9919c17f6b6ae555fc` | Guarded ticket/slice/slice-set/use-case record operations through shared owners. |
| T3.3 | done | `e5de10f0eb5e1a44cd4cc787186ab10a5da60bb3`, `76c4862df1b7dd68eb366d1395c2e1ec81ea0c69` | Durable outbox, reply/action guards and bounded Unicode delivery. |
| T3.4 | done | `e5de10f0`, `5507e704`, `c0385ee781452a7dfdad5aadeeb971e3b62cb60e` | Real-store crash/replay proof and changed-argument rejection. |

Validation: plan section 8; `TestTelegramOperation`, including focused race mode.

Task validation: record owner, application and live HTTP fixtures passed in normal and race mode, including compatibility with receiver/conversation and migration checks. Strict focused lint and repository Go-style/package checks passed. Exact commands and fixture boundaries are recorded in the [Slice 3 report](../report/slices/telegram-conversational-workflow/slice-3-operations.md).

Candidate: `c0385ee781452a7dfdad5aadeeb971e3b62cb60e` implementation. Canonical PR: [#392](https://forge.adrianpk.com/adrianpk/chr/pulls/392) on `feat/slice-3-telegram-operations`. Verified merge: `697a3ba4b3e33c459c343e140d0337860f453918`, 2026-10-16; the controller receipt records all four ordered checks passed on that exact clean HEAD.

## Slice 4: Conversational Authoring

State: delivered

Prerequisite: verified Slice 3 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T4.1 | done | `b54d526fd0471eff7c4df2ef769ba0abe66476ce` | Bounded typed natural-language intent and deterministic command parity. |
| T4.2 | done | `f7408ea2fa51d4a66ce5be156d2aa192160ad37f` | Real configured-provider authoring with exact question/turn binding. |
| T4.3 | done | `4d1932fedf6ac79e659ce2b475b4261b822f50e6` | Explain/prepare/develop intent, retained scope and bounded authority. |
| T4.4 | done | `e4aa7bf44bc36ba8a5f1ca0edb82978f79bb0923`, `37a1a0aad92404cf086bf202f812bde81d39f333`, `392ce4b1f8bb4180ce6f227ac3340aacaec29b77` | Canonical artifacts, native preparation, exact scope decisions and replay checks. |

Validation: plan section 9; `TestTelegramAuthoring`, including focused race mode.

T4.1 task checks: the combined `TestTelegramAuthoring`, conversation, operation,
receiver and migration selection passed in normal and race modes. Strict focused
lint reported zero issues. The [Slice 4 draft report](../report/slices/telegram-conversational-workflow/slice-4-authoring.md)
records exact commands and fixture boundaries.

T4.2 task checks: configured-provider authoring and existing shared/CLI authoring
regressions passed in normal and race modes; strict focused lint, Go signature,
package README and diff checks passed. Fixtures cover real prompts/answers,
generated artifacts, guarded refinement, interrupted journal recovery, original
receipt replay, independent owner exclusion and retained provider attempts.

T4.3 task checks: combined Telegram authoring/conversation/operation/receiver
regressions passed in normal and race modes. Additional original-instruction
recovery checks passed in both modes after that addition; strict focused lint,
Go signature, package README and diff checks passed. Shared-operation fixtures
prove real explanations, native backlog descriptions with canonical scope/authority,
zero preparation invocations, guarded sources and original receipt replay.
T4.4 adapter checks: the combined Telegram regression passed in normal
and race modes, including real HTTP generation/refinement, canonical questions,
slash/NL parity, exact raw capture/answers, lost-acknowledgement recovery and
post-generation authorization guards. Strict focused lint, Go signature, package
README and diff checks passed. T4.4 also connects slash/NL explanations to frozen
canonical context, with independent retained provider attempts/results, guarded
delivery and no artifact/work effects. Its real HTTP/reopened-store fixtures
cover malformed output exhaustion, unchanged sources/intent, original result
replay and current operator/root checks in both normal and race modes. Native
prepared-work and scope decisions now pass the same combined normal/race checks,
strict lint and style checks. Real native-owner fixtures prove first-scope
authority, changed/missing authority, button/reply/text parity, stale prompts and
lost-acknowledgement recovery preserving independent source/ticket edits.

Implementation: `392ce4b1f8bb4180ce6f227ac3340aacaec29b77`. Canonical PR: [#393](https://forge.adrianpk.com/adrianpk/chr/pulls/393). Verified merge: `ca9fb09cb27d075aae3e24e6122b2679796385ec`, 2026-10-16; all four ordered controller checks passed at that exact clean HEAD.

## Slice 5: Native Work Continuation

State: delivered

Prerequisite: verified Slice 4 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T5.1 | done | `4456f3f1bbef9b0b8368294f12e40c9cc9ca15ff` | Exact native previews/admission for all five extents and RequestKey replay. |
| T5.2 | done | `d94c581ded73f5a5fbe9ba45192097ebf1634ee0` | Bounded asynchronous continuation under existing workflow ownership. |
| T5.3 | done | `3eb2a912f0c13058906b3e62ffc3316b14e172f4` | Unattended policy-driven progress, review and integration. |
| T5.4 | done | `b94e60267f7b8eeafb21e93f3cd653fafad1de1a` | Evidence-bound results and shared CLI/TUI/Web invocation references. |

Validation: plan section 10; `TestTelegramWork`, including focused race mode.

T5.1 task checks: all five native extent previews/admissions, typed natural-language parity, changed source/policy/target rejection, root/operator guards, frozen namespace and lost native acknowledgement recovery passed. The combined Telegram/shared-work regression and its focused race selection passed; final `TestTelegramWork` normal/race checks, strict focused lint and repository style checks passed. The [Slice 5 report](../report/slices/telegram-conversational-workflow/slice-5-work.md) records exact commands and fixture boundaries.

T5.2 task checks: all five extents execute one bounded native checkpoint; the real foreground receiver answers HTTP status during a blocked worker. Shared live owners, operator stops, human review, merger decisions, capacity waits and reopened abandoned claims retain native fences and budgets. Combined Telegram/shared/daemon normal regressions and the focused concurrency selection passed under race instrumentation. Strict focused lint and repository style checks passed; exact commands and fixture boundaries are in the [Slice 5 report](../report/slices/telegram-conversational-workflow/slice-5-work.md).

T5.3 task checks: actual foreground intake reaches trusted review, one guarded publication to a temporary bare target and confirmed native receipt without another message. Transactional phase/outbox projection coalesces never-attempted routine progress, enforces the thirty-second interval at delivery reservation, preserves uncertain sends and terminal identity, and rolls back on capacity exhaustion. Focused normal/race and existing receiver/conversation/operation regressions passed; strict focused lint and repository style checks passed. The report records exact commands and evidence boundaries.

T5.4 task checks: real CLI, TUI and Web reads expose the same Telegram-created invocation and accepted native evidence without writing work. Actual human review projects its candidate, open PR and exact approval hash without claiming integration. An uncertain HTTP send followed by SQLite reopen repeats only the labelled notice and preserves canonical workflow bytes. Focused normal/race, strict focused lint and repository style checks passed. User Guide and the existing Telegram changelog entry describe the usable workflow.

Canonical PR: [#394](https://forge.adrianpk.com/adrianpk/chr/pulls/394) on `feat/slice-5-telegram-work`. Verified merge: `c321da6917e369049583eece5a1a07951c1cc22b`, 2026-10-16; the controller receipt records all four ordered checks passed on that exact clean HEAD.

## Slice 6: Guarded Controls and Capacity

State: in_progress

Prerequisite: verified Slice 5 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T6.1 | validated | `a3e5e95be3b45aaa66cc0c1bc0744b73adac6770` | Shared guarded pause/cancel/resume/continue/reconcile, mapped actor, exact invocation/revision and asynchronous stop checkpoints. |
| T6.2 | validated | `76b55b28ed770b13fc40da53ff39e42dbe822ae9` | Exact review/candidate, entry/budget and question guards; atomic native request receipts; lost-ack and partial-journal recovery; mapped actors. |
| T6.3 | validated | `aaca31360292b67e72509557fa4b650021593aaf`, `b2f03c1159ea56ab4e482a2519489b1eec42e736`, `dcf386f2423a669707d1609a92eeee9a784453a5`, `27165b596e763e5117177a83592044540855a93f` | Bounded asynchronous intake/application/delivery, shared model slots, 32 reserved pending inputs, callback/text/reply priority, native lifecycle receipts and guarded recovery through expired/replaced prompts; late reply binding, guarded current-state diagnostics, 64 reserved control request records and durable capacity rejection are validated. |
| T6.4 | ready | pending | Shared serial/parallel capacity and cross-adapter ownership checks. |

Validation: plan section 11; `TestTelegramControl`, including focused race mode.

T6.3 progress: the receiver shares 16 application slots across inbox and native
work calls, with separate decision and urgent control capacity. A real HTTP
model-load fixture proves status and draft cancellation remain responsive.
SQLite fixtures cover the 32-input reserve within the 1,024 ceiling, atomic
capacity rejection, priority restoration, two shared model slots and context
guards. Native lifecycle receipts now recover interrupted stop/resume, task/closing
reconciliation and continuation acceptance without repeating work. Recovery
now includes receipt-only recovery through expired/replaced Telegram prompts
and text/callback/reply parity for accepted and unaccepted stale actions.
T6.3 completion adds late reply binding, guarded stale-state diagnostics,
reserved request metadata, durable rejection and real activation-bound evidence.
Focused normal/race and strict checks passed; T6.4 is ready.
Exact commands and results belong to the [Slice 6 report](../report/slices/telegram-conversational-workflow/slice-6-controls.md).

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
| TG-04 | validated | [Slice 4 authoring and scope fixtures](../report/slices/telegram-conversational-workflow/slice-4-authoring.md) |
| TG-05 | validated | [Slice 4 authoring and scope fixtures](../report/slices/telegram-conversational-workflow/slice-4-authoring.md) |
| TG-06 | validated | [Slice 2 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-2-conversations.md) |
| TG-07 | validated | [Slice 2 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-2-conversations.md) |
| TG-08 | validated | [Slice 2 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-2-conversations.md) |
| TG-09 | validated | [Slice 4 authoring and scope fixtures](../report/slices/telegram-conversational-workflow/slice-4-authoring.md) |
| TG-10 | validated | [Slice 4 authoring and scope fixtures](../report/slices/telegram-conversational-workflow/slice-4-authoring.md) |
| TG-11 | validated | [Slice 5 exact extent fixtures](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
| TG-12 | validated | [Slice 5 admission replay and guard fixtures](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
| TG-13 | validated | [Slice 5 unattended foreground integration fixture](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
| TG-14 | pending | pending |
| TG-15 | pending | pending |
| TG-16 | pending | pending |
| TG-17 | pending | pending |
| TG-18 | validated | [Slice 3 owner, SQLite and HTTP crash fixtures](../report/slices/telegram-conversational-workflow/slice-3-operations.md) |
| TG-19 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-20 | validated | [Slice 3 uncertain result and prompt delivery fixtures](../report/slices/telegram-conversational-workflow/slice-3-operations.md) |
| TG-21 | pending | pending |
| TG-22 | pending | pending |
| TG-23 | pending | pending |
| TG-24 | pending | pending |
| TG-25 | pending | [Slice 2 session/topic evidence](../report/slices/telegram-conversational-workflow/slice-2-conversations.md); broader journey remains pending |
| TG-26 | validated | [Slice 5 validation, human wait and confirmed receipt projections](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
| TG-27 | pending | pending |
| TG-28 | pending | pending |
| TG-29 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-30 | validated | [Slice 5 CLI, TUI and Web observation fixture](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
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

Current state: Slices 1 through 5 are delivered through canonical PR #390/#391/#392/#393/#394; the latest verified merge is `c321da6917e369049583eece5a1a07951c1cc22b`. Slice 6 is active on its recorded branch. T6.1 through T6.3 have focused lifecycle, authority, durable receipts, stale-reply, activation-load and reserved-capacity evidence. T6.4 shared cross-adapter capacity certification is ready. Final Slice 6 controller validation, live activation and exhaustive certification remain pending.
