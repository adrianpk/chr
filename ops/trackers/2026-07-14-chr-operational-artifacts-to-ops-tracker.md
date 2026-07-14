# CHR Operational Artifacts to Ops Tracker

Status: planned
Plan: `ops/plans/2026-07-14-chr-operational-artifacts-to-ops-plan.md`
Ticket: `ops/tickets/open/20260714151004-migrate-chr-operational-artifacts-to-ops.md`
Active Slice: Slice 1 — Ops artifact boundary
Expected Branch: `fix/slice-1-ops-artifact-boundary`
Expected PR Title: `feat(slice-1): ops artifact boundary`
Expected Report: `.prj/reports/slices/chr-operational-artifacts-to-ops/slice-1-ops-artifact-boundary.md`

## Slice 1: Ops artifact boundary

Status: planned

### Tasks

- [ ] T1.1 Define code-level path helpers for CHR private state versus operational artifacts.
- [ ] T1.2 Retarget active operational artifact producers/readers to `ops/<namespace>/<type>/...`.
- [ ] T1.3 Preserve `.chr/` configuration paths and update docs/scripts/tests.
- [ ] T1.4 Validate with focused scans, `make verify`, ops layout validation, and diff hygiene.

### Validation

- Pending.

### Commits

- Pending.

### PR

- Pending.

### Risks

- Broad existing UI/test expectations may still assume path-shaped IDs.
- Some legacy QA fixture content may intentionally document historical paths;
  distinguish historical evidence from active behavior.
