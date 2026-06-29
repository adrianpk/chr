# Slice 1: Runtime Config Model

Status: delivered
Slicing: Agent Runtime Boundary
Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
Branch: `feat/slice-1-runtime-config-model`
PR: `#29`

## Purpose

Add the first typed runtime configuration contract for project-local
`.chr/config.yml` files.

## Delivered Behavior

`LoadWorkspaceConfig` now returns a `RuntimeConfig` alongside the existing ticket
key prefix. Missing runtime fields use deterministic defaults, and configured
runtime values are read from the `runtime:` section in `.chr/config.yml`.

## Implementation Notes

The config loader remains a small line-oriented parser rather than introducing a
YAML dependency in this slice. It recognizes the current project config shape
and preserves compatibility with the existing top-level `key_prefix` fallback.

## Contracts Added or Changed

Runtime config fields:

- `runtime.image`
- `runtime.dockerfile`
- `runtime.workdir`
- `runtime.artifact_dir`
- `runtime.validation_command`

Default runtime config:

```yml
runtime:
  image: chr-agent-runtime:local
  dockerfile: .chr/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chr/runtime/runs
  validation_command: make verify
```

## Files of Interest

- `internal/chr/config.go`
- `internal/chr/config_test.go`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

The loader intentionally supports the known config contract only. Broader YAML
features remain out of scope unless future config needs justify a parser
dependency.
