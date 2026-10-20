# Ticket-Native Work Execution

Status: accepted

Date: 2026-10-08

Approved: 2026-10-08

Product: chr

Model: [Ticket-Native Work Model](ticket-native-workflow-model.md)

Plan: [Ticket-Native Work Delivery](../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../tracker/ticket-native-workflow.md)

This specification defines the approved chr contract. Commands,
configuration keys, records, and guarantees below are requirements for delivery,
not claims about the current executable. Specification approval and implementation
authorization remain separate decisions.

## 1. Outcome and Scope

chr must execute approved work through durable tickets, from selection through
verified integration. A user can resolve an independent ticket, one slice, one
slice set, the slices of a use case, or a filtered queue of loose tickets. The
same state and rules serve automated runners and manual Codex/Pi work.

For planned work, every task has a ticket. A slice groups ordered task tickets
and remains exactly one pull-request and integration unit. A slice-set execution
can create and integrate several consecutive slice PRs without an intermediate
human approval when its review policy permits that behavior.

This is prospective. Existing plans, trackers, reports, and ticket history remain
readable. Historical conversion, release publication, and promotion to `main`
are outside this execution contract. The integration target is explicit; the
normal development target is `dev`.

## 2. Verified chr Baseline and Required Changes

The baseline inspected for this specification is commit
`d36dee206c3e4ad3ab6a6230365fd0812b5923a9` on `dev`.

| Capability | Current chr evidence | Required contract |
| --- | --- | --- |
| Ticket identity and classification | [Ticket](../../../../internal/core/ticket.go) has `key`, `use_case`, `project_area`, priority, severity, and a textual `slice`. | Preserve these concepts; add stable hierarchy identity and task order. |
| Plan materialization | [Slice detail](../../../../internal/core/plan_slice_detail.go) contains tasks; [plan service](../../../../internal/app/service_plan.go) creates slice tickets. | Materialize one ticket for every task and retain one PR for the owning slice. |
| Work orchestration | [Work service](../../../../internal/app/service_work.go) owns ticket execution and serialized integration cycles. | Extend the same operation boundary to slice, set, and use-case execution. |
| Selection | [Loop query](../../../../internal/core/ticket_loop_query.go) supports combined filters and deterministic ordering. | Preserve rank semantics and add explicit hierarchy exclusion and prerequisites. |
| Review defaults | [Configuration](../../../../internal/core/config.go) and the standard embedded profile default to `human_review`. | New workflow projects default to `trusted_auto_merge`; existing effective policy is preserved on upgrade. |
| Concurrency | [Operation preflight](../../../../internal/core/work_selection_preflight.go) uses exclusive `agent-operation.lock` creation; the work service holds it for the selected operation. | Current same-project work is serial. Parallel independent work requires shared capacity reservations and exclusive work claims. |
| Lock scope | The operation lock resolves under the checkout's active ops namespace. | Worktree and namespace changes must not bypass repository-wide worker capacity or target publication exclusion. |
| Runtime isolation | Project-mounted execution remains the fallback; isolated worktrees are supported. | New workflow execution uses isolated worktrees; parallel workers cannot share a mutable checkout. |
| Durable records | Ticket Markdown is durable; SQLite provides ticket indexes and execution/loop records; `ops/<namespace>/` contains shared artifacts. | Shared ticket and workflow records own recoverable work state. Derived indexes and runtime records cannot override them. |
| Surfaces | CLI and TUI invoke shared work operations; Web exposes board and authoring operations. Telegram notifications and command records exist. | Deliver the surface contract in section 11; notifications alone are not an execution adapter. |

Runner implementations and model backends are interchangeable executors. Their
existence does not establish a worker pool or safe parallel execution.

## 3. Ownership, Identity, and Ordering

The hierarchy is:

```text
project + namespace
  use case (optional association)
    slice set
      slice, ordered within the set
        task ticket, ordered within the slice
        task ticket, ordered within the slice
  independent ticket (optional use-case association)
```

Every ticket belongs to one stable project ID and namespace. A checkout path,
title, folder name, or remote alias is not identity. A slice set may exist without
a use case. A standalone ticket needs neither a set nor a slice.

