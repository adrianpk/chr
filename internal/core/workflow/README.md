# `internal/core/workflow`

## Purpose

The package owns executable workflow records and the coordinator that advances
them through admission, execution, review, integration and closure.

## Boundary

`Store` owns repository locking, journals, guarded source-image checks, atomic
multi-record writes and recovery. Every method of the store stays with its type.
Core composition and application adapters call this owner directly. The package
cannot import root core or application adapters.

Consumer-owned contracts in `integrate`, `merge`, `review`, `forge` and `closeout`
serve live validation, evidence, decision, response and gate operations. Bound
implementations retain private claim/prepared/provider or journal identities;
the merger evidence contract is implemented directly by the store. Stage policy
extraction uses these operation boundaries without moving mutation authority.

Shared retained values live in `workflow/state`. That package has a separate,
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
