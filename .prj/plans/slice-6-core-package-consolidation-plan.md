# Slice 6 Core Package Consolidation Plan

Status: delivered

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Tracker: `.prj/trackers/slice-6-core-package-consolidation-tracker.md`

## Purpose

Consolidate core ChrysoCoord behavior under `internal/chc` before adding user environment, project registry, daemon, and global index behavior.

The current `internal/workspace` package name conflicts with the emerging model where a project workspace is only one core `chc` concept.

## Scope

This slice covers:

- moving project workspace initialization from `internal/workspace` to `internal/chc`,
- exposing a project workspace abstraction as `chc.Workspace`,
- exposing construction through `chc.NewWorkspace(...)`,
- preserving existing `chc init` behavior,
- updating tests and imports,
- keeping one-shot CLI behavior unchanged.

## Task Breakdown

### T6.1 Move Workspace Initialization Into Core Package

Move the existing project workspace initialization implementation into `internal/chc`.

Expected behavior:

- keep `.chc/` initialization behavior unchanged,
- keep embedded asset materialization behavior unchanged,
- keep profile behavior unchanged,
- keep current tests passing after import/package updates.

### T6.2 Introduce Workspace-Oriented API Names

Introduce names that match the core model.

Expected behavior:

- expose `Workspace` as the project-local `.chc/` abstraction,
- expose `NewWorkspace(...)`,
- keep constructor/test seams compatible with table-based tests,
- update CLI wiring to use `chc.NewWorkspace(...)`.

### T6.3 Documentation and Tracker Alignment

Update project documentation and slice tracker references to reflect the package consolidation.

Expected behavior:

- docs no longer describe `internal/workspace` as the active core package,
- tracker records validation,
- no runtime behavior changes.

## Out of Scope

This slice does not cover:

- user environment path resolver,
- project registry,
- SQLite index,
- daemon command,
- filesystem watcher,
- Telegram integration.

## Validation

Required checks:

```sh
make verify
```

Manual behavior check:

```sh
tmpdir="$(mktemp -d)"
go run ./cmd/chc init --dir "$tmpdir"
test -d "$tmpdir/.chc"
test -f "$tmpdir/AGENTS.md"
```
