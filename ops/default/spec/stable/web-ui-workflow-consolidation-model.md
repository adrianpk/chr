# Web UI and Workflow Model

Status: Approved

Date: 2026-10-18

Approved: maintainer, 2026-10-18; the explicit request to proceed with Looper accepts this set and its implementation design.

Specification: [Web UI and Workflow Consolidation](web-ui-workflow-consolidation.md).

This companion defines the approved durable changes and read contracts. Existing
Markdown/ops artifact and native workflow owners remain authoritative. SQLite
indexes and Web view models remain rebuildable projections. The approved
set implements and tests these changes with isolated fixtures; applying a
reconciliation to live user data requires its own explicit operation.

## 1. Identity, Scope and Write Binding

| Value | Business meaning | Stored or derived representation |
| --- | --- | --- |
| Project | Existing registered project | Existing project identity and root; never a display-name-derived identity |
| Named scope | Project-wide or one existing use case | Existing project ID plus empty or validated case slug; name is a display value |
| Artifact | One actual document, including multiple specs in one scope | Existing ops namespace, type, ID and validated project-relative body reference |
| Expected revision | Exact content and metadata observed by the caller | Artifact source hash/revision or native ticket revision, compared inside the owning mutation |
| Operation identity | One logical write across duplicate transport submissions | Existing authoring/request receipt identity, bound to target and payload hash |
| Observation | Read result with semantic revision | Rebuildable; excludes volatile render/observation time from the change hash |

Use existing safe path resolvers. Reject traversal, symlink escapes, mismatched
project/namespace/type, unknown case associations and cross-document revision
bindings. Compare revisions after exclusion is acquired. A Web preflight alone
cannot protect against a concurrent writer. Reusing an operation identity with
a different target or payload is a conflict, not an idempotent retry.

## 2. Project Dump Reference and Reconciliation

Extend the existing `ProjectMetadata` owner and frontmatter with one optional
`project_dump_ref` string. It points to the canonical project-relative ops dump
body. Namespace and artifact identity are derived from that validated reference;
do not maintain an independent pointer in every user's namespace or in a cache.
Preserve existing project metadata fields and source-hash write checks.

| Data | Durable representation | Invariant |
| --- | --- | --- |
| Canonical general destination | `project_dump_ref` in existing project metadata | At most one logical general destination for this project |
| Editable current body | Canonical dump body within its existing ops artifact | Save replaces this body; capture extends it |
| Raw/history components | Existing related files and provenance in the owning artifact | Components do not become extra general catalog entries |
| Prior general artifacts | Retained original bodies/metadata and canonical-reference relationship in artifact metadata | Old references remain resolvable; originals are not deleted |
| Reconciliation provenance | Source references/hashes, target identity/hash and operation receipt | Repeatable and auditable without model-dependent content rewriting |

Under project authoring exclusion:

1. Follow a valid existing project reference. A missing target or a conflicting
   set of candidates is an explicit reconciliation error, not permission to
   allocate another destination.
2. With no pointer and one unambiguous existing logical general dump, adopt
   that identity without moving its namespace or replacing its content.
3. With no general dump, allocate one in the default ops namespace, write its
   components and publish the project reference as one recoverable mutation.
4. With multiple general dumps, build a reconciliation preview. Propose the
   earliest valid creation timestamp, then relative path as a deterministic
   tie-breaker, for the retained canonical identity. Explicitly report missing
   or invalid timestamp evidence; never treat the newest file as the answer.
5. Retain every source's original bytes and provenance. Compose source bodies
   in chronological/path order with source attribution; do not deduplicate or
   summarize away material by an LLM. Mark superseded logical artifacts with
   the canonical reference and keep their original content available as history.
6. Apply reconciliation only through an explicit shared migration/mutation
   operation with observed source hashes. Publish content, alias metadata and
   project reference through the existing recovery journal. A retry resumes
   or returns the same result without incorporating the sources twice.

The canonical pointer is discoverable after index rebuild and namespace
selection changes. All general capture paths resolve it, including authoring
sessions and Telegram continuations; per-session artifact allocation cannot
bypass this rule. Old references resolve through a bounded alias check, reject
cycles and do not expose superseded entries as independent active dumps.
The original namespace still governs ordinary case-scoped captures.

Add expected revision and operation identity to shared Save/capture requests
where missing. Preserve unknown frontmatter and artifact metadata on a body
save. Use atomic/recoverable writes for body plus metadata/history. If an index
refresh fails after commit, return committed-content plus projection-stale
information so retry does not duplicate an append or discard a successful save.

## 3. Quiz Runtime, Answer Entries and Mode Queue

Reuse `QuizRuntimeState`, `QuizEntry`, `QuizQuestion` and their repository.
Retain `idle`, `active`, `done`, `cancelled`, `blocked` and Deep/Fast vocabulary.
Answered entries, not artifact metadata or runtime labels, determine whether a
scope has answers. Sort answer history by answered time and stable entry ID.

