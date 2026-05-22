# UX Operation Surface Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
Ticket: `.prj/tickets/open/20260522071859-align-cli-authoring-and-tui-operation-surface.md`
Feedback draft: `.prj/drafts/chc-cli-output-polish-notes.md`

## Slicing Checklist

- [x] Capture lab feedback in a durable draft.
- [x] Commit captured feedback to `dev`.
- [x] Create ticket for the aligned UX operation surface.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: CLI Output Readability | Delivered | `feat/slice-1-cli-output-readability` | `feat(slice-1): cli output readability` | `.prj/reports/slices/ux-operation-surface/slice-1-cli-output-readability.md` | PR #134 merged. |
| Slice 2: Workspace Bootstrap Contract Cleanup | Delivered | `feat/slice-2-workspace-bootstrap-contract` | `feat(slice-2): workspace bootstrap contract cleanup` | `.prj/reports/slices/ux-operation-surface/slice-2-workspace-bootstrap-contract.md` | PR #135 merged. |
| Slice 3: Protocol Document Boundary Cleanup | Reviewing | `feat/slice-3-protocol-doc-boundaries` | `feat(slice-3): protocol doc boundaries` | `.prj/reports/slices/ux-operation-surface/slice-3-protocol-doc-boundaries.md` | PR pending review. |
| Slice 4: Authoring Session Ergonomics | Planned | `feat/slice-4-authoring-session-ergonomics` | `feat(slice-4): authoring session ergonomics` | `.prj/reports/slices/ux-operation-surface/slice-4-authoring-session-ergonomics.md` | IDs, slugs, list/show UX. |
| Slice 5: Context-Aware Quiz Generation Foundation | Planned | `feat/slice-5-context-aware-quiz-generation` | `feat(slice-5): context-aware quiz generation` | `.prj/reports/slices/ux-operation-surface/slice-5-context-aware-quiz-generation.md` | Quiz engine boundary. |
| Slice 6: TUI Specification Set | Delivered | `docs/slice-6-tui-specification-set` | `docs(slice-6): tui specification set` | `.prj/reports/slices/ux-operation-surface/slice-6-tui-specification-set.md` | Spec-only slice delivered directly on `dev`. |
| Slice 7: TUI Read-Only Shell | Planned | `feat/slice-7-tui-read-only-shell` | `feat(slice-7): tui read-only shell` | `.prj/reports/slices/ux-operation-surface/slice-7-tui-read-only-shell.md` | First Bubble Tea implementation. |
| Slice 8: TUI Action Wiring | Planned | `feat/slice-8-tui-action-wiring` | `feat(slice-8): tui action wiring` | `.prj/reports/slices/ux-operation-surface/slice-8-tui-action-wiring.md` | Safe selected-record actions. |

## Slice 1: CLI Output Readability

Pre-flight:

- Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
- Tracker: `.prj/trackers/slicing-ux-operation-surface-tracker.md`
- Slicing: `UX Operation Surface`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-cli-output-readability`
- Expected PR title: `feat(slice-1): cli output readability`
- Expected report: `.prj/reports/slices/ux-operation-surface/slice-1-cli-output-readability.md`

Candidate tasks:

- [x] Add or consolidate human-output formatting helpers.
- [x] Render runtime validation checks without duplicated `ok`.
- [x] Render authoring session and quiz creation with compact defaults.
- [x] Add headers to list outputs that currently rely on positional fields.
- [x] Prefer project-relative paths for human output.
- [x] Add focused CLI tests.
- [x] Update affected docs/runbook.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `fa8bfc9` `feat(cli): add human output helpers`; `ded68ac` `refactor(cli): split runtime check output`
- T1.2: `a07a94c` `feat(cli): compact runtime validation output`; `ded68ac` `refactor(cli): split runtime check output`
- T1.3: `d71a917` `feat(cli): improve authoring command output`
- T1.4: `d71a917` `feat(cli): improve authoring command output`
- T1.5: `a07a94c` `feat(cli): compact runtime validation output`; `d71a917` `feat(cli): improve authoring command output`
- T1.6: `a07a94c` `feat(cli): compact runtime validation output`; `d71a917` `feat(cli): improve authoring command output`
- T1.7: `03eb554` `docs(cli): describe readable command output`
- T1.8: pending report commit
- T1.9: `make verify` pass

## Slice 2: Workspace Bootstrap Contract Cleanup

Candidate tasks:

- [x] Inventory generated `.chc/config.yml` fields.
- [x] Decide invariant versus real user decision for each field.
- [x] Simplify generated config output.
- [x] Rewrite or remove `manifest.md` as config-like surface.
- [x] Preserve compatibility with existing managed projects.
- [x] Update docs and tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `cb694b1` `feat(init): simplify workspace bootstrap assets`
- T2.2: `cb694b1` `feat(init): simplify workspace bootstrap assets`
- T2.3: `cb694b1` `feat(init): simplify workspace bootstrap assets`
- T2.4: `cb694b1` `feat(init): simplify workspace bootstrap assets`
- T2.5: `cb694b1` `feat(init): simplify workspace bootstrap assets`
- T2.6: `06d15c9` `docs(init): clarify generated config boundary`
- T2.7: pending report commit
- T2.8: `make verify` pass

## Slice 3: Protocol Document Boundary Cleanup

Candidate tasks:

- [x] Review generated `.chc/agents/*.md` assets.
- [x] Keep project documentation guidance separate from `.chc/` operational
  contracts.
- [x] Move artifact layout explanation into proper reference/workspace docs.
- [x] Adjust context composition if protocol candidates change.
- [x] Update generated asset tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `e432a82` `feat(init): separate protocol docs from workspace contracts`
- T3.2: `e432a82` `feat(init): separate protocol docs from workspace contracts`
- T3.3: `e432a82` `feat(init): separate protocol docs from workspace contracts`
- T3.4: no context candidate change required
- T3.5: `e432a82` `feat(init): separate protocol docs from workspace contracts`
- T3.6: pending report commit
- T3.7: `make verify` pass

## Slice 4: Authoring Session Ergonomics

Candidate tasks:

- [ ] Evaluate authoring session ID prefix naming.
- [ ] Add session list headers and updated-time display.
- [ ] Add safe slug/topic lookup where appropriate.
- [ ] Allow promote flows to resolve session from artifact metadata where safe.
- [ ] Preserve old `auth-...` session compatibility.
- [ ] Update docs and tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1:
- T4.2:
- T4.3:
- T4.4:
- T4.5:
- T4.6:
- T4.7:
- T4.8:

## Slice 5: Context-Aware Quiz Generation Foundation

Candidate tasks:

- [ ] Define quiz generation request/result models.
- [ ] Define provider interface and deterministic fake provider.
- [ ] Keep static template mode as explicit fallback.
- [ ] Add spec-target question coverage for useful spec generation.
- [ ] Add context inputs from `.chc/`, docs, tree, tickets, and Go inventory.
- [ ] Add output contract validation.
- [ ] Update docs and tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1:
- T5.2:
- T5.3:
- T5.4:
- T5.5:
- T5.6:
- T5.7:
- T5.8:
- T5.9:

## Slice 6: TUI Specification Set

Candidate tasks:

- [x] Create TUI product role and scope spec.
- [x] Create TUI navigation and layout spec.
- [x] Create TUI entity screens spec.
- [x] Create TUI authoring and quiz UX spec.
- [x] Create TUI ticket execution and evidence UX spec.
- [x] Create TUI action and safety model spec.
- [x] Cross-link specs from plan/tracker as needed.
- [x] Add slice implementation report.
- [x] Run documentation validation.

Task commits:

- T6.1: specs for product surface, navigation, records, authoring, execution,
  actions, and implementation boundary.
- T6.2: slice report and tracker update.

## Slice 7: TUI Read-Only Shell

Candidate tasks:

- [ ] Add Bubble Tea dependency after confirming repo dependency policy.
- [ ] Add `chc tui` command entrypoint.
- [ ] Implement project resolution and navigation model.
- [ ] Implement read-only lists for tickets, authoring sessions, reports, and
  runtime status.
- [ ] Implement detail panes for selected records.
- [ ] Add tests for state/model functions where practical.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1:
- T7.2:
- T7.3:
- T7.4:
- T7.5:
- T7.6:
- T7.7:
- T7.8:

## Slice 8: TUI Action Wiring

Candidate tasks:

- [ ] Wire runtime validation action.
- [ ] Wire ticket creation/status/comment actions.
- [ ] Wire ticket start dry-run action.
- [ ] Wire authoring session/quiz actions.
- [ ] Wire report/evidence display actions.
- [ ] Add confirmation model for mutating and long-running operations.
- [ ] Add tests where practical.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1:
- T8.2:
- T8.3:
- T8.4:
- T8.5:
- T8.6:
- T8.7:
- T8.8:
- T8.9:
