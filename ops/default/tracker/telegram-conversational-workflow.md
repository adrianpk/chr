# Telegram Conversation Tracker

Status: Blocked

Repair integration checkpoint: 2026-10-21. The maintainer authorized integrating
the pending admission and executor handoff repairs through the canonical Slice 8
PR. The report is reviewing for that bounded change. T8.1-T8.3, real journey
evidence and final closure remain pending; canonical repair integration alone
does not complete those tasks or change the retained QA invocations.

Repair PR: [#426](https://forge.adrianpk.com/adrianpk/chr/pulls/426).

Date: 2026-10-16

Approved: maintainer, 2026-10-16; specification scope and delivery design accepted through the request for spec, plan and tracker.

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../plan/telegram-conversational-workflow.md)

Spec: [Telegram Conversation and Workflow Control](../spec/stable/telegram-conversational-workflow.md)

Model: [Telegram Conversation and Workflow Model](../spec/stable/telegram-conversational-workflow-model.md)

Implementation authorization: Granted by the maintainer on 2026-10-16 for this exact set, canonical slice PRs and Looper-owned Rebase + Fast-forward merges into dev.

Active slice: Slice 8, Live journey and closure; first live canonical dump confirmed; blocked on remaining live input and evidence

Next slice: None; final slice active, followed by exact-candidate scheduled certification and closure

Progress: 7 of 8 slices integrated; 26 of 29 tasks done and 3 not delivered; 28 criteria have deterministic fixture evidence, TG-01 has live reception evidence and 3 criteria remain pending.

Full gate: `make verify`

Exhaustive certification: pending the scheduled nightly for the final integrated candidate.

## State and Evidence Rules

Task states: `pending`, `ready`, `in_progress`, `validated`, `blocked`, `done`.
Slice 1 through Slice 7 tasks are done after verified PR #390/#391/#392/#393/#394/#395/#396 integration. Slice 8 is active;
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

State: delivered

Prerequisite: verified Slice 5 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T6.1 | done | `a3e5e95be3b45aaa66cc0c1bc0744b73adac6770` | Shared guarded pause/cancel/resume/continue/reconcile, mapped actor, exact invocation/revision and asynchronous stop checkpoints. |
| T6.2 | done | `76b55b28ed770b13fc40da53ff39e42dbe822ae9` | Exact review/candidate, entry/budget and question guards; atomic native request receipts; lost-ack and partial-journal recovery; mapped actors. |
| T6.3 | done | `aaca31360292b67e72509557fa4b650021593aaf`, `b2f03c1159ea56ab4e482a2519489b1eec42e736`, `dcf386f2423a669707d1609a92eeee9a784453a5`, `27165b596e763e5117177a83592044540855a93f`, `8476423bc7aa84d116273e7df120a7916b96284f`, `8b01210ac76dd5a20cbc14e4ec2e34472adc723a`, `a09bf959dc8770174fb6987849c5a104126dbdd9` | Bounded asynchronous intake/application/delivery, shared model slots, 32 reserved pending inputs, callback/text/reply priority, native lifecycle receipts and guarded recovery through expired/replaced prompts; late reply binding, guarded current-state diagnostics, 64 reserved control request records and durable capacity rejection are validated. |
| T6.4 | done | `8346ab32efee666563fddc19ee335b4d86915e6d`, `a4faf4367bdaecf28577d4ed569da7d24952e725` | Two Telegram conversations, daemon and manual callers share serial/default and eight-worker native limits; actual source overlap, duplicate rejection, capacity reuse, linked-root reservations and target publication exclusion are certified. Claimed-worker stop polls retain owner and credential guards without repository-wide invocation enumeration. |

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
Focused normal/race and strict checks passed. T6.4 adds actual shared-worker overlap, overflow and duplicate rejection, slot reuse and native target publication exclusion. Stop polling uses the claimed owner under the existing repository/journal guards and verifies its credential; identity-only observers retain public scope resolution.
Exact commands and results belong to the [Slice 6 report](../report/slices/telegram-conversational-workflow/slice-6-controls.md).

