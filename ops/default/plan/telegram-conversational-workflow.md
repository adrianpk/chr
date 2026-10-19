# Telegram Conversation Delivery

Status: Blocked

Date: 2026-10-16

Approved: maintainer, 2026-10-16; specification scope and delivery design accepted through the request for spec, plan and tracker.

Delivery set: telegram-conversational-workflow

Spec: [Telegram Conversation and Workflow Control](../spec/stable/telegram-conversational-workflow.md)

Model: [Telegram Conversation and Workflow Model](../spec/stable/telegram-conversational-workflow-model.md)

Tracker: [Telegram Conversation Tracker](../tracker/telegram-conversational-workflow.md)

Implementation authorization: Granted by the maintainer on 2026-10-16 for this exact set, its canonical slice PRs and Looper-owned Rebase + Fast-forward merges into dev.

Active slice: Slice 8, Live journey and closure; pending delivery, retained controller blocked

Next slice: none; Slice 8 is the final pending unit

Slice strategy: behavior-first

Reason: extend the existing application and workflow owners through usable Telegram capabilities. Each slice delivers its contracts, persistence, behavior and regression checks together; no separate skeleton or intentionally failing test slice is needed.

Full gate: `make verify`

Exhaustive certification: the existing scheduled Forgejo nightly, running `make nightly` against the exact integrated `dev` candidate.

## 1. Outcome and Scope

Deliver one complete text conversation adapter: configure a dedicated bot,
discuss a change, prepare canonical artifacts, authorize an exact native work
extent, retain progress and controls, and observe actual integration. Ordinary
language and explicit commands use the same typed application operations.

The accepted specification and model remain normative. All 32 acceptance
criteria are assigned below. Serial execution remains the default; configured
parallel projects retain the existing shared capacity and integration queue.
Developing Chryso through this adapter is supported and preferred when useful,
without making that workflow mandatory.

This set contains eight ordered slices and 29 tracker tasks. A task is a
commit-sized unit when practical; a slice is the branch and PR unit. The linked
tracker records the task list, commits, validation and integration. Do not
create another ticket batch merely to duplicate these planned tasks.

The maintainer first requested the planning artifacts, then instructed this
set to proceed on 2026-10-16. That instruction authorizes implementation,
the recorded slice lifecycle and its bounded validation. Under Looper, the controller owns canonical slice
PR integration into `dev`; workers prepare candidates and do not merge them.
Keep one development worker and at most one open canonical slice PR for this
set. Product-level parallel execution is exercised through isolated fixtures.

## 2. Entry Conditions and External Prerequisites

1. Commit the spec, model, plan and tracker together on `dev` before execution.
   Re-read the current repository rules, installed `looper --help`, Git state
   and existing controller ownership before configuring the future run.
2. Planning baseline: `644750de3295450284b0d985ea137a3ce5adf56a`.
   The native work engine, shared application service, authoring operations,
   existing notification sender and saved-update importer are already present.
   Verify those contracts at activation rather than rebuilding their owners.
3. The live receiver, conversation tables, real conversational intent resolver
   and its complete shared authoring path are not yet delivered. Existing
   notification sending and saved-update parsing are not receiver evidence.
4. Slices 1-7 use deterministic HTTP/model boundaries, real temporary SQLite
   databases and project stores. They do not require a bot secret, external
   repository writes or a live model subscription to pass their regressions.
5. The maintainer has a dedicated new Chryso bot. Its token, allowed numeric
   identities and operator/project mapping must be supplied through private
   local setup before live activation. Never request or commit its token in
   chat, artifacts, argv, PR bodies or test output.
6. Execution authorization includes the bounded Slice 8 validation: dedicated
   receiver/service activation, live bot messages, configured model use and
   writes to the already designated disposable QA repositories. Use host-owned
   access and supplied private bot/operator/project configuration. Missing
   readiness is an external blocker, never fixture evidence substituted for
   live evidence. Do not create accounts, repositories or credentials under
   this plan without their own explicit authorization.
7. The full scheduled gate must be ready for the final candidate. Scheduling
   and runner setup remain owned by the existing nightly workflow. Preserve
   unrelated schedule changes and active validations; this set does not alter
   other projects' schedules or run an exhaustive suite interactively.

