# Slice 5: Forgejo Woodpecker Integration

Status: delivered

Slicing: Dev to Main Agent Evidence Gate
Spec: `ops/specs/2026-06-01-dev-main-agent-evidence-gate.md`
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`
Branch: `feat/slice-5-forgejo-woodpecker-integration`
PR: #268

## Scope

Slice 5 adds the first Forgejo/Woodpecker integration path for the dev-to-main
gate. The implementation keeps Forgejo API access out of the first version and
uses Woodpecker pipeline status as the feedback mechanism.

## Changes

- Added `chr gate woodpecker init`.
- Added a reusable Woodpecker snippet that runs `chr gate dev-main`.
- Added `--dry-run` support that never writes `.woodpecker.yml`.
- Added safe missing-config creation.
- Added non-destructive existing-config behavior.
- Added marker-based updates for generated Woodpecker configs.
- Added helper and CLI tests for the setup path.

## Feedback Mechanism

The selected feedback mechanism is Woodpecker CI status. The generated workflow
fails when `chr gate dev-main` exits non-zero, which gives Forgejo a promotion
signal without requiring Forgejo API comments or status writes in the first
implementation.

Gate evidence remains local to `.chr/default/reports/gates/`.

## Validation

- `go test ./internal/chr ./internal/chrapp` passed.
- `make verify` passed.
