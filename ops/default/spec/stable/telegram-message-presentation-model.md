# Telegram Presentation and Document Delivery Model

Status: accepted

Approved: maintainer, 2026-10-21; companion to the accepted presentation specification.

Specification: [Telegram Message Presentation](telegram-message-presentation.md)

This companion describes the durable extension needed for document delivery.
Field and storage names below define the accepted model. Use the existing
receiver user database and notification event database under their current
owners; do not create another transport scheduler or project registry.

## 1. Presentation Value

| Field | Meaning |
| --- | --- |
| version | Presentation schema version, initially 1 |
| family | Inventory family 01 through 48, or a newly recorded family |
| template_revision | Deterministic renderer/template identity |
| initial_text | Human event, project and work context |
| content_kind | Plain text, literal Markdown, code or list |
| content | Bounded main content or document excerpt |
| secondary_text | Optional consequence, wait or next action |
| document_id | Optional reference to the complete frozen snapshot |
| action_ref | Existing retained action reference, when applicable |
| scope_digest | Existing prepared-operation binding, when applicable |
| plain_fallback | Bounded fallback of the same visible content |

The value contains no credentials. It does not copy editable workflow state or
create execution authority. Action choice keys remain unchanged; visible labels
are rendered separately. Store the immutable presentation and its hash before
the first send. Reopening uses that value rather than rerendering current state.

The renderer derives wire text/caption and supported formatting entities from
these blocks. Counts and entity offsets use the final visible string after
truncation. Formatting decisions must not change the document snapshot or scope.

## 2. Document Snapshot

| Field | Meaning |
| --- | --- |
| id | Stable owner-scoped snapshot identity |
| owner_kind / owner_id | Receiver delivery/request or notification event |
| sender_scope / recipient_scope | Existing configured sender and destination binding |
| source_ref / source_hash | Authorized canonical artifact or identified export input |
| content_hash | Hash of the complete attached bytes |
| filename / media_type | Safe human filename and actual file type |
| byte_count | Original byte count, at most 2 MiB |
| payload | Immutable document bytes; nullable only after eligible cleanup |
| created_at / purged_at | Creation and optional completed cleanup time |

Use a snapshot table in each existing owning database. Reuse the same value
contract, validation and uploader across both paths. A snapshot belongs to
exactly one original owner and destination; an identical content hash does not
grant access from another project, peer or notification destination.

Resolve source references through existing authorized artifact owners. Reject
path escape, unexpected symlinks and credential/configuration sources. A display
filename is a basename, not a local filesystem path. Canonical references remain
available for audit without being pasted into routine conversation text.

Admission reserves bytes and the owner's presentation atomically. The 64 MiB
snapshot quota counts all retained payload bytes in that database. Deduplicate
only within a valid ownership binding. Capacity failure cannot partially offer
an action that depends on a missing document.

## 3. Stored Mapping

| Existing owner | Additive mapping |
| --- | --- |
| Receiver outbox | Nullable versioned presentation manifest and document reference; retain the existing delivery ID, event key, peer, config digest, state, attempts and reply association |
| Receiver actions | Optional presentation binding/version added to the current guard; document hash is covered through the sealed manifest |
| Notification event repository | Nullable immutable Telegram presentation manifest and document reference on the existing audit record; retain event identity, status, channel and dispatch policy |
| Document snapshot repository | Owner-scoped immutable bytes and metadata, in the same database as the owner |
| Send acknowledgement | Existing owner retains Telegram message ID and any returned file identity as transport evidence, outside the immutable content hash |

For receiver version 1, the action guard covers the existing prepared-operation
guard and presentation hash. Existing version-0 actions keep their old guard
algorithm. Do not rewrite their prepared input, consumed answer, native receipt
or callback token during migration.

A NULL/version-0 presentation means an existing queued text response or
notification. It follows its retained legacy payload and guards. Unknown future
versions are refused before send or action application.

Physical migration numbers and SQL must follow the then-current schema owner.
Tests must prove old-row reads, atomic new writes, rollback and future-version
refusal. Do not assign a schema number from this planning baseline.

## 4. Delivery and Recovery

A normal text presentation uses sendMessage. A document presentation uses
sendDocument with its bounded caption and optional existing action keyboard.
Upload the retained bytes; do not ask Telegram to fetch a private artifact URL.

| Situation | Required outcome |
| --- | --- |
| First send | Persist manifest/snapshot before IO and consume the existing owner's attempt allowance |
| Confirmed acknowledgement | Retain transport evidence and mark delivery under the current owner |
| Lost acknowledgement | Preserve uncertainty and the same snapshot/action; any retry may duplicate the same notice |
| Reopen | Resume original delivery state and counters; never allocate a replacement mutation |
| Formatting rejection | Retain plain fallback and retry the same document/action within the current allowance |
| Source file changed | Keep the old frozen bytes; a new view/request is required to bind a different scope |
| Source unavailable before snapshot | Return a bounded failure; do not expose a required-document confirmation |
| Revoked destination/configuration | Block transmission and apply existing revocation rules |
| Domain operation already succeeded | Preserve its receipt even if snapshot preparation or sending later fails |

An acknowledged send is not a mutation receipt. A valid bound callback remains
subject to the original operation checks even if delivery acknowledgement was
uncertain. Do not replace a lost acknowledgement with invented success or an
unrelated new confirmation.

Notification dispatch keeps its present trigger and retry policy. Any supported
explicit retry reuses the event's manifest and document. This model does not
introduce automatic notification replay through the conversational receiver.

## 5. Retention and Queries

Retain bytes while any original delivery is unresolved or a related action can
still be used. Only terminal, unreferenced snapshots at least seven days old
are eligible for bounded cleanup. Cleanup preserves content hashes and original
receipts until their existing owner can compact them. A later request for
purged content must explicitly obtain a fresh authorized view.

Required indexed queries are snapshot by owner/destination, retained byte
usage, eligible terminal snapshots and existing due deliveries with their
presentation/document. Unique ownership prevents a second snapshot or manifest
from replacing the first on request replay. Cleanup, action expiry and delivery
reservation serialize with the relevant existing transaction owner.

A test must demonstrate each of these invariants: immutable content, bounded
storage, no cross-destination reuse, no live-action cleanup, unchanged old
records and recovery of the same guarded operation.