The set owns its slices and their positive, unique `slice_order`. Each grouped
ticket owns its `slice_set_id`, `slice_id`, and positive `task_order`. Task order
is unique within a slice. Orders may have gaps; reading never renumbers them.
Set and slice identities survive title changes and project relocation.

The existing `slice` string is a display/source label, not proof of membership.
Legacy `use_case` values must resolve explicitly before they can supply stable
ownership. Titles and similar filenames never create associations implicitly.

A set's optional use-case association is authoritative for its grouped tickets.
A duplicated ticket association must agree with it. A standalone ticket can have
its own use-case association, but use-case execution excludes it from slice work.

Priority, severity, ticket creation time, and worker availability never reorder
tasks or slices. Reordering, removal, or reassociation after execution starts
requires an explicit revision-guarded amendment at a quiescent checkpoint.
Integrated history and earlier closure receipts remain immutable.

## 4. State Authority and Manual Work

| Operating mode | Authority | Tracker role | Exit condition |
| --- | --- | --- | --- |
| `tickets` | Ticket records own task state; the owning set record owns slice contracts and integration receipts. | Required order-only companion for revision-bound plans; optional v1 view for existing workflows. Neither imports editable execution state. | Normal target mode. |
| `tracker_compat` | One explicitly selected existing tracker owns the unadopted legacy workflow. | Existing execution map. Ticket projections are read-only and excluded from ticket-native selection. | Explicit adoption of selected active work, with identities, revisions, and delivery evidence reconciled. |

New executable work uses `tickets`. Transitional mode is limited to workflows
that have not adopted tickets. There is one authority for each selected workflow;
the two modes cannot execute the same work simultaneously.

In ticket mode, every task and later repair or iteration is represented by a
ticket before execution. Repair of the same obligation creates another attempt
on that ticket; an added obligation creates a new linked ticket. Neither a
tracker checkbox nor an agent statement completes a ticket.

Revision-bound plans retain an internal `chr-tracker/v2` Markdown companion for
every generated revision. Before approved materialization, it declares stable
planned identities and order without allocated tickets or execution authority.
Approved native materialization binds canonical tickets and publishes the set,
map, provenance and receipt through the existing recoverable workflow journal.
A retry reuses the same allocation. A later draft does not change an active set;
applying approved changes uses its guarded amendment operation and retains
completed tickets, receipts and earlier revision maps.

V2 hashes and renders project/namespace, plan identity/revision/body hash,
materialization/set identity, ordered slice/task IDs and titles, and canonical
ticket references. Status, phase, claims, attempts, review, delivery/closure state
and whole-record revisions/hashes are excluded. Legacy progress cells stay blank.
Ticket progress leaves map bytes and freshness unchanged. The current map is
included in worker input; live tickets and receipts determine remaining work.
A missing or stale required map blocks preparing execution context until repair.
It never rolls back a successfully committed ticket transition.

Tracker export flows from native records to the projection only. Conflicting
edits are reported and never imported. Replacement requires the exact observed
output hash; repair launches no work. Existing optional v1 views retain their
progress projection, and their failure does not prevent direct native execution.
Reading legacy artifacts or `tracker_compat` state never upgrades their schema
or adopts them implicitly.

Manual runners use the same claim, checkpoint, validation, and reconciliation
operations. They can report a commit and evidence for host verification. They
cannot bypass a live owner, merge barrier, or required review by editing a status.
Existing artifacts remain untouched unless their particular adoption is approved.

## 5. Selection and Continuation

Every invocation names a project, namespace, explicit extent, and identity where
required. Missing identity is an input error; it never expands to a project queue.

| Extent | Admitted work | Completion meaning |
| --- | --- | --- |
| `ticket` | One exact ticket. Grouped tickets retain their task prerequisites. | The selected task is validated; an independent ticket additionally requires integration. |
| `slice` | Remaining tasks of one slice, in task order. | That slice is integrated and its ticket finalization is durable. |
| `slice_set` | Remaining slices of one set, in slice order. | Every slice is integrated and the set closing gate is confirmed. |
| `use_case` | Sets belonging to one use case. | Every selected set closes; unsliced tickets are excluded and counted. |
| `loose` | Eligible independent tickets matching the requested filters. | Selected tickets integrate or the invocation stops at a blocker. |

