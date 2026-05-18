# Active Project Persistence Plan

Status: active

Slicing: `.prj/plans/slicing-user-project-context-plan.md`
Tracker: `.prj/trackers/active-project-persistence-tracker.md`

## Purpose

Persist the user's selected project in user-level state so later commands can operate on an active project without requiring repeated project flags.

## Scope

This slice covers:

- an active project state model;
- core persistence for storing and loading the active project;
- replacing the active project selection;
- empty-state behavior when no project has been selected;
- development-mode behavior under the existing user environment paths;
- focused table-based tests.

## Out of Scope

This slice does not cover:

- `chc project select`;
- `chc project current`;
- command-level project context resolution;
- ticket commands;
- daemon behavior;
- project path recovery.

## Design

The active project state belongs to user-level configuration, not project-local `.chc/` state.

Initial persistence should be a small file under `UserEnvironment.ConfigDir()`. This keeps active selection as user preference, while project registry rows remain in the user-level SQLite database.

The stored value should identify the selected project by stable project id. Name and path can change later, but the id is already derived from the absolute path and is the registry primary key.

## Tasks

### T1 Active Project Model

Add a core model for active project state.

Expected behavior:

- represent an active project id;
- reject empty project ids when setting active state;
- return a distinct empty-state result when no active project exists.

### T2 Active Project Store

Add a core store backed by `UserEnvironment.ConfigDir()`.

Expected behavior:

- create the config directory when writing;
- write active project state atomically enough for local CLI use;
- load active project state;
- return empty state when the file does not exist;
- wrap filesystem errors with useful context.

### T3 Tests

Add focused tests for the active project store.

Required cases:

- load empty state;
- set and load active project id;
- replace active project id;
- reject empty active project id;
- use development-mode paths without writing real user state.

## Validation

Required checks:

```sh
make verify
```
