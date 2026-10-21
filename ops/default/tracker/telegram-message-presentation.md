# Telegram Message Presentation Tracker

Status: approved

Approved: maintainer, 2026-10-21; specification, model, slice map and Looper execution accepted.

Full gate: `make verify`

Spec: [Message presentation](../spec/stable/telegram-message-presentation.md)

Model: [Document delivery](../spec/stable/telegram-message-presentation-model.md)

Plan: [Slice map and validation](../plan/telegram-message-presentation.md)

Inventory: [48-family review](../report/telegram-message-review.md)

Implementation and merge authorization: granted for this delivery set and
Looper-owned Rebase + Fast-forward slice PR merges into dev.
Active slice: Slice 1 of 5, Readable text and conversation messages.

## Canonical Slice Map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Readable text and conversation messages | feat/slice-1-telegram-message-composition | feat(slice-1): compose readable Telegram messages | ops/default/report/slices/telegram-message-presentation/slice-1-message-composition.md |
| Slice 2 | Document excerpts and complete attachments | feat/slice-2-telegram-document-delivery | feat(slice-2): deliver Telegram document previews and attachments | ops/default/report/slices/telegram-message-presentation/slice-2-document-delivery.md |
| Slice 3 | Clear guarded confirmations and controls | feat/slice-3-telegram-action-presentation | feat(slice-3): clarify Telegram confirmations and controls | ops/default/report/slices/telegram-message-presentation/slice-3-action-presentation.md |
| Slice 4 | Progress, notifications and failure consistency | feat/slice-4-telegram-status-presentation | feat(slice-4): unify Telegram progress and notification presentation | ops/default/report/slices/telegram-message-presentation/slice-4-status-presentation.md |
| Slice 5 | Complete coverage and real client acceptance | test/slice-5-telegram-presentation-acceptance | test(slice-5): certify Telegram message presentation | ops/default/report/slices/telegram-message-presentation/slice-5-presentation-acceptance.md |

## Tasks

All entries are approved obligations, with no implementation commit, PR, merge
or runtime validation claimed. The plan owns exact branch names, PR titles,
report paths and focused commands.

| Task | State | Commit | Required outcome |
| --- | --- | --- | --- |
| T1.1 | pending | pending | Shared optional blocks and bounded rich/plain renderer |
| T1.2 | pending | pending | Readable conversation/text families 01-20 |
| T1.3 | pending | pending | Renderer, transport and conversation regressions |
| T2.1 | pending | pending | Durable receiver snapshots, migrations and quota |
| T2.2 | pending | pending | Document upload, caption, acknowledgement and fallback |
| T2.3 | pending | pending | Primary artifact/export families 21-24 and 28 |
| T2.4 | pending | pending | Exact bytes, bounds, rollback and recovery evidence |
| T3.1 | pending | pending | Human guarded previews and controls |
| T3.2 | pending | pending | Original operation, document and action bindings |
| T3.3 | pending | pending | Stale scope, ambiguity and replay regressions |
| T4.1 | pending | pending | Progress, waits, failures and partial/final results |
| T4.2 | pending | pending | Notification sender and event snapshot integration |
| T4.3 | pending | pending | Both-path faults, legacy records and truthful outcomes |
| T5.1 | pending | pending | Complete producer and variant coverage |
| T5.2 | pending | pending | Real mobile/desktop and natural-language journey |
| T5.3 | pending | pending | User documentation, cycle evidence and exact nightly |

## Message-Family Coverage

Numbers retain the source review's identities. These are required outcomes,
not implementation evidence. Evidence is pending for every row. Record the
actual source owner, test/fixture and applicable client result as each row is
delivered; family numbers alone do not prove dynamic-variant coverage.

