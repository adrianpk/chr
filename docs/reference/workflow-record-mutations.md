# Workflow Record Mutation Receipts

`WorkflowStore.ApplyRecordMutation` and the central application service's
`ApplyWorkflowRecordMutation` provide request-bound creation and replacement of
native ticket and plan records. A plan record owns its slice-set descriptors.
Work admission retains the existing `WorkflowStartRequest.RequestKey` contract.

## Request identity and guards

A `WorkflowMutationBinding` contains an opaque request ID and operator identity.
Request IDs contain 1–128 ASCII letters, digits, dots, underscores, colons or
hyphens, starting with a letter or digit. Operators contain 1–128 bytes without
NUL, CR or LF. Leading and trailing whitespace is normalized on these fields
and the creation title. Body bytes remain unchanged.

The normalized argument digest includes the project, namespace and complete
mutation request. A valid identity is reserved before semantic preparation.
A failed preparation retains that reservation: changed arguments require a new
request ID. Identity and encoding bounds are checked before reservation.

Creation uses expected revision zero and the existing namespace-global artifact
allocator. The request includes the canonical ticket key where applicable.
The owner reserves the artifact path before journaling any record effect.
Replacement requires the observed positive revision and complete source hash,
and passes the same lifecycle, snapshot, coordinator and ownership checks as
ordinary workflow writes.

## Durable owner state

The existing repository lock excludes concurrent mutations. Immutable request
reservations live under `run/operation-bindings/` in the namespace. Original
confirmed results live under `run/operation-receipts/`. Filenames hash the
opaque request ID. Records contain schema, operator, target scope, operation,
argument digest, observed guards, state and result references.

The existing workflow transition journal contains the intended record images
and result receipt in one transition. Recovery accepts only an observed before
or intended after image, completes that transition and returns its original
result. A divergent image preserves the unfinished journal and blocks the
operation with `ErrWorkflowConflict`. Pending allocation ownership remains
retained until that journal finishes.

A successful replay returns its original confirmed result even if a later
independent request has changed the record. Observe the current record before
issuing a new revision-guarded mutation. Receipts contain no adapter routing
metadata; callers can reuse the same binding through any application adapter.

## Bounds

Normalized arguments are at most 64 KiB. A namespace retains at most 4,096
request reservations, including failed preparations. Existing reservations and
confirmed receipts remain replayable at capacity. The existing 1-MiB workflow
file/journal limit and record lifecycle bounds also apply. Capacity failure
creates no new canonical record.

## Use-case metadata

`CaseMetadataService.SaveBound` accepts the same caller-neutral binding with a
project, namespace and metadata image. Creation allocates the existing case
slug under owner exclusion. Updates require the observed source hash and exact
case path. The owner stores relative references in receipts and resolves them
against the current project root for callers.

The adjacent runtime `case-mutations` directory retains immutable bindings,
original receipts and a two-image pending journal. Existing metadata saves use
the same owner lock and recover that journal before writing. Divergent pending
images block writes and preserve the journal. Binding and argument bounds match
the native record contract. `SaveContext` carries caller cancellation through
ordinary metadata writes.