For use-case execution, sets follow explicit unique use-case-local order when
all selected sets declare it; otherwise all selected sets follow immutable
set `created_at`, then stable set ID. Mixed explicit and absent orders reject
admission until the ordering contract is resolved.
Slices inside each set always follow `slice_order`. Global ticket creation order
has no role. The invocation records which set-order rule it resolved.

A grouped single-ticket request executes only that task. It does not silently
execute siblings or publish a partial slice. Its result distinguishes
`task_validated` from `slice_integrated`. A later slice/set invocation reuses the
validated task. Integrating a slice requires the slice extent or a broader extent.

Initial admission freezes ordered member identities, revisions, and contracts.
An accepted execution transition advances the invocation's expected revision;
its own state updates do not invalidate that frozen membership snapshot.
Changes to execution-owned fields or descriptions require reconciliation;
unrelated display edits may survive only through a guarded update that proves
the frozen execution contract did not change. By default, later arrivals are
excluded from that invocation and reported. A new explicit invocation can admit
them. An explicit `include_later_tickets` policy can instead admit matching ready
work strictly after the confirmed frontier while the invocation is unfinished.
It never inserts work before that frontier, changes filters, reopens a completed
set, or bypasses prerequisites. Resume retains that original admission policy.
This policy does not authorize changing an approved slice contract: adding tasks
to an existing slice still needs its quiescent, revision-guarded amendment.

Equivalent unfinished requests attach to the same invocation. Resume by execution
ID restores the original extent, filters, target, membership, and cursor. It
cannot broaden the scope. Overlapping but different requests observe ownership
conflicts instead of launching duplicate workers. Completed and cancelled
invocations remain terminal on retry.

A bounded call can return a resumable checkpoint before exhausting the selection.
Exhaustion completes that invocation; it does not subscribe to future tickets.

## 6. Task, Slice, and Set Lifecycle

Before a claim, chr validates expected outcome, scope, acceptance criteria,
validation commands, hierarchy, prerequisites, observed revisions, and target.
Explicitly selected unprepared work returns a blocker. Queue discovery excludes
unprepared work and reports the reason and count.

For one slice, chr must:

1. Reconstruct tickets, attempts, branch/worktree, PR, and integration evidence.
2. Reconcile retained results and skip already verified work.
3. Confirm earlier slices integrated into the selected target and all explicit
   dependencies satisfied.
4. Claim the slice and the first incomplete task atomically.
5. Execute remaining tasks in task order on one cumulative slice branch. Validate
   each task and record its attempt, commits, and result before the next task.
6. Validate the completed slice against its aggregate acceptance contract and the
   exact integration candidate, beyond its individual task checks.
7. Create or reuse the slice's single canonical PR and apply review policy.
8. Integrate that candidate, verify the provider receipt and actual target
   history, and durably finalize every task against the shared slice receipt.
9. Refresh authoritative target and work state before selecting the next slice.

Task validation permits the next task in that slice; it does not mark the task
`done`. Grouped tasks remain `in_progress` with `task_phase: validated` until the
aggregate candidate is ready. They enter `ready_for_review` or `ready_for_merge`
with their common slice branch and PR. All become `done` only after integration
proof is durable. Existing status guards must recognize slice-owned PR evidence.

Independent tickets retain one ticket as their integration unit. Slice failures
retain validated tasks and commits. Repair invalidates affected task evidence
and aggregate validation; unrelated task evidence is reusable only when its
recorded inputs still match. A new candidate revision uses the same branch/PR,
subject to provider identity guards.

A closed, unmerged PR is a blocker. Reopening that PR or correcting an erroneous
recorded reference requires explicit reconciliation and retains the original
observation. Ordinary retries cannot create replacement or per-task PRs.

