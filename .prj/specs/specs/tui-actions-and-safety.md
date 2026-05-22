# TUI Actions and Safety Model

Status: initial

Sources:

- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`
- `.prj/specs/specs/tui-ticket-execution-and-evidence.md`
- `.agents/INVARIANTS.md`
- `.agents/VERSIONING.md`

## Purpose

Define how the TUI should expose actions without making project mutation easy to
trigger accidentally.

The TUI should feel faster than the CLI for daily use, but it must remain
explicit about file changes, runtime execution, and workflow state changes.

## Action Classes

### Read Actions

Read actions inspect existing state:

- refresh view;
- open detail;
- filter list;
- show source path;
- show equivalent command;
- show report;
- show validation evidence.

Read actions should be quick and should not require confirmation.

### Local Artifact Actions

Local artifact actions create or update Markdown/state files:

- create ticket;
- add ticket comment;
- update ticket status;
- create authoring session;
- create context-aware quiz;
- save quiz answer;
- promote artifact;
- rebuild index.

These actions should show a short confirmation or review screen when they create
new durable files or modify existing ones.

### Runtime Actions

Runtime actions can run Docker, Codex, validation, or ticket loops:

- runtime validate;
- execute ticket;
- dry-run execution;
- run filtered loop;
- retry failed execution;
- apply result;
- emit notification.

Runtime actions require confirmation.

The confirmation should state whether project files may be modified.

### External Actions

External actions leave the local project boundary:

- send Telegram notification;
- send email;
- clone/push through git;
- open external URLs;
- call an LLM provider.

External actions require explicit configuration and clear user-visible state.

Future integrations should not appear as active controls before the backend is
implemented and configured.

## Confirmation Model

Confirmation screens should include:

- action name;
- target record;
- expected file changes;
- expected runtime/external side effects;
- validation command when relevant;
- cancel key;
- confirm key.

Danger should not be communicated only through color.

The confirm key for mutating/runtime actions should be more deliberate than
Enter when practical.

## Dirty Worktree Handling

Before runtime execution or result application, the TUI should surface dirty
worktree state when available.

The TUI should not silently revert user changes. It should not offer destructive
git cleanup actions unless a separate explicit workflow is specified.

When dirty state exists, the TUI should show:

- summary count;
- changed files when available;
- whether the selected action can continue;
- evidence path where runtime output will land.

## Long-Running Actions

Long-running actions should create an operation record in TUI state and, when
the backend supports it, map to a durable run/loop record.

The TUI should show:

- running state;
- elapsed time;
- latest phase;
- cancel availability;
- log/evidence path;
- completion result.

If the TUI process exits while an operation is running, durable backend state
should remain recoverable through CLI or a later TUI session.

## Error Recovery

Every failed action should provide:

- human-readable summary;
- technical detail path or expandable detail;
- related record;
- recommended inspection target;
- whether retry is safe.

The TUI should not automatically retry mutating operations unless the user
selects retry.

## Command Equivalence

For each action that maps to a command, the TUI should be able to show an
equivalent command or conceptual command.

This supports:

- learning the CLI;
- copying commands into runbooks;
- debugging action behavior;
- keeping TUI behavior aligned with CLI behavior.

The command display should use final-user-style commands by default. Developer
flags such as `--dev-root` should appear only in a developer isolation context.

## Acceptance Criteria

- Read actions do not require confirmation.
- Mutating artifact actions show enough context before writing.
- Runtime actions require explicit confirmation.
- External actions are unavailable until implemented and configured.
- Dirty worktree state is surfaced before risky execution.
- Failed actions provide inspectable evidence and next inspection targets.
- The TUI never offers destructive git actions as a casual shortcut.
