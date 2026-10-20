# Ticket-Native Work Model

Status: accepted

Date: 2026-10-08

Approved: 2026-10-08

Specification: [Ticket-Native Work Execution](ticket-native-workflow.md)

All fields and schema additions below are part of the approved contract and
remain implementation requirements. They extend chr's ticket,
ops artifact, execution and integration models. They do not introduce a second
application service, private scheduling database, or independent adapter state.

## 1. Storage and Authority

Use chr's project-local `ops/<namespace>/` layout and existing artifact envelope
allocator. Canonical artifact metadata keeps `ops-layout/v1`; new workflow domain
objects have an explicit `chr-workflow/v1` discriminator. Do not reinterpret an
unversioned plan, tracker, or arbitrary Markdown file as an executable set.

| Record | Durable owner | Derived representation |
| --- | --- | --- |
| Task ticket | Canonical `ticket/<id>/metadata.json` domain object and `artifact.md` body, or an explicitly adopted existing Markdown ticket with frontmatter and body. | Existing ticket index and board rows. |
| Slice set and slices | `plan/<id>/metadata.json` workflow object; `artifact.md` owns the plan narrative. | Plan view and optional tracker. |
| Invocation | `run/<id>/extension/chr/workflow.json`. | Run envelope/body, CLI/TUI/Web status, existing loop rows. |
| Attempt and claim | Invocation attempt record plus repository coordination lock/reservation. | Existing execution rows and runtime logs. |
| Integration intent/receipt | Immutable receipt and journal beneath the owning invocation's `evidence/` and `extension/chr/`. | Common slice delivery references on task tickets. |
| Set closure | Immutable owning-set closure receipt with a guarded reference on the set record. | Set status and tracker completion. |

One ticket uses exactly one adopted representation. Adopting an existing ticket
preserves its key, creation instant, body, compatible unknown fields and original
location. Canonical envelope identity/revision/timestamps are not duplicated in
its domain object. Markdown stores the equivalent values in frontmatter.
Existing `.chr/` configuration and protocols remain at their owned locations;
this specification does not relocate project configuration.

Membership has one owner at each level: the set owns ordered slice descriptors;
task tickets own task-to-slice associations and order. A tracker or plan body
cannot provide an alternate editable task list. Rebuild discovers membership
from tickets and validates it against the slice contract's immutable task-input
fingerprint before execution or closure.

## 2. Identity and Ticket Fields

Identity is `(project_id, namespace, id)`. For existing tickets, `id` is their
stable chr `key`; adopting a canonical envelope retains that key as the domain
ticket key instead of substituting an allocated artifact directory ID. Artifact
ID and ticket key are separately typed references where both exist.

Keep current ticket fields and append these execution fields in this order:

| Field | Stored shape and invariant |
| --- | --- |
| `workflow_schema` | `chr-workflow/v1` for adopted/new executable records. |
| `project_id` | Stable registered project identity; required, never derived from path. |
| `namespace` | Required active ops namespace; references cannot cross it implicitly. |
| `revision` | Positive monotonic revision; observed value required for updates. |
| `use_case` | Existing optional stable use-case reference. For grouped tickets it is derived from the set or agrees with it. No title-based resolution. |
| `slice_set_id` | Owning set identity, absent for independent tickets. |
| `slice_id` | Owning slice identity, absent for independent tickets. |
| `task_order` | Positive integer unique within a slice; absent for independent tickets. |
| `depends_on` | Bounded typed `{kind, id}` references to tickets/slices in this project and namespace. |
| `task_phase` | `pending`, `running`, `validated`, `waiting_slice`, `blocked`, or `integrated`. |
| `execution_contract_hash` | Digest of outcome, scope, acceptance, validation, associations, order and dependencies. |
| `active_attempt_id` | Optional current attempt reference, guarded by claim generation. |
| `task_evidence` | References to immutable task validation, cumulative input/output trees and accepted commits. |
| `slice_delivery_ref` | Common immutable slice receipt reference; absent until verified. Independent tickets use their own receipt. |

The association triple is all absent or all present. Grouped association changes
require a quiescent explicit amendment. A stable `slice_id` resolves to exactly
one set descriptor; task membership cannot reference an arbitrary textual `slice`.
Existing display/source `slice` remains readable and does not assign ownership.

