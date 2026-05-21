# Slice 7 User Environment Paths Plan

Status: delivered

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Tracker: `.prj/trackers/slice-7-user-environment-paths-tracker.md`

## Purpose

Add the user-level path resolver needed by the future project registry, global ticket index, and daemon.

The resolver must keep user-level `chc` state separate from project-local `.chc/` state and must support development mode so work on this repository does not write to real user config or data paths.

## Scope

This slice covers:

- `chc.UserEnvironment`,
- `chc.NewUserEnvironment(...)`,
- normal user path resolution for config, data, cache, and state,
- development-mode path resolution into a local sandbox,
- focused tests for path resolution.

## Task Breakdown

### T7.1 User Environment Model

Create the user environment abstraction and constructor options.

Expected behavior:

- expose config, data, cache, and state directory paths,
- keep construction deterministic in tests,
- avoid touching the filesystem during construction.

### T7.2 Normal and Development Path Resolution

Implement path resolution for normal mode and development mode.

Expected behavior:

- normal mode uses platform/user directory inputs,
- development mode resolves paths under a caller-supplied sandbox root,
- development paths do not point at real user config or data directories.

### T7.3 Documentation and Validation

Update documentation and tracker state.

Expected behavior:

- document the user environment path contract,
- run validation,
- do not add registry, SQLite, daemon, watcher, or Telegram behavior.

## Out of Scope

This slice does not cover:

- project registry,
- global SQLite index,
- daemon command,
- filesystem watcher,
- Telegram integration,
- CLI flags for daemon or project registration.

## Validation

Required checks:

```sh
make verify
```
