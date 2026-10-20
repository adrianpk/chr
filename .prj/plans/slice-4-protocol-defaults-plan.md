# Slice 4 Protocol Defaults Plan

Status: delivered

## Purpose

Add the first chr protocol defaults to the `standard` workspace profile.

The defaults should give a managed Go project a coherent agent protocol without copying this repository's `.agents/` files mechanically.

## Scope

This slice covers:

- adding `.chr/agents/*.md` defaults to the `standard` profile,
- adding `.chr/agents/playbooks/*.md` defaults to the `standard` profile,
- keeping `personal` and `minimal` unchanged unless needed for validation,
- verifying recursive profile materialization writes the protocol files,
- documenting that `standard` includes protocol defaults.

## Candidate Files

Root protocol files:

- `ARCHITECTURE.md`
- `CONVENTIONS.md`
- `DOCUMENTING.md`
- `INVARIANTS.md`
- `LIFECYCLE.md`
- `LINTING.md`
- `PLAYBOOKS.md`
- `ROUTING.md`
- `SLICES.md`
- `VERSIONING.md`

Playbook files:

- `dump.md`
- `planned-delivery-workflow.md`
- `quiz.md`
- `release-alignment.md`
- `slice-delivery.md`
- `specification-workflow.md`
- `ticket-workflow.md`
- `version-control-workflow.md`

## Out of Scope

This slice does not cover:

- ticket model implementation,
- ticket commands,
- Docker runtime setup,
- Codex harness execution,
- final personal profile content.

## Validation

Required checks:

```sh
make verify
```

Manual behavior checks:

```sh
go run ./cmd/chr init --dir "$tmpdir" --profile standard
test -f "$tmpdir/.chr/agents/ARCHITECTURE.md"
test -f "$tmpdir/.chr/agents/playbooks/ticket-workflow.md"
```