Candidate: `feat/slice-6-telegram-controls`; PR: [#395](https://forge.adrianpk.com/adrianpk/chr/pulls/395); verified merge: `eecdfd1e85c9593715c5e00502d0bfd0bf23c1ea`. Controller receipt `62QTJGFYZPHBQEWSAVCVDYQUYM` passed all four configured commands at that exact HEAD. Normal core/app: 34.374/47.842 seconds; race core/app: 138.627/128.905 seconds.

## Slice 7: Recovery and User Operation

State: delivered

Prerequisite: verified Slice 6 integration.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T7.1 | done | `0c2fee3f1ad68bdba578bae9cd75605057d53b05`, `8fd3713efaf1` | Startup receipt replay, dead-owner recovery, bounded owned shutdown, independent worker/validation preservation and fixed running-image proof. |
| T7.2 | done | `b2647940ccdd` | Cross-restart retry budgets, attributed offline cycles, gap epochs, exact retirement and bounded authority-preserving retention. |
| T7.3 | done | `5546d2ab4b7726951ba885c8932ca88bdc3ea1b7` | Optional user service, local retry/status/doctor/help and operational diagnostics. |

Validation: plan section 12; `TestTelegramRecovery`, including focused race mode.

Canonical PR: [#396](https://forge.adrianpk.com/adrianpk/chr/pulls/396). Verified merge: `c14f1aea640657e27c97bb64e40b46f30f698451`, 2026-10-16T22:44:10Z. Controller receipt `JR526N5RLBBVIBT62W5YOZMNB6` passed all four ordered checks at that exact clean HEAD; normal core/app 2.719/21.135 seconds, race 56.358/57.946 seconds. Acta verified merge event: `01M53E66GA9C9VH1CKRKYRDM5B`. Evidence: [Slice 7 report](../report/slices/telegram-conversational-workflow/slice-7-recovery.md).

## Slice 8: Live Journey and Closure

State: blocked

Prerequisite: verified Slice 7 integration plus private live readiness and
recorded bot/model/service/QA activation authority from plan section 2.

| Task | State | Commit | Work |
| --- | --- | --- | --- |
| T8.1 | blocked | `b72a6955cf7c` admission correction; executor handoff correction prepared | Genuine clarification, spec generation, TGQA-1 preparation and exact admission reached a real Codex worker. The worker made the correct edit without a commit; native execution paused before host validation. Scoped commit instructions, their regression and make verify pass locally; receiver activation, live integration and a complete natural-language journey remain pending. |
| T8.2 | blocked | isolated source QA setup complete | The source checkout is separately registered against its own QA target; the receiver is fixed at `14c563d40f6a`. Authorized-user input for the actual source-development journey is missing. |
| T8.3 | in_progress | `b72a6955cf7c` User Guide/reference/changelog; installed help inspected; acceptance traceability audited | All 32 criterion entries, report references and seven canonical merge identities are checked. Complete live evidence, the cycle report and nightly-bound closure remain. |

Validation: plan section 13; `TestTelegramJourney`, including applicable focused
race mode, separate authorized live proof and final exact-candidate nightly.

Candidate: `test/slice-8-telegram-delivery`; PR and merge pending. The [Slice 8 report](../report/slices/telegram-conversational-workflow/slice-8-delivery.md) separates deterministic preparation from live readiness and scheduled certification.

Live progress: `tgc-38356dda2b04b98de3d419985fa97481` is bound to
`telegram-journey-qa`. The configured user's real `/develop` request and its
delivered project-bound acknowledgement validate TG-01. A genuine `/done`
confirmed `authoring.continue` and created the canonical dump. The remaining
live phases require model clarification, work preparation and any bound
admission answers before native integration. The second source QA request is
still pending. Keep TG-27, TG-28 and TG-32 pending until their complete live
journeys and evidence exist. The existing receiver is managed outside this
worker and remains active with its fixed image.

Task validation: combined journey, prepared-work, admission and automatic
continuation regressions pass in normal and race modes. The actual asynchronous
receiver reaches a confirmed fixture receipt after preparation and SQLite reopen.
Strict lint reports zero issues. The report records exact commands, timings and
the live-evidence boundary. TG-27, TG-28 and TG-32 remain pending.

Audit, 2026-10-19: the retained controller reports 7/8 slices integrated and
stopped after repeated turns without observable progress. Its later passing
local validation does not establish the required live journeys or Slice 8
delivery. Preserve the checkpoint, conversation and budgets; this status audit
does not restart execution. No Slice 8 report or verified integration is
present on `dev`.

## Acceptance Evidence

Traceability audit at `0274135cc98b4640624ab6fc634cb1b6bc69d6d5`: all 32
criteria appear exactly once, report links and anchors resolve, and the seven
delivered slice identities and merge commits match the canonical forge. Their
configured test selections remain nonempty in the candidate source. This audit
preserves the existing fixture/live evidence boundary and does not rerun tests.

The plan's section 14 maps every criterion to its primary tasks. Add immutable
report/test/live references in the Evidence column as they are verified.

| Criterion | State | Evidence |
| --- | --- | --- |
| TG-01 | validated | [Slice 8 live reception](../report/slices/telegram-conversational-workflow/slice-8-delivery.md#live-qa-progress); conversation `tgc-38356dda2b04b98de3d419985fa97481`, fixed receiver `14c563d40f6a3bc0c41f7e3a8b4fc52e1c2cb166`. |
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
| TG-14 | validated | [Slice 6 exact review and candidate guards](../report/slices/telegram-conversational-workflow/slice-6-controls.md) |
| TG-15 | validated | [Slice 6 native lifecycle, grant and decision receipts](../report/slices/telegram-conversational-workflow/slice-6-controls.md) |
| TG-16 | validated | [Slice 6 shared capacity and publication exclusion](../report/slices/telegram-conversational-workflow/slice-6-controls.md) |
| TG-17 | validated | [Slice 6 bounded control activation and model load](../report/slices/telegram-conversational-workflow/slice-6-controls.md) |
| TG-18 | validated | [Slice 3 owner, SQLite and HTTP crash fixtures](../report/slices/telegram-conversational-workflow/slice-3-operations.md) |
| TG-19 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-20 | validated | [Slice 3 uncertain result and prompt delivery fixtures](../report/slices/telegram-conversational-workflow/slice-3-operations.md) |
| TG-21 | validated | Slice 7 reopened poll/model/delivery exhaustion and explicit transport retry fixtures; `b2647940ccdd`. |
| TG-22 | validated | [Slice 7 startup/shutdown ownership and restart evidence](../report/slices/telegram-conversational-workflow/slice-7-recovery.md) |
| TG-23 | validated | [Slice 7 real HTTP identity/webhook/remote conflicts and local exclusion](../report/slices/telegram-conversational-workflow/slice-7-recovery.md) |
| TG-24 | validated | Slice 7 retention-gap HTTP poll, exact sparse retirement, safe compaction and payload/receipt/pending/outbox boundaries; `b2647940ccdd`. |
| TG-25 | validated | [Slice 7 combined Unicode/reply/topic/replaced-prompt and uncertain-ack race evidence](../report/slices/telegram-conversational-workflow/slice-7-recovery.md), with earlier conversation/operation fixtures. |
| TG-26 | validated | [Slice 5 validation, human wait and confirmed receipt projections](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
| TG-27 | pending | pending |
| TG-28 | pending | pending |
| TG-29 | validated | [Slice 1 deterministic evidence](../report/slices/telegram-conversational-workflow/slice-1-receiver.md) |
| TG-30 | validated | [Slice 5 CLI, TUI and Web observation fixture](../report/slices/telegram-conversational-workflow/slice-5-work.md) |
| TG-31 | validated | [Slice 7 installed command help, provider/readiness and sanitized diagnostics](../report/slices/telegram-conversational-workflow/slice-7-recovery.md) |
| TG-32 | pending | pending |

## External Prerequisite Register

| Requirement | State | Evidence or action |
| --- | --- | --- |
| Looper execution authority | granted | Maintainer instructed this exact prepared set to proceed on 2026-10-16. |
| Dedicated bot credentials and exact identity bindings | isolated QA binding and fixed executable ready | Connection doctor succeeds; an isolated dev-root permits only the two disposable QA clones. Native work doctors report ready. Fixture details are recorded in the Slice 8 report. |
| Live service/messages/model/QA authority | granted for bounded plan validation | Use the dedicated bot, configured model and existing designated QA repositories; private readiness remains required. |
| Authorized-user live development messages | first request and capture completion received | The genuine QA development request has a delivered project-bound acknowledgement and a confirmed canonical dump. Live model clarification, native admission/integration and the second source QA request remain pending. The externally managed receiver stays active. |
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

Current state: Slices 1 through 7 are delivered through canonical PR #390/#391/#392/#393/#394/#395/#396; the latest verified merge is `c14f1aea640657e27c97bb64e40b46f30f698451`. Slice 8's combined fixture and preparation/admission correction pass focused normal/race checks. The fixed receiver, two registered QA clones and isolated target refs are prepared. Authorized-user development messages, real model/integration evidence and final scheduled certification remain pending. No Slice 8 PR is ready.

## Continuation Audit: 2026-10-20

Current `dev` through `6037036b2fba` is synchronized into the Slice 8 branch.
The capture evidence and pending admission correction are preserved alongside
the canonical merged Slice 7 evidence. The receiver remains on its fixed
`14c563d40f6a` executable and matching private descriptor digest. Read-only
status reports ready, one conversation, no pending updates/deliveries and no
exhausted model requests.

The retained Looper run `run-5ffcddf0ca78104f80d09ad5` remains blocked after
repeated turns without observable progress. Its last exact-head validation
passed on `4e4fbf8fadef`; that receipt does not validate this newer composition.
No controller is resumed, reset or reconfigured by this synchronization.

Remaining dependencies are genuine configured-user conversation input, real
model clarification and ordinary QA integration, separately bound source
self-development, complete live/visual evidence and the scheduled exhaustive
certificate for the final integrated candidate. The source message-review
report is available at `ops/default/report/telegram-message-review.md`; its
replacement proposals are separate from this slice's runtime scope.
