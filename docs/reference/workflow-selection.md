# Workflow Selection

`WorkflowStore.Preview` and the application operation `PreviewWorkflow` resolve
read-only ticket-native work in one project and ops namespace. The filesystem
owns selection. Trackers and SQLite do not supply execution authority.

`WorkflowPreviewRequest.Selector` names `ticket`, `slice`, `slice_set`,
`use_case` or `loose`. Every extent except `loose` needs an exact identity.
Missing identities reject rather than selecting the backlog. Unadopted tickets
are counted as `unprepared`. Use-case work counts and excludes unsliced tickets;
loose work excludes every slice-owned ticket and integrated work.

Pages contain at most 128 task identities with their revisions, source images,
execution contracts, membership, status, phase, readiness and first blocker.
`Total` and mutually exclusive `Excluded` counts describe the whole extent.
`NextOffset` is -1 at exhaustion. Continue with the same normalized request and
`SnapshotHash`; changed records, proof bytes or request scope reject continuation.
An empty page is a selection result, not proof that a set closing gate passed.

Sets in a use case follow homogeneous explicit unique `use_case_order`, or all
follow immutable creation time then set ID. Mixed ordering rejects. Slices and
tasks follow positive local orders, including gaps. Rank, creation of unrelated
tickets and display labels do not affect hierarchy order. Slice task fingerprints
must match authoritative membership.

Prerequisites include earlier tasks, earlier slice integration, and explicit
typed references. Missing, cyclic and forward dependencies reject. Unmet work
reports `outside_selection` or `awaiting_selected_prerequisite` and its owner;
it never adds another task to the extent. Same-slice validation requires a
`chr-task-validation/v1` evidence file with the ticket key, execution contract
hash, full branch ref, immutable commit/tree and successful host-validation fact.
The commit must belong to the requested current cumulative lane. Cross-slice or
independent prerequisites need the recorded confirmed integration receipt. Its
canonical repository, result tree and ancestry must prove reachability on
`Target.Commit`. Provider reconciliation owns creation of those receipts.

Loose requests reuse native `TicketLoopFilters`, status and order names, with
optional independent `MinPriority` and `MinSeverity`. The default eligible status
is `ready_for_dev`. Filters combine AND across fields and OR within value lists;
every tag must match. Areas are trimmed and compared case-sensitively. Both ranks
use `critical > high > normal > low > unclassified`. Empty ranks are unclassified;
that value sorts last and never meets a minimum. Unknown nonempty ranks reject.
Combined order compares both ranks before creation time and natural ticket key.
Unknown creation times follow known times. Integrated tickets cannot become
eligible through a status filter. Loose options on hierarchy extents reject.

A pending transition journal blocks preview until an explicit domain operation
reconciles it; preview does not replay pending writes.

Observation uses repository workflow exclusion, a ten-second shared deadline,
4096 record/slice/evidence-reference bounds, 1 MiB files and at most 128 distinct
prerequisite observations per page and at most 128 combined references per task. A proof-observation limit returns a blocker.
Cancellation stops the preview without claiming work. Execution, invocation
retention, publication and adapter controls remain separate operations.

## Bound admission

`PreviewSelectedWork` returns the normalized request, exact target, snapshot,
optional selected-ticket guard, effective execution limits and publication
policy. `StartBoundWorkExtent` accepts an optional `WorkflowBoundAdmission` with
a stable `RequestKey` and both displayed policies. In this mode, checkpoint
steps must be zero: admission and continuation are separate calls.

The native invocation retains `admission_guard` with the displayed snapshot,
selected revision and policies. The guard participates in the request hash.
New admission rechecks it under workflow exclusion, together with the caller's
last authority/target check. A stale preview rejects; it does not expand or
refresh membership. A repeated key returns the original invocation before
rechecking mutable source images, without launching another worker. Changing
the scope or guard under that key remains a conflict. Existing callers without
the optional binding retain their bounded start-and-continue behavior.
