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


The 895-line Web entrypoint file is split by assembly, page handlers, artifact
handlers, write handling, page composition, access control and request logging.
All 53 complete declaration texts are preserved across seven files, largest
209 lines; the web-moves ledger retains exact hashes. HTTP route registration,
origin/token checks, login ownership, bounded form parsing, redirects, fragments
and logging remain under their original contracts. Existing full Web package
checks and strict lint pass after the move.

All 50 assigned baseline rows have final dispositions: 47 fixed occurrences,
two retained cohesive five-input panel renderers, and the one exact imposed
Bubble Tea Update interface. No assigned declaration remains in temporary debt.
The global ceiling remains seven with 425 exact debt entries and one external
exception; enforcing later slices is still the recorded policy.


## Slice 5 measured boundary

The same 128-ticket fixture and three 200ms samples per case compare baseline
production ad40e8b089b7 (benchmark fixture introduced in c016b6c2a5a0) with clean
production 739cdf1b3e1b. Model remains 49,832 bytes; the measured improvement comes
from retaining its pointer rather than shrinking unrelated value records.
Median Update allocation bytes change from 229,376 to zero for idle, 345,162 to
1,096 for pane focus, and 172,032 to zero for resize. Allocation counts change
from 4/15/3 to 0/9/0 respectively. Median times change from 51,532/88,149/41,288 ns
to 14.61/847.7/13.98 ns for these isolated event-loop paths.

Median plain/shell View time changes from 225,533/692,490 ns to 122,040/497,802 ns.
Their allocation counts remain 703/1,404 and bytes remain approximately
435/550 KB. Rendering allocations are not claimed fixed. These measurements
cover the named paths on the same machine; they do not justify blanket pointer
changes for the remaining general hugeParam or range-copy signals. The raw
sample ledger names the production revisions and retains all 30 observations.

After the mechanical proofs, one moved case helper's signature header is put on
a single line to satisfy the repository's signature-format rule. Its types and
body are unchanged. The historical move hashes retain their original scope.


## Slice 6 compatibility and dependency evidence

Gate, project identity and notification persistence now have direct package
callers and enforced production import bounds. Config parsing consumes an explicit
path and named parsed-line state. Git refs/ranges retain distinct base/head and
partial evidence on error. Report classification consumes loaded execution
reports; artifact allocation and metadata refresh remain application composition.
Gate report JSON, validation command execution and failure evidence retain their
original contracts. The two five-input report constructors remain cohesive.

Project registry and active selection preserve original SQLite columns/order,
normal absence, active file bytes and atomic replacement. The shared database
owner preserves schema, migration order, connection/timestamp handling and joined
transaction failures. RootBindingDigest preserves project ID, registered absolute
path and resolved path separated by NUL bytes. Root conversation mapping preserves
the original safe unavailable-root diagnostic.

Notification transactions still own admission, reserve, completion and retention.
TelegramLookup separates delivery/Frozen from failure; legacy rows stay untouched.
TelegramCompletion carries the original event/target, acknowledgement, safe failure
code/certainty and time. Conditional SQL preserves target/hash/owner bindings,
attempt counts, the single definite-format fallback and uncertainty across reopen.
Acknowledgement still clears failure/uncertainty without granting another attempt.
The shared quota predicate retains its exact schema-level conditions inside each
owner's existing transaction; receiver mutation ownership remains separate.

Shared message declarations have 27 explicit identifier-normalized token proofs
including literals and tags. Hash-sensitive storage and record fixtures retain
original bytes, field order, omission/null semantics, entities and document hashes.
Content hashing has one SHA-256/hex owner, namespace spelling one pure owner and
conditional writes one ErrConflict identity. No schema or retained-hash migration
is introduced. Mechanical gate/project/notification moves have 240 complete
unchanged declaration-text hashes before qualified-name/result conversions.

All six assigned authoring tuple returns now use precise named outcomes. Artifact
reads preserve aggregate byte counts and partial construction; file images preserve
separate body/metadata bytes; reconciliation preserves partial mutation results;
receipt reads preserve absence/found/error; Markdown/field parsing preserves raw
body and found state. Both mutation callbacks explicitly assign their outer
result/error after receipt lookup. Replay, locks, checksums and operation identities
therefore retain the existing owner. Those transactions remain authoring behavior
under the recorded Slice 9 package owner; their Slice 6 signatures are complete.