After all slices integrate, the set closing gate validates the exact observed
target and a frozen member/plan snapshot in an isolated checkout. A failed gate
does not rerun integrated tasks. An explicitly waived closing gate records a
reason and still requires integration proof for every member. An empty set is
complete only after its explicit closure decision. Reopening a completed set
requires an observed revision and approved changed plan; prior invocations and
receipts remain terminal.

## 7. Dependencies and Loose-Ticket Rules

Earlier tasks in a slice are implicit task prerequisites. Earlier slices in a
set are implicit integration prerequisites. Explicit dependencies are typed
ticket or slice references within the same project and namespace. Missing
references, self-dependencies, cycles, and forward task dependencies conflict
with admission and are rejected. This requires dependency validation, not a
general-purpose parallel DAG scheduler.

Within a slice, a task dependency needs validated evidence from the current
cumulative branch. A dependency outside that slice needs verified integration
on the selected target. A status label alone satisfies neither condition.
Dependencies outside an invocation block it; chr does not add them to its scope.
Resume observes them again. A blocker stops that invocation before later work;
it does not stop unrelated independent lanes.

Loose selection excludes all slice-owned tickets, including when a use-case
filter matches them. The default eligible status is `ready_for_dev`. Already
admitted work resumes from its checkpoint; users cannot make a completed ticket
eligible by adding `done` to a status filter. Legacy aliases remain readable,
but execution requires explicit identity and evidence normalization.

Keep `project_area` as chr's canonical free-form classification. Trim values;
compare them exactly and case-sensitively, as the current query does. Filters
combine with AND across fields and OR within an area's or rank's value list;
every requested tag must match. Exact priority/severity lists and optional
minimum thresholds use the same independent rank tables:

```text
critical > high > normal > low > unclassified
```

Unclassified includes an absent legacy value. It sorts last and never satisfies
a minimum threshold. Invalid nonempty ranks reject input. Default loose order is
`created_at`, then natural ticket key. Explicit priority or severity order uses
highest rank first, then oldest known `created_at`, then natural key. Missing
creation times follow known times. Combined priority/severity order compares both
ranks before those tie-breakers. Selection occurs again after each integration.

Required examples include frontend tickets filtered and ordered by severity,
persistence tickets filtered and ordered by priority, and all tickets of one
area ordered by either rank. Priority does not preempt a live task or reorder a
slice. A loose invocation is serial; separate independent invocations can overlap
only under section 9.

## 8. Review, Integration, and Effective Defaults

| Review mode | Required behavior |
| --- | --- |
| `human_review` | Retain the exact candidate and wait for its explicit approval. |
| `agentic_review` | Run a separate review role; passing evidence allows automatic integration. Blocking findings enter bounded repair; insufficient evidence or human-required findings stop. |
| `trusted_auto_merge` | Integrate after required host validation and repository gates without a separate review stop. |

Automatic review or merge policy does not authorize unapproved implementation,
override target protections, waive validation, or authorize a release. Workers
produce candidate changes; the host operation owns PR creation and publication.

New projects initialized for workflow version 1 use `tickets`, isolated worktrees,
and `trusted_auto_merge`. Standard and minimal profiles, runtime/bootstrap setup,
and missing-field fallbacks for that version must agree. New configuration fields are
listed in the model; existing configuration must not be described as supporting
them before delivery.

Configuration precedence is explicit invocation restriction, explicit project
value, then versioned project default. An invocation can restrict an automatic
mode to human review; it cannot silently relax a project requirement. Resume
retains the original authorization scope. Publication re-resolves current policy
and repository capabilities; a policy change invalidates an old candidate grant.

Projects without the version marker retain `human_review` when review mode is
absent and preserve explicit existing values. Re-running initialization does not
add the marker or reset a policy. An explicit upgrade records the previous
effective review policy and state-authority mode before enabling version 1.
Changing an existing project to automatic integration requires an explicit
policy decision; it is never a side effect of filling a template field.

Integration supports the existing method names `merge`, `rebase`, and `squash`.
The version-1 default is `rebase`, consistent with current merge
preflight. It requires a verified linear fast-forward result. A provider that
cannot guarantee the configured result blocks; there is no fallback method.
Merge proof binds source/base and resulting tree; rebase proof binds base and
ordered commit trees; squash proof binds base and exactly one combined commit.
Original source ancestry alone is insufficient for rewritten histories.