No important product behavior decision remains open for planning. Private live
configuration and activation authority are execution prerequisites, not an
excuse to defer the deterministic implementation slices.

## 3. Ownership and Delivery Rules

- Keep command assembly and adapter-facing service operations in `internal/app`;
  keep reusable domain behavior and persistence with their existing
  `internal/core` owners. Put transport-specific mappings in dedicated files.
- Extend the current user database with additive migrations. Preserve the
  current registry, existing rows, `--dev-root` behavior and future-schema
  refusal. Do not introduce a separate project registry or transport scheduler.
- Use caller-neutral request bindings under each shared mutation owner.
  SQLite acknowledgement, a process-local map or a Telegram send receipt is
  insufficient proof that a domain mutation happened exactly once.
- Call typed service methods. Never dispatch CLI strings, shell commands,
  model-supplied tool names or another development controller from the bot.
- Before exposing any mutating operation, deliver its authentication, exact
  target guard, durable request identity, shared receipt and recovery path.
  Report not-yet-delivered operations accurately during intermediate slices.
- Keep every slice compiling and passing its checks. Do not use empty test
  selections, skipped assertions or unimplemented success responses to satisfy
  an acceptance criterion.
- Use the exact identities below. One PR targets `dev` for each slice. Use one
  Conventional Commit per tracker task when practical and record exceptions.
  Additional report/reference commits remain traceable to that slice.
- Keep reports at the recorded paths; create each report with its real slice,
  not as a prefilled success artifact. Record tested HEAD, command order,
  evidence boundary, canonical PR and verified integration.
- Follow the repository slice PR template: `Summary`, `Changes`, `Validation`.
  End Changes with an immutable report link and Validation with
  `git diff --check`. List only commands that actually passed.
- A focused defect correction belongs to the active task. Material scope or
  slice-identity changes require an explicit plan revision before execution.
  `main`, release tags and publication are outside this delivery authority.

## 4. Canonical Slice Map

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

## 5. Validation and Completion

Before a task commit, format changed Go files and run the smallest relevant
behavioral regression. Do not invent a new exhaustive campaign. Use real
temporary storage for persistence assertions and deterministic local servers
for Telegram HTTP boundaries. Assert effects and observable state rather than
duplicating implementation details.

For each clean immutable slice candidate, record these commands in order:

1. `make verify`.
2. The focused non-short selection specified for that slice.
3. The corresponding focused race selection when shared ownership,
   asynchronous work, shutdown or database concurrency changed.
4. `git diff --check` last.

New test-family prefixes below are implementation obligations. Use concise,
named table cases and repository-required explanatory comments. The report
must show that each selected family executed its required scenarios; zero
matches do not count. Add existing affected tests to the same bounded selection
when needed. Do not blanket-skip cheap Telegram checks in short mode.

The future Looper configuration uses `--gate 'make verify'`. Never put
`go test ./...`, `go test -race ./...`, full acceptance campaigns or `make nightly`
in the interactive controller gate. Reuse validation only for the same exact
HEAD and ordered command list; preserve an active validation owner.

After the last slice integrates, record the exact `dev` candidate and wait for
its scheduled Forgejo nightly. Keep the set `awaiting_nightly` until that
candidate is green. A controller can park at that checkpoint; do not manufacture
progress or start another exhaustive run while waiting. Local Looper success
and passing `make verify` do not establish exhaustive certification.

If the nightly requires code changes, use the prescribed
`fix/telegram-conversational-workflow-validation` branch and one corrective PR
to `dev`; run focused checks and wait for the next scheduled nightly on the
corrected candidate. Do not add a ninth slice automatically. Close the set only
after recording the exact green candidate and real live evidence. Later
closure-only document commits must identify the tested candidate and may not
change runtime, test or gate inputs.

## 6. Slice 1: Dedicated Live Receiver

Prerequisite: committed planning baseline and execution authorization.

Deliver a useful, authenticated read-only bot through the real transport
adapter, with a durable polling cursor. Mutating project operations remain
unavailable until their shared replay contract is delivered.

