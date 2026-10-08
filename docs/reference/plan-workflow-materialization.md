# Plan Workflow Materialization

The application operation service accepts explicit approved
`GeneratedWorkflowPlan` input through `PlanTrackingRequest.Workflow` or
`MaterializeWorkflowPlan`. Existing requests without this input retain the
legacy tracking path. Listing or reading plans never adopts records.

The input has a stable materialization request ID, source plan reference and
observed SHA-256, use-case association, closing gate, and ordered slices. Each
slice has a stable ID, a display label, positive local order, aggregate contract
and validation gate. Each task has a separate stable input ID, display label,
positive slice-local order, outcome, scope, acceptance, validation and native
classification. IDs must not equal display labels. References are relative to
the selected ops namespace root. Commands are ordered argv arrays.

Materialization allocates the existing ops envelope identities and native ticket
keys separately. It writes one task ticket per task, a set-owned descriptor per
slice, immutable contract files and an immutable materialization receipt. Task
tickets own membership. Slice descriptors retain its ordered fingerprint, not
another editable task list. Task dependencies naming input task IDs resolve to
the allocated native keys; other references must resolve in the same namespace.

The complete namespace update is validated and journaled before replacement.
An interrupted journal replays matching images. A conflicting manual image
blocks replay. Retrying the same request returns the original set, ticket keys
and receipt; a changed approved input or membership requires reconciliation.
Recovery releases only an allocation lock matching its recorded ownership.
Record and journal limits also apply to materialization.

`AdoptWorkflowPlan` accepts a selected legacy plan envelope, its observed
metadata/body hash, and an exact mapping from approved tasks to observed legacy
Markdown tickets. The plan retains its envelope ID and creation time. Tickets
retain their keys, paths, creation times, bodies and compatible unknown fields.
The approved task body must match the selected original body. Adoption observes
legacy execution records and holds legacy operation exclusion across linked
worktrees, together with repository workflow exclusion. Ambiguous keys, foreign
scope, unresolved associations, active attempts and unfinished legacy execution
block adoption. Existing backlog and ready-for-development tickets can be
adopted; other lifecycle states require explicit evidence reconciliation.

An adoption receipt binds the selected images. Matching interrupted adoption
replays and matching retries return the existing records. Adoption never infers
integration or closure receipts from legacy status.
