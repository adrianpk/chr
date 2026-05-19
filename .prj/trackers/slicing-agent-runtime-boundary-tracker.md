# Agent Runtime Boundary Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`
Priority draft: `.prj/drafts/chc-command-surfaces-and-priority.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm `Automatic Reconciliation` is completed.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Runtime Config Model | Delivered | `feat/slice-1-runtime-config-model` | `feat(slice-1): runtime config model` | `.prj/reports/slices/agent-runtime-boundary/slice-1-runtime-config-model.md` | Repository PR #29. |
| Slice 2: Runtime Asset Template | Delivered | `feat/slice-2-runtime-asset-template` | `feat(slice-2): runtime asset template` | `.prj/reports/slices/agent-runtime-boundary/slice-2-runtime-asset-template.md` | Repository PR #30. |
| Slice 3: Runtime Init CLI | Delivered | `feat/slice-3-runtime-init-cli` | `feat(slice-3): runtime init cli` | `.prj/reports/slices/agent-runtime-boundary/slice-3-runtime-init-cli.md` | Repository PR #31. |
| Slice 4: Runtime Plan Contract | Reviewing | `feat/slice-4-runtime-plan-contract` | `feat(slice-4): runtime plan contract` | `.prj/reports/slices/agent-runtime-boundary/slice-4-runtime-plan-contract.md` | Compute Docker boundary paths and mounts. |
| Slice 5: Runtime Validate CLI | Planned | `feat/slice-5-runtime-validate-cli` | `feat(slice-5): runtime validate cli` | `.prj/reports/slices/agent-runtime-boundary/slice-5-runtime-validate-cli.md` | Validate runtime readiness without agent execution. |

## Slice 1: Runtime Config Model

Pre-flight:

- Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
- Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
- Slicing: `Agent Runtime Boundary`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-runtime-config-model`
- Expected PR title: `feat(slice-1): runtime config model`
- Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-1-runtime-config-model.md`

Candidate tasks:

- [x] Extend `WorkspaceConfig` with runtime fields.
- [x] Provide safe runtime config defaults.
- [x] Parse runtime config from `.chc/config.yml`.
- [x] Add focused config tests.
- [x] Document runtime config fields.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `a56e0e5` `feat(runtime): add workspace runtime config model`
- T1.2: `a56e0e5` `feat(runtime): add workspace runtime config model`
- T1.3: `a56e0e5` `feat(runtime): add workspace runtime config model`
- T1.4: `ca1b5e0` `test(runtime): cover workspace runtime config`
- T1.5: `26a3621` `docs(runtime): document workspace runtime config`
- T1.6: `6fcc8a8` `docs(runtime): report runtime config model slice`
- T1.7: `docs(runtime): record runtime config validation`

## Slice 2: Runtime Asset Template

Pre-flight:

- Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
- Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
- Slicing: `Agent Runtime Boundary`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-runtime-asset-template`
- Expected PR title: `feat(slice-2): runtime asset template`
- Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-2-runtime-asset-template.md`

Candidate tasks:

- [x] Add embedded agent-runtime Dockerfile template.
- [x] Add default runtime config fields to profiles.
- [x] Ensure init creates runtime template safely.
- [x] Add focused init tests.
- [x] Document runtime template ownership.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `a6bd87e` `feat(runtime): add agent runtime Dockerfile template`
- T2.2: `af3466d` `feat(runtime): add default runtime profile config`
- T2.3: `521e576` `test(runtime): cover runtime template init`
- T2.4: `521e576` `test(runtime): cover runtime template init`
- T2.5: `47f777e` `docs(runtime): document runtime template ownership`
- T2.6: `64164ef` `docs(runtime): report runtime asset template slice`
- T2.7: `docs(runtime): record runtime asset validation`

## Slice 3: Runtime Init CLI

Pre-flight:

- Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
- Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
- Slicing: `Agent Runtime Boundary`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-runtime-init-cli`
- Expected PR title: `feat(slice-3): runtime init cli`
- Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-3-runtime-init-cli.md`

Candidate tasks:

- [x] Add runtime command group.
- [x] Add `chc runtime init`.
- [x] Resolve project context through the existing resolver.
- [x] Support dry-run and skip-existing behavior.
- [x] Add focused CLI tests.
- [x] Document command behavior.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `b2bc397` `feat(cli): add runtime init command`
- T3.2: `b2bc397` `feat(cli): add runtime init command`
- T3.3: `b2bc397` `feat(cli): add runtime init command`
- T3.4: `aa549e0` `feat(runtime): add runtime asset initializer`
- T3.5: `f7b65a4` `test(runtime): cover runtime init command`
- T3.6: `5b5b2c3` `docs(runtime): document runtime init command`
- T3.7: `0d581a6` `docs(runtime): report runtime init cli slice`
- T3.8: `docs(runtime): record runtime init validation`

## Slice 4: Runtime Plan Contract

Pre-flight:

- Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
- Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
- Slicing: `Agent Runtime Boundary`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-runtime-plan-contract`
- Expected PR title: `feat(slice-4): runtime plan contract`
- Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-4-runtime-plan-contract.md`

Candidate tasks:

- [x] Add side-effect-free runtime planner.
- [x] Resolve project and workspace runtime paths.
- [x] Define Docker mount and workdir contract.
- [x] Define result artifact path contract.
- [x] Add focused planner tests.
- [x] Document runtime path semantics.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `00a109d` `feat(runtime): add runtime plan contract`
- T4.2: `00a109d` `feat(runtime): add runtime plan contract`
- T4.3: `00a109d` `feat(runtime): add runtime plan contract`
- T4.4: `00a109d` `feat(runtime): add runtime plan contract`
- T4.5: `5500682` `test(runtime): cover runtime plan contract`
- T4.6: `e863658` `docs(runtime): document runtime plan semantics`
- T4.7: `6e6a1cc` `docs(runtime): report runtime plan contract slice`
- T4.8: `docs(runtime): record runtime plan validation`

## Slice 5: Runtime Validate CLI

Pre-flight:

- Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
- Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
- Slicing: `Agent Runtime Boundary`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-runtime-validate-cli`
- Expected PR title: `feat(slice-5): runtime validate cli`
- Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-5-runtime-validate-cli.md`

Candidate tasks:

- [ ] Add `chc runtime validate`.
- [ ] Resolve project context and runtime config.
- [ ] Verify required runtime files and directories.
- [ ] Print computed Docker build/run plan.
- [ ] Support dry-run validation without Docker execution.
- [ ] Add focused CLI tests.
- [ ] Document validation behavior.
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
- T5.8: pending
- T5.9: pending

## Validation Record

- Planning validation: documentation diff reviewed.
- Slice 1 validation: `go test ./internal/chc` pass.
- Slice 1 validation: `make verify` pass.
- Slice 2 validation: `go test ./internal/chc` pass.
- Slice 2 validation: `make verify` pass.
- Slice 3 validation: `go test ./internal/chc ./cmd/chc` pass.
- Slice 3 validation: `make verify` pass.
- Slice 4 validation: `go test ./internal/chc` pass.
- Slice 4 validation: `make verify` pass.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- This slicing prepares a Docker runtime boundary but does not execute agents.
- Each implementation slice must include a Markdown implementation report before PR creation.
