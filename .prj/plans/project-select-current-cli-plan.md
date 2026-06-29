# Project Select and Current CLI Plan

Status: delivered

Slicing: `.prj/plans/slicing-user-project-context-plan.md`
Tracker: `.prj/trackers/project-select-current-cli-tracker.md`

## Purpose

Expose active project selection and inspection through CLI commands.

This slice lets the user select a registered project once and inspect the active project later.

## Scope

This slice covers:

- `chr project select <name-or-path>`;
- resolving registered projects by name;
- resolving registered projects by path when unambiguous;
- persisting the selected project id through `ActiveProjectStore`;
- `chr project current`;
- stable, parseable command output;
- focused CLI tests.

## Out of Scope

This slice does not cover:

- reusable project context resolution for all commands;
- ticket commands;
- automatic ticket reconciliation;
- daemon behavior;
- project removal or disabling;
- stale path recovery beyond clear command errors.

## Design

`project select` should use the project registry as the source of truth for selectable projects.

Selection input can be:

- a registered project name;
- a registered project path.

The selected state should store the registry project id through `ActiveProjectStore`.

`project current` should load the active project id and resolve it back to the registry row before printing. If no active project exists, it should return a clear error.

## Tasks

### T1.1 Project Registry Lookup Support

Add registry lookup needed by the CLI commands.

Expected behavior:

- get a project by id;
- get a project by path;
- keep existing list and name lookup behavior unchanged;
- return clear not-found errors.

### T1.2 Project Select Command

Add `chr project select <name-or-path>`.

Expected behavior:

- require exactly one selector argument;
- resolve selector by registered name first;
- resolve selector by registered path when name lookup misses;
- store the selected project id;
- print stable output for the selected project;
- support `--dev-root`.

### T1.3 Project Current Command

Add `chr project current`.

Expected behavior:

- reject positional arguments;
- load active project state;
- resolve the active project id through the registry;
- print stable output for the active project;
- return a clear error when no project is selected;
- support `--dev-root`.

### T1.4 Tests and Tracker

Add focused tests and update tracker state.

Required cases:

- select by name;
- select by path;
- current returns selected project;
- current without selection fails clearly;
- unknown selection fails clearly;
- command behavior uses development-mode state.

## Validation

Required checks:

```sh
make verify
```
