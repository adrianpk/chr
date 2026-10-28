# `internal/core/workflow`

## Purpose

The package owns executable workflow records and the coordinator that advances
them through admission, execution, review, integration and closure.

## Boundary

`Store` owns repository locking, journals, guarded source-image checks, atomic
multi-record writes and recovery. Closing history, receipt and retry policy now
belong to closeout through a read-only evidence adapter. Frozen input assembly
and actual-member policy use the stage InputSource and shared native record data.
Remaining Store methods
are still being reconciled under Slice 15; their count does not establish final
persistence ownership.
Core composition and application adapters call this owner directly. The package
cannot import root core or application adapters.

Consumer-owned contracts in `integrate`, `merge`, `review`, `forge` and `closeout`
serve live validation, evidence, decision, response, checkout and command operations. Bound
implementations retain private claim/prepared/provider or journal identities;
the merger evidence contract is implemented directly by the store. Stage policy
extraction uses these operation boundaries without moving mutation authority.

Shared retained values, including native records/snapshots and delivery proofs,
live in `workflow/state`. Record/snapshot validation stays in explicit host
functions at the original boundaries; the shared data owns no validation or IO.
That package has a separate,
narrow import contract and cannot import this coordinator or its configuration.
Ticket projections and SQL indexing use the existing ticket and database owners;
the parsed workspace configuration has its own owner under `workspace/config`.

## API

`NewWorkflowStore` constructs a store over an explicit layout and project ID.
Its operations retain the existing identity, revision, receipt and recovery
guards. `ReadLegacyTicketFile` refuses adopted sources. `ReplaceLegacyTicketFile`
takes the existing repository or legacy file lock, rechecks adoption and the
observed image, then atomically replaces an unadopted source.

## Validation

Tests live beside the owning behavior. Wire fixtures retain exact baseline bytes;
Git, filesystem, SQLite and process fixtures exercise durable boundaries without
exporting private state to tests.

```sh
go test ./internal/core/workflow -run '^TestWorkflowRoundTrip$' -count=1
```

## Publication operation values

Transient publication results distinguish `SubmissionIntent` dispatch permission
from `NativePublicationResult` transport output, canonical observation and
receipt-backed finalization. Private values bind each callback to its current
entry/budget/journal image and pair read-only canonical/Git observation ports.
Contexts remain explicit and Store retains claim, journal, lock and receipt
transactions. Internal `Handled` and `Observed` flags preserve native/publication
barriers on partial errors; they do not assert completed delivery.

## Provider protocol boundary

`workflow/forge.Client` owns provider request construction, bounded PR discovery,
JSON decoding and structured PR creation. The host adapter resolves enabled
configuration and canonical URLs/repositories, chooses the bounded command runner
and retains prerequisite diagnostics before delegating IO. Immutable PR inputs
live in `workflow/state`. Guarded PR intents, uncertain outcome reconciliation,
source identity checks, approval and receipt transactions retain their host owner.

Native publication policy and observation classification live in forge. The host
keeps callback-bearing transport requests and projects only immutable identity.
Its stateless native proof adapter passes the explicit expected query and proof
to the existing integration validator. Enabled configuration, current host grants,
process lifetime, intent fences and receipt transactions retain their owners.
Native retained values and the history-divergence sentinel live in state; direct
consumers use those owners without compatibility aliases.

## Closing execution boundary

`closeout.Execute` owns ordered command policy and unconfirmed receipts through
consumer-owned checkout and command ports. The host supplies the watched finite
deadline, supervises processes and retains current-input/candidate checks and
atomic confirmation. Named closing admission and proof results preserve optional
claims, retained completion, partial proof and hash outcomes. The closing command
exchange binds one process, permit and bounded output pair to its exact argv and
output base. Unused projection context forwarding is removed; journaled set and
invocation changes still commit under the existing lock.

