# Slice 3 Root AGENTS Plan

Status: active

## Purpose

Materialize a root-level `AGENTS.md` when `chc init` initializes a managed project.

External agent tooling may look for `AGENTS.md` at the project root, so the root file must exist outside `.chc/` even though ChrysoCoord-owned workspace state lives under `.chc/`.

## Scope

This slice covers:

- adding root `AGENTS.md` to workspace profile assets,
- writing root-level files during `chc init`,
- preserving existing root `AGENTS.md` without overwrite,
- reporting root file creation and skips in init output,
- tests for root file materialization and non-overwrite behavior.

## Out of Scope

This slice does not cover:

- full protocol defaults under `.chc/agents/`,
- root `AGENTS.md` content finalization beyond a useful first template,
- ticket model,
- agent execution,
- Docker runtime setup.

## Validation

Required checks:

```sh
make verify
```

Manual behavior checks:

```sh
go run ./cmd/chc init --dir "$tmpdir"
go run ./cmd/chc init --dir "$tmpdir" --dry-run
```

Both checks must avoid creating `.chc/` in the ChrysoCoord repository itself.
