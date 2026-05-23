# Markdown-Only Workspace Model Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`

## Slicing Checklist

- [x] Capture Markdown-only persistence direction.
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
| Slice 1: Markdown-Only Contract | Delivered | `feat/slice-1-markdown-only-contract` | `feat(slice-1): markdown-only contract` | `.prj/reports/slices/markdown-only-workspace-model/slice-1-markdown-only-contract.md` | PR #149. |
| Slice 2: Remove Project And Case Indexes | Active | `feat/slice-2-remove-project-case-indexes` | `feat(slice-2): remove project case indexes` | `.prj/reports/slices/markdown-only-workspace-model/slice-2-remove-project-case-indexes.md` | Remove relational persistence for Project/Cases only. |
| Slice 3: Workspace Graph Builder | Planned | `feat/slice-3-workspace-graph-builder` | `feat(slice-3): workspace graph builder` | `.prj/reports/slices/markdown-only-workspace-model/slice-3-workspace-graph-builder.md` | Runtime in-memory graph. |
| Slice 4: Filesystem Refresh Semantics | Planned | `feat/slice-4-filesystem-refresh-semantics` | `feat(slice-4): filesystem refresh semantics` | `.prj/reports/slices/markdown-only-workspace-model/slice-4-filesystem-refresh-semantics.md` | Manual editor/file changes. |
| Slice 5: TUI Project And Cases Rewire | Planned | `feat/slice-5-tui-project-cases-rewire` | `feat(slice-5): tui project cases rewire` | `.prj/reports/slices/markdown-only-workspace-model/slice-5-tui-project-cases-rewire.md` | Project/Cases screens use Markdown and graph. |
| Slice 6: Context Assembly From Graph | Planned | `feat/slice-6-context-assembly-from-graph` | `feat(slice-6): context assembly from graph` | `.prj/reports/slices/markdown-only-workspace-model/slice-6-context-assembly-from-graph.md` | LLM context selection. |

## Slice 1: Markdown-Only Contract

Candidate tasks:

- [x] Update Project spec to declare Markdown as source of truth.
- [x] Update Cases spec to declare Markdown as source of truth.
- [x] Define ephemeral workspace graph terminology.
- [x] Define case definition vs case association.
- [x] Define unresolved case reference behavior.
- [x] Define manual filesystem edit behavior.
- [x] Define case-rooted and project-wide context assembly direction.
- [x] Update interactive design notes.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.0: `bb47052` created and published slicing plan and tracker.
- T1.1: `b844ab7` updates Project and Cases specs.
- T1.2: `b844ab7` updates interactive design notes.
- T1.3: `2a47d8f` adds slice implementation report and tracker status.
- T1.4: `make verify` passed on 2026-05-23.

## Slice 2: Remove Project And Case Indexes

Candidate tasks:

- [x] Remove project metadata repository/index code.
- [x] Remove case metadata repository/index code.
- [x] Stop TUI executors from updating SQLite for Project/Cases.
- [x] Remove or neutralize project/case index migrations according to migration
      compatibility constraints.
- [x] Keep Markdown parse/render/write services.
- [x] Update tests for file-only Project/Cases persistence.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `5dd9c40` removes project/case SQLite repositories and migrations.
- T2.2: `5dd9c40` rewires TUI executors to Markdown-only refresh/save paths.
- T2.3: `5dd9c40` updates tests for seven remaining SQLite migrations.
- T2.4: pending commit adds slice implementation report and validation.

## Slice 3: Workspace Graph Builder

Candidate tasks:

- [ ] Add workspace graph model and builder.
- [ ] Scan `.chc/` Markdown files.
- [ ] Parse frontmatter into lightweight nodes.
- [ ] Classify supported document kinds.
- [ ] Build case-rooted views.
- [ ] Build project-wide views.
- [ ] Track unresolved case references.
- [ ] Track invalid frontmatter diagnostics.
- [ ] Add graph fixture tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending.
- T3.2: pending.
- T3.3: pending.
- T3.4: pending.

## Slice 4: Filesystem Refresh Semantics

Candidate tasks:

- [ ] Define refresh/reload behavior in code contracts.
- [ ] Implement graph rebuild on refresh.
- [ ] Update graph after app writes.
- [ ] Surface invalid Markdown/frontmatter diagnostics.
- [ ] Remove database resync terminology from runtime paths.
- [ ] Add runtime guardrails for active edits when source files changed.
- [ ] Add tests for external edits and refresh.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending.
- T4.2: pending.
- T4.3: pending.
- T4.4: pending.

## Slice 5: TUI Project And Cases Rewire

Candidate tasks:

- [ ] Load Project from Markdown service.
- [ ] Load Cases from workspace graph.
- [ ] Show case documents and associated artifacts from graph views.
- [ ] Show unresolved case references in Cases.
- [ ] Preserve inline editing for Project and Cases.
- [ ] Rebind refresh behavior to graph rebuild.
- [ ] Remove index/database wording from Project and Cases UI.
- [ ] Add TUI tests for graph-backed Project/Cases behavior.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending.
- T5.2: pending.
- T5.3: pending.
- T5.4: pending.

## Slice 6: Context Assembly From Graph

Candidate tasks:

- [ ] Add case-rooted context request model.
- [ ] Add project-wide context request model.
- [ ] Collect selected case document.
- [ ] Collect case-scoped dumps, quizzes, specs, and plans.
- [ ] Collect project-wide material when no case applies.
- [ ] Read document bodies lazily during context assembly.
- [ ] Report missing or unresolved references.
- [ ] Add context assembly tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending.
- T6.2: pending.
- T6.3: pending.
- T6.4: pending.

## Open Decisions

- Whether historical project/case index migrations can be removed outright or
  should remain as inert compatibility migrations.
- Exact frontmatter field names for artifact case association when existing
  artifacts do not already have a compatible field.
- Whether runtime source change guardrails use `mtime`, hash, or both.
- How much unresolved-reference detail belongs directly in the TUI versus a
  diagnostics view.

## Validation Log

- 2026-05-23: `make verify` passed.
- 2026-05-23: `make verify` passed for Slice 2.
