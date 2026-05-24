# TUI Dump Workflows Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`

## Active Slice

Slicing name: TUI Dump Workflows

Active slice number: Slice 6

Active slice short name: full dump markdown editing

Expected branch name: `feat/slice-6-full-dump-markdown-editing`

Expected PR title: `feat(slice-6): full dump markdown editing`

Expected report path:
`.prj/reports/slices/tui-dump-workflows/slice-6-full-dump-markdown-editing.md`

Prerequisite PRs or delivered work:

- Workspace index and semantic cache slicing through PR #172.
- Application operation service consolidation slicing through its delivered
  slices.
- TUI pane navigation and split behavior commit `d10056c`.
- Slice 1 dump service contract through PR #175.
- Slice 2 scoped dump read model through PR #176.
- Slice 3 create dumps from TUI through PR #177.
- Slice 4 append/edit dump workflow through PR #178.
- Slice 5 polish dump workflow through PR #179.

## Slicing Checklist

- [x] Capture dump workflow direction.
- [x] Capture service-boundary requirement.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [x] Validate completed slicing with `make verify`.
- [x] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Dump Service Contract | Delivered | `feat/slice-1-dump-service-contract` | `feat(slice-1): dump service contract` | `.prj/reports/slices/tui-dump-workflows/slice-1-dump-service-contract.md` | PR #175. |
| Slice 2: Scoped Dump Read Model | Delivered | `feat/slice-2-scoped-dump-read-model` | `feat(slice-2): scoped dump read model` | `.prj/reports/slices/tui-dump-workflows/slice-2-scoped-dump-read-model.md` | PR #176. |
| Slice 3: Create Dumps From TUI | Delivered | `feat/slice-3-create-dumps-from-tui` | `feat(slice-3): create dumps from tui` | `.prj/reports/slices/tui-dump-workflows/slice-3-create-dumps-from-tui.md` | PR #177. |
| Slice 4: Append And Edit TUI Dumps | Delivered | `feat/slice-4-append-edit-tui-dumps` | `feat(slice-4): append and edit tui dumps` | `.prj/reports/slices/tui-dump-workflows/slice-4-append-edit-tui-dumps.md` | PR #178. |
| Slice 5: Polish Dump Workflows | Delivered | `feat/slice-5-polish-dump-workflows` | `feat(slice-5): polish dump workflows` | `.prj/reports/slices/tui-dump-workflows/slice-5-polish-dump-workflows.md` | PR #179. |
| Slice 6: Full Dump Markdown Editing | Reviewing | `feat/slice-6-full-dump-markdown-editing` | `feat(slice-6): full dump markdown editing` | `.prj/reports/slices/tui-dump-workflows/slice-6-full-dump-markdown-editing.md` | Tuning slice. |

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

- [x] Add append request state for the selected dump.
- [x] Execute append through the dump service boundary.
- [x] Add open/edit operation for the selected dump.
- [x] Refresh dump state after append/edit.
- [x] Preserve existing scope and frontmatter during service writes.
- [x] Add tests for append/edit requests and refresh behavior.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `de4988e` adds append/open dump service behavior, TUI actions, adapter wiring, and tests.

## Slice 5: Polish Dump Workflows

Candidate tasks:

- [x] Add command palette entries for dump workflows.
- [x] Add clear empty states for project-wide and case-scoped dumps.
- [x] Add clear status messages for dump operations.
- [x] Verify pane navigation and shortcuts remain consistent.
- [x] Add integration-style TUI tests for the full dump workflow surface.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `138dbc5` polishes dump commands, empty state copy, edit naming, and workflow tests.

## Slice 6: Full Dump Markdown Editing

Candidate tasks:

- [x] Add service-owned editor launcher contract for dump Markdown editing.
- [x] Execute `Edit dump` through that service boundary and refresh dump state
      when the editor exits.
- [x] Preserve selected dump state after edit when the same dump still exists.
- [x] Remove or demote append as a user-facing TUI workflow where it conflicts
      with full-document editing.
- [x] Add tests for editor launch requests, refresh behavior, and selection
      preservation.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `d6e2b96` adds service-owned dump editor launch, TUI editor
  execution, refresh behavior, append demotion, and tests.
- T6.2: `7312fd6` review correction aligns `Edit dump` with internal Markdown
  editing and moves the external editor to `alt+e`.

Validation:

- `go test ./internal/chcapp ./internal/chctui`
- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `make verify`

## Open Decisions

- Whether `.prj/dumps` should remain as a legacy read source or be removed from
  active dump loading.
- None for this slice.
- Whether dump preview should come from the structural index, lazy file body
  reads, or service-provided summaries.

## Completion Notes

- Post-slice tuning needed: yes.
- Slice 6 covers conversion of the current path-resolution edit command into
  real full-Markdown editing through a service-owned editor or launcher
  abstraction.
- A second tuning pass should use captured screenshots to refine Dumps layout,
  grouping, labels, and selected/focused state after full editing is available.
