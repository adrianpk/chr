# Code Quality Remediation Compatibility Model

Status: accepted
Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`
Specification: [Code Quality Remediation](code-quality-remediation.md)
Plan: [Single-cycle plan](../plan/code-quality-remediation.md)

## Purpose

Define the persistence and value-model invariants that structural remediation must preserve. This is a compatibility model for existing entities, not a new runtime schema. Existing domain model documents remain authoritative for lifecycle and approval behavior.

## Owners and representations

| Value or record | Current owner | Proposed grouping | Stored/wire constraints | Owning slices |
| --- | --- | --- | --- | --- |
| WorkflowInvocation | core/workflow_invocation.go and guarded workflow files | identity, admitted inputs, execution policy, process/attempt ownership, integration, outcome/timing | Existing JSON tags, ordering on hashed paths, omitempty, revisions and retained references remain identical | 3, 12, 15 |
| WorkflowMergerInput | core/workflow_merger.go | role identity, immutable Git input, source/evidence references, validation contract, reservations/budget | Preserve protocol version, complete original payload and binding hashes; do not hash a new nested envelope | 3, 13 |
| IndexedTicket | core/indexed_ticket.go plus index/repository mappings | identity/source, presentation, review, merge, timestamps/index metadata | SQLite projection remains rebuildable; canonical Markdown/native records remain authoritative; preserve null timestamp and status meanings | 3, 11 |
| WorkflowWorkStatus | app/service_workflow_status.go | identity/selection, progress, review/integration, evidence/pagination | Keep public JSON keys, omission/default behavior and read-only semantics; no TUI/Web types in service results | 3, 5, 12 |
| GuidedUseCaseView | webui/guided_use_case.go | existing page/current-step, question, documents, progress, lifecycle actions | Presentation-only; original operation/revision/request identities stay bound to each action | 5 |
| Model | tui/model.go and owned view states | navigation, interaction, layout, background work, per-view state | In-memory owner; pointer conversion is not permission for goroutines to mutate shared state | 5 |
| Execution rows | core/execution_repository.go | one table-specific row, columns, values and scan targets | Same table/schema, key, update predicates, transactions and field order | 3, 11 |
| Remote command rows | core/remote_command_record_repository.go | one table-specific row and ordered binding | Same request/idempotency key, status/response meaning and null handling | 3, 11 |
| Telegram notice/document/action rows | core/telegram_* and retained outboxes | narrow row/request values with one transactional owner | Same quotas, bytes, retry identity, acknowledgement, lease and restart semantics | 3, 7 |

## Field completeness

At Slice 3 entry, capture the actual baseline field/tag/serialization manifest from the four durable/public Go types above before editing them. Include promoted fields, nil pointers, empty/nil collections, zero times, default-valued state and optional fields. For each grouping, map every old field to exactly one new internal field and its unchanged external field. Review the schema, SQL scans, row bindings and hashes together; a field-count or JSON-key-only assertion is insufficient.

IndexedTicket has 40 fields; WorkflowInvocation 48; WorkflowMergerInput 39; WorkflowWorkStatus 37 at the source scan. These are inventory signals, not target field limits. A grouping must name an actual responsibility and must not create a large generic options value.

## State machines and outcomes

| Owner | Typed value | Persisted behavior to preserve |
| --- | --- | --- |
| Ticket | Existing ticket status type, extended only where needed | Current backlog/readiness/review/integration meanings and native authority |
| Invocation | Invocation lifecycle state | Existing pending, running, waiting, blocked, paused and terminal values and guards |
| Attempt/claim | Attempt and claim phase | Generation/lease identity, consumed attempts and quiescence |
| Publication/review | Provider and review state | An open PR, accepted review, publication and confirmed integration remain different facts |
| Telegram | Operation/inbox/outbox delivery state | Receipt-backed transitions and duplicate/lost-ack recovery |
| Preparation | Named preparation mode | Existing automatic/explicit behavior, selected caller authority and defaults |

No single global status enum replaces these distinct machines. String spellings remain unchanged, and typed conversion must not silently introduce a new rejection of previously supported data. Existing invalid/corrupt-state handling remains explicit.

## Lookup and operation result values

Each converted tuple must state whether its flag meant found, changed, cached, consumed, authorized or complete. Use a narrow named result for normal outcomes. Preserve errors.Is identity and wrapping when an error is already part of the contract. Do not manufacture an ErrNotFound flow for a normal absent result or discard a cleanup callback when consolidating outputs.

## Invariants, writes and reads

1. Workflow files and canonical ticket/artifact records remain the source of truth; SQLite and workspace indexes do not become a replacement authority.
2. Locks, transactions, optimistic revisions, request keys, receipts and hashes keep their current atomic boundary. A stage's narrow port must not split a multi-record mutation into independently observable writes.
3. Old committed authoring inputs and invocation receipts remain readable and verifiable after a restart. No migration rewrites historical hashes or creates an extra work invocation.
4. SQL identifiers come from static owned column declarations. Values stay bound. Table-specific update filters, nullable fields and default behavior stay explicit.
5. Map nil versus empty, omitted versus present and zero versus absent deliberately. Check exact bytes where a hash depends on them; compare semantic content only where the existing contract does.
6. Preserve finite attempts, deadlines, review/publication permissions and failure recovery. Contexts are explicit call parameters, not durable or reusable operation fields.

## Compatibility evidence required

Use existing real SQLite, filesystem, Git and retained-receipt fixtures. Extend them with distinct per-column values and old-record reopening. Record serializer/hash comparisons before and after grouping, round-trip insert/update/rollback outcomes, and the unchanged state-transition cases. This planning model is not evidence that those checks have run.

## Slice 3 frozen serialization evidence

Entry candidate: `f24fc8533f92`; production records match the verified Slice 2
candidate `de3691d3b0155f850006f242ee2d99062a995ce0`.

Before grouping, the exact declared field names, Go types, JSON tags (including
omitempty), anonymous fields and order were captured in the
[core field manifest](../../../internal/core/testdata/quality-records/field-manifest.json)
and [status field manifest](../../../internal/app/testdata/quality-records/field-manifest.json).
The adjacent frozen JSON fixtures cover zero/nil values, non-nil empty
collections and populated distinct values with fixed nonzero timestamps. They
include the outer workflowInvocationFile envelope so grouping cannot silently
lose owner/request metadata through promoted serialization methods.

WorkflowInvocation persists through workflowInvocationChange, which merges its
JSON fields and preserves supported unknown fields before indented encoding.
WorkflowMergerInput binds its complete JSON bytes through workflowMergerInputHash.
IndexedTicket remains a rebuildable SQLite projection with default JSON field
names; WorkflowWorkStatus remains a read-only public JSON projection. These
contracts keep the same key names, order, omission, nil/empty and timestamp
behavior. The byte fixtures are frozen inputs, not regenerated by validation.

The [field mapping](../plan/code-quality-remediation-record-fields.csv) maps each
original direct field to its cohesive value group. Groups occupy contiguous
positions in the original declaration, retain all original leaf tags and use
native value embedding. No serialization method is promoted onto the outer
invocation file. All keyed constructors were updated with the same expressions;
existing promoted reads and scan targets retain their original leaves.

The [wide SQL inventory](../plan/code-quality-remediation-sql-rows.csv) covers
every production SQL call above 16 fields, including expanded argument/scan
helpers. Eight tables are affected; 11 calls bind or scan directly, and three
expanded boundaries cover notification creation and ticket upsert/read.

The frozen [legacy SQLite fixture](../../../internal/core/testdata/quality-storage/legacy.sqlite)
was written before changing those SQL mappings. Its adjacent provenance records
the producer, SHA-256, eight table counts and clean foreign-key check. Expected
records retain distinct same-type fields, fixed timestamps, NULL optionals, and
empty versus nil remote loop lists. Validation copies and opens this database;
it never regenerates the fixture.

All eight affected tables now have one named row holder and one static ordered
column definition. Execution, remote command, notification event and ticket
rows pair values and scan targets; the conversation row preserves its existing
COALESCE projection. Inbox and document rows retain table-specific write-only
admission shapes. The receiver document owner_kind remains the static receiver
literal, and both document tables share only their identical quota predicate.
Updates retain their exact SET/WHERE contracts; ID-last arguments derive from
the same row values. Existing enum/time/JSON decoding and NULL conversion remain
unchanged. No schema or migration changes accompany these mappings.

The real-SQLite compatibility tests compare complete stored table images against
the old writers, preventing symmetric bind/scan swaps from hiding a regression.
They cover insert and update/upsert paths, nil/empty JSON lists, nullable times,
COALESCE defaults and omitted admission defaults. Distinct peer identifiers and
independent source/content hashes exercise same-type ordering separately.
Injected failures after document insertion, notification insertion and ticket
upsert prove rollback preserves the captured table and durable owner images.
The completed legacy fixture passes SQLite foreign_key_check and opens through
the existing migrator without changing those rows. All new fixtures are local
and test-owned.

## Slice 4 typed command boundary

The process entrypoint groups its three streams in app.IO and constructs one
commandFactory. All 134 command constructors inherit those streams and typed
assets through the same value; actions use method bindings rather than root
Metadata lookup. The [constructor inventory](../plan/code-quality-remediation-command-constructors.csv)
records the complete tree. Seventy-four forwarding action factories are removed.

The [wiring transitions](../plan/code-quality-remediation-slice-4-wiring.csv)
preserve the original identity of each affected signature allowance. Two
authoring input violations resolve incidentally when command IO is captured;
the remaining authoring work retains its approved owner. Existing unresolved
result shapes keep exact explicit identities until their assigned conversion.
A real command regression supplies invalid legacy root metadata and verifies
that bound migration assets determine the action's behavior. CLI names, flags,
output streams, action contexts and dependency ownership remain explicit.

The [mechanical move ledger](../plan/code-quality-remediation-slice-4-moves.csv)
records all 292 declarations split from ticket.go, author.go, maintenance.go
and service_work.go. The SHA-256 for every declaration is unchanged by this
move, including comments and bodies. The 29 resulting files separate command
branches, preflight/workspace/execution mechanics, authoring modes, maintenance
areas and work selection/review/merge/cycle stages. All remain in internal/app;
no new service type or package changes ownership at this mechanical step.
Exact temporary debt identities follow their declarations, including later
authoring-owned tuples. Semantic ticket-operation changes follow separately.

Ticket selection, preflight, workspace/execution mechanics, result application,
legacy loop execution and manual review/merge now belong to service_ticket
files on the existing central service. All stage contexts remain call arguments.
One ticketLoop value retains the run repository, record, events and partial
result. Legacy loop requests return selection and execution separately so a
failed allocated run still produces its original CLI evidence before returning
the error. CLI handlers parse requests and render the service result.

CLI, TUI, Web and Telegram work callers share the same preflight, execution and
result stages through their existing service operations. Manual review, agentic
review and trusted review keep distinct approvals and outcomes. Manual and
automatic merge retain their own lock and refresh scopes while using the same
core merge authority. Native and unadopted legacy selection guards, worktree
preparation, finite iteration limits, configured merge strategies and nullable
explicit base overrides retain their original meanings.

The [operation ownership ledger](../plan/code-quality-remediation-slice-4-operations.csv)
and [signature transitions](../plan/code-quality-remediation-slice-4-operations-debt.csv)
record the concrete moves and typed input replacements. No compatibility
forwarder is retained for the removed ticket preflight wrapper or CLI
merge-configuration tuple.

Command resolution uses `commandFactory.projectService` as the single CLI-to-service constructor. Its service already contains the resolved project; returning that project separately duplicated ownership. The unused `service.dumpDirectory` accessor is removed; dump tests continue to exercise the live listing/content path. Telegram-private error names remain assigned to Slice 7.

The [Slice 4 row dispositions](../plan/code-quality-remediation-slice-4-dispositions.csv)
account for all 69 assigned A/B/C rows: 47 fixed occurrences and 22 retained
cohesive five-input contracts. Every current declaration is at most five
parameters and has at most two results, with error/bool second when present.
No assigned declaration remains in temporary signature debt. Exact transitions
for the final 31 fixes are in the adjacent slice-4-signatures ledger; the global
manifest now has 472 entries with the ceiling still seven and one external
exception.

Named ticket results retain the execution record on application/report failure,
partial integration state and selected queue, original eligible count, refreshed
ticket, stop decision and preparation release callback. Failed preparations
still release their lock; successful callers retain deferred release ownership.
Web dispatch retains handled separately from response/error. Evidence parsing
retains nonfatal diagnostics separately from file-read failures. Provider text
results retain response IDs on failed responses. Notification type and severity
remain paired; unknown status keeps an empty kind.

Daemon iterations use the existing daemonWorkflowRunner, which owns output,
environment, assets and state repository; the operation context and iteration
bound stay explicit. Gate report source groups creation time and report paths,
while refs, evidence, validation and governing configuration remain separate.
Conversation intent derives its bound context in the operation caller and
validates the bound execution project without a context-bearing result object.


## Slice 5 model ownership

The populated benchmark fixture measures the installed Bubble Tea event-loop
interface before and after receiver changes, using idle, focus and resize
messages plus sized and unsized board rendering. The baseline production
revision is ad40e8b089b7; the benchmark fixture is introduced in the T5.1 task
commit. Three samples per case are retained in the slice-5-benchmarks ledger.
This measures the stated paths rather than all large-value warnings.

NewModel returns *Model, and only that pointer implements tea.Model. The event
loop retains one mutable owner across updates; read helpers also use pointer
receivers to avoid repeated whole-model copies. All 85 original fields and their
types are assigned to 17 explicit navigation, layout, interaction, background and
per-view state owners in the model-fields ledger. Existing field semantics and
initial values are retained.

Commands capture their immutable requests and executor interfaces before they
leave the event loop. The one refresh closure that previously captured the
value receiver now captures its executor directly. No command closes over the
mutable Model pointer; completion messages are applied only on the event loop.
Existing sequence/revision guards, busy flags and cancelled/stale acknowledgments
retain their original authority. Tests use the pointer contract explicitly and
still exercise modal precedence, edits, retained bindings and stale replies.


Internal update stages mutate that same pointer and return tea.Cmd or
updateOutcome{Command, Handled}. The only Model result tuple left is the
installed Bubble Tea Update interface. All 37 converted internal helper
signatures have exact before/after entries in the slice-5-updates ledger; no
adapter or internal interface exemption is added. Modal dispatch precedence,
command scheduling and every handled/unhandled branch retain their ordering.
The production closure scan finds no command reading Model state after dispatch.


The Slice 5 mechanical move ledger records all 193 declarations from model.go
and shell.go, with exact text hashes. Sixteen responsibility-owned files separate
model assembly, event dispatch, layout input, navigation, ticket data/Markdown,
board selection and per-view shell rendering. The largest is 434 lines. All
declarations, comments and bodies are unchanged during that move; four exact
signature-debt identities follow their new files without a new allowance.


The D01 Spec/Plan catalog detail/list rule now has one presentation owner in
`document_catalog.go`. Per-kind adapters retain their title/scope/status mapping,
action hints and busy messages. Spec generation and feature actions remain
separate from plan generation, regeneration and guarded tracking materialization.
The actual D11 baseline ranges (shell.go:636-678) are the Spec/Plan workspace
layouts, despite the source inventory's ticket-column label. They now share
panel sizing, focus and maximization through that same catalog view. Compact
ticket columns already had a common renderer; their wide input remains assigned
to T5.5. Frozen complete frames verify both kinds and every retained pane mode.

Main view dispatch, ordered modal guards and interaction key hints are data,
with the original precedence retained separately where dispatch and hint order
differ. The three corresponding cyclop suppressions are removed. The plan TUI
adapter shares success/failure logging and projection reload while its smaller
switch keeps the three distinct service request types; its cyclop/funlen/wsl
suppressions are removed. All operation failures still return zero adapter
results and retain their original error identities and log fields. Real guided
TUI journey/parity tests use and assert the single pointer owner without copying
it back into a value fixture.


## Slice 5 guided Web sections

GuidedUseCaseView has 47 original fields. Seven contiguous question, page, ideas,
lifecycle, run, step and diagnostic groups preserve every original name, Go type,
tag and encoding order. Result, work, artifacts, enablement and resumable links
retain their existing section types/fields. The guided-fields ledger accounts
for every field; frozen zero, non-nil empty and distinct populated JSON images
and a direct field manifest prove exact compatibility before and after grouping.

Leaf templates now receive only their form bindings, retry form, contextual
actions, status, recovery, progress or result section. The page compositor keeps
its existing Detail/Board/quiz/dump inputs. Each control retains its own revision,
native revision, request, question and input digest; a retry uses the original
pending action/content and the same displayed owner identity. Visibility flags,
HTML classes, routes and form field spellings remain unchanged. The section
methods do not invoke operations or advance work.

WorkflowWorkStatus already uses the cohesive adapter-neutral bindings,
lifecycle, selection, outcome, scope, counts, control, delivery and page values
established in Slice 3. Those are reused directly rather than regrouped again.
The existing frozen work-status images still prove original public JSON bytes,
omission/default semantics and field order; no service result imports TUI/Web
presentation types.


The final ten owned signature violations are converted in the Slice 5 signatures
ledger. Document scanning owns its project/family and accumulated views, with
context passed into each WalkDir callback. Observability construction receives
loaded daemon/execution/loop/runtime/event records, and compact ticket rendering
receives one display row with its original header alignment. No other large
value is changed to a pointer merely because of the general hugeParam signal.

Named geometry returns distinguish left/right pane widths, top/bottom heights
and the shell body offset/height. Dump panels retain their title/content pairing.
Markdown summaries retain title/status/preview defaults. Retained harness reads
keep pagination and nonfatal diagnostics separate from absence, and ticket write
projections retain both the correct PageData and partial-template selection.
The global manifest now has 425 entries and the same one external exception.

Three definition-only compatibility helpers (renderDocumentsByFamily,
selectedDocumentView and renderDocumentFamilyWorkspace) and their unused
suppressions are removed after checking all production/test references. Active
legacy document navigation/rendering and documentsByFamily remain live. The
asynchronous refresh regression mutates/renders the event-loop owner while IO is
blocked, then applies the returned message on that owner; focused race validation
covers this test with the existing guided, harness and workflow ownership checks.
