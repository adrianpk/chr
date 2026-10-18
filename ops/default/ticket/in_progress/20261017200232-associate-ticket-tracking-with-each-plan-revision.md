---
id: TKT-20261017200232
title: "Require an internal tracker for deterministic ticket order"
status: in_progress
kind: task
severity: medium
priority: normal
scope: domain
tags: tracking, tickets, plans
source: chat
reported_at: 2026-10-17T20:02:32Z
commits: []
branch: feat/slice-9-plan-order-tracker
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Require an internal tracker for deterministic ticket order

## Context and Verified Baseline

Part of [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
The maintainer requires CHR to retain an internal tracking document for every
plan generation or regeneration. Its purpose is to give the agent the
deterministic order for processing the plan's tickets. Task progress is recorded
in the tickets, rather than marked as processed in the tracking document.
Defining the ticket sequence is its only responsibility. There is no need to
keep processing state synchronized in this artifact. Keep the existing tracker
name and compatible document structure during the transition; renaming it is
separate future work.
The tracker is not a user-facing Web document or editor. It remains a Markdown
artifact in the project that someone working manually can inspect.

The current
[ticket-native contract](../../spec/stable/ticket-native-workflow.md)
currently keeps a tracker as an optional generated view. Ticket records own task
state; the owning set and its slices, together with ticket membership and order,
organize execution. The new requirement makes the internal document necessary
for this plan workflow while retaining those owners. Tracker text must not
introduce a conflicting editable order or a second task-state authority.

Verified source paths:

1. `internal/core/workflow_tracker.go` exports, inspects and repairs a tracker
   derived from authoritative records. `workflow_tracker_image.go` renders
   slice and ticket tables with state and delivery references.
2. `internal/app/service_plan_workflow.go` materializes approved workflow input
   into the set and task tickets. It does not automatically export that tracker.
3. The older `MaterializePlanTracking` path in `service_plan.go` writes a tracking
   document with ticket references, but returns the existing scoped tracker
   when one is present. This is not a guarantee of a companion bound to each
   regenerated plan revision.

Capture the maintainer's target behavior. Do not implement it or change the
accepted workflow contract as part of ticket capture. The existing optional
tracker contract must be aligned during the separately authorized delivery.

## Required Behavior

1. Automatically associate a generated tracking document with each generated
   or regenerated plan revision. Bind it to the actual plan identity and
   revision or content hash, rather than inferring the relationship from a
   reused filename or use-case title.
2. Make the tracker a required internal implementation map pointing to
   canonical tickets. List slices by slice_order and their tickets by
   task_order, with ticket titles and resolvable references. The sequence must
   agree with the native workflow records and remain deterministic when ticket
   files are created or discovered in a different order.
3. Supply that ordered map to the agent's implementation context so it can
   identify the ticket sequence. Read current progress and record completion
   through the tickets and shared workflow operations. Do not maintain a
   separate set of processed checkboxes or editable plan entries in the tracker.
   Remove the need to synchronize processing, phase or completion state into
   this document at task checkpoints. The agent consults current ticket state
   to determine what remains within the declared sequence.
4. Regeneration must reconcile the associated work rather than blindly
   allocating duplicate tickets or resetting completed tasks. Preserve the
   association and delivery history of earlier plan revisions.
5. Retain provenance and stale/conflict handling for the ordered map itself:
   its plan binding, ticket membership, references and ordering. Processing,
   phase and delivery-state changes alone must not require a tracker rewrite,
   export or repair, or invalidate its unchanged order. A missing or structurally
   stale map must not be presented as current. Reuse the existing tracker owner
   while removing its processing-state synchronization responsibility.
6. Keep the tracker out of the Web UI. Do not show it as a Plans card, expose a
   tracker detail view or editor, or ask users to create it with a Create
   Tracking form. Its creation belongs to the plan workflow, not a separate
   manual Web step. Ordinary plan cards and their read-only details remain.
7. Keep the generated Markdown artifact in the project for manual inspection.
   It is not a manually maintained document. Direct file edits must not change
   ticket state or override canonical execution ordering.
8. Retain the tracker name and compatible existing document structure for this
   transition. Do not rename its artifact family, paths, commands or schemas
   as part of moving processing state to tickets. If legacy progress columns
   remain in the structure, they are unused: do not maintain them or read them
   as task state. A more accurate name can be considered separately later.

## Refinement Before Execution

The need for the tracker and its exclusion from the Web UI are settled by this
capture. Align companion creation with plan approval and ticket materialization
without treating a regenerated draft as authorization to execute new work.
Define how a plan revision refreshes its ticket map and interacts with an
already active or delivered set. Use the existing guarded amendment, ordering
and history contracts; do not invent an independent Markdown scheduler.
The tracker must not need refreshing merely because a ticket advances through
its lifecycle. Only changes to the declared map or its plan binding require
that map to be updated.

Keep this ticket separate from
[plan generation placement and source prerequisites](../solved/20261017195741-contextualize-plan-generation-and-check-source-readiness.md).

## Acceptance

1. Initial generation and regeneration can each be traced to their own plan
   revision and associated tracking document.
2. Once tickets are materialized, tracker references resolve to the canonical
   tickets and its explicit sequence agrees with slice_order and task_order.
   Different file discovery or ticket creation order does not change that
   sequence. The agent receives the current ordered ticket map.
3. Repeating creation or export does not duplicate tickets or erase delivery
   history. Completing work updates ticket state without marking a separate
   tracker task as processed. Advance and complete tickets while leaving the
   plan binding, membership and order unchanged: the tracker remains unchanged
   and valid, without a required export or repair. Subsequent work observes
   the live ticket state rather than legacy tracker progress cells.
4. Tracker edits cannot change ticket state, membership or execution order.
   Preserve stale/conflict checks for changes to the ordered map, while
   distinguishing those from processing-state changes. Verify the agreed
   approval boundary when implementation is authorized.
5. The project contains the associated Markdown tracker for manual inspection.
   The Web UI shows no tracker card, tracker detail/editor or Create Tracking
   form. Verify that plan browsing still opens the ordinary read-only plan.
6. Existing tracker naming and compatible structure are retained. No rename
   migration or maintenance of legacy progress fields is required for this
   transition.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).

## Slice 9 Execution

Activated T9.1 after verified PR #405 integration. The [Slice 9 report](../../report/slices/web-ui-workflow-consolidation/slice-9-plan-order-tracker.md) owns implementation and validation evidence.

T9.1 completed: both generation paths require structured unbound companions;
recoverable publication retains revision history and protects edited current
output. Root-bound receipts acknowledge explicit operation retries without
another model call or revision. Focused plan race checks and affected-package
strict lint passed. T9.2 is active; approved materialization and execution-context
binding remain for T9.3.