The 19-row disposition ledger records 16 fixed occurrences and three cohesive
five-input contracts. Temporary debt decreases from 425 to 410 because the watch
items and already-retired rows are distinct from outstanding violations. Both
locally owned notification interfaces change with their implementations. The
existing external-interface exception remains unchanged. Static import fixtures
cover forbidden aliases, descendants, generated sources and build variants; real
SQLite/HTTP/Git/child/filesystem fixtures cover the behavior of direct consumers.


## Slice 7 shared binding prerequisite

workflow/state.MutationBinding preserves the original request_id/operator fields,
JSON tags and order. Its validation uses the exact original predicates and error.
ValidRequestKey retains whitespace, length and control-character rules.
MutationReceiptPath retains the same SHA-256/hex receipt filename and slash path.
Root workflow and Telegram callers share this pure owner; receipt serialization,
verification, containment, locks and durable mutation remain in the existing store.
Two outstanding debt signatures explicitly follow the qualified type name, with
no allowance or ceiling change. Frozen record and real receipt regressions
protect the representations before the Telegram package move.


Telegram directory capabilities replace the host-wide UserEnvironment dependency
at each actual consumer. Configuration/credentials, OS identity/exclusion and SQL
inspection receive only their corresponding paths. Store construction still
performs the same schema checks, migration, private creation and connection setup;
its retained environment supplies only exclusion paths. Read-only inspection uses
the original database reader without configuration, state creation or locks.
Receiver/configuration/token, recovery-environment and storage regressions pass.


Three non-short Telegram assertions failed identically on unchanged predecessor
c7d7040fde4385a64b20322da1bc5874a38c2deb and the Slice 7 binding prerequisite.
The capacity diagnostic and ticket-readiness prompt now assert the current full
wording while retaining artifact immutability, unconsumed action and actual
native-work completion checks. The uncertain-action-send fixture checks exact
visible retry content, then reads SQLite before/after reopen: exactly one reply,
the same original request/receipt/delivery ID, confirmed request state and one
uncertain-to-delivered retry (attempt one to two). One application and one canonical
record remain required. This replaces stale visible opaque-ID assertions with
stronger evidence at their durable owner; production behavior is unchanged.
The focused three-test command passes, including the real HTTP/native journey.


Shared project-directory validation retains filepath.Abs, directory existence/type
checks and original errors at project.ResolveDirectory. Namespace selection retains
explicit input, OPS_USER and default precedence, trimming and invalid-name errors
at namespace.Resolve. Existing operational layouts still validate the directory
before selecting its namespace. Telegram allocation/target construction calls the
same two shared operations in that exact order after root binding; artifact path
composition stays with OpsLayout. Original filesystem, namespace precedence,
registry and Telegram conversation/target tests pass with direct consumers.


D06 shares intake mechanics through two independent typed route records. Exact
retained operation lookup still precedes parsing; an unrelated retained operation
does not change recognition. Unrecognized input still falls through, malformed
recognized input still queues its own diagnostic, and current configuration and
permitted projects populate the actual parsed command before dispatch. Work and
explanation preserve their distinct reservation, authority, model, document and
confirmation paths. Test-only help forwarding and guided result aliases are
removed while their live renderer/export guards and behavioral tests remain.


Authoring Intent keeps the exact explain/prepare/develop values and field types;
ValidateSessionID keeps the exact identifier predicate and error text. The two
structured model failure identities are each defined once at the neutral model
owner and used directly by provider/retained operation consumers. The shared
workspace DirectoryName stays .chr. These moves change no JSON tags, field order,
retained references, checksum bytes, provider slots or retry allowance. Frozen
records, old SQLite images, authoring scope, model and Telegram regressions pass.