Immediately before publication, chr rechecks policy, candidate revision, canonical
PR repository/number/head/base, target SHA, validation evidence, and any approval
binding. A moved target requires preparing and validating the new combined result.
It cannot reuse approval for different content. Approval is for one exact
candidate and does not approve later slices.

Publication must atomically guard the source and validated target base, through
the provider's capability or an equivalent guarded fast-forward operation on
the same canonical PR. A local target lock cannot exclude external writers.
If the selected provider/method cannot protect that boundary, chr blocks it;
checking the target only before an unguarded API call is insufficient.

Record submission intent before provider IO. A timeout or lost response retains
`awaiting_observation` and excludes another publisher on that target until its
outcome is reconciled. A successful command is not completion. Fetch actual
target state and verify the canonical receipt and method-specific history/tree
proof. External integration can satisfy the same proof without another worker
or submission.

## 9. Multiple Runners and Publication Exclusion

The target contract allows independent lanes to execute concurrently with an
explicit repository capacity. Its default is one worker. A lane is one slice
or one independent ticket. A slice/set/use-case invocation advances one lane
at a time; there is no parallel task execution within a slice or speculative
execution of later slices before integration.

Two independent use cases or loose-ticket invocations may overlap when capacity
is greater than one, their dependencies permit it, and their claims do not
overlap. Different projects can run independently. Two namespaces or linked
worktrees of the same repository share its configured capacity and target lock.
Different configuration values for that repository are a conflict, not separate
pools. Separate clones or hosts must use one registered coordinator for the
same target; unsupported distributed ownership returns a blocker.

Every worker has an isolated cumulative lane branch/worktree, immutable input
snapshot, attempt ID, expiring ownership lease, fencing generation, and a
process-held capacity reservation. A lease expiring while a process remains
alive does not release its worker slot. A replaced or expired owner cannot
apply results, release a replacement claim, or publish. Manual takeover needs
the same guarded handoff after the previous owner is quiescent.

Claim creation and record updates are serialized across processes. Claims cover
the task, its slice, and the ordered frontier of its set/use case as applicable.
An explicit ticket request cannot bypass another operation's slice claim.
Capacity limits apply to every work-start entrypoint, including legacy runners,
manual harness adapters, and daemon activity.

Only one integration publisher may own a canonical repository/remote-target
branch at a time, across project aliases, namespaces, CLI/TUI/Web processes, and
worktrees. Independent targets may integrate independently. Waiting for human
review releases worker capacity while retaining candidate ownership. It does
not hold the publication lock until a submission is authorized. An uncertain
submitted result keeps target exclusion even after worker/lease expiry.

If lane A integrates while lane B develops, B may finish local work. Before B
publishes, chr refreshes the target, prepares B against that target, validates
the combined result, and repeats review when its bound inputs changed. A conflict
blocks B; A remains integrated. Workers never force-push the shared target or
edit another lane's worktree. Resource scarcity produces `waiting_capacity`,
distinct from dependency, review, validation, and merge blockers.

The current project/namespace-wide lock must remain effective until replacement
claims, capacity, target exclusion, and crash recovery pass their acceptance
tests. Deleting or narrowing that lock alone does not deliver parallelism.

## 10. Recovery, Bounds, and Observability

Every transition has a stable operation ID and observed revision. Selection and
claim snapshots use the same process-safe lock as shared domain updates.
Updates use compare-and-swap revisions, atomic replacement, and a durable
transition journal for multi-file changes. SQLite and tracker
projections can be rebuilt from shared records and receipts without choosing
new work or inventing completion events.

Restart reconciles partial writes, live/dead ownership, retained candidate and
PR identities, submitted operations, actual target history, and ticket finalization
before admitting another task. Incompatible manual edits return
`reconciliation_required`; chr never overwrites them with an old worker result.
Finalizing several task tickets after a slice merge is replayable and cannot
create another PR, merge, or receipt.