Existing `priority`, `severity`, `project_area`, `type`, `tags`, `branch`, review
fields, merge fields, timestamps and body remain CHR-native. Their owning
serializer preserves compatible unknown fields and descriptions. Shared slice
branch/PR/merge projections may appear on several task tickets, but the delivery
receipt owns their immutable facts. Conflicting projections block reconciliation.

### Lifecycle Mapping

| Task phase | Ticket status | Required evidence |
| --- | --- | --- |
| `pending` | `backlog` or `ready_for_dev` | Approved contract before ready admission. |
| `running` | `in_progress` | Current claim and attempt. |
| `validated` or `waiting_slice` | `in_progress` | Accepted task result bound to cumulative branch trees; no completion claim. |
| `waiting_slice` | `ready_for_review` | All task evidence, aggregate validation, slice branch and one canonical PR. |
| `waiting_slice` | `ready_for_merge` | Applicable review decision, candidate and policy binding. |
| `blocked` | Existing `blocked` value | Reason, affected phase and next-action owner. Validated siblings retain their own phase. |
| `integrated` | `done` | Verified shared delivery receipt and replayable finalization. |

Independent tickets use the same status vocabulary, but their task candidate is
also the integration unit. Legacy aliases are accepted for reading, never as
sufficient integration proof. Validation status and delivery status are separate.

## 3. Slice Set and Slice Descriptor

| Set field | Stored shape and invariant |
| --- | --- |
| `workflow_schema` | `chr-workflow/v1`. |
| `id`, `project_id`, `namespace`, `revision` | Stable envelope identity and revision. |
| `use_case` | Optional stable use-case association. |
| `use_case_order` | Optional positive integer unique among this use case's sets. All active selected sets either have explicit orders or use chronological fallback; mixed ordering is invalid. |
| `created_at` | Immutable UTC time used by chronological ordering. |
| `status` | `proposed`, `ready`, `running`, `closing`, `completed`, or `blocked`. |
| `plan_ref`, `plan_hash` | Owning narrative reference and approved content digest. |
| `slices` | Ordered descriptors below; each identity belongs to this set only. |
| `closing_validation` | Required flag, ordered argv commands, timeout, explicit reason if waived. |
| `closure_ref`, `closure_history` | Confirmed current receipt and immutable prior closures after explicit reopen. |

| Slice field | Stored shape and invariant |
| --- | --- |
| `id`, `slice_order` | Stable unique ID and positive set-local unique order. |
| `title` | Editable label, never identity. |
| `contract_ref`, `contract_hash` | Scope, aggregate acceptance and validation contract. |
| `task_input_hash` | Approved ordered task identities and execution-contract digests; detection snapshot, not an editable second membership list. |
| `depends_on` | Explicit typed dependency references, validated with implicit predecessor edges. |
| `phase` | `pending`, `developing`, `validating`, `reviewing`, `ready_for_merge`, `merging`, `awaiting_observation`, `integrated`, or `blocked`. |
| `branch`, `candidate_ref`, `pull_request` | Retained cumulative branch, immutable candidate revision and canonical PR identity. |
| `delivery_ref` | Verified integration receipt shared by every task. |

Empty slices reject preparation. Empty sets require an explicit closing contract.
Membership additions cannot change a running invocation's frozen scope. Amendment
increments the set revision and requires no live owner; reopening retains all
earlier integration/closure evidence and never changes terminal invocation scope.

## 4. Invocation, Attempt, and Claim

An invocation stores `id`, project/namespace, request idempotency key, authority
mode, selector kind/identity, normalized filters, ordering rule, later-admission
policy, canonical target,
frozen ordered member identities/revisions/hashes, confirmed cursor, counts,
active slice/task, resolved configuration snapshot, stop intent and blocker.
Accepted domain transitions advance the expected current revisions while
retaining the initial admission snapshot. External conflicting writes do not.
With explicit later admission, newly accepted identities/revisions are appended
strictly after the confirmed frontier and within the original selector/filters.
Closure always observes all actual members, including ones outside a frozen
selection; omitted members prevent whole-set completion.

Statuses are `pending`, `running`, `waiting_capacity`, `waiting_review`,
`waiting_integration`, `blocked`, `completed`, `paused`, and `cancelled`.
`waiting_integration` distinguishes authorized submission from observation using
the candidate phase; status alone cannot license a retry.