Guided Telegram association now receives only the validated session identity,
project, namespace and topic. The app calls the unchanged complete guided domain
validator before projection, then the store retains its exact binding, peer and
scope checks. The invalid-schema regression uses a separate current conversation
and proves failed attachment cannot change peer focus or either canonical run;
removing the guard makes that test fail. Legacy update decoding remains transport
owned, while generic remote-command classification and record construction move
to app composition with their existing tests and time/error semantics.


Mechanical Telegram relocation preserves all 453 production declarations across
71 files byte-for-byte, recorded in slice-7-moves.csv. The 51 moved tests keep
private persistence checks at their owner; mixed domain integration tests retain
root placement and use independent fixture SQL connections. Frozen legacy bytes,
identity columns, cross-table rollback, receiver HTTP/actions/documents, lost-ack
restart and actual app/native integration regressions pass after relocation.
Naming cleanup and all assigned signature dispositions remain pending; no final
slice validation or PR readiness is claimed by this task checkpoint.


Receiver tuple results now carry explicit records for answer recognition,
conversation/target binding, optional retained operations/conversations, routing
replies, multipart body/type, provider-call reservations, lanes and work pages.
The prior data and flag assignments still occur before each original error
branch. Named callback outputs still update their outer variables; work-page
iteration retains a named error so deferred rows.Close failures remain visible.
No read miss becomes an error and no read grants authority. Original retained
HTTP/SQLite/recovery/control/domain-app and frozen identity regressions pass;
37 exact migration allowances disappear without increasing the ceiling. App
result tuples and long operation inputs remain assigned to this active slice.


HTTP sender requests now group each original action, rich-text, document, upload
and bounded-body operation. Context and credentials stay explicit call arguments.
Original label omission is still nil [][]string; an explicit label list remains
one element, even when empty. Methods preserve the original validation order,
request fields, timeouts and safe error/uncertainty classification. Multipart
bytes, caption/entity fallback, missing file acknowledgement, malformed envelope
and uncertain delivery regressions pass at real HTTP/SQLite boundaries. Six
more exact allowances are removed; core debt is now 367 repository-wide.


App Telegram result records preserve parser recognition independently of errors;
target Needed retains its existing guard semantics. Guided Start stays distinct
from recognition, and receipt Found remains distinct from read failure. Guided
progress/facts, prepared input/current configuration, intent context/digest,
harness retained paging and host finished/quiesced channels keep their original
paired values. The native control owner still exposes its existing contract;
only its Telegram composition adapter projects the named receipt result. The
complete selected app Telegram/guided regression set and pinned lint pass; 30
further exact allowances disappear. App mutation and dispatch inputs, queue rows
and typed state cleanup remain active work before final slice validation.


Typed queue rows align the original response-part and work-notice columns with
their ordered values; action admission/read, receiver delivery and document image
scans use typed row owners. Document retention accepts its existing Slice 3
snapshot row directly. Fixed SQL literals, defaults, NULL behavior and quota
reservation remain unchanged. A bounded static comparison records 174 unchanged
SQL expressions in slice-7-storage.csv against cb4affbe32cf. Existing row identity,
cross-table rollback, original multipart bytes, caption fallback, finite quota,
action guards, receiver and retained recovery regressions and pinned lint pass.
Two more allowances disappear, leaving 335 repository-wide. All 67 assigned
result rows and 52 baseline B rows now conform to the approved policy; 28 A
inputs and the remaining state/error/close work still precede PR readiness.


All 28 remaining assigned long inputs now group exactly one existing completion,
projection, retained reply, intake batch/update, image binding or app dispatch.
Context and active SQL transactions remain explicit arguments; no capability,
guard, source, timestamp or cursor mutation is relocated across its original
boundary. Result flags retain their distinct named meanings. All 156 assigned
A/B/C rows now meet five parameters excluding the receiver and the final result
policy. The non-short receiver/domain/app Telegram/guided suite, frozen wire and
SQLite fixtures and pinned lint pass; 28 allowances disappear, leaving 307 for
other recorded owners. State/error names, close handling, final reporting and
clean-head controller evidence remain before opening the Slice 7 PR.


