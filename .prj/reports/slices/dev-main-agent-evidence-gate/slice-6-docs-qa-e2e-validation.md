# Slice 6: Docs QA and End To End Validation

Status: delivered
Slicing: Dev to Main Agent Evidence Gate
Plan: `ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md`
Tracker: `ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md`
Branch: `feat/slice-6-docs-qa-e2e-validation`
PR: #269

## Purpose

Document the completed dev-to-main gate, add a repeatable QA scenario for local
gate behavior, and validate the completed slice set.

## Delivered Behavior

- Added command reference coverage for `chr gate dev-main`.
- Added command reference coverage for `chr gate woodpecker init`.
- Added gate config and report schema reference.
- Added a Woodpecker how-to guide.
- Added `make qa-gate` as a repeatable local gate scenario.
- Validated the slice branch with the normal verification suite.
- Ran the gate against a temporary worktree of this repository.

## Implementation Notes

The QA scenario uses a temporary Go project and a temporary Go build cache so it
does not depend on the user's global Go cache or leave managed-project artifacts
inside the chr repository.

The self-gate validation used a temporary worktree to avoid writing `.chr/`
artifacts into the active checkout.

## Contracts Added or Changed

- New reference doc: `docs/reference/dev-main-gate.md`.
- New how-to doc: `docs/how-to/use-dev-main-gate-with-woodpecker.md`.
- New QA target: `make qa-gate`.
- New QA helper: `scripts/qa/gate.sh`.

## Files of Interest

- `docs/reference/dev-main-gate.md`
- `docs/reference/commands.md`
- `docs/how-to/use-dev-main-gate-with-woodpecker.md`
- `scripts/qa/gate.sh`
- `Makefile`
- `docs/index.md`

## Validation

- `make verify` (pass)
- `make qa-gate` (pass)
- `GOCACHE=/tmp/chr-gate-self.aRZ3ao/.go-build-cache go run /workspaces/chr/cmd/chr gate dev-main --dir /tmp/chr-gate-self.aRZ3ao --base origin/dev --head HEAD` (pass)

## Risks and Follow-ups

None recorded.
