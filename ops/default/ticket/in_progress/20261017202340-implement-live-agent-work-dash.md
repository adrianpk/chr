---
id: TKT-20261017202340
title: "Implement Dash as a live view of agent work"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: dash, observability, workflow
source: chat
reported_at: 2026-10-17T20:23:40Z
commits: []
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Implement Dash as a Live View of Agent Work

## Outcome and Authorization

Part of [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
Dash is the operational view of what CHR's agent harness is doing as it
processes tickets and their owning workflow. Show the ready queue, current
execution and phase, recent activity and eventual result using the same domain
data that informs the TUI. The maintainer requested a ticket definition based
on the existing code and gallery, not application implementation in this session.

The internal tracker only defines ticket order under its
[separate requirement](20261017200232-associate-ticket-tracking-with-each-plan-revision.md).
Dash reads actual progress from tickets, invocations, attempts, events and
receipts. It must not recreate processing-state maintenance in that Markdown
artifact or depend on the development controller used to build CHR.

## Verified Baseline and References

A read-only request to the running development service confirmed:

```text
GET /dash
HTTP 500
load dash
```

The route has a partial implementation. Its underlying load failure has not
been diagnosed in this capture.

1. `internal/tui/dash.go` renders Queue, Current and Recent panels, with
   readiness, queue selection/order, current loop and execution, latest phase
   and message, processed/blocked/failed counts and recent results.
2. `internal/tui/workflow_actions.go` adds native workflow status to the TUI.
3. `internal/app/app.go` loads observability from execution/loop repositories
   and agent event streams. In ticket authority mode it also loads the native
   workflow page.
4. `internal/app/web_views.go` currently starts Dash loading with the Board
   operation and builds a queue from ready_for_dev tickets. It obtains
   readiness and observability, but does not expose the loaded native workflow
   page in Dash. Queue readiness must be checked against the active workflow's
   selection rules rather than inferred solely from a ticket status string.
5. `internal/webui/assets/templates/shell.html` renders Readiness, Queue,
   Current and Recent panels. Current depends on a latest event phase. The
   projection also includes LatestLoop and LatestExecution, but these results
   are not rendered by the current Dash template.
6. `internal/app/service_workflow_selection.go` and
   `service_workflow_status.go` provide shared native selection and inspection
   paths to reuse for the Web view.

Gallery examples establish the intended operational composition:

1. [Ready queue](../../../../assets/img/gallery/chr-agent-work-11-dash-ready-queue-before-run.png):
   queue on the left, current work and recent results on the right.
2. [Running action](../../../../assets/img/gallery/chr-agent-work-14-dash-work-running-start.png):
   an action is active. Do not reproduce the blocking presentation that hides
   the monitoring panels while work is running.
3. [Completed run](../../../../assets/img/gallery/chr-agent-work-16-dash-loop-succeeded.png):
   final status, processed/blocked/failed counts and last execution remain
   visible after the queue drains.

These are historical screenshots illustrating states, not proof of current
runtime health or a requirement to copy terminal formatting and generated IDs.

## Required Information

### Queue and Readiness

1. Show the selected project's eligible pending work, with readable ticket
   keys, titles and relevant scope. Preserve the shared selection rules,
   prerequisites and exclusion between loose tickets and slice-owned work.
2. Reflect deterministic slice/task ordering. A UI priority filter or display
   sort must not redefine the declared order of a slice set or imply that an
   ineligible task is ready to execute.
3. Show readiness as a compact Ready or Blocked summary with understandable
   reasons. Branch or workspace details belong here when they explain a
   blocker; repeated paths and protocol dumps are not the main content.
4. Distinguish no pending work, no eligible work because of a prerequisite,
   and an unavailable data source. Do not present a read failure as an empty
   queue. Make additional entries reachable when a shared read is bounded.

### Current Work

1. Show the active ticket by title and key, its use case or set when relevant,
   current invocation/execution state, phase and latest meaningful event.
   Surface tool or validation activity from structured events where available.
2. Read native workflow progress in ticket authority mode. A legacy loop row
   or the last historical event must not override the active native invocation.
   The page must work without launching or keeping a TUI process open.
3. Show real scope counters when available: eligible, processed, blocked and
   failed, with labels matching their actual definitions. Distinguish a
   successful harness attempt from reviewed/integrated ticket completion.
4. Preserve useful current status during a long-running command or validation
   with no new event. Show when the state or latest activity was observed;
   silence alone does not establish failure, completion or a blocked run.
5. When multiple invocations or runners are active, keep their ticket, phase,
   events and results associated correctly. Show the active contexts or a
   clear selector instead of combining unrelated latest values into one run.

### Recent Activity and Results

1. Show a bounded, readable timeline with timestamps, phase/action, ticket
   reference and event message, with newest activity easy to find. Repeated
   refreshes must not duplicate entries.
2. Retain the latest completed or interrupted result after there is no active
   work. Show actual success, failure, stop or blocked status, its reason and
   available counters. Link to the existing run or ticket detail for deeper
   inspection rather than filling the main view with internal IDs and paths.
3. Do not show No current work signal merely because an execution has no latest
   event phase. Use the current execution/workflow records as well as events.
   Distinguish no activity yet from a known completed run.

## Live Web Behavior and Layout

1. Use Dash as the heading, without a project-name prefix. Translate the TUI's
   three-panel organization into the existing Web theme: a queue column on the
   left, current work at the upper right, and activity/results below it.
   Keep readiness compact near the relevant current-work context. Stack the
   sections coherently on narrow screens.
2. Make the current phase and ticket the primary information. Use short status
   labels and readable events rather than a terminal dump, a screen full of
   generated identifiers or unrelated artifact statistics.
3. Update without manual reloads or full-page navigation. Reuse the existing
   bounded fragment-refresh mechanism and changed-revision handling where
   suitable. Preserve the user's scroll/selection and avoid replacing content
   unnecessarily when the observed state has not changed.
4. Pause or limit polling for a hidden page as the shell already does. When
   connection or refresh fails, retain the last observation with an explicit
   stale/unavailable indication. On recovery or page reopen, load persisted
   state rather than requiring a new run to reconstruct progress.
5. Opening or refreshing Dash is observational. It must not claim work, start
   a harness, materialize tickets, refresh tracker processing state or change
   lifecycle records.

## Shared Ownership and Scope

Repair and complete the shared Web read projection rather than building a new
execution engine or parsing tracker checkboxes. Extend adapter-neutral shared
read results if required. Preserve the canonical owners of ticket lifecycle,
invocation state, event provenance and delivery receipts.

The existing TUI exposes execution actions, while the Web currently includes
Preview Work and Stop Daemon forms. This ticket's requested outcome is live
monitoring; those controls do not authorize inventing a new execution-control
workflow or treating daemon stop as equivalent to stopping one ticket/run.
Any retained control must keep its existing target, confirmation and guards,
and refreshes must never invoke it.

Coordinate the common data-loading repair with the
[Board ticket](20261017201015-restore-board-page-loading.md) without duplicating
its Kanban or ticket-editor work. Do not add unrelated service management,
configuration editors or a Web-visible tracker document.

## Acceptance

1. `/dash` and its refresh render current canonical data without HTTP 500.
   Cover native ticket workflow mode and the supported legacy observation path.
2. Verify no activity yet, eligible queue, active execution, a long quiet
   validation, blocked work, failure/stop and a completed run with an empty
   queue. Each state remains understandable without reading raw IDs.
3. Advancing a persisted event or workflow phase updates the open page without
   a hard reload. Repeated refreshes preserve position and do not duplicate
   events. A completed run remains visible without a latest phase event.
4. Test two active contexts so a newer event in one cannot be attached to the
   other ticket or overwrite its current state. A display filter must not
   change execution order or readiness semantics.
5. Reopening the page, a temporary read outage and observation after a worker
   restart recover from persisted records. Stale data is labeled, not presented
   as fresh or converted into invented success/failure.
6. Viewing and refreshing do not mutate tickets, tracker content, claims or
   invocation state. Scope counters and completion labels agree with their
   authoritative records and receipt boundaries.
7. Capture desktop and narrow-screen Web evidence for queued, running and
   completed states. Use focused read-model, refresh and presentation checks
   when implementation is authorized; do not start live production work merely
   to capture a screenshot.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).