Extend the persisted state or its owned companion with these semantic fields;
they must share the quiz repository's commit/recovery boundary:

| Field | Purpose |
| --- | --- |
| Session identity and revision | Bind actions to this retained quiz, independently of scope reuse |
| Active question identity | Bind Answer/Skip to the displayed question |
| Fast batch identity and source/context hash | Recognize the one successfully generated flat batch |
| Ordered Fast questions and consumed identities | Resume the remaining queue after answers, skips, mode changes and restart |
| Suspended Fast queue | Preserve pending Fast work while Deep is selected |
| Action identity/result binding | Replay an action without recording a second answer or consuming a second question |

Use the existing `Candidates` representation where it can express these
invariants. Extend it with explicit batch/mode ownership where it cannot; do not
build a second browser-owned quiz repository. Retained schema changes need a
versioned decoder and an explicit legacy interpretation. Never invent answers
or discard entries when older state lacks new queue fields.

The shared action request includes scope, session, expected revision, active
question where relevant and operation identity. Commit answer/skip entries and
the next-question state consistently. If model continuation fails after an
answer is recorded, retain that answer and a resumable blocked/continuation
state; a retry must not record it again.

Switching mode preserves the active question. The specification defines which
queue is suspended/resumed. Only explicit Answer persists the textarea as an
answer. Browser draft values are transient interaction state, distinct from
QuizEntry and retained answers. Cancel and Done preserve existing entries while
performing their different native transitions.

## 4. Spec Documents and Lifecycle

Each manual create allocates one spec artifact through the existing ops owner.
A save targets that artifact, rather than the latest spec for its scope.
Generated specs use the same kind and shared save/lifecycle rules, preserving
existing generation identities where already allocated.

| Business field | Durable field/owner |
| --- | --- |
| Identity, namespace, type | Existing artifact envelope |
| Title, project, optional case | Existing document frontmatter and artifact metadata projection |
| Lifecycle | Document `status`: `draft` or `approved`; normalize existing `accepted`/`final` on read |
| Body | Markdown body, excluding managed frontmatter and retained provenance |
| Revision/source hash | Existing artifact metadata and exact observed document bytes |
| Creation/update times | Existing lifecycle metadata, retained and advanced by shared writes |
| Origin and generation sources | Existing producer/provenance and source references/hashes |

Document lifecycle is authoritative in its shared writer, not in a Web badge.
Do not conflate generic envelope status such as created with spec approval.
Update any duplicated envelope presentation through that same writer; it does
not create a second lifecycle authority.

Manual create supports draft or approved directly. Generated create remains
draft unless an existing explicit approval operation says otherwise. Save
retains selected identity, scope, provenance and state, incrementing the source
revision. A source change invalidates dependent generation observations without
rewriting historical plans or automatically authorizing their execution.

Concrete-document listing enumerates actual spec artifacts in the applicable
workspace, without merging independent documents merely because scopes match.
Maintain the existing latest-in-scope query separately for callers that need it.
Missing-source diagnostics are not invented spec cards.

## 5. Generation Context and Readiness Observation

Build an immutable observation of selected project/case, canonical general
dump, matching case dump material, actual quiz answers and applicable spec.
Each input has a validated reference and exact content hash. Ignore metadata
scaffolding and skipped answers when counting usable discovery. Do not include
unrelated cases merely because they are recent or share a directory.

The readiness decision contains `sufficient`, readable missing-information
reasons, and the source observation hash. Absence of useful discovery fails
before semantic model evaluation. Otherwise use a bounded structured assessment
through the configured generation provider. Validate its output and use normal
capacity/time limits. No independent provider or new background assessment loop
is introduced. A cached decision, if used, is keyed by source, scope, policy
version and model; it remains discardable and cannot authorize execution.

Stage model output against the observation. Before publication, recheck input
hashes and the target plan revision under the existing authoring owner. A
changed input returns a stale result, retaining the staged candidate as private
recovery material where supported. It does not overwrite the current plan or
publish a plan generated from an unacknowledged source revision.

## 6. Plan Revision and Order-Only Companion

Reuse the existing plan artifact, workflow input and materialization identity.
Add or retain the following bindings in owning artifact metadata/retained files:

| Binding | Required value |
| --- | --- |
| Plan identity/revision | Stable artifact identity plus validated body hash and revision |
| Source observation | Selected spec/discovery references and content hashes used for this revision |
| Tracker reference | Exact companion document reference and output hash for this plan revision |
| Intended order | Stable slice/task identities, titles and explicit positive order values |
| Materialization binding | Existing approved materialization/set identity and canonical ticket references, once allocated |

Hash the plan body independently of generated tracker metadata to avoid a
circular hash dependency. Retain revision-specific companion copies inside the
plan artifact's existing `files/` area, for example
`files/plan-revisions/<revision>-<plan-body-hash>/tracker.md`, and register them
in its file manifest. A fresh regeneration has a new revision even when its
body bytes match an earlier plan; a retry of the same committed operation
reuses its revision. Preserve the existing current tracker path for compatibility.
Shared metadata binds that view to the exact plan revision; filenames alone do not.

