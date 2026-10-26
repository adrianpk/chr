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
