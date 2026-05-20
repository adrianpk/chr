# Slice 6: Registered Project Loading

Status: delivered
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-6-daemon-project-loading`
PR: #69

## Purpose

Load registered projects for daemon coordination before daemon work starts.

## Delivered Behavior

The core package now includes daemon project loading.

The loader:

- reads projects from the user-level project registry;
- returns only enabled projects;
- validates each project path;
- classifies projects as `ready`, `missing`, or `invalid`.

## Implementation Notes

Ready projects must have an existing project directory and a `.chc/` workspace
directory.

Missing or invalid projects are returned as classified rows instead of failing
the whole load operation. This lets later daemon run-loop behavior skip or
report unavailable projects without stopping coordination for every project.

## Contracts Added or Changed

Added core loader:

- `LoadDaemonProjects`;
- `DaemonProject`;
- `DaemonProjectStatus`.

## Files of Interest

- `internal/chc/daemon_project_loader.go`
- `internal/chc/daemon_project_loader_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./internal/chc -run 'TestLoadDaemonProjects'`: pass
- `make verify`: pass

## Risks and Follow-ups

The loader does not execute loops or refresh indexes. Slice 7 uses this loading
boundary in the daemon foreground run loop.