| Task | Required work |
| --- | --- |
| T1.1 | Implement private setup/config loading, masked token input, environment references, explicit replacement, local status and connection doctor. Preserve existing configuration and `--dev-root`; setup does not start work or send messages. |
| T1.2 | Implement bounded long polling, getMe/getWebhookInfo readiness, message/callback decoding, exact sender/chat/topic authorization and bot-scoped OS ownership. Reject unsupported identities before interpretation or lookup. Never remove webhooks or take over a receiver. |
| T1.3 | Add versioned receiver/inbox storage and the initial durable read-response outbox. Commit complete batches and next offsets atomically, preserve sparse IDs/digests, enforce intake bounds and retain poll/delivery retry state across reopen. |
| T1.4 | Wire help, authorized project listing and receiver status through a typed operation registry and the central service. Produce bounded text responses through the outbox; preserve existing notification and fixture-import behavior. |

Acceptance: TG-02, TG-03, TG-19 and TG-29; transport foundation for TG-01/TG-23.

Required focused command:
`go test ./internal/core ./internal/app -run '^Test(TelegramReceiver|TelegramSender|RenderTelegramNotification|DecodeTelegramUpdates|RemoteCommandRecordFromTelegramUpdate)' -count=1 -timeout=3m`.

Run the same selection with `-race`. Prove invalid credentials/configuration,
private/group authorization, oversized responses, cursor transaction rollback,
duplicate/sparse updates, reopening and singleton contention. Confirm that
saved handoffs and help cannot admit work.

## 7. Slice 2: Conversation Routing

Prerequisite: Slice 1 integrated.

Deliver independent project-bound conversations, deterministic targeting and
bounded text capture without project effects.

| Task | Required work |
| --- | --- |
| T2.1 | Persist peer selection, conversation phase and frozen project/root/namespace/intent. Implement exact project precedence, project mismatch diagnostics, sessions/focus and ambiguity handling; recheck current allowlists and registry bindings. |
| T2.2 | Bind replies to exact peers, topics, authoring questions and conversation revisions. Implement multi-part capture, done/cancel and duplicate-safe append within SQLite transactions; invalidate replaced/terminal prompts. |
| T2.3 | Add deterministic conversation commands, generated registry help and bounded session/capture views. Preserve separate questions across concurrent peers and sessions; unknown slash commands never call a model. |

Acceptance: TG-06, TG-07 and TG-08; session/topic part of TG-25.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramConversation' -count=1 -timeout=3m`.

Run the same selection with `-race`. Cover every targeting precedence, two
projects, two users sharing a group, separate topics, multiple pending sessions,
root changes, stale reply bindings and 16/64-KiB plus 32-part boundaries.

## 8. Slice 3: Replay-safe Operations

Prerequisite: Slice 2 integrated.

Deliver useful record operations with shared-owner replay guarantees before
any conversation can mutate canonical project artifacts.

| Task | Required work |
| --- | --- |
| T3.1 | Extend shared application/domain mutation contracts with opaque request identities and normalized argument digests. Reserve artifact IDs before writes, journal observed before/intended after images under existing owner exclusion and reconcile the same receipt after a crash. Keep transport metadata out of domain records. |
| T3.2 | Expose the specified list/show/create/update operations for tickets, slices, slice sets and use cases through their existing owners. Require actual lifecycle authority and observed revisions; bound lists and return typed unsupported/conflict results accurately. |
| T3.3 | Complete durable outbox delivery, exact message/reply associations and single-use action records. Share callback and text-answer request bindings, enforce expiry/revision/peer guards and split plain text within both Unicode limits. Bound retries and label uncertain sends. |
| T3.4 | Prove crash recovery between domain write, receipt, SQLite acknowledgement and send acknowledgement. Verify changed-argument conflicts, no duplicated artifact/grant/action, pending-record preservation and caller-neutral reuse of the shared replay contract. |

Acceptance: TG-18 and TG-20; record-mutation and output foundation for
TG-04/TG-14/TG-15/TG-25.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramOperation' -count=1 -timeout=3m`.

