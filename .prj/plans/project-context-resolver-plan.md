# Project Context Resolver Plan

Status: delivered

Slicing: `.prj/plans/slicing-user-project-context-plan.md`
Tracker: `.prj/trackers/project-context-resolver-tracker.md`

## Purpose

Add a reusable command boundary for resolving which project a command should operate on.

This slice makes later ticket, execution, and loop commands use one project resolution path instead of duplicating current-directory, active-project, and explicit directory handling.

## Scope

This slice covers:

- a CLI-side project context type;
- explicit `--dir` resolution;
- active project fallback;
- current-directory fallback for commands that allow it;
- `.chc/` workspace validation before returning a context;
- clear missing-context errors;
- integrating `ticket new` with the resolver while preserving current-directory behavior;
- focused CLI tests.

## Out of Scope

This slice does not cover:

- project path recovery for moved or deleted registered projects;
- project disable/remove commands;
- ticket list or ticket show;
- automatic ticket reconciliation;
- agent runtime or execution behavior;
- daemon behavior.

## Design

The resolver belongs at the CLI boundary because it combines command flags, user environment construction, active project state, and registry lookup.

Resolution order:

1. explicit `--dir` when the command exposes and receives it;
2. active project from `ActiveProjectStore`;
3. current directory when the command allows current-directory fallback.

The resolver should validate that the returned project directory has a `.chc/` workspace.

For `ticket new`, the existing `--dir` behavior remains available. When `--dir` is omitted, the command can use the active project. If no active project exists, it falls back to the current directory to preserve existing behavior.

## Tasks

### T1.1 Project Context Type and Resolver

Add CLI-side project context helpers.

Expected behavior:

- return a project directory;
- support explicit `--dir`;
- support active project fallback;
- support current-directory fallback when enabled;
- return clear errors when no project can be resolved;
- validate `.chc/` before returning.

### T1.2 Ticket New Integration

Update `chc ticket new` to use the resolver.

Expected behavior:

- `--dir <path>` keeps working;
- omitted `--dir` uses the active project when one exists;
- omitted `--dir` falls back to current directory when no active project exists;
- ticket creation still reads `.chc/config.yml` from the resolved project.

### T1.3 Tests and Tracker

Add focused tests and update tracker state.

Required cases:

- explicit `--dir` ticket creation still works;
- active project ticket creation works without `--dir`;
- current-directory fallback still works without active project;
- missing context fails clearly when current-directory fallback is not allowed by the helper;
- tracker records task commits and validation.

## Validation

Required checks:

```sh
make verify
```