Pause/cancel record durable stop intent and wait for a safe checkpoint. They do
not free a live reservation early, discard evidence, or undo integration. A
cancelled invocation is terminal; completing observation of a previously
submitted candidate can repair delivery facts without resuming that invocation.

The model defines finite worker, queue, file, attempt, command, and call budgets.
Budget exhaustion returns an explicit blocker or resumable checkpoint. Resume
does not reset consumed automatic-retry budgets. Review repair is limited to
three attempts per integration unit in an invocation, including candidate
revisions created by those repairs; extending the budget is explicit.

Preview and status must show extent, exclusions, hierarchy order, authoritative
mode, resolved review/merge policy and its source, target, selected task/slice,
ownership/capacity waits, PR, evidence, confirmed integrations, remaining work,
blocker, and next-action owner. Public events include stable invocation,
slice/ticket, attempt, and receipt IDs, but no private ownership tokens.

## 11. Shared Operation and Adapter Contract

Selection, scheduling, lifecycle, repair, reconciliation, review, integration,
and closure belong in `internal/app` behind the central operation service, with
domain validation and persistence boundaries in `internal/core`. Adapters parse
input and render results. They do not keep independent queues or schedulers.

The target CLI extends the existing `chr work` group:

```text
chr work ticket <key>
chr work slice <slice-id>
chr work slice-set <set-id>
chr work use-case <use-case-id>
chr work loose --project-area frontend --severity high --order severity
chr work loose --project-area persistence --priority high --order priority
chr work loose --project-area frontend --order priority
chr work loose --project-area frontend --order severity
chr work continue <execution-id>
chr work status <execution-id>
```

Mutating operations support a read-only preview. Existing `work run-all` remains
a compatibility entrypoint; in ticket mode it selects loose work and explicitly
reports excluded slice-owned tickets. Transitional mode retains its existing
scope and reports that mode. Neither command silently includes a whole use case.

| Surface | Required delivery scope |
| --- | --- |
| CLI | Explicit extents, preview, status, continuation, pause/cancel, exact-candidate approval and manual reconciliation. |
| TUI | Invoke the same operations from selected tickets, slices, sets, and use cases; show progress, waits, evidence and confirmation bindings. |
| Web | Add the same bounded operation controls to existing ticket/case views, with revision guards, visible failures, status and continuation. |
| Daemon | Resume only retained authorized invocations; use the same ownership and capacity rules. Startup is not authorization to execute the backlog. |
| Telegram | Existing notifications remain informational. A complete workflow-control adapter is a separate scope decision and is not required for this delivery. |
| Manual Codex/Pi | Shared claim/checkpoint and evidence-reconciliation protocol; no independent tracker scheduler or unguarded completion writer. |

Missing surface operations are explicit delivery gaps. Displaying a status or
sending a notification never certifies execution parity.

## 12. Acceptance Criteria

