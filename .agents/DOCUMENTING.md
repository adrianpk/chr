# Go Documentation Guidelines

This repository uses Diataxis under `docs/` for durable user-facing and technical documentation:
https://diataxis.fr

## Categories

- `docs/tutorials/` for learning paths.
- `docs/how-to/` for task procedures.
- `docs/reference/` for contracts and exact behavior.
- `docs/explanation/` for rationale and tradeoffs.

The top level of `docs/` should contain only the four Diataxis quadrant
directories above. If a quadrant needs internal structure, add subdirectories
inside that quadrant instead of adding new top-level directories under `docs/`.

## Go examples

- Keep setup and execution examples in Go project terms (`go test ./...`, `go run ./cmd/...`, `go build ./...`).
- Keep command examples aligned with actual `cmd/` entrypoints.
- Keep package references aligned with Go import paths and package names.
- API/reference docs should use Go names, signatures, structs, interfaces, and error contracts.
- Exported package contracts must have useful Godoc. This includes exported
  symbols in `internal/` packages when they are consumed across package
  boundaries.
- Exported struct fields and interface methods must be documented when they are
  part of an input, output, config, or boundary contract.
- Packages with exported contracts should include a local `README.md` when the
  package needs a durable boundary, behavior, or usage explanation beyond
  Godoc.
- Package READMEs should avoid mixing unrelated documentation styles. Prefer a
  consistent structure: purpose, boundary, API, types, behavior, and examples.

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
