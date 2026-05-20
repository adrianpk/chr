# Daemon Coordination Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm prerequisite foreground loop slicing is completed.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Daemon Command Skeleton | Planned | `feat/slice-1-daemon-command-skeleton` | `feat(slice-1): daemon command skeleton` | `.prj/reports/slices/daemon-coordination/slice-1-daemon-command-skeleton.md` | Command surface only. |
| Slice 2: Daemon Lifecycle State Model | Planned | `feat/slice-2-daemon-lifecycle-state` | `feat(slice-2): daemon lifecycle state` | `.prj/reports/slices/daemon-coordination/slice-2-daemon-lifecycle-state.md` | State contract. |
| Slice 3: Daemon State Repository | Planned | `feat/slice-3-daemon-state-repository` | `feat(slice-3): daemon state repository` | `.prj/reports/slices/daemon-coordination/slice-3-daemon-state-repository.md` | User database persistence. |
| Slice 4: Daemon Status Reporting | Planned | `feat/slice-4-daemon-status-reporting` | `feat(slice-4): daemon status reporting` | `.prj/reports/slices/daemon-coordination/slice-4-daemon-status-reporting.md` | Read lifecycle state. |
| Slice 5: Single-Instance Guard | Planned | `feat/slice-5-daemon-single-instance` | `feat(slice-5): daemon single instance` | `.prj/reports/slices/daemon-coordination/slice-5-daemon-single-instance.md` | Prevent duplicate owners. |
| Slice 6: Registered Project Loading | Planned | `feat/slice-6-daemon-project-loading` | `feat(slice-6): daemon project loading` | `.prj/reports/slices/daemon-coordination/slice-6-daemon-project-loading.md` | Load registered projects. |
| Slice 7: Daemon Foreground Run Loop | Planned | `feat/slice-7-daemon-foreground-run-loop` | `feat(slice-7): daemon foreground run loop` | `.prj/reports/slices/daemon-coordination/slice-7-daemon-foreground-run-loop.md` | Supervised foreground daemon. |
| Slice 8: Graceful Stop and Stale Recovery | Planned | `feat/slice-8-daemon-stop-recovery` | `feat(slice-8): daemon stop recovery` | `.prj/reports/slices/daemon-coordination/slice-8-daemon-stop-recovery.md` | Stop and recovery. |

## Slice 1: Daemon Command Skeleton

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-daemon-command-skeleton`
- Expected PR title: `feat(slice-1): daemon command skeleton`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-1-daemon-command-skeleton.md`

Candidate tasks:

- [ ] Add `chc daemon` command group.
- [ ] Add `daemon status`, `daemon start --dry-run`, and `daemon stop --dry-run`.
- [ ] Resolve user environment for daemon commands.
- [ ] Keep daemon commands side-effect free in this slice.
- [ ] Add focused CLI tests.
- [ ] Document the daemon command boundary.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T1.1: pending
- T1.2: pending
- T1.3: pending
- T1.4: pending
- T1.5: pending
- T1.6: pending
- T1.7: pending
- T1.8: pending

## Slice 2: Daemon Lifecycle State Model

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-daemon-lifecycle-state`
- Expected PR title: `feat(slice-2): daemon lifecycle state`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-2-daemon-lifecycle-state.md`

Candidate tasks:

- [ ] Add daemon state model.
- [ ] Track status, PID, started time, heartbeat time, and stop request time.
- [ ] Validate lifecycle transitions.
- [ ] Add focused model tests.
- [ ] Document lifecycle fields.
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

## Slice 3: Daemon State Repository

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-daemon-state-repository`
- Expected PR title: `feat(slice-3): daemon state repository`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-3-daemon-state-repository.md`

Candidate tasks:

- [ ] Add SQLite migration.
- [ ] Add daemon state repository.
- [ ] Support get, upsert, clear, and heartbeat updates.
- [ ] Add repository tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending
- T3.2: pending
- T3.3: pending
- T3.4: pending
- T3.5: pending
- T3.6: pending

## Slice 4: Daemon Status Reporting

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-daemon-status-reporting`
- Expected PR title: `feat(slice-4): daemon status reporting`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-4-daemon-status-reporting.md`

Candidate tasks:

- [ ] Read daemon state from the repository.
- [ ] Print stopped, running, stale, and stopping states.
- [ ] Include PID and heartbeat details when available.
- [ ] Add focused CLI tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending
- T4.2: pending
- T4.3: pending
- T4.4: pending
- T4.5: pending
- T4.6: pending

## Slice 5: Single-Instance Guard

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-daemon-single-instance`
- Expected PR title: `feat(slice-5): daemon single instance`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-5-daemon-single-instance.md`

Candidate tasks:

- [ ] Add a user-level lock or ownership guard.
- [ ] Detect existing running daemon state.
- [ ] Reject duplicate start attempts.
- [ ] Add stale-state classification hooks.
- [ ] Add focused guard tests.
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

## Slice 6: Registered Project Loading

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-daemon-project-loading`
- Expected PR title: `feat(slice-6): daemon project loading`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-6-daemon-project-loading.md`

Candidate tasks:

- [ ] List enabled registered projects.
- [ ] Validate project paths before daemon work.
- [ ] Classify missing or invalid project paths.
- [ ] Add focused project-loading tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending
- T6.2: pending
- T6.3: pending
- T6.4: pending
- T6.5: pending
- T6.6: pending

## Slice 7: Daemon Foreground Run Loop

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-daemon-foreground-run-loop`
- Expected PR title: `feat(slice-7): daemon foreground run loop`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-7-daemon-foreground-run-loop.md`

Candidate tasks:

- [ ] Add foreground `daemon start` execution path.
- [ ] Update heartbeat state while running.
- [ ] Periodically refresh registered projects.
- [ ] Invoke existing foreground-safe loop boundaries.
- [ ] Support a small testable iteration limit.
- [ ] Add focused run-loop tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending
- T7.6: pending
- T7.7: pending
- T7.8: pending

## Slice 8: Graceful Stop and Stale Recovery

Pre-flight:

- Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
- Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
- Slicing: `Daemon Coordination`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-daemon-stop-recovery`
- Expected PR title: `feat(slice-8): daemon stop recovery`
- Expected report: `.prj/reports/slices/daemon-coordination/slice-8-daemon-stop-recovery.md`

Candidate tasks:

- [ ] Persist stop requests.
- [ ] Make the run loop observe stop requests.
- [ ] Mark stopped daemon state on graceful exit.
- [ ] Classify stale process state.
- [ ] Expose recovery behavior in status output.
- [ ] Add focused stop and stale recovery tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1: pending
- T8.2: pending
- T8.3: pending
- T8.4: pending
- T8.5: pending
- T8.6: pending
- T8.7: pending
- T8.8: pending

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Notification and remote command surfaces remain outside this slicing.