| ID | Required observable result |
| --- | --- |
| A01 | Materializing two slices with three tasks each creates six task tickets, stable hierarchy/order, and at most one canonical PR per executed slice. Retrying materialization creates no duplicates. |
| A02 | Ticket or set title changes and checkout relocation preserve identity. Duplicate task/slice orders, foreign ownership, and half-populated associations reject without partial writes. |
| A03 | A set run executes task order and integrates slice 1 before starting slice 2. Automatic policy completes consecutive PRs without per-PR human interaction. |
| A04 | Interleaved creation of two use cases' tickets never changes each use case's set/slice order. Unsliced associated tickets are excluded from use-case execution. |
| A05 | Each of the four loose-ticket examples returns the specified AND-filtered order. Rank ties and missing ranks/timestamps resolve deterministically; slice-owned tickets cannot enter that queue. |
| A06 | An exact grouped-ticket request validates only that task, creates no partial-slice PR, and reports that integration remains pending. |
| A07 | Unmet, missing, cyclic, or outside-selection dependencies prevent execution and identify the blocker. Resume after verified prerequisite integration proceeds without expanding scope. |
| A08 | A task validated locally is not `done`. One verified slice receipt finalizes all its task tickets. Crash after any partial finalization replays once without another merge. |
| A09 | Duplicate equivalent requests and competing CLI/TUI/Web processes create one owner. A different overlapping selector cannot start the same ticket or slice. |
| A10 | Capacity two permits independent lanes to overlap; capacity one serializes them. Worktrees/namespaces share the capacity. Tasks and dependent slices remain serial at either capacity. |
| A11 | Two ready lanes targeting one branch submit serially. Target movement invalidates stale combined validation/approval; a conflicting lane blocks without changing another lane's result. |
| A12 | Expired/replaced owners cannot commit domain results or release newer claims. A still-running process retains its capacity reservation after durable lease expiry. |
| A13 | Restart after PR creation reuses its identity; restart after merge verifies actual delivery and skips implementation. Closed/unmerged or mismatched PRs require reconciliation. |
| A14 | Lost publication responses retain target exclusion. Reconciliation confirms actual delivery or establishes an unpublished result before a bounded retry; no blind duplicate submission occurs. |
| A15 | Human, agentic, and trusted policies have distinct observable stops. Changed candidate/policy bindings cannot reuse approval. Provider protection and unsupported methods remain blockers. |
| A16 | Standard/minimal initialization and all bootstrap/fallback paths agree on version-1 defaults. Existing absent and explicit policies survive reinitialization/upgrade. Invalid config never falls back to automation. |
| A17 | Manual execution and automated execution obey the same ownership, revisions, task lifecycle, candidate validation, and delivery proof. A manual status edit alone cannot advance the cursor. |
| A18 | Tracker edits cannot change ticket-native selection. Projection failure preserves domain progress and reports staleness; repair rebuilds the projection without running work. |
| A19 | Default frozen membership excludes later work. Explicit later admission accepts only matching ready work after the confirmed frontier; neither mode hides actual members from closure. Reordered/reassociated work blocks reconciliation. Reopening preserves earlier closure and terminal invocations. |
| A20 | Set closure binds all actual members, plan, validation contract and target SHA/tree. Gate failure retries closure only; target/input movement cannot establish closure. |
| A21 | Cancellation at execution, review and uncertain publication checkpoints preserves evidence and ownership rules. A retry of a terminal invocation does not restart it. |
| A22 | Killing a cooperating writer between journaled file replacements restores one consistent transition or reports conflicting edits, without silently mixing revisions. |
| A23 | CLI, TUI, Web and daemon exercise the same operation contract; unsupported Telegram controls are reported accurately. |
| A24 | Limits and retry budgets reject oversized input and return bounded checkpoints. Repeated continuation does not reset retry limits or duplicate events. |

Validation must include real filesystem records, subprocess claim races, linked
Git worktrees, a bare remote, target movement, crash injection and lost-response
recovery. Provider doubles prove adapter contracts, not live provider behavior.
Each supported provider/method requires a disposable live PR integration check
before claiming that path delivered. Future implementation runs `make verify`
and race-sensitive tests; this specification-only change requires document checks.

## 13. Accepted Decisions

The requested workflow, per-task tickets, one PR per slice, prospective adoption,
and independent chr contract are approved requirements. The maintainer accepted
the following design decisions with this specification on 2026-10-08:

| Decision | Accepted contract |
| --- | --- |
| Automatic default | `trusted_auto_merge` for newly initialized version-1 projects; preserve existing effective policy on upgrade. |
| Parallel capacity | Default one worker; support opt-in capacity up to eight independent lanes only after ownership/recovery acceptance passes. |
| Surface scope | CLI, TUI, Web, daemon and manual protocol; notifications only for Telegram. |
| State transition | Ticket-native new work; selected, quiescent adoption for legacy active work; no historical batch conversion. |
| Continuation | Frozen explicit selection by default; explicit later admission can extend only an unfinished selection after its confirmed frontier. |
| Integration strategy | Existing names and `rebase` default, with verified fast-forward semantics and capability checks. |

## Open Questions

No behavior or model decisions remain open for this approved scope. The linked
plan defines implementation sequencing and provider rollout. Plan approval,
versioning on `dev`, and explicit activation precede implementation. A separate
Telegram workflow-control adapter remains outside this delivery.
