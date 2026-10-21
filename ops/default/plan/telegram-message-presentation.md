# Telegram Message Presentation Delivery

Status: approved

Approved: maintainer, 2026-10-21; specification, model, slice map and Looper execution accepted.

Full gate: `make verify`

Date: 2026-10-21

Delivery set: telegram-message-presentation

Spec: [Message presentation](../spec/stable/telegram-message-presentation.md)

Model: [Document delivery model](../spec/stable/telegram-message-presentation-model.md)

Tracker: [Delivery and coverage](../tracker/telegram-message-presentation.md)

Input: [48-family review](../report/telegram-message-review.md)

Planning scope: requested by the maintainer on 2026-10-21.

Implementation and merge authorization: granted for this delivery set and
Looper-owned Rebase + Fast-forward slice PR merges into dev.

Active slice: Slice 1 of 5, Readable text and conversation messages.

## 1. Outcome and Strategy

Deliver readable Telegram responses and notifications, bounded literal
Markdown/document views, complete attachments and truthful action/result copy.

Slice strategy: behavior-first

Reason: each slice applies a shared capability to actual message families and
delivers visible, testable behavior. A shared-type-only or deliberately failing
test slice is unnecessary. Both sending paths use the same presentation
contract while keeping their current authority, persistence and dispatch owners.

The set contains five ordered slices and 16 tasks. The tracker maps every
initial inventory family to one rollout owner. Additional discovered producers
must receive a row and evidence before coverage closes. Do not create another
ticket batch for these tasks.

## 2. Baseline and Activation

Planning baseline: dev at 94f6096c77ebe35860cf9a3bf2166893569b873b.

The receiver, conversations, operations and controls from the previous
Telegram set are integrated. Its Slice 8 preparation/admission and executor
handoff corrections remain on a separate pending branch at this checkpoint.

Activation baseline: dev at 94f6096c77ebe35860cf9a3bf2166893569b873b, verified
equal to origin/dev before committing this package. The earlier Telegram set
has seven of eight slices integrated and its controller remains blocked.
Its pending Slice 8 corrections are on a separate branch and are not integrated
at activation. Slices 1-2 can proceed against current dev. Before Slice 3,
consume the corrections through their canonical integration or record the
dependency as blocked. Do not copy them into this set or acquire the retained
controller's work.

The maintainer authorized the accepted package and Looper execution on
2026-10-21. Commit and push these artifacts before starting the new controller.
Use a separate local configuration for this set, preserving the earlier
controller and QA invocations. Real client acceptance requires user participation
and candidate-bound evidence. Do not restart the QA receiver or manufacture
inbound messages to bypass that checkpoint.

## 3. Canonical Slice Map

Report paths below are future implementation outputs, not existing evidence.

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Readable text and conversation messages | feat/slice-1-telegram-message-composition | feat(slice-1): compose readable Telegram messages | ops/default/report/slices/telegram-message-presentation/slice-1-message-composition.md |
| Slice 2 | Document excerpts and complete attachments | feat/slice-2-telegram-document-delivery | feat(slice-2): deliver Telegram document previews and attachments | ops/default/report/slices/telegram-message-presentation/slice-2-document-delivery.md |
| Slice 3 | Clear guarded confirmations and controls | feat/slice-3-telegram-action-presentation | feat(slice-3): clarify Telegram confirmations and controls | ops/default/report/slices/telegram-message-presentation/slice-3-action-presentation.md |
| Slice 4 | Progress, notifications and failure consistency | feat/slice-4-telegram-status-presentation | feat(slice-4): unify Telegram progress and notification presentation | ops/default/report/slices/telegram-message-presentation/slice-4-status-presentation.md |
| Slice 5 | Complete coverage and real client acceptance | test/slice-5-telegram-presentation-acceptance | test(slice-5): certify Telegram message presentation | ops/default/report/slices/telegram-message-presentation/slice-5-presentation-acceptance.md |

Each slice uses one branch and one PR targeting dev. Follow the existing
repository PR/report contracts and the delegated Looper integration rules.
The table defines the approved branch names, titles and report paths.

## 4. Slice Tasks

### Slice 1: Readable Text and Conversation Messages

| Task | Required outcome |
| --- | --- |
| T1.1 | Implement shared optional message blocks and deterministic rich/plain rendering in existing Telegram owners. Bound Unicode, escape dynamic values and separate literal Markdown/code from ordinary text. |
| T1.2 | Apply the presentation to families 01-20. Use human project/topics, acknowledge retained initial requests, explain optional detail entry and show one justified next step or target choice. Keep actual operations and authorization unchanged. |
| T1.3 | Add focused renderer, transport and conversation regressions for those variants. Preserve help formatting, reply routing and unsupported/unauthorized behavior. |

Acceptance: TMP-01, TMP-03, TMP-06; initial TMP-02/TMP-11 coverage.

Focused selector: ^TestTelegramPresentation(Text|Format|Conversation)

### Slice 2: Document Excerpts and Complete Attachments

