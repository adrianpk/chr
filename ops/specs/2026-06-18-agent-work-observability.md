# Agent Work Observability

Status: draft

Date: 2026-06-18

Model: `ops/specs/2026-06-18-agent-work-observability-model.md`

## Summary

ChrysoCoord should show useful live feedback while long-running agent work is
executing.

Today, when a user starts `Run all ready tickets` from the TUI Dash, the UI
enters a modal work state and renders a static message:

```text
Work Action Running
  Running all ready tickets...

  Please wait for this action to finish.
```

The work may be progressing correctly, but the user has to infer progress from
external side effects such as Git diffs, ticket files, or runtime logs. This is
especially weak when multiple Codex-backed tickets are running in sequence and
each execution may take a long time.

The desired behavior is a modal but observable work screen: navigation remains
locked while work is running, but the Dash shows a human-readable event stream
and provides a path to deeper execution/container logs.

## Goals

- Replace the static work-running message with a live human-readable event
  stream.
- Persist agent-work events as append-only durable evidence.
- Keep the TUI work action modal while work is running.
- Make it clear that navigation and exit are intentionally locked during active
  foreground work.
- Provide two feedback levels:
  - a concise event stream for normal users;
  - a deeper execution detail/log path for debugging.
- Let persisted events help a future agent or user understand where an
  interrupted workflow stopped.
- Keep event contracts structured enough for TUI rendering, CLI inspection, and
  recovery summaries.

## Non-Goals

- Do not make foreground work freely navigable in this iteration.
- Do not make `F10` quit active foreground work.
- Do not stream arbitrary unstructured Codex narration directly into the primary
  TUI surface.
- Do not require a database-backed event store.
- Do not implement a full terminal log viewer before the durable event stream
  exists.
- Do not change ticket implementation behavior or result application semantics.

## Desired Behavior

### Work Running Screen

When a work action is running, Dash should show:

- the active action label;
- a clear lock notice;
- recent human-readable events;
- the current ticket or execution when known;
- an optional hint for deeper logs when an execution exposes a log path.

Example:

```text
Work Action Running
Navigation is locked until this action finishes.

11:32:01 loop started: 5 ready tickets
11:32:04 CHC-1 started: Game outcome rules
11:32:07 CHC-1 branch prepared
11:32:11 CHC-1 Codex worker started
11:34:22 CHC-1 worker finished
11:34:33 CHC-1 validation passed: go test ./...
11:34:34 CHC-1 moved to reviewing
11:34:36 CHC-2 started: Move state transitions
```

The screen should keep updating while the foreground work action runs.

### Event Stream

Events should be emitted by ChrysoCoord-owned runtime boundaries at lifecycle
points such as:

- loop started;
- ticket selected;
- execution created;
- branch or worktree prepared;
- runtime/container prepared;
- worker started;
- worker finished;
- changed files discovered;
- validation started;
- validation passed or failed;
- result applied;
- ticket moved to `reviewing` or `blocked`;
- loop finished, failed, or stopped.

The event stream should be written durably before or at the same time the user
visible state changes. The event stream is evidence, not only UI cache.

### Feedback Levels

The normal Dash view should show concise events. It should avoid raw log spam.

The expanded/debug path should expose execution details:

- ticket key;
- execution id;
- loop id when available;
- branch or worktree refs;
- container id when available;
- changed files;
- validation command;
- result path;
- runtime log path;
- validation log path.

The first implementation exposes deeper paths through CLI commands:

```text
chc work logs <execution-id>
chc work events <event-stream-path>
```

The TUI does not need to embed a full log viewer in the first slice set.

### Modal Execution

Foreground work remains modal.

While `workBusy` is active:

- navigation keys remain blocked;
- `F10` does not exit;
- the UI explains why it is locked;
- the event stream continues updating.

This preserves the existing safety posture and avoids accidental interruption of
expensive agent work.

### Recovery and Readback

Persisted event files should be useful after a crash or interruption. A user or
agent should be able to inspect the latest event stream and identify:

- the latest loop;
- the latest ticket;
- whether an execution started;
- whether validation ran;
- whether result application completed;
- the latest known terminal status.

Recovery does not need to resume work automatically in this iteration.

## Acceptance Criteria

- Agent work emits structured append-only events for loop and ticket execution
  lifecycle points.
- Event files survive process exit and can be read independently of the TUI.
- Dash renders recent work events while foreground work is active.
- The static "Please wait" screen is replaced by a live event stream and lock
  explanation.
- Navigation remains locked during active foreground work.
- Existing `chc work run-all` and TUI Dash work actions still execute tickets in
  the same order and with the same result application semantics.
- The TUI or CLI exposes a stable way to locate detailed runtime logs for the
  active or recent execution.
- Tests cover event persistence, event rendering, and at least one work-loop
  lifecycle emission path.

## Open Questions

- Whether the command for detailed logs should be `chc work logs`,
  `chc runtime logs`, or report-based lookup.
- Whether event files should be grouped only by loop/execution id or also have a
  project-level latest stream index.
- Whether container identifiers are always available from every runtime backend.
- Whether old event streams should participate in existing maintenance cleanup
  commands.