`CloseSet` and `CloseInvocationSet` are coordinator operations over the guarded
store. A private `workflowClosingHost` owns the watched context, exact checkout,
process permits, bounded outputs and waits. The finite Workspace/Commands adapter
binds that host to the original private claim. Store retains admission,
credentials, process registration/completion, confirmation and recovery under the
same critical sections. The application and continuation callers use these
operations directly; no Store forwarding methods or promoted store fields are
introduced.

## Task execution boundary

`ExecuteTask`, `ExecuteTaskWithHarness`, `ExecuteSlice` and `ValidateSlice` are
coordinator operations. `PrepareManualTask`, `ExecuteManualTask` and
`ReconcileTask` keep manual handoff, process execution and verified acceptance
separate. The application service retains its adapter-neutral operation API and
calls these functions directly.

A private `workflowTaskHost` owns task and aggregate deadlines, ordered gate
commands and partial execution results. The shared stop watcher owns its ticker,
goroutine and cancellation; it reads intent through the original guarded Store
path. Store retains input freezing, credentials, process records, replay checks,
candidate fences and complete result/claim transactions. A successful manual
worker exit still requires explicit candidate reconciliation and host validation.

## Review, merger and publication lifecycle

`PrepareIntegration`, `ReviewSlice`, `ReviewSliceWithHarness`, `ExecuteMerger` and
`ExecuteMergerRepair` coordinate the existing guarded operations. Private review
and merger hosts own role deadlines, stop watches, workspace handles, supervised
commands and partial execution evidence. `MergerRepairPorts` pairs the provider
fence with the trusted repair Git capability. An accepted repair still requires
candidate validation; it does not grant review or publication.

Review admission and human-only restriction policy use `workflow/review` directly.
Store reads the configured policy and applies the returned admission under its
existing lock. A private `workflowPreparationHost` executes ordinary or candidate
repair Git preparation under the original admission guard and lease; Store
retains accepted repair reads, current source fences and complete claim/input
and result transactions.

`SubmitIntegration`, `ReconcileSubmission` and `PublishNativeSubmission` keep
bounded transport and observation sequencing outside Store. Store retains the
original intent/exclusion, BeforeSend checks, response/observation journals,
claims, budgets, explicit approval and complete acceptance transactions. Actual
service and continuation callers use these operations directly.

`AdvanceWorkflowSliceIntegration`, `ContinueWorkflow` and
`ReconcileWorkflowNativeIntegration` are coordinator operations over those same
guarded owners. The private continuation host runs only the accepted bounded
step loop. It preserves membership refresh, retry authority, barriers and partial
step counts. Explicit native reconciliation selects one observation; continuation
selects bounded polling. Both re-read current guards between Git reconciliation,
canonical observation and receipt confirmation. The poll-only forwarding method
is removed; Store retains journal/projection, acceptance and checkpoint commits.

`ControlWorkflowLifecycle` freezes the original request before guarded acceptance
and runs continuation only after that acceptance returns. Store owns the original
minute-bounded lock, request reservation, revision/authorization checks and atomic
action/receipt commit. It cancels the acceptance context before continuation uses
the caller context. Verified receipt replay returns immediately; it never starts
another checkpoint. The existing frozen request/digest value crosses this boundary
without changing JSON, hashes, partial results or detached settlement.

## Bounded operation results

Capacity observations pair the retained process with liveness; owned worker
reservations pair the slot and shared guard through the existing cleanup owner.
Continuation results retain accepted steps, barrier reasons, membership-change
observations and role counts. Control receipt lookups distinguish normal absence
from verified original receipt recovery without reserving or executing new work.
Frozen requests pair the exact original JSON digest with their request value.

Amendment preparation retains partial materialization and its atomic file set.
Frontier proof uses one retained invocation/receipt/candidate unit; target and
proof observations survive their original errors. Upgrade preparation retains
journal/image outcomes while journal presence remains distinct from validity.
All writes, locks, authority checks and recovery retain their existing owners.