Run the same selection with `-race`. Use real reopened databases and project
stores, including a successful write whose acknowledgement is lost, a divergent
journal image, duplicate callbacks, uncertain delivery and multibyte chunking.
No SQLite transaction may span an external model, agent or forge call.

## 9. Slice 4: Conversational Authoring

Prerequisite: Slice 3 integrated.

Deliver natural-language explanation, clarification and real artifact
preparation using configured model providers through shared authoring owners.

| Task | Required work |
| --- | --- |
| T4.1 | Implement bounded typed intent resolution from the same declarative registry as slash commands. Validate schemas/authority outside the model, limit calls/context/deadlines, retain model retry state and preserve deterministic commands during provider failure. |
| T4.2 | Complete adapter-neutral authoring continuations for dump, quiz, spec and plan generation/refinement. Bind real provider prompts and answers to canonical question/turn revisions; remove dependence on fake question or summary inputs from this path. |
| T4.3 | Preserve explain, preparation-only and bounded development intent. Materialize canonical work descriptions and retained scope/authorization, preview changes and ask only for missing or changed authority. Preparation never starts work; execution wiring follows in Slice 5. |
| T4.4 | Connect the preparation journey to existing ops writers with provenance, guarded replacement and request replay. Test actual generated artifacts through a deterministic model HTTP boundary, slash/NL parity and clear failures without losing prior context. |

Acceptance: TG-04, TG-05, TG-09 and TG-10.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramAuthoring' -count=1 -timeout=3m`.

Run the same selection with `-race`. Prove strict malformed/unknown model output,
ambiguous targets, independent questions, provider timeout/capacity failure,
guarded edits, crash/retry after artifact generation and explain/prepare requests
that create zero invocations. No live provider access is needed for these tests.

## 10. Slice 5: Native Work Continuation

Prerequisite: Slice 4 integrated.

Deliver accepted work from conversation to actual shared execution and
automatic review/integration without requiring another Telegram message.

| Task | Required work |
| --- | --- |
| T5.1 | Wire exact preview and StartBoundWorkExtent admission for ticket, slice, slice-set, use-case and loose extents. Retain selection revisions, membership and effective policy; use the existing RequestKey and reject changed input or a stale preview. |
| T5.2 | Host bounded asynchronous calls to the existing continuation service, one shared checkpoint per activation. Reattach/reconcile retained invocations and continue only eligible authorized work, preserving pauses, waits, validation ownership and budgets. Do not create a transport-owned scheduler. |
| T5.3 | Project native progress and policy-driven review/integration into conversations and coalesced outbox events. Automatic policy continues unattended; human policy waits on its exact candidate. Keep status responsive during a long activation. |
| T5.4 | Return evidence-bound outcomes and canonical references observable through CLI/TUI/Web. Distinguish worker exit, checked candidate, open PR, confirmed receipt and remaining extent; prove that notification failure cannot rerun work. |

Acceptance: TG-11, TG-12, TG-13, TG-26 and TG-30; asynchronous part of TG-17.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramWork' -count=1 -timeout=3m`.

Run the same selection with `-race`. Use bounded temporary workflow fixtures
for all five extents, replayed admission, changed membership, automatic review,
human waiting, integration observation, interrupted continuation and another
adapter observing the same invocation. These are not live-forge certification.

## 11. Slice 6: Guarded Controls and Capacity

Prerequisite: Slice 5 integrated.

Deliver responsive exact-target controls and decisions while other
conversations and adapters use the same native capacity.

| Task | Required work |
| --- | --- |
| T6.1 | Wire pause, cancel, resume, continue and reconcile through ControlWorkExtent and its shared contracts. Derive the actor from authorized local mapping, address exact invocations and preserve operator stop intent through asynchronous checkpoints. |
| T6.2 | Wire human review approval, grants and decision answers to exact displayed candidate/review, invocation, integration, budget and question revisions. Reuse shared single-use request receipts; stale controls display current state without silently reapplying. |
| T6.3 | Complete text/callback parity, expiration, replaced prompts and priority handling for status/stop/decision requests under model or worker load. Enforce the reserved control allowance and all caller bounds without bypassing persistence. |
| T6.4 | Verify serial default and configured capacity up to eight across Telegram, daemon and other adapters. Prevent duplicate activations/leases and competing publication; several conversations do not allocate independent engine limits. |

