# TUI Ticket Execution and Evidence

Status: initial

Sources:

- `.prj/specs/specs/ticket-execution-context.md`
- `.prj/specs/specs/ticket-resolution-lifecycle.md`
- `.prj/specs/specs/execution-evidence-and-handoff.md`
- `.prj/specs/specs/tui-product-surface.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-loop.md`

## Purpose

Define how the TUI should expose ticket execution, Ralph-loop progress, runtime
validation, and evidence review.

The TUI should make execution observable and recoverable. It should not hide
attempts behind a single success/failure message.

## Execution Entry Points

The TUI should eventually expose:

- execute selected ticket;
- dry-run selected ticket execution;
- validate runtime;
- run a filtered ticket loop;
- stop or pause a loop when the runtime supports it;
- inspect latest run for a ticket;
- inspect latest loop record.

Early implementation may make some actions read-only or unavailable.

## Execution Confirmation

Before running ticket execution, confirmation should show:

- selected ticket key and title;
- current ticket status;
- runtime image;
- validation command;
- sandbox boundary summary;
- expected artifact location;
- dirty worktree warning when known;
- whether the action may modify project files.

The confirmation should require an explicit key action. It should not execute
because the user pressed Enter on the ticket detail view.

## Runtime Validation View

Runtime validation should show compact checks:

```text
checks:
  ok      workspace     .chr
  ok      dockerfile    .chr/runtime/agent/Dockerfile
  ok      artifact_dir  .chr/runtime/runs
  missing auth          ~/.codex/auth.json
```

The exact labels may change, but the output should avoid duplicated status
words and should remain understandable without color.

Validation detail should show:

- runtime image;
- Dockerfile path;
- mount summary;
- workdir;
- validation command;
- log path;
- result path.

Paths should be project-relative when possible.

## Execution Progress

While a ticket is running, the TUI should show:

- execution ID;
- ticket key;
- attempt number;
- current phase;
- elapsed time;
- last log line or summarized status;
- validation state when available.

The user should be able to navigate away and return to the run screen without
losing the run state.

## Ralph Loop Display

When a Ralph loop is involved, the TUI should display attempts as a sequence:

- attempt number;
- objective summary;
- action taken;
- validation command;
- validation result;
- blocker or next step;
- artifact path.

The user should see whether the loop is improving, blocked, or repeating.

Threshold behavior should be visible. If an execution reaches a configured
attempt or blocker threshold, the TUI should show the comment, notification, or
handoff material produced by the system.

## Evidence Review

Evidence detail should include:

- result envelope status;
- changed file summary;
- validation output summary;
- runtime log link/path;
- ticket comments added by the execution;
- checklist changes;
- report path;
- handoff summary when present.

The TUI should not require the user to open `result.json` manually to
understand whether execution succeeded.

## Loop View

Ticket loops should be inspectable as their own records.

Loop list columns:

- loop ID;
- status;
- filter summary;
- ordered by;
- processed count;
- failed count;
- updated.

Loop detail fields:

- source command or selected filters;
- ticket selection result;
- per-ticket execution outcome;
- blocker threshold;
- notifications emitted;
- report path.

## Failure States

The TUI should distinguish:

- runtime validation failure;
- Docker startup failure;
- missing credentials/configuration;
- Codex worker failure;
- validation command failure;
- result contract parse failure;
- ticket application failure;
- loop threshold reached;
- user cancellation.

Each failure should show:

- short reason;
- related log/report path;
- next useful inspection action;
- whether the project may have been modified.

## Acceptance Criteria

- A selected ticket can show latest execution and evidence.
- Runtime validation output is compact and readable.
- Execution confirmation makes mutation risk explicit.
- Ralph-loop attempts are visible as records, not buried in logs.
- Failed executions leave inspectable evidence and clear next actions.
- Loop execution can be inspected separately from individual ticket details.