| Task | Required outcome |
| --- | --- |
| T2.1 | Add the receiver's immutable presentation/document mapping, additive migration, transactional byte quota and bounded safe retention under existing database owners. |
| T2.2 | Add bounded sendDocument upload, caption formatting, original acknowledgement handling and plain fallback through the existing delivery loop. |
| T2.3 | Apply document-aware results to families 21-24 and 28, choosing the primary relevant artifact or identified export instead of dumping earlier files or raw JSON into chat. |
| T2.4 | Prove exact-byte attachments, mutable-source isolation, Unicode cuts, old-row compatibility, rollback, quota, stale/revoked destinations and recovery without resetting retry budgets. |

Acceptance: TMP-04, TMP-05, TMP-08, TMP-09, TMP-10.

Focused selector: ^TestTelegramPresentation(Document|Delivery|Migration)

Run the same focused selection with race instrumentation because persistence,
delivery reservation and cleanup ownership change.

### Slice 3: Clear Guarded Confirmations and Controls

| Task | Required outcome |
| --- | --- |
| T3.1 | Migrate families 25-27, 29-32 and 35-41 to concise scope/consequence views and accurate action labels. Cover backlog readiness transitions from the actual integrated admission owner. |
| T3.2 | Bind document-backed caption, attachment and keyboard to the original sealed operation. Preserve legacy action guards, peer/project/source checks, expiration and replay receipts. Add follow-up buttons only through existing typed operation owners and their guards. |
| T3.3 | Prove full scope remains inspectable before confirmation, stale/missing documents cannot authorize work, name ambiguity is resolved safely and duplicate/lost acknowledgements cannot repeat mutations. |

Acceptance: TMP-06, TMP-07, TMP-08, TMP-09, TMP-10.

Focused selector: ^TestTelegramPresentation(Action|Control|Confirmation)

Run that selection with race instrumentation for concurrent answers and
delivery/action recovery.

### Slice 4: Progress, Notifications and Failure Consistency

| Task | Required outcome |
| --- | --- |
| T4.1 | Migrate families 33-34 and 42-46. Explain actual progress, waits, partial/final results and failures using existing truthful states, throttle and coalescing rules. Keep revoked/unauthorized paths silent where required. |
| T4.2 | Migrate families 47-48 through the notification event repository and sender. Apply the shared renderer and frozen document mapping without a receiver, another scheduler or altered notification triggers. |
| T4.3 | Prove both paths handle long/dynamic content, failed uploads, unchanged legacy records, uncertain sends and bounded fallback without false delivery/integration claims. |

Acceptance: TMP-01, TMP-02, TMP-03, TMP-08, TMP-09, TMP-10.

Focused selector: ^TestTelegramPresentation(Progress|Notification|Failure)

Run that selection with race instrumentation where snapshot/receipt ownership
changes.

### Slice 5: Complete Coverage and Real Client Acceptance

| Task | Required outcome |
| --- | --- |
| T5.1 | Reconcile the actual producer inventory, all 48 matrix rows and rare/dynamic branches. Record explicit keep/omit outcomes, existing regression reuse and any additional family before closing coverage. |
| T5.2 | With authorized QA activation, inspect real mobile and desktop messages and attachment opening. Complete one user-operated natural-language journey with a fixed candidate receiver, truthful execution states and a canonical integration receipt. |
| T5.3 | Update the existing User Guide, transport reference, reader-facing changelog and cycle report. Record exact candidate, actual evidence, remaining limitations and scheduled nightly outcome before closure. |

Acceptance: TMP-11, TMP-12 and final audit of TMP-01 through TMP-10.

Focused selector: ^TestTelegramPresentation(Coverage|Journey)

Actual client evidence is additional to fixture results. Do not invent inbound
messages, screenshots or approvals to fill the matrix. Existing live QA and
active native invocations remain owned by their current workflow.

## 5. Validation and Delivery Gate

For each exact clean slice candidate:

1. Run make verify.
2. Run go test ./internal/core ./internal/app -run '<slice selector>' -count=1 -timeout=3m.
3. Add -race to that focused command when required above.
4. Run git diff --check last.

Selectors are test obligations to implement; zero matching tests are not a pass.
Reuse existing tests where they establish the contract. Add affected legacy
regressions to the same bounded package invocation when the change requires
them, recording their names and reason rather than growing a blanket suite.

Use real temporary SQLite databases and local HTTP servers. Live Telegram calls
are reserved for authorized acceptance. The expected local cost is the existing
short development gate plus the selected fixture group, with the current
three-minute package timeout; measure changes before expanding that budget.
Do not run full repository or full race suites interactively.

The final exact dev candidate must pass make verify and the bounded combined
TestTelegramPresentation selection, including focused race coverage of changed
ownership. Reuse evidence only for the identical commit and command order.
The existing scheduled nightly owns exhaustive certification. Do not dispatch
it under this planning request. Retain awaiting_nightly until that exact
candidate is green.

## 6. Completion and Handoff

Every family row needs its real disposition, source owner and regression
reference. Every acceptance criterion needs concrete evidence. Renderer tests
cannot stand in for mobile/desktop inspection, and worker exit or a send receipt
cannot stand in for integration.

Each slice produces its own report after implementation. The final cycle report
is ops/default/report/telegram-message-presentation-cycle.md. It is not
pre-created as successful evidence.

The specification/model are accepted, the plan/tracker are approved and
Slice 1 is active. Execution starts only from the committed and pushed package.
Record real implementation and acceptance results as they become available;
approval does not establish delivery or client evidence.
