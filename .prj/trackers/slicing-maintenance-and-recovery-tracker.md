# Maintenance and Recovery Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm existing index, project validation, runtime validation, and
  reporting prerequisites.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Database Migration Status | Delivered | `feat/slice-1-db-migration-status` | `feat(slice-1): db migration status` | `.prj/reports/slices/maintenance-and-recovery/slice-1-db-migration-status.md` | Repository PR #126 merged. |
| Slice 2: Project Registry Diagnostics | Planned | `feat/slice-2-project-registry-diagnostics` | `feat(slice-2): project registry diagnostics` | `.prj/reports/slices/maintenance-and-recovery/slice-2-project-registry-diagnostics.md` | Pending. |
| Slice 3: Project Registry Repair | Planned | `feat/slice-3-project-registry-repair` | `feat(slice-3): project registry repair` | `.prj/reports/slices/maintenance-and-recovery/slice-3-project-registry-repair.md` | Pending. |
| Slice 4: Execution Artifact Drift | Planned | `feat/slice-4-execution-artifact-drift` | `feat(slice-4): execution artifact drift` | `.prj/reports/slices/maintenance-and-recovery/slice-4-execution-artifact-drift.md` | Pending. |
| Slice 5: Runtime Artifact Cleanup Preview | Planned | `feat/slice-5-runtime-artifact-cleanup-preview` | `feat(slice-5): runtime artifact cleanup preview` | `.prj/reports/slices/maintenance-and-recovery/slice-5-runtime-artifact-cleanup-preview.md` | Pending. |
| Slice 6: Diagnostic Bundle | Planned | `feat/slice-6-diagnostic-bundle` | `feat(slice-6): diagnostic bundle` | `.prj/reports/slices/maintenance-and-recovery/slice-6-diagnostic-bundle.md` | Pending. |
| Slice 7: Maintenance Documentation | Planned | `feat/slice-7-maintenance-docs` | `feat(slice-7): maintenance docs` | `.prj/reports/slices/maintenance-and-recovery/slice-7-maintenance-docs.md` | Pending. |

## Slice 1: Database Migration Status

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-db-migration-status`
- Expected PR title: `feat(slice-1): db migration status`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-1-db-migration-status.md`

Candidate tasks:

- [x] Add migration status model.
- [x] Compare migration assets with `schema_migrations`.
- [x] Add `chc maintenance db status`.
- [x] Add focused tests.
- [x] Document the command.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `9637c96`
- T1.2: `9637c96`
- T1.3: `9637c96`
- T1.4: `9637c96`
- T1.5: `827bfc7`
- T1.6: `docs(slice-1): record db migration status`
- T1.7: `make verify`

## Slice 2: Project Registry Diagnostics

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-project-registry-diagnostics`
- Expected PR title: `feat(slice-2): project registry diagnostics`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-2-project-registry-diagnostics.md`

Candidate tasks:

- [ ] List registry diagnostics for registered projects.
- [ ] Include path existence, `.chc/` presence, enabled state, and active marker.
- [ ] Keep diagnostics read-only.
- [ ] Add focused tests.
- [ ] Document registry diagnostics.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1: pending
- T2.2: pending
- T2.3: pending
- T2.4: pending
- T2.5: pending
- T2.6: pending
- T2.7: pending

## Slice 3: Project Registry Repair

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-project-registry-repair`
- Expected PR title: `feat(slice-3): project registry repair`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-3-project-registry-repair.md`

Candidate tasks:

- [ ] Add explicit registry repair operations.
- [ ] Support disabling stale project registrations.
- [ ] Support updating registered project paths.
- [ ] Add dry-run output.
- [ ] Add focused tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending
- T3.2: pending
- T3.3: pending
- T3.4: pending
- T3.5: pending
- T3.6: pending
- T3.7: pending

## Slice 4: Execution Artifact Drift

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-execution-artifact-drift`
- Expected PR title: `feat(slice-4): execution artifact drift`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-4-execution-artifact-drift.md`

Candidate tasks:

- [ ] Inspect execution records for artifact references.
- [ ] Report missing run directories and execution artifacts.
- [ ] Keep the command read-only.
- [ ] Add focused tests.
- [ ] Document execution artifact drift.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending
- T4.2: pending
- T4.3: pending
- T4.4: pending
- T4.5: pending
- T4.6: pending
- T4.7: pending

## Slice 5: Runtime Artifact Cleanup Preview

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-runtime-artifact-cleanup-preview`
- Expected PR title: `feat(slice-5): runtime artifact cleanup preview`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-5-runtime-artifact-cleanup-preview.md`

Candidate tasks:

- [ ] Identify unreferenced runtime artifact directories.
- [ ] Print cleanup preview output.
- [ ] Avoid deleting by default.
- [ ] Add focused tests.
- [ ] Document cleanup safety boundaries.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending
- T5.2: pending
- T5.3: pending
- T5.4: pending
- T5.5: pending
- T5.6: pending
- T5.7: pending

## Slice 6: Diagnostic Bundle

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-diagnostic-bundle`
- Expected PR title: `feat(slice-6): diagnostic bundle`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-6-diagnostic-bundle.md`

Candidate tasks:

- [ ] Gather command-readable local state summaries.
- [ ] Include selected reports and maintenance diagnostics.
- [ ] Avoid copying secrets or credentials.
- [ ] Write deterministic Markdown bundle under `.chc/maintenance/`.
- [ ] Add focused tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending
- T6.2: pending
- T6.3: pending
- T6.4: pending
- T6.5: pending
- T6.6: pending
- T6.7: pending

## Slice 7: Maintenance Documentation

Pre-flight:

- Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
- Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
- Slicing: `Maintenance and Recovery`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-maintenance-docs`
- Expected PR title: `feat(slice-7): maintenance docs`
- Expected report: `.prj/reports/slices/maintenance-and-recovery/slice-7-maintenance-docs.md`

Candidate tasks:

- [ ] Add how-to documentation for inspecting and recovering local state.
- [ ] Update command reference.
- [ ] Link related index, runtime, report, and registry references.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending

## Validation Record

No slices delivered yet.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- Existing `index rebuild` and `index status` commands already cover ticket
  index repair and parse error listing; this set builds around them instead of
  duplicating them.