Before materialization, tracker entries have stable planned identities and
explicitly absent ticket references. The map is non-executable. On approved
materialization, reuse the existing materialization identity to allocate or
reuse canonical tickets and bind the map atomically/recoverably. Repeating this
step returns the same tickets and map, not duplicate allocations. Keep earlier
revision companions and their history when a later plan is generated.

Introduce a versioned order-only projection contract, such as
`chr-tracker/v2`, retaining the tracker artifact family and compatible table
layout. Its structural source hash covers:

1. Project, namespace, plan identity, revision and plan body hash.
2. Approved set/materialization binding when present.
3. Slices with stable identity, order and title.
4. Tasks with stable identity, slice association, task order, title and resolved
   ticket references when materialized.

Exclude whole-record hashes/revisions, status, processing phase, claims,
attempts, review results, delivery receipts and closure state. Output bytes
must also exclude those changing values; retained progress cells are blank.
Generation time changes only when the structural map changes, not on a ticket
checkpoint. Thus a valid map remains byte-identical while tickets advance.

Freshness compares the structural source hash and observed output hash.
Observed-hash replacement protects user-edited exports. Existing v1 files remain
readable historical views; the newly required plan workflow explicitly exports
the v2 map. Do not reinterpret or rewrite frozen tracker-compatible workflows
during a read, ticket transition or upgrade startup.

Publication stages plan, companion and their bindings before replacing current
references. Failure leaves the old plan/map pair usable or a recoverable
transaction with an explicit error; never report complete success with a
required companion missing. A failure of an optional index projection after
the durable commit remains a projection error, not permission to duplicate work.

Regenerating a draft prepares a new revision/map without changing an active
set. Applying approved changes to materialized work uses the existing guarded
amendment operation and quiescence/claim checks. Retain completed task identities
and receipts; no regeneration resets progress, silently reorders a claimed
task or overwrites an earlier revision's map.

## 7. Board Editing and Observation Models

Extend existing central ticket operations and native mutation guards. The Web
form never writes raw ticket files as an independent persistence path.

| Field group | Write ownership |
| --- | --- |
| Title/body/synopsis | Existing ticket document owner; derive a synopsis deterministically if required |
| Classification and named case | Shared enum validation and supported association rules |
| Assignee/owner | Descriptive metadata only; preserve execution claim identity separately |
| Status/resolution | Shared native or supported legacy transition rules and terminal-state requirements |
| Branch/PR/commits | Existing delivery-reference writer; verified receipts cannot be fabricated by form input |
| Comments | Existing append/comment owner with its own operation binding |
| Unknown metadata and discussion | Preserve through all body or metadata edits |

Repository development tickets use their namespaced status-directory contract;
native runtime tickets use their canonical workflow lifecycle. Map display
columns through the owning model, without treating `in_progress`, board labels
and native task phases as interchangeable strings. Fields without a compatible
write path must gain shared support or be explicitly read-only with a reason.

Board is a deduplicated read projection with column totals, loaded-page entries,
attention cases and continuation. Keep its identity stable across native and
derived index rows. Native set membership/order and active claims remain owned
by workflow records. Visual drag/drop updates status only through the same
guarded request used by the form; it never changes slice/task order.

Dash observation joins queue, invocation, attempt, ticket and event identities
within one project and authority mode. It keeps stable event identifiers for
deduplication and result records for completed work. These are derived views,
not new durable execution records. Observation timestamps and stale flags are
presentation metadata, not evidence of lifecycle progress.

## 8. Migration, Compatibility and Invariants

1. Implement the shared writes and versioned readers before exposing controls.
   No schema or artifact migration runs as a consequence of GET or polling.
2. Exercise canonical-ops-only, supported legacy and mixed projections. Repair
   indexes from authoritative documents; never require users to copy tickets
   back into obsolete storage to make Board or Dash load.
3. Reconcile old general dumps through the explicit preview/apply operation,
   with retained originals and expected hashes. The pointer and alias relation
   are recoverable after interruption. Never silently discard content.
4. Write spec state and plan/tracker bindings through the same journaled owners
   as their bodies. Extend existing readers for older documents without
   guessing approval or changing active work during migration.
5. Update the ticket-native tracker contract, source-hash tests and relevant
   adapter documentation when the new workflow is implemented. Keep existing
   tracker naming; separate version compatibility from a future rename.
6. Validate file manifests, revision/hash rules and index rebuilds together.
   Only add a database migration if the existing canonical repository actually
   persists a new field there; derived indexes must not acquire write authority.
7. Use fixtures for interrupted publication, stale writes, retry and concurrent
   namespaces/runners. Verify that ticket lifecycle progress never changes the
   structural tracker hash, output bytes or required freshness state.
