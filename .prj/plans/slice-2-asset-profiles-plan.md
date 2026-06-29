# Slice 2 Asset Profiles Plan

Status: delivered

## Purpose

Move the asset profile work out of direct `dev` history and back into the required slice workflow.

## Scope

This slice covers:

- embedded workspace profile assets,
- `standard`, `personal`, and `minimal` init profiles,
- `chr init --profile`,
- recursive profile file materialization into `.chr/`,
- validation for unknown profiles,
- documentation for profile usage.

## Out of Scope

This slice does not cover:

- root `AGENTS.md` generation,
- protocol default content,
- ticket model,
- Docker agent runtime,
- Codex harness execution.

## Workflow

1. Revert the unapproved asset profile commits from `dev`.
2. Push corrected `dev`.
3. Create a slice branch from corrected `dev`.
4. Reapply the asset profile changes on the slice branch.
5. Run validation.
6. Push the branch.
7. Open a PR against `dev`.
8. Wait for user review and approval before merge.

## Validation

Required checks:

```sh
make verify
```

Manual behavior checks:

```sh
go run ./cmd/chr init --dir "$tmpdir" --dry-run
go run ./cmd/chr init --dir "$tmpdir" --profile minimal
go run ./cmd/chr init --dir "$tmpdir" --profile invalid
```
