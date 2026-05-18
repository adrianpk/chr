# Slice 9 Project Registry Repository Plan

Status: active

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`
Tracker: `.prj/trackers/slice-9-project-registry-repository-tracker.md`

## Purpose

Add the core project registry repository backed by the global user SQLite database.

The registry stores local projects known by the user daemon. It does not replace project-local `.chc/` state and does not add CLI commands yet.

## Scope

This slice covers:

- `chc.Project` model,
- project registry repository interface and SQLite implementation,
- opening the global database from `UserEnvironment.DataDir()`,
- running embedded migrations before registry use,
- adding projects,
- listing projects,
- getting projects by name,
- focused repository tests with temporary SQLite databases.

## Task Breakdown

### T9.1 Project Model

Add the core `Project` model.

Expected behavior:

- represent id, name, path, enabled, registered timestamp, and last indexed timestamp,
- keep the model independent from SQL row scanning details,
- keep timestamps deterministic in tests.

### T9.2 Project Registry Repository

Add the SQLite-backed project registry repository.

Expected behavior:

- open or create the global database under `UserEnvironment.DataDir()`,
- run migrations before repository operations,
- insert project rows,
- list projects in stable order,
- get a project by name,
- surface uniqueness errors as normal operation errors.

### T9.3 Validation and Documentation

Document the project registry repository and validate the slice.

Expected behavior:

- document the registry database contract,
- update tracker,
- run `make verify`.

## Out of Scope

This slice does not cover:

- `chc project add` or `chc project list` CLI commands,
- ticket indexing,
- daemon command,
- filesystem watcher,
- Telegram integration.

## Validation

Required checks:

```sh
make verify
```
