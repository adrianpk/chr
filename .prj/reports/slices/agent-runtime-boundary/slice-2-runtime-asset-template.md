# Slice 2: Runtime Asset Template

Status: reviewing
Slicing: Agent Runtime Boundary
Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
Branch: `feat/slice-2-runtime-asset-template`
PR: pending

## Purpose

Provide managed projects with a default, editable Dockerfile for future agent
ticket execution.

## Delivered Behavior

`chc init` now materializes `.chc/runtime/agent/Dockerfile` from each workspace
profile. Profile `config.yml` files also include runtime defaults that point to
that Dockerfile and to the runtime artifact directory.

Existing runtime Dockerfiles are preserved by the existing non-overwrite init
behavior.

## Implementation Notes

The template is intentionally minimal: Ubuntu 24.04 with shell, certificates,
Git, curl, and Make. It is a project-owned starting point, not a production
image or a full agent harness installer.

## Contracts Added or Changed

New initialized project asset:

```text
.chc/runtime/agent/Dockerfile
```

Profile config files now include:

```yml
runtime:
  image: chc-agent-runtime:local
  dockerfile: .chc/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/runtime/runs
  validation_command: make verify
```

## Files of Interest

- `cmd/chc/assets/workspace/profiles/*/runtime/agent/Dockerfile`
- `cmd/chc/assets/workspace/profiles/*/config.yml`
- `internal/chc/workspace_init_test.go`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

The Dockerfile is a generic starter template. Project-specific dependencies,
agent harness installation, and runtime validation commands remain later
runtime-boundary work.