An attempt stores `id`, invocation ID, task/lane identity, generation, replacement
link, owner token hash, lease expiry, process reservation identity, input ticket
revision and contract hash, prerequisite snapshot, observed target SHA, cumulative
base tree, owned branch/worktree, outcome, immutable evidence, start/finish times
and consumed repair/retry budget. Private tokens are excluded from public events.

Repository coordination uses the canonical Git common directory and normalized
remote repository identity. Capacity reservations are process-held and shared
across linked worktrees and namespaces. Record claims and result application use
fencing generations. Expiry prevents result application but cannot release a slot
while its worker process is alive. No cooperative writer can replace a live claim.
The worker or its surviving supervisor holds the process reservation for the
worker's actual lifetime; a host crash cannot leave an uncounted live child.

Claim resource keys cover repository lane capacity; exact ticket; owning slice;
and ordered set/use-case frontier. Target publication keys additionally include
canonical remote repository and target branch, independent of namespace/project
aliases. Different hosts/clones cannot infer shared locking from matching paths;
they require the same registered coordinator or fail admission.

## 5. Candidate, Submission, and Receipt

A candidate revision contains the integration-unit identity, ordered task
revisions/contracts, source branch/commit/tree, observed target/base, method,
prepared result tree/history fingerprint, validation references, review result,
policy source/revision, publication source/base guard and optional exact-candidate
approval binding. Provider capability must protect that guard atomically, rather
than merely report the base observed before submission.

An integration journal records an idempotent submission ID, candidate revision,
canonical PR identity, source and base guards, method, prepared result, policy
binding, attempt number, submission time, response or observation state and target
exclusion. Record it atomically before provider IO. Unknown outcomes cannot be
cleared by lease expiry, cancellation or an application restart.

A receipt stores the immutable integration-unit/candidate/submission identity,
canonical provider/repository/PR identity, method, submitted source/base, prepared
tree/history proof, provider result commit, verified tree, observed target SHA,
verification reference, validation references and confirmation time.

Method verification is explicit: merge preserves submitted source/base with the
validated result tree; rebase produces the prepared ordered linear history on
the validated base; squash produces exactly one combined commit on that base.
The result must be reachable from the observed target. Later target commits do
not invalidate a receipt that remains reachable with matching proof.

Ticket finalization is a journaled transition keyed by receipt and member key.
Each task receives its common PR, branch, merge facts, receipt reference and final
status once. Partial finalization blocks successor admission until replay finishes.
No per-task PR is allocated during this fan-out.

## 6. Closure and Projection Records

A closure receipt binds set identity/revision, ordered slice IDs and receipts,
every actual task ID/revision, plan/closing-contract hashes, exact target commit
and tree, validation references, waiver reason when applicable and confirmation
time. Changed inputs or target during validation prevent confirmation.

Tracker projection records retain source receipt/revision hashes, output hash,
generation time and `current`/`stale` state. A failed export never rolls back a
domain receipt. A manually changed projection cannot be replaced silently.
Transition journals and receipts can repair run/SQLite/tracker projections
without claiming work.

## 7. Guards, Indexes, Queries, and Migration

All cooperating record writers use repository coordination and one process-safe
record mutation lock. Selection/claim snapshots take that lock too; reads cannot
treat a pending multi-file journal as committed state. Updates require observed
revision plus claim generation
when work is owned. Replace individual files atomically; journal multi-record
changes with expected old/new hashes before the first replacement. Replay only
matching pre/post images. A third incompatible image requires reconciliation.

Rebuildable SQLite extensions to the existing ticket index carry project,
namespace, key, representation/body reference, revision, execution hash,
set/slice IDs, task order, phase and delivery reference. Derived workflow tables
index sets/slices, invocations and receipt references. Do not create a second
SQL-only owner for any shared lifecycle field. Existing execution records retain
attempt diagnostics and link to their shared run record.

Required uniqueness and query contracts are:

- Unique ticket key and adopted representation per project/namespace.
- Unique slice ID within a project/namespace; unique `(set_id, slice_order)` and
  `(project_id, namespace, slice_id, task_order)` for grouped tickets.
