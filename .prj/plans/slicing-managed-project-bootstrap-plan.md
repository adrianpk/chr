# Managed Project Bootstrap Slicing Plan

Status: completed

## Purpose

Plan the `Managed Project Bootstrap` slicing from the feature slicing set map.

This slicing covers the use case:

> A user initializes a Go project so it can be managed by `chr`.

## Source

Primary source:

- `.prj/drafts/chr-feature-slicing-set-map.md`

Related delivered plans:

- `.prj/plans/slice-1-urfave-cli-bootstrap-plan.md` was represented by the first delivered PR, before the current numbered plan files started.
- `.prj/plans/slice-2-asset-profiles-plan.md`
- `.prj/plans/slice-3-root-agents-plan.md`
- `.prj/plans/slice-4-protocol-defaults-plan.md`

## Current State

The slicing is delivered.

Already delivered:

- CLI entrypoint using `urfave/cli`;
- `chr init`;
- default current-directory initialization;
- optional `--dir` override;
- `standard`, `personal`, and `minimal` profile selection;
- embedded workspace assets;
- root-level `AGENTS.md` materialization;
- default protocol files in the `standard` profile;
- recursive profile materialization;
- dry-run support;
- unknown profile rejection;
- basic safety behavior for preserving existing root files.

Bootstrap behavior gaps were closed by the safety pass. Further project registry, ticket, runtime, daemon, or execution work belongs to separate slicing sets.

## Slicing Scope

This slicing includes:

- CLI bootstrap;
- project-local `.chr/` initialization;
- embedded profile assets;
- root `AGENTS.md`;
- default protocol assets;
- profile-specific defaults;
- initialization dry-run and safety checks;
- bootstrap validation and documentation needed to make `chr init` dependable.

## Out of Scope

This slicing does not include:

- user-level project registry;
- active project selection;
- ticket creation, listing, indexing, or execution;
- automatic reconciliation;
- Docker runtime setup;
- agent execution;
- daemon behavior;
- notifications.

## Slices

### CLI Bootstrap

Status: delivered.

Delivery:

- root `chr` command;
- `urfave/cli` command structure;
- `init`, `project`, and `ticket` command families available.

### Project-Local Initialization

Status: delivered.

Delivery:

- `chr init`;
- creates `.chr/` under the target project;
- defaults to the current directory;
- supports `--dir`.

### Asset Profiles

Status: delivered.

Delivery:

- embedded workspace assets;
- `standard`, `personal`, and `minimal` profiles;
- recursive profile materialization;
- unknown profile validation.

### Root Agents File

Status: delivered.

Delivery:

- root-level `AGENTS.md` materialization;
- generated root entrypoint points into the managed protocol;
- existing root file preservation.

### Protocol Defaults

Status: delivered.

Delivery:

- standard profile protocol files;
- standard profile playbooks;
- generated `.chr/agents/` defaults.

### Bootstrap Safety Pass

Status: delivered.

Goal:

Close initialization safety and inspectability gaps discovered after the first bootstrap implementation.

Candidate scope:

- verify dry-run output remains useful for existing and new paths;
- verify existing `.chr/` content is not overwritten unexpectedly;
- verify root `AGENTS.md` preservation behavior is documented and tested from the CLI boundary;
- expose clearer errors when the target directory is invalid or not writable;
- keep behavior deterministic and small.

The safety pass was the final implementation slice for this slicing.

## Acceptance Criteria

The slicing is complete when:

- `chr init` can initialize a Go project from the current directory;
- `chr init --dir <path>` can initialize a target directory;
- `chr init --profile standard` installs standard protocol defaults;
- `chr init --profile personal` and `chr init --profile minimal` install their profile assets;
- `chr init --dry-run` reports planned work without writing files;
- existing root files are not overwritten silently;
- invalid profile and invalid target failures are clear;
- validation passes with `make verify`.

## Validation

Required validation for this slicing:

```sh
make verify
```

Manual smoke checks:

```sh
tmpdir="$(mktemp -d)"
chr init --dir "$tmpdir"
test -d "$tmpdir/.chr"
test -f "$tmpdir/AGENTS.md"
```

```sh
tmpdir="$(mktemp -d)"
chr init --dir "$tmpdir" --dry-run
test ! -e "$tmpdir/.chr"
```

```sh
tmpdir="$(mktemp -d)"
chr init --dir "$tmpdir" --profile standard
test -f "$tmpdir/.chr/agents/CONVENTIONS.md"
```

## Risks

- Bootstrap behavior is foundational; accidental overwrite behavior would be high-impact.
- Profile assets can drift from the code contract if tests only cover synthetic assets.
- More bootstrap options can become configuration noise if not tied to a real initialization problem.

## Notes

- No quiz was needed for this plan because the use case and delivered behavior are already documented.
- The next implementation work from this slicing should be a small bootstrap safety pass, not a broad initialization redesign.
