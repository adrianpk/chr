# Slice 1: Guarded Workflow Records

Status: reviewing
Delivery set: ticket-native-workflow
Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)
Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)
Branch: `feat/slice-1-workflow-records`
PR: pending

## Purpose

Persist explicitly versioned workflow records with scoped identities, guarded
writes and recoverable shared transitions. Files remain authoritative.

## Delivered Behavior

- Native task, set, slice and integration-receipt contracts reject unsupported
  schemas, partial associations, foreign ownership, duplicate keys/orders,
  dependency cycles and unsupported lifecycle claims.
- Canonical records use allocated `ops-layout/v1` envelopes with a
  `chr-workflow/v1` domain. Explicit Markdown adoption preserves the ticket key,
  creation instant, original path, exact body and compatible native fields.
- Writes check observed revision and source image. Envelope metadata, domain
  extension fields, slice extension fields and closing-validation extensions
  survive round trips. Receipts are immutable after creation.
- One process-held lock in the Git common directory excludes cooperating
  writers across linked worktrees and namespaces. Legacy replacements recheck
  their observed image under that lock and cannot overwrite newly adopted data.
- Multi-file transitions journal pre/post images before the first replacement.
  Matching interrupted writes replay; incompatible edits block further writes
  and reads. Replacement files and parent directories are synchronized.
- Additive SQLite migration 14 extends the existing ticket index and derives
  set, slice and receipt queries. Namespace-scoped uniqueness, stable bounded
  pages and snapshot hashes prevent stale-cache selection. Rebuilding a deleted
  database preserves shared state and task ordering.
- Admission rejects more than 4096 records, pages over 128 results and files or
  transition journals over 1 MiB with explicit errors.

## Implementation Notes

`WorkflowStore` owns persistence and repository coordination in `internal/core`.
Its constructor performs no IO. `WorkflowIndex` is implemented by the existing
SQLite ticket index and returns validated file records. Generic ticket-index
mutations invalidate the corresponding workflow snapshot.

Subprocess fixtures exit after the first replacement to exercise recovery.
A separate process holds the lock while a linked worktree attempts a read.
Competing writers admit one observed revision. Tests also exercise explicit
adoption, title changes, checkout relocation, namespace collisions, pagination,
cache deletion, manual conflicts and admission bounds.

## Contracts Added or Changed

- `WorkflowIdentity`, `WorkflowTask`, `WorkflowSet`, `WorkflowSlice`,
  `WorkflowReceipt`, `WorkflowSnapshot` and typed dependency references.
- `WorkflowStore.Create`, `Adopt`, `Read`, `Write`, `Transition`, `Snapshot`
  and `CreateReceipt`.
- `WorkflowIndex.RebuildWorkflowIndex` and `QueryWorkflow` with exact snapshot
  binding and bounded continuation offsets.
- Existing legacy ticket mutations now use repository exclusion and image
  guards. Review-result application carries the caller's context through lock
  acquisition. Existing unadopted ticket behavior remains available.

## Files of Interest

- `internal/core/workflow_record.go`, `workflow_validation.go`,
  `workflow_snapshot.go`: domain contracts and scoped invariants.
- `internal/core/workflow_store.go`, `workflow_codec.go`,
  `workflow_markdown.go`: canonical and adopted representations.
- `internal/core/workflow_lock.go`, `workflow_transition.go`,
  `workflow_legacy_write.go`: exclusion, durable replay and legacy boundaries.
- `internal/core/workflow_index.go`, `workflow_index_query.go` and
  `cmd/chr/assets/migration/sqlite/20261008000014-add_workflow_records.sql`:
  the rebuildable projection.

## Validation

Successful local task checks:

- `make verify`
- `go test ./internal/core -count=1`
- `go test -race ./internal/core -run '^TestWorkflow' -count=1`
- `git diff --check`

The required final sequence is `make verify`,
`go test ./internal/core -count=1`, `go test -race ./internal/core -count=1`
and `git diff --check`. The controller records that sequence against the exact
clean PR-reference follow-up HEAD before handoff. These checks cover Slice 1;
the composed delivery-set gate belongs to the approved final validation phase.

Acceptance: A02 and A22; record, file and pagination foundations for A24.

## Risks and Follow-ups

An incompatible third journal image requires explicit reconciliation and blocks
further shared operations. Process exclusion uses the host's Git common
directory and advisory file locking. Cross-host coordination is not provided
by a matching path. Subsequent approved slices compose these persistence
contracts with materialization, execution and integration verification.
