# CLI TUI Surface Parity Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`

## Active Slice

Slicing name: CLI TUI Surface Parity

Active slice number: Slice 5

Active slice short name: parity docs

Expected branch name: `feat/slice-5-parity-docs`

Expected PR title: `feat(slice-5): parity docs`

Expected report path:
`.prj/reports/slices/cli-tui-surface-parity/slice-5-parity-docs.md`

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
| Slice 2: Service Boundaries | Delivered | `feat/slice-2-service-boundaries` | `feat(slice-2): service boundaries` | `.prj/reports/slices/cli-tui-surface-parity/slice-2-service-boundaries.md` | PR `#205`; shared daemon stop and authoring session boundaries; `make verify` passed. |
| Slice 3: Missing CLI Actions | Delivered | `feat/slice-3-missing-cli-actions` | `feat(slice-3): missing cli actions` | `.prj/reports/slices/cli-tui-surface-parity/slice-3-missing-cli-actions.md` | PR `#206`; added CLI entrypoints for service-backed operational TUI actions; `make verify` passed. |
| Slice 4: Parity Tests | Delivered | `feat/slice-4-parity-tests` | `feat(slice-4): parity tests` | `.prj/reports/slices/cli-tui-surface-parity/slice-4-parity-tests.md` | PR `#207`; added CLI command and TUI adapter parity tests; `make verify` passed. |
| Slice 5: Parity Docs | Reviewing | `feat/slice-5-parity-docs` | `feat(slice-5): parity docs` | `.prj/reports/slices/cli-tui-surface-parity/slice-5-parity-docs.md` | Published final CLI/TUI parity documentation; `make verify` passed. |

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

- [x] T2.1 Identify matched actions with divergent implementation paths.
- [x] T2.2 Move duplicated logic into shared service methods where needed.
- [x] T2.3 Preserve current CLI and TUI behavior.
- [x] T2.4 Update matrix statuses for resolved divergences.
- [x] T2.5 Add slice implementation report.
- [x] T2.6 Run validation.

Task commits:

- `5ef3814 docs(cli-tui): correct matched service paths`
- `17fedd1 refactor(daemon): share stop request boundary`
- `93efc94 refactor(authoring): share session creation service`
- `89535d6 docs(cli-tui): record resolved service boundaries`
- `bcb1a72 chore(cli-tui): fix service boundary lint`
- `pending commit`

Validation:

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 3: Missing CLI Actions

Candidate tasks:

- [x] T3.1 Add missing project and use-case metadata commands.
- [x] T3.2 Add missing dump commands.
- [x] T3.3 Add missing quiz commands.
- [x] T3.4 Add missing spec, plan, and tracking commands.
- [x] T3.5 Add missing ticket and work commands where applicable.
- [x] T3.6 Update matrix statuses for implemented CLI commands.
- [x] T3.7 Add slice implementation report.
- [x] T3.8 Run validation.

Task commits:

- `a67df45 feat(cli): add tui parity commands`
- `9e2f4ac test(cli): cover parity command entrypoints`
- `ca1c70b docs(parity): mark implemented cli actions`
- `pending commit`

Validation:

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 4: Parity Tests

Candidate tasks:

- [x] T4.1 Add CLI tests for new commands.
- [x] T4.2 Add service tests for shared logic.
- [x] T4.3 Add TUI adapter dispatch tests for critical actions.
- [x] T4.4 Update matrix with test evidence.
- [x] T4.5 Add slice implementation report.
- [x] T4.6 Run validation.

Task commits:

- `41f45c6 test(cli): cover parity command paths`
- `71e9d90 test(tui): cover service adapter parity`
- `fa9f2dc docs(parity): record test evidence`
- `pending commit`

Validation:

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 5: Parity Docs

Candidate tasks:

- [x] T5.1 Update the user guide.
- [x] T5.2 Update command reference.
- [x] T5.3 Update TUI reference.
- [x] T5.4 Update architecture explanation.
- [x] T5.5 Finalize matrix statuses and remaining gap notes.
- [x] T5.6 Add slice implementation report.
- [x] T5.7 Run validation.

Task commits:

- `48af80a docs(user-guide): document parity commands`
- `a3ebb34 docs(commands): document parity command groups`
- `32179a7 docs(tui): link cli equivalents`
- `a573347 docs(architecture): explain operation surfaces`
- `e038eeb docs(parity): finalize gap notes`
- `pending commit`

Validation:

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.
