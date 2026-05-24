# TUI Dump Workflows Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`

## Active Slice

Slicing name: TUI Dump Workflows

Active slice number: Slice 3

Active slice short name: create dumps from tui

Expected branch name: `feat/slice-3-create-dumps-from-tui`

Expected PR title: `feat(slice-3): create dumps from tui`

Expected report path:
`.prj/reports/slices/tui-dump-workflows/slice-3-create-dumps-from-tui.md`

Prerequisite PRs or delivered work:

- Workspace index and semantic cache slicing through PR #172.
- Application operation service consolidation slicing through its delivered
  slices.
- TUI pane navigation and split behavior commit `d10056c`.
- Slice 1 dump service contract through PR #175.
- Slice 2 scoped dump read model through PR #176.

## Slicing Checklist

- [x] Capture dump workflow direction.
- [x] Capture service-boundary requirement.
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
| Slice 1: Dump Service Contract | Delivered | `feat/slice-1-dump-service-contract` | `feat(slice-1): dump service contract` | `.prj/reports/slices/tui-dump-workflows/slice-1-dump-service-contract.md` | PR #175. |
| Slice 2: Scoped Dump Read Model | Delivered | `feat/slice-2-scoped-dump-read-model` | `feat(slice-2): scoped dump read model` | `.prj/reports/slices/tui-dump-workflows/slice-2-scoped-dump-read-model.md` | PR #176. |
| Slice 3: Create Dumps From TUI | Reviewing | `feat/slice-3-create-dumps-from-tui` | `feat(slice-3): create dumps from tui` | `.prj/reports/slices/tui-dump-workflows/slice-3-create-dumps-from-tui.md` | PR #177. |
| Slice 4: Append And Edit TUI Dumps | Pending | `feat/slice-4-append-edit-tui-dumps` | `feat(slice-4): append and edit tui dumps` | `.prj/reports/slices/tui-dump-workflows/slice-4-append-edit-tui-dumps.md` | Dumps can grow and remain editable. |
| Slice 5: Polish Dump Workflows | Pending | `feat/slice-5-polish-dump-workflows` | `feat(slice-5): polish dump workflows` | `.prj/reports/slices/tui-dump-workflows/slice-5-polish-dump-workflows.md` | Commands, empty states, messages, and coverage. |

## Slice 1: Dump Service Contract

Candidate tasks:

- [x] Add service-owned dump operation contracts.
- [x] Add dump scope model where empty case slug means project-wide.
- [x] Move canonical dump path/layout decisions out of the TUI boundary.
- [x] Align active dump loading with the `.chc/dumps` product layout or record
      explicit compatibility handling.
- [x] Add focused tests for service-level dump scope and path behavior.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `311402f` adds dump service contracts and canonical dump directory.
- T1.2: `45b897a` aligns active dump document loading with `.chc/dumps`.
- T1.3: `66cdb9e` adds service scope tests, report, and validation.

## Slice 2: Scoped Dump Read Model

Candidate tasks:

- [x] Add a dump-specific read model carrying title, status, scope, path,
      updated time, and preview.
- [x] Add service query or adapter result for scoped dump lists.
- [x] Render project-wide and use-case-scoped dumps in the Dumps screen.
- [x] Keep the TUI from parsing Markdown or constructing dump paths.
- [x] Add TUI tests for project-wide and use-case-scoped dump display.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `4cc14c7` adds scoped dump read model, adapter mapping, TUI rendering, and tests.

## Slice 3: Create Dumps From TUI

Candidate tasks:

- [x] Add a create-dump TUI action and confirmation state.
- [x] Default creation scope from selected project/case context.
- [x] Execute creation through the dump service boundary.
- [x] Refresh dump state after creation.
- [x] Add tests for project-wide and use-case-scoped creation requests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `12095bf` adds dump creation service behavior, TUI action flow, adapter wiring, and tests.

## Slice 4: Append And Edit TUI Dumps

Candidate tasks:

- [ ] Add append request state for the selected dump.
- [ ] Execute append through the dump service boundary.
- [ ] Add open/edit operation for the selected dump.
- [ ] Refresh dump state after append/edit.
- [ ] Preserve existing scope and frontmatter during service writes.
- [ ] Add tests for append/edit requests and refresh behavior.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- Pending.

## Slice 5: Polish Dump Workflows

Candidate tasks:

- [ ] Add command palette entries for dump workflows.
- [ ] Add clear empty states for project-wide and case-scoped dumps.
- [ ] Add clear status messages for dump operations.
- [ ] Verify pane navigation and shortcuts remain consistent.
- [ ] Add integration-style TUI tests for the full dump workflow surface.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- Pending.

## Open Decisions

- Whether `.prj/dumps` should remain as a legacy read source or be removed from
  active dump loading.
- Exact shape of append input in the TUI before a full text editor exists.
- Whether edit/open should use `$EDITOR`, an existing command runner, or a
  service-owned abstraction that later supports non-terminal adapters.
- Whether dump preview should come from the structural index, lazy file body
  reads, or service-provided summaries.

## Completion Notes

- Pending.
