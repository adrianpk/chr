# Slice 1: Todolist Fixture Project

Status: delivered
Slicing: MVP Agent Proving Ground
Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
Branch: `feat/slice-1-proving-ground-fixture`
PR: #72

## Purpose

Create the controlled project used to prove that `chr` can run a sandboxed
Codex agent against a real code task.

## Delivered Behavior

The repository now contains a versioned todolist fixture source under:

```text
.prj/fixtures/mvp-agent-proving-ground/todolist
```

The fixture materializes into the ignored local workspace:

```text
_workspace/testing/projects/todolist
```

The fixture is intentionally not green. It has:

- a backend bug where task IDs are reused after deletion;
- a missing `CompleteTask` feature;
- tests that expose both problems;
- compile, lint, and test validation commands;
- `.chr/` runtime config;
- a hand-authored sandbox Dockerfile for agent execution.

## Implementation Notes

`_workspace/` remains ignored. The committed fixture source lives under `.prj/`
and is copied into `_workspace/testing/projects/todolist` by:

```sh
./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh
```

The fixture's `make compile` and `make lint` commands pass. Its `make verify`
fails by design until later proving-ground slices run the agent to fix the bug
and implement the feature.

## Contracts Added or Changed

New fixture convention:

- committed fixture source lives under `.prj/fixtures/mvp-agent-proving-ground/`;
- generated proving projects live under `_workspace/testing/projects/`;
- fixture validation is `make verify`;
- fixture runtime image is `chr-todolist-agent:local`;
- fixture runtime Dockerfile is `.chr/runtime/agent/Dockerfile`.

## Files of Interest

- `.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/internal/todo/store.go`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/internal/todo/store_test.go`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chr/config.yml`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chr/runtime/agent/Dockerfile`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chr/tickets/TODO-1-fix-task-id-reuse.md`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chr/tickets/TODO-2-complete-task.md`

## Validation

- `./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh`: pass
- `make compile` in `_workspace/testing/projects/todolist`: pass
- `make lint` in `_workspace/testing/projects/todolist`: pass
- `make verify` in `_workspace/testing/projects/todolist`: expected failure
  exposing task ID reuse and missing completion behavior
- `make verify`: pass

## Risks and Follow-ups

The sandbox Dockerfile installs Codex through npm during image build. This is
acceptable for the proving-ground window but should be revisited before
generalizing runtime setup.
