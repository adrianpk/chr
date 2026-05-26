# CLI TUI Surface Parity Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`

## Active Slice

Slicing name: CLI TUI Surface Parity

Active slice number: Slice 2

Active slice short name: service boundaries

Expected branch name: `feat/slice-2-service-boundaries`

Expected PR title: `feat(slice-2): service boundaries`

Expected report path:
`.prj/reports/slices/cli-tui-surface-parity/slice-2-service-boundaries.md`

Prerequisite PRs or delivered work:

- TUI operational console.
- TUI dump, quiz, spec, plan, board, and work action surfaces.
- Plan tracking and ticket materialization from the Plans tab.
- Current documentation refresh for TUI and command surfaces.

## Slicing Checklist

- [x] Capture CLI/TUI parity objective.
- [x] Define complete ordered slice map.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Inventory | Delivered | `feat/slice-1-cli-tui-inventory` | `feat(slice-1): inventory` | `.prj/reports/slices/cli-tui-surface-parity/slice-1-inventory.md` | PR `#204`; matrix created; `make verify` passed. |
| Slice 2: Service Boundaries | Planned | `feat/slice-2-service-boundaries` | `feat(slice-2): service boundaries` | `.prj/reports/slices/cli-tui-surface-parity/slice-2-service-boundaries.md` | Align shared service paths for existing operations. |
| Slice 3: Missing CLI Actions | Planned | `feat/slice-3-missing-cli-actions` | `feat(slice-3): missing cli actions` | `.prj/reports/slices/cli-tui-surface-parity/slice-3-missing-cli-actions.md` | Add CLI entrypoints for operational TUI actions. |
| Slice 4: Parity Tests | Planned | `feat/slice-4-parity-tests` | `feat(slice-4): parity tests` | `.prj/reports/slices/cli-tui-surface-parity/slice-4-parity-tests.md` | Add tests for CLI/TUI/service parity. |
| Slice 5: Parity Docs | Planned | `feat/slice-5-parity-docs` | `feat(slice-5): parity docs` | `.prj/reports/slices/cli-tui-surface-parity/slice-5-parity-docs.md` | Publish final parity contract. |

## Slice 1: Inventory

Candidate tasks:

- [x] T1.1 Audit TUI direct keybindings by view.
- [x] T1.2 Audit command palette commands and availability gates.
- [x] T1.3 Audit current CLI commands by domain.
- [x] T1.4 Map each action to service or adapter paths.
- [x] T1.5 Create `docs/reference/cli-tui-action-matrix.md`.
- [x] T1.6 Create slice implementation report.
- [x] T1.7 Run validation.

Task commits:

- `f8e1c9c docs(cli-tui): inventory direct tui bindings`
- `4f1a90e docs(cli-tui): inventory command palette actions`
- `0a40856 docs(cli-tui): inventory cli commands`
- `5c5ce0b docs(cli-tui): map service parity gaps`
- `6c9e769 docs(cli-tui): link parity matrix`
- `pending commit`

Validation:

- Pending.

## Slice 2: Service Boundaries

Candidate tasks:

- [ ] T2.1 Identify matched actions with divergent implementation paths.
- [ ] T2.2 Move duplicated logic into shared service methods where needed.
- [ ] T2.3 Preserve current CLI and TUI behavior.
- [ ] T2.4 Update matrix statuses for resolved divergences.
- [ ] T2.5 Add slice implementation report.
- [ ] T2.6 Run validation.

Task commits:

- `f8e1c9c docs(cli-tui): inventory direct tui bindings`
- `4f1a90e docs(cli-tui): inventory command palette actions`
- `0a40856 docs(cli-tui): inventory cli commands`
- `5c5ce0b docs(cli-tui): map service parity gaps`
- `6c9e769 docs(cli-tui): link parity matrix`
- `36dc004 docs(cli-tui): report inventory slice`
- `pending commit`

Validation:

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 3: Missing CLI Actions

Candidate tasks:

- [ ] T3.1 Add missing project and use-case metadata commands.
- [ ] T3.2 Add missing dump commands.
- [ ] T3.3 Add missing quiz commands.
- [ ] T3.4 Add missing spec, plan, and tracking commands.
- [ ] T3.5 Add missing ticket and work commands where applicable.
- [ ] T3.6 Update matrix statuses for implemented CLI commands.
- [ ] T3.7 Add slice implementation report.
- [ ] T3.8 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 4: Parity Tests

Candidate tasks:

- [ ] T4.1 Add CLI tests for new commands.
- [ ] T4.2 Add service tests for shared logic.
- [ ] T4.3 Add TUI adapter dispatch tests for critical actions.
- [ ] T4.4 Update matrix with test evidence.
- [ ] T4.5 Add slice implementation report.
- [ ] T4.6 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 5: Parity Docs

Candidate tasks:

- [ ] T5.1 Update the user guide.
- [ ] T5.2 Update command reference.
- [ ] T5.3 Update TUI reference.
- [ ] T5.4 Update architecture explanation.
- [ ] T5.5 Finalize matrix statuses and remaining gap notes.
- [ ] T5.6 Add slice implementation report.
- [ ] T5.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.