- Unique live claim per protected resource; fencing generations always increase.
- One canonical PR per slice; one confirmed receipt per candidate/submission.
- Queries for ordered tasks, ordered slices, use-case sets, missing prerequisites,
  loose eligible tickets, live ownership, submitted-unobserved publication and
  unfinished ticket finalization. Every query validates shared input revisions.

SQL migrations are additive to the repository-owned SQLite migration sequence.
New indexes mirror these queries; partial uniqueness excludes absent standalone
associations. File-domain validation enforces the same rules before SQL refresh.
Rebuilding a deleted cache must preserve selection and completion from shared
records. Losing runtime diagnostics cannot authorize another implementation.

Do not bulk-convert existing data. Explicit adoption resolves stable project,
namespace, use-case identity, hierarchy/order, contract readiness and current
delivery while retaining original content. Selected migration is quiescent,
journaled and idempotent; inferred legacy status alone creates no receipt.
Upgrade snapshots prior effective defaults before writing any version marker.

## 8. Configuration and Bounds

These are approved additions to chr project configuration, pending delivery. Existing
`review.mode` and `work.workspace_mode` keep their current names.

| Key or bound | Version-1 default | Allowed values / behavior |
| --- | --- | --- |
| `workflow.version` | `1` for new projects | Absence selects existing defaults; never inferred from a filename. |
| `work.state_authority` | `tickets` | `tickets` or explicit `tracker_compat`; one owner per selected workflow. |
| `review.mode` | `trusted_auto_merge` | Existing three modes. Legacy absence remains `human_review`. |
| `work.workspace_mode` | `worktree` | `project` allowed only for explicitly selected serial legacy execution. |
| `work.merge_strategy` | `rebase` | Existing `merge`, `rebase`, `squash`; provider capability/result proof required. |
| `work.max_active_workers` | `1` | `1..8`; shared repository pool. |
| `work.ready_queue_capacity` | `64` | `1..1024` unfinished admitted lanes; overflow rejects new admission without dropping accepted work. Accepted work may wait for worker capacity within this bound. |
| `work.worker_timeout_seconds` | `3600` | `1..86400`; expiry does not free a live process reservation. |
| `work.validation_timeout_seconds` | `3600` | `1..86400`; ordered commands share the gate deadline. A smaller contract deadline wins. |
| `work.checkpoint_step_limit` | `128` | `1..128` accepted task/closure transitions per call; checkpoint on exhaustion. |
| `work.include_later_tickets` | `false` | Captured at admission; explicit `true` admits matching ready work strictly after the confirmed frontier of an unfinished invocation. |
| `work.max_worker_attempts` | `3` | `1..128` per task/invocation, including crash replacement; explicit extension is required after exhaustion. |
| `work.max_integration_attempts` | `3` | `1..10` submissions per integration unit/invocation; candidate revisions inherit consumed attempts. Uncertain observation allocates no fresh submission. |
| Automatic review repairs | `3` | Per integration unit/invocation; candidate revisions inherit consumed repair budget. |
| Records / namespace | `4096` | Reject exceeding admission; never silently truncate. |
| Query page | `128` | Stable bounded pagination over the validated snapshot. |
| Record/journal/evidence file | `1 MiB` | Reject oversized files with an explicit diagnostic. |
| Validation command count | `16` | Ordered, deduplicated exact argv; a shell must be explicitly named. |
| Arguments / command | `64` | Command line at most `4096` bytes; explicit timeout and shared cancellation. |
| Closure attempts / set invocation | `128` | Persist the counter; restart cannot reset it. |

Invalid, conflicting, or unsupported configuration blocks execution. Config
resolution reports the effective value and source. Resume keeps admission scope
and budget consumption; publication uses the current policy binding. Dynamic
capacity reduction waits for existing reservations and does not terminate work.

## 9. Consistency Checks and Approval Boundary

The implementation must reject duplicate or foreign identities, inconsistent
association triples, unknown schema versions, ambiguous task authority, changed
execution contracts, stale revisions, stale fencing generations, mismatched PRs,
invalid dependency graphs, unverifiable completion and mixed set-order rules.

This model was approved with the behavior specification on 2026-10-08. That
approval covers the defaults, opt-in concurrency, supported surfaces, admission
policy and prospective adoption boundary in specification section 13. Planning
was authorized separately; runtime implementation and historical migration were
not authorized by model approval.
