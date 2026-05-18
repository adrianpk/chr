# Project Validation and Recovery Plan

Status: active

Slicing: `.prj/plans/slicing-user-project-context-plan.md`
Tracker: `.prj/trackers/project-validation-recovery-tracker.md`

## Purpose

Make stale registered project paths understandable and recoverable at command boundaries.

This slice closes the `User Project Context` slicing by making project context failures explicit when a registered or active project path is missing, no longer contains `.chc/`, or is disabled.

## Scope

This slice covers:

- clear missing project path errors;
- clear registered path without `.chc/` errors;
- clear disabled project errors;
- validation behavior for active project resolution;
- validation behavior for `project select`;
- focused CLI tests.

## Out of Scope

This slice does not cover:

- project remove or disable commands;
- automatic repair of moved project paths;
- ticket listing or ticket showing;
- automatic ticket reconciliation;
- daemon behavior.

## Design

Existing command behavior should remain conservative:

- never silently select a different project;
- never auto-remove or auto-disable stale registry rows;
- fail with an actionable error when the selected or active project cannot be used.

The command boundary should distinguish:

- registered project path no longer exists;
- registered project path exists but lacks `.chc/`;
- registered project is disabled;
- no active project exists.

This slice should prefer improving validation helpers and command errors over adding new user-facing commands.

## Tasks

### T1.1 Project Workspace Validation Helpers

Add focused validation helpers for registered project paths.

Expected behavior:

- validate that a project path exists;
- validate that the path has a `.chc/` directory;
- return contextual errors that include the project name or path.

### T1.2 Apply Validation to Selection and Active Context

Use the validation helpers in command paths that resolve registered projects.

Expected behavior:

- `project select` rejects missing registered paths clearly;
- `project select` rejects registered paths without `.chc/` clearly;
- active project resolution rejects missing active project paths clearly;
- active project resolution rejects active paths without `.chc/` clearly;
- disabled projects remain rejected distinctly.

### T1.3 Tests and Tracker

Add focused tests and update tracker state.

Required cases:

- select rejects registered project whose path was deleted;
- select rejects registered project whose `.chc/` was removed;
- active project command rejects deleted active project path;
- active project command rejects active project path without `.chc/`;
- disabled project rejection remains distinct where directly exercised;
- tracker records task commits and validation.

## Validation

Required checks:

```sh
make verify
```
