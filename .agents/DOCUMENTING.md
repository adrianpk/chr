# Go Documentation Guidelines

This repository uses Diataxis under `docs/` for durable user-facing and technical documentation:
https://diataxis.fr

## Categories

- `docs/tutorials/` for learning paths.
- `docs/how-to/` for task procedures.
- `docs/reference/` for contracts and exact behavior.
- `docs/explanation/` for rationale and tradeoffs.

## Go examples

- Keep setup and execution examples in Go project terms (`go test ./...`, `go run ./cmd/...`, `go build ./...`).
- Keep command examples aligned with actual `cmd/` entrypoints.
- Keep package references aligned with Go import paths and package names.
- API/reference docs should use Go names, signatures, structs, interfaces, and error contracts.
- Public exported APIs should have useful Godoc when they are part of a package contract.

## Contract documentation

- Document stable wire formats, manifest fields, config fields, persistence schemas, and public package contracts under `docs/reference/`.
- Document architectural rationale and tradeoffs under `docs/explanation/`.
- Keep generated or embedded asset documentation close to the owning package or under `docs/reference/` when it is a user-facing contract.

## Operational documents

- `ops/specs/` stores formal specs and drafts.
- `ops/conv/quiz/` stores raw and curated quiz artifacts.
- `ops/conv/dump/` stores raw and curated knowledge dumps.
- `ops/plans/` stores cycle plans.
- `ops/trackers/` stores execution trackers.
- `ops/artifacts/` stores durable cycle outputs and reports.
- Do not mix operational delivery records into `docs/` unless they become durable product documentation.

## General rules

- Keep all repository docs in English.
- Each document should have one primary intent.
- Split documents when intent changes.
- Preserve technical meaning when reorganizing docs.
- Avoid title/heading suffixes in parentheses.
- Prefer semantic metadata fields over title decoration when state needs to be explicit.
- For structural documentation updates, complete one consistency pass before commit.

## Scope boundary

Diataxis applies to `docs/*` only.
Operational instructions in `.agents/*` are outside Diataxis.
