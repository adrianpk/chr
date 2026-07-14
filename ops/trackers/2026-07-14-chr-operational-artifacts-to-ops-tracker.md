# CHR Operational Artifacts to Ops Tracker

Status: implemented
Plan: `ops/plans/2026-07-14-chr-operational-artifacts-to-ops-plan.md`
Ticket: `ops/default/ticket/00053-migrate-chr-operational-artifacts-to-ops/artifact.md`
Active Slice: Slice 1 — Ops artifact boundary
Expected Branch: `fix/slice-1-ops-artifact-boundary`
Expected PR Title: `feat(slice-1): ops artifact boundary`
Expected Report: `.prj/reports/slices/chr-operational-artifacts-to-ops/slice-1-ops-artifact-boundary.md`

## Slice 1: Ops artifact boundary

Status: implemented

### Tasks

- [x] T1.1 Define code-level path helpers for CHR private state versus operational artifacts.
- [x] T1.2 Retarget active operational artifact producers/readers to `ops/<namespace>/<type>/...`.
- [x] T1.3 Preserve `.chr/` configuration paths and update docs/scripts/tests.
- [x] T1.4 Validate with focused scans, `make verify`, ops layout validation, and diff hygiene.

### Validation

- `make verify` — passed.
- `GOCACHE=/tmp/chr-ops-boundary-go-cache go test ./...` — passed.
- `python3 /home/adrian/Projects/ops/default/spec/scripts/validate_ops_layout.py --root .` — passed.
- `git diff --check` — passed.
- Active operational `.chr/default` text scan — no matches.

### Commits

- `c63e92d1` — `feat(paths): migrate operational artifacts to ops envelopes`.

### PR

- `#361` — `feat(slice-1): ops artifact boundary`.

### Risks

- Broad existing UI/test expectations may still assume path-shaped IDs.
- Some legacy QA fixture content may intentionally document historical paths;
  distinguish historical evidence from active behavior.