| Family | Required presentation/outcome | Owning slice | Evidence |
| --- | --- | --- | --- |
| 01 | Readable help; advanced material conditional | Slice 1 | pending |
| 02 | Relevant command/unsupported-action guidance | Slice 1 | pending |
| 03 | Human project list and empty state | Slice 1 | pending |
| 04 | Selected-project acknowledgement | Slice 1 | pending |
| 05 | Safe project/reply/root ambiguity and recovery | Slice 1 | pending |
| 06 | Receiver health and next responsible party | Slice 1 | pending |
| 07 | Request acknowledged; extra capture details optional | Slice 1 | pending |
| 08 | Concrete start/capacity limits | Slice 1 | pending |
| 09 | Human session topics, states and targeting choices | Slice 1 | pending |
| 10 | Clear focus effect or actual unavailable state | Slice 1 | pending |
| 11 | Stale/conflicting prompt and preserved work | Slice 1 | pending |
| 12 | Details saved; one useful next action | Slice 1 | pending |
| 13 | Readable retained capture and inspection | Slice 1 | pending |
| 14 | Capture/finished-state errors with unchanged context | Slice 1 | pending |
| 15 | Saved request and truthful generation state | Slice 1 | pending |
| 16 | Explicit unstarted-capture discard | Slice 1 | pending |
| 17 | Relevant syntax/example only | Slice 1 | pending |
| 18 | One concrete natural-language target question | Slice 1 | pending |
| 19 | Interpretation failure/retry exhaustion guidance | Slice 1 | pending |
| 20 | Actual clarification question and clear answer route | Slice 1 | pending |
| 21 | Primary generated artifact excerpt and full file | Slice 2 | pending |
| 22 | Authoring wait/failure from real owner state | Slice 2 | pending |
| 23 | Saved result after conversation advances | Slice 2 | pending |
| 24 | Bounded explanation, with full content when needed | Slice 2 | pending |
| 25 | Short ticket scope and bound complete document | Slice 3 | pending |
| 26 | Prepared backlog result and correct next step | Slice 3 | pending |
| 27 | Explicit missing intent/context for preparation | Slice 3 | pending |
| 28 | Readable/paginated records and identified full export | Slice 2 | pending |
| 29 | Readable proposed record change and consequence | Slice 3 | pending |
| 30 | Actionable malformed/stale record failure | Slice 3 | pending |
| 31 | Exact work selection, policy, validation and scope | Slice 3 | pending |
| 32 | Admission/current-owner/capacity state | Slice 3 | pending |
| 33 | Real progress, review and partial/final result | Slice 4 | pending |
| 34 | Durable progress/recovery without a duplicate action | Slice 4 | pending |
| 35 | Exact control scope, effect and bounds | Slice 3 | pending |
| 36 | Confirmed/replayed/stale control outcome | Slice 3 | pending |
| 37 | Decision question and supported bound choices | Slice 3 | pending |
| 38 | Readable harness selections and actual readiness | Slice 3 | pending |
| 39 | Explicit harness-change consequence | Slice 3 | pending |
| 40 | Original receipt, expired action and safe retry | Slice 3 | pending |
| 41 | Human button labels; original tokens/choice keys | Slice 3 | pending |
| 42 | Queue rejection versus waiting and actual acceptance | Slice 4 | pending |
| 43 | Uncertain send and bounded plain fallback | Slice 4 | pending |
| 44 | Bounded text, Unicode and full-document route | Slice 4 | pending |
| 45 | Unavailable service and preserved state | Slice 4 | pending |
| 46 | Silent unauthorized rejection; local outage diagnosis | Slice 4 | pending |
| 47 | Concise workflow notifications and full details | Slice 4 | pending |
| 48 | Clearly labelled notification test | Slice 4 | pending |

## Acceptance Evidence

| Criterion | Owner tasks | State |
| --- | --- | --- |
| TMP-01 | T1.1, T4.1 | pending |
| TMP-02 | T4.2, T4.3 | pending |
| TMP-03 | T1.1, T1.3, T4.3 | pending |
| TMP-04 | T2.3 | pending |
| TMP-05 | T2.1, T2.2, T2.4 | pending |
| TMP-06 | T1.2, T3.1 | pending |
| TMP-07 | T3.2, T3.3 | pending |
| TMP-08 | T2.4, T3.3, T4.3 | pending |
| TMP-09 | T2.4, T3.3, T4.3 | pending |
| TMP-10 | T2.1, T2.4, T3.3, T4.2 | pending |
| TMP-11 | T5.1 | pending |
| TMP-12 | T5.2, T5.3 | pending |

## Completion State

- Planning artifacts: accepted; local links, whitespace, family/task coverage and acceptance mappings checked.
- Specification/model promotion and implementation approval: completed on 2026-10-21.
- All five slices, their PRs and reports: pending.
- Mobile/desktop and user-operated journey evidence: pending.
- Exact final candidate and scheduled nightly: pending.

The existing conversational workflow and its active QA evidence retain their
own controller, tracker and invocation states. This set uses a separate Looper
configuration. Its Slice 3 depends on canonical disposition of the earlier
pending corrections; real client and journey evidence remain required.
