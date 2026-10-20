# Plan Workflow Materialization

The application operation service accepts explicit approved
`GeneratedWorkflowPlan` input through `PlanTrackingRequest.Workflow` or
`MaterializeWorkflowPlan`. Existing requests without this input retain the
legacy tracking path for historical body-only plans. Newly generated plans
already retain order companions; the legacy tracking action reports that native
approval is required without allocating tickets. Listing or reading plans never
adopts records.

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

For a revision-bound generated plan, supply `plan_revision` from its observed
binding, use its immutable `body_ref` as `plan_ref`, and its `plan_body_hash` as
`plan_hash`. The approval must preserve the retained scope, planned IDs, positive
orders and titles. Missing bindings, edited companions and mismatched hierarchy
fail before ticket-key reservation. A revision already bound to another set
cannot allocate another set.

Materialization writes bound `order.json`, `tracker.md`, `binding.json` and
checksummed `projection.json` below that revision, retaining the original
`planned-*` images. It registers v2 at the existing set tracker path and updates
the scoped current companion only when that revision is still current. Retained
file registration stays in the canonical plan manifest. New draft revisions
keep their own unbound maps and do not replace earlier bindings. Authoring
exclusion precedes native workflow exclusion; pending binding journals block a
competing plan publication until recovery.

The complete namespace update is validated and journaled before replacement.
An interrupted journal replays matching images. A conflicting manual image
blocks replay. Retrying the same request returns the original set, ticket keys
and receipt; a changed approved input or membership requires reconciliation.
Recovery releases only an allocation lock matching its recorded ownership.
Record and journal limits also apply to materialization.

Initial native authoring also journals the repository's coordinator marker.
The canonical target must retain that registration before execution begins.
Copied authority cannot register another local execution pool. See
[Coordinator registration](workflow-execution.md#coordinator-registration).

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

Revision-bound amendments require a new observed plan binding and the existing
set revision/source guards. They preserve materialization identity, existing
contracts, task identities, completed work and receipt history. Direct set/task
record writes cannot change that bound structural order. The new binding and
companions commit with the amendment. Execution input carries the current
set-specific map; missing or edited required images block preparation until
native tracker repair. Repair preserves known structural generation time and
requires observed-hash permission for edited output, including retained maps.
Legacy adoption keeps its existing explicit path; a new revision-bound plan uses
native materialization rather than implicit adoption.