Receiver health, operation, delivery, action, conversation and inbox states now
have separate string types and the exact original constants. Type-checked
replacement updates 161 actual state references; SQL, JSON keys and native
workflow status vocabulary remain unchanged. App rendering/model projections
convert the same stored strings at their boundary. Four app error identities
use Error names with unchanged definitions, messages and matching behavior.

Read cursor cleanup joins Close errors into named error results, including early
scan exits. One documented HTTP read-body helper preserves the observed read and
acknowledgement outcome; cleanup cannot turn an acknowledged send into a retry.
The unused serve constructor arguments were already removed by Slice 4's typed
factory. Distinct-column SQLite tests cover new action/notice/response writes and
action/delivery/document reads after updates, including different source/content
hashes and conversation revisions. Swapping positive topic/user IDs makes the
column-name assertions fail; the original mapper was restored. Focused storage,
wire, HTTP, recovery lifecycle and lost-ack tests plus pinned lint pass.

## Slice 8 quiz and provider compatibility

The quiz-domain runtime/action/model and generated artifact values moved without
renaming serialized fields, changing tags, reordering stored fields or changing
state spellings. Observation's unexported filesystem images were never part of
JSON; those images remain private beside the original recoverable authoring
journal. Domain snapshots retain Exists, State, Entries and Revision.

Named receipt lookup retains Result on failure and Found separately from error.
Schema, checksum, digest, operation-ID checks and revision restoration are the
original rules. The journal still writes runtime, raw/curated images, metadata
and acknowledgement in one guarded recovery operation. Existing action recovery,
binding, torn-observation and exact-entry round-trip tests pass.

The provider-neutral schema keeps the original name, description and schema map.
The OpenAI wrapper alone retains Type, Name, Description, Strict, Schema in their
original encoded order, with Store=false and Strict=true. Existing request-body
fixtures pass. Provider provenance is explicit in response decoding; harness and
API identities remain their actual selected providers.

Host delegates existing generation identity/admission, output validation and
trace policy. Bound output and raw HTTP output/accounting remain separate named
fields. Quiz/structured body limits, deadlines, finite clarification/Fast bounds
and the 90-second journal transition context are unchanged. Real HTTP checks
verify authorization and absence of the configured key from provider and opted-in
trace payloads. The original close-error policy remains justified after checked
bounded reads; it does not trigger another acknowledged request.

## Slice 10 harness identity and host compatibility

Core harness identity/profile/selection values have no workflow or configuration
coordinator dependency. Host executable admission and retained generation IO,
runtime plans/launch commands, daemon persistence and trace retention live in
cohesive children. Native configuration retains full-image revision, authorized
selection changes/recovery and independent generation admission fences.

Bindings, capabilities, semantic profile revisions, generation operation/launch
records, metadata and daemon SQL preserve original fields, JSON order/tags, hash
bytes, absence distinctions and state spellings. Runtime/workspace defaults,
canonical and private artifact paths, Docker argv, executable/input/output bounds,
private resource rules and provider errors retain their original meanings.
The declaration move ledger reconciles 327 current source/test declarations and
constant specifications; its current hashes resolve to the actual declarations.
The redundant response trace facade was removed after direct callers moved.

Generation binding returns operation values and selected adapters. Each caller
passes or derives its context; cancellation of one caller cannot poison a fresh
call to the retained operation. The Pi bridge owns and defers its cancellation;
process exchange owns both bounded spools. Runtime holds shared references to an
already admitted slot/guard while native workflow owns inherited FD3/FD5
validation, slot identity/limits and coordinated conversion. No capacity pool or
retained authority was duplicated.

All 38 assigned signatures have precise dispositions: 31 fixed, seven cohesive
five-input contracts retained. Two additional owned resolver contracts return
one preparation value plus error. The global exact debt count is 207 with the
same ceiling seven and one external exception. Boundary dispositions distinguish
actual binary data/dispatch from redundant forwarding and record real regressions.

## Slice 11 ticket record and parse compatibility

Ticket records and indexed projection groups retain their original field order,
text/JSON bytes, timestamps, optional metadata and legacy status aliases under
the ticket owner. Renderer/parser/validation APIs have direct consumers; native
guarded status mutation retains its actual edit authority. Shared integer decoding
belongs to the existing metadata codec with the same empty and error outcomes.