Acceptance: TG-14, TG-15, TG-16 and TG-17; guarded question part of TG-25.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramControl' -count=1 -timeout=3m`.

Run the same selection with `-race`. Cover stale approval after candidate/base
change, duplicate grants/answers, unauthorized callback actors, explicit pauses,
worker/model saturation, concurrent adapter admission and shared target queues.
Do not substitute transport-local mutexes for canonical ownership evidence.

## 12. Slice 7: Recovery and User Operation

Prerequisite: Slice 6 integrated.

Deliver reliable operation of the complete adapter across restart, exhaustion,
retention pressure and installation changes. Earlier slices already provide
their required correctness; this slice proves the combined lifecycle.

| Task | Required work |
| --- | --- |
| T7.1 | Reconcile unfinished requests, retained authoring and invocations on startup. Implement bounded SIGTERM shutdown of intake and owned activations, dead-owner verification and fixed running-build behavior. Preserve other adapters' workers, validations, pauses and budgets. |
| T7.2 | Complete cross-restart poll/model/delivery exhaustion and explicit local retry cycles, remote retention-gap epochs, exact retired-ID intervals and payload/receipt/outbox quotas. Backpressure never deletes unresolved authority or makes retired updates executable. |
| T7.3 | Finish optional user-service installation, retry commands, sanitized status/doctor/help and private configuration documentation. Report bot identity changes, webhook conflicts, another local/remote receiver and project/provider readiness precisely. Service installation never implicitly starts work. |

Acceptance: TG-21, TG-22, TG-23, TG-24 and TG-31; complete TG-25.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramRecovery' -count=1 -timeout=3m`.

Run the same selection with `-race`. Use real reopen/process boundaries where
needed, injected clock/network failures, each specified quota, existing/future
schema versions and a long-outage sparse stream. Exercise shutdown ownership
and failure reporting without signaling unrelated developer services.

## 13. Slice 8: Live Journey and Closure

Prerequisite: Slice 7 integrated, private live setup ready and its explicit
activation/QA authority recorded. Deterministic checks can proceed while a live
prerequisite is blocked; the slice cannot claim delivered without that proof.

| Task | Required work |
| --- | --- |
| T8.1 | Run one bounded authorized live bot/model journey against a designated QA project: request a small real change, clarify, prepare, admit the exact extent and observe native integration. Record build, sanitized request/invocation references, source/base/result and confirmed receipt separately from fixture evidence. |
| T8.2 | Demonstrate the same journey with a separately registered Chryso source checkout in the designated QA repository. Make a small real source/documentation change through the installed receiver, keeping its build fixed while candidate source changes. Use no external development controller for that invocation; never give Chryso and Looper ownership of the same production PR. Label QA integration accurately. |
| T8.3 | Complete installed help, the existing User Guide and reference documentation, a concise reader-facing changelog entry, acceptance evidence and the cycle report. Audit all 32 criteria, integrate the slice, record the exact candidate and close only after its scheduled nightly is green. |

Acceptance: TG-01, TG-27, TG-28 and TG-32, with the complete prior matrix audited.

Required focused command:
`go test ./internal/core ./internal/app -run '^TestTelegramJourney' -count=1 -timeout=3m`.

Run the same selection with `-race` for its asynchronous/recovery scenarios.
Keep a deterministic end-to-end fixture for normal development; the authorized
live demonstration is distinct evidence and is not rerun in every PR/nightly.
The scheduled nightly still runs all applicable existing regression and
designated forge certification suites.

Place the cycle report at
`ops/default/report/telegram-conversational-workflow-cycle.md` when real evidence
exists. It records exact integrated candidate, scheduled run/result, live proof,
all canonical PRs and remaining limitations. Product documentation shows normal
Chryso operation; Looper belongs only to this delivery's development procedure.

## 14. Acceptance Ownership

The owner below must supply concrete evidence. Supporting earlier or later
checks may strengthen it but cannot replace the required result. The tracker
starts with every criterion pending.

