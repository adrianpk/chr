# Slice 2: Workspace Bootstrap Contract Cleanup

Status: delivered
Slicing: UX Operation Surface
Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
Tracker: `.prj/trackers/slicing-ux-operation-surface-tracker.md`
Branch: `feat/slice-2-workspace-bootstrap-contract`
PR: #135

## Purpose

Remove pseudo-configurable internal invariants from generated workspace
surfaces while preserving runtime defaults and compatibility for existing
managed projects.

## Delivered Behavior

- New generated workspace config no longer includes `workspace_root`.
- New generated workspace config no longer includes `canonical_assets`.
- New generated workspace config no longer repeats default runtime paths,
  workdir, artifact directory, or validation command.
- New generated workspace config uses the notification field shape actually
  parsed by the config loader.
- Generated workspace manifests now read as human orientation instead of a
  parallel config-like identity block.

## Implementation Notes

- Runtime defaults remain owned by `chr` through `DefaultRuntimeConfig`.
- Existing configs that specify runtime fields remain supported by the loader.
- Existing configs that include old top-level invariant fields remain harmless;
  they are ignored by the loader as before.
- Init tests now assert the generated config and manifest boundaries directly.

## Contracts Added or Changed

- Generated `.chr/config.yml` contains only user-owned initial decisions:
  ticket key prefix and disabled Telegram notification state.
- Generated `.chr/manifest.md` describes useful workspace locations without
  exposing `.chr` root or Markdown canonical assets as config-like values.
- Runtime settings remain configurable when explicitly present, but generated
  workspaces rely on binary defaults.

## Files of Interest

- `cmd/chr/assets/workspace/profiles/standard/config.yml`
- `cmd/chr/assets/workspace/profiles/minimal/config.yml`
- `cmd/chr/assets/workspace/profiles/personal/config.yml`
- `cmd/chr/assets/workspace/profiles/standard/manifest.md`
- `cmd/chr/assets/workspace/profiles/minimal/manifest.md`
- `cmd/chr/assets/workspace/profiles/personal/manifest.md`
- `internal/chr/workspace_init_test.go`
- `docs/reference/runtime-config.md`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chr -run 'TestInit|TestLoadWorkspaceConfig'` - pass
- `make verify` - pass

## Risks and Follow-ups

- Older historical plans and reports still mention the previous generated
  config shape as historical context.
- A later protocol-boundary slice still needs to clean generated agent protocol
  docs separately.