A private timestamp/review-round parse result replaces the five-result helper.
All four source fields are parsed in the original order; any parse failure still
returns no usable data. The indexed SQL row assigns its own nullable/required
time fields in the original order and returns the same errors. No SQL column,
transaction, cache invalidation, schema, null handling or durable value changes.
timePtr is test-only. Constructors accept the existing database.DataDirectory
port without changing open/configure/migration or failure-cleanup behavior.

Native workflow projection methods now belong to an explicit storage composition
that embeds the existing ticket index and holds the same SQLite handle. Native
source locks, integration/cache hashes, query validation and corruption recovery
remain at the workflow owner. Ticket updates retain their original transactional
workflow-snapshot invalidation. No second database pool, cache authority or
forwarding API was introduced; application assembly uses the actual native owner.

Ticket projection storage, row/tag binding, document discovery, reconciliation,
index status and key reservation belong to `core/ticket`. Native projection writes
reuse the same transactional row/tag operation on their existing transaction;
workflow metadata, source authority and commit remain with native orchestration.
Public index upserts retain validation and snapshot invalidation in their original
order. No schema, migration, SQL binding or null/time contract changes.

Ticket file creation stores an explicit key allocator, parameterized by the actual
configured prefix. It retains cancellation and title/synopsis admission before
allocation, UTC time/render/write order and file modes. Application creation passes
its already reserved durable key. Native cross-namespace key observation retains
its complete envelope/domain decoding and bounded path traversal with native
workflow code; ticket-owned filename/text observation preserves number, presence
and read-error meanings. No native schema copy is introduced in the ticket owner.

Readiness Git observation keeps operation name and independent presence in
GitOperationState plus error. Internal repair outcomes retain both the exact repair
record and whether it was applied; worktree status classification retains separate
non-nil runtime/user lists. Normalized worktree configuration keeps the existing
branch policy and workspace together with zero results on admission failures.
These in-memory results do not alter diagnostic/metadata wire fields, Git command
order, filesystem effects, repair permission guards or workspace authority.

Readiness/repair/worktree lifecycle and branch-policy data/defaults/validation are
owned by `core/execution`, with direct native configuration/admission consumers.
The root coordinator is excluded by dependency checks. Check, repair, Git adapter,
status and diagnostic concerns share the same settings normalization after D03
removal. Original serialized fields/specs, IO order and safe-repair/workspace
authority remain unchanged; native fixture helpers stay test-only.

Execution output file ownership travels as one value; opening the validation file
still closes the runtime log on failure, and the caller retains deferred close
order. Runner messages group phase/level/text independently from execution/loop
identity and preserve schema, UTC timestamp and refs. Full and bounded recent
event readers share a named valid/malformed-line result with unchanged diagnostics,
ordering and limits. Execution/loop SQL constructors consume DataDirectory without
changing directory/open/configure/migrate/cleanup semantics.

Execution/loop records, result/failure contracts, runner, event streams, artifact
inspection/path policy and bounded read-only SQL observation now have the execution
owner. Native ticket-file queue policy retains its coordinator. Naming uses the
actual package owner without root aliases. All 156 moved declarations preserve
their original strings, fields, SQL order, nulls, limits, resource cleanup and error
semantics after qualification/naming; the move ledger now covers 723 preimages.
The signature ledger records all 21 completed fixes; ceiling seven retains 187
exact debts and one external exception. Execution component tests use public owner APIs, real SQLite and shipped
migrations; environment and native observation/evidence/cleanup integration tests
retain their coordinator fixtures. Parser/validation placement and the remaining
Slice 11 contracts are still pending.

Ticket parsing helpers retain field presence, empty non-nil lists and next-line
indexes in named outcomes. Native metadata writers retain byte/line ranges and
-1 absence sentinels; document sections retain description/discussion bytes. Key
comparison retains original prefix/integer admission and fallback ordering. These
seven result changes leave 180 exact debts at ceiling seven; actual field/wire/
SQL contracts and guarded write/transition admission remain unchanged.