| Criterion | Primary tasks | Required evidence boundary |
| --- | --- | --- |
| TG-01 | T8.1 | Authorized real bot reception and correct project-bound reply. |
| TG-02 | T1.1 | Setup preservation, private files and secret-free IO. |
| TG-03 | T1.2 | Rejection before model, lookup disclosure or mutation. |
| TG-04 | T4.1, T4.4 | Slash and natural language resolve to the same typed operation. |
| TG-05 | T4.1, T4.4 | Invalid or ambiguous resolution creates no guessed action. |
| TG-06 | T2.1 | Frozen project bindings and isolated peers. |
| TG-07 | T2.2, T2.3 | Independent sessions and exact reply/focus targeting. |
| TG-08 | T2.2 | Duplicate-safe capture, limits and safe cancellation. |
| TG-09 | T4.2, T4.4 | Real authoring operation/provider path and canonical artifacts. |
| TG-10 | T4.3 | Preparation creates no invocation; bounded development retains authority. |
| TG-11 | T5.1 | All five native extent kinds use exact accepted selections. |
| TG-12 | T5.1 | RequestKey replay attaches; changed arguments fail. |
| TG-13 | T5.2, T5.3 | No-message continuation through automatic integration. |
| TG-14 | T6.2, T6.3 | Human review binds the displayed immutable candidate. |
| TG-15 | T6.1, T6.2 | Shared controls, grants and question-answer guards. |
| TG-16 | T6.4 | Serial/parallel cross-adapter ownership and queue proof. |
| TG-17 | T6.3 | Responsive status/stop during model and worker saturation. |
| TG-18 | T3.1, T3.4 | Reopened stores after each effect/ack crash boundary. |
| TG-19 | T1.3 | Atomic batch/cursor persistence and sparse replay. |
| TG-20 | T3.3, T3.4 | Uncertain delivery repeats no mutation or authority. |
| TG-21 | T7.2 | Retained poll/model/delivery budgets and explicit exhaustion. |
| TG-22 | T7.1 | Restart preserves canonical waits, pauses and validation ownership. |
| TG-23 | T7.3 | Bot/webhook/receiver conflict diagnostics without takeover. |
| TG-24 | T7.2 | Pinned unresolved state and exact retired identity rejection. |
| TG-25 | T3.3, T6.3, T7.2 | Unicode, splitting, topics and stale prompt handling. |
| TG-26 | T5.4 | Actual integration receipt distinguished from intermediate success. |
| TG-27 | T8.1 | Real bounded request-to-integration bot journey. |
| TG-28 | T8.2 | Chryso source development using Chryso with a fixed receiver build. |
| TG-29 | T1.4, T3.4 | Existing sender/import regressions and no historical execution. |
| TG-30 | T5.4 | Existing adapters observe/control the same canonical invocation. |
| TG-31 | T7.3, T8.3 | Installed help, readiness diagnostics and documented bounds. |
| TG-32 | T8.1, T8.2, T8.3 | Exact-build live evidence recorded separately from fixtures. |

## 15. Closure Checklist

- [ ] All 29 tasks have implementation commits and passing focused evidence.
- [ ] All eight canonical slice PRs are verified integrated into `dev`.
- [ ] Every slice report contains real delivered behavior and exact validation.
- [ ] All TG-01 through TG-32 criteria have concrete evidence references.
- [ ] The authorized live and self-development journeys are recorded honestly.
- [ ] The scheduled nightly is green for the exact integrated candidate.
- [ ] The cycle report, tracker and plan distinguish tested code from later closure-only documentation.
- [ ] The delivery set is closed with no pending work or unreported blocker.

Current state, audited 2026-10-19: Slices 1 through 7 are delivered through
canonical PR #390/#391/#392/#393/#394/#395/#396. The Slice 7 merge is
`c14f1aea640657e27c97bb64e40b46f30f698451`; its four ordered controller checks
passed on that exact HEAD. All 26 tasks in those slices are done. Slice 8's
three tasks remain pending; no Slice 8 report or verified delivery is recorded.
The retained controller stopped after repeated turns without observable
progress. This audit does not resume it or change its configuration or budgets.
Live journey evidence and exact-candidate scheduled exhaustive certification
remain required before closure.
