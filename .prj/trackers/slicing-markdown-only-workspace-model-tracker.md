# Markdown-Only Workspace Model Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`

## Slicing Checklist

- [x] Capture Markdown-only persistence direction.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [x] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Markdown-Only Contract | Delivered | `feat/slice-1-markdown-only-contract` | `feat(slice-1): markdown-only contract` | `.prj/reports/slices/markdown-only-workspace-model/slice-1-markdown-only-contract.md` | PR #149. |
| Slice 2: Remove Project And Case Indexes | Delivered | `feat/slice-2-remove-project-case-indexes` | `feat(slice-2): remove project case indexes` | `.prj/reports/slices/markdown-only-workspace-model/slice-2-remove-project-case-indexes.md` | PR #150. |
| Slice 3: Workspace Graph Builder | Delivered | `feat/slice-3-workspace-graph-builder` | `feat(slice-3): workspace graph builder` | `.prj/reports/slices/markdown-only-workspace-model/slice-3-workspace-graph-builder.md` | PR #151. |
| Slice 4: Filesystem Refresh Semantics | Delivered | `feat/slice-4-filesystem-refresh-semantics` | `feat(slice-4): filesystem refresh semantics` | `.prj/reports/slices/markdown-only-workspace-model/slice-4-filesystem-refresh-semantics.md` | PR #152. |
| Slice 5: TUI Project And Cases Rewire | Delivered | `feat/slice-5-tui-project-cases-rewire` | `feat(slice-5): tui project cases rewire` | `.prj/reports/slices/markdown-only-workspace-model/slice-5-tui-project-cases-rewire.md` | PR #153. |
| Slice 6: Context Assembly From Graph | Delivered | `feat/slice-6-context-assembly-from-graph` | `feat(slice-6): context assembly from graph` | `.prj/reports/slices/markdown-only-workspace-model/slice-6-context-assembly-from-graph.md` | PR #154. |

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
- T2.4: `4d307d6` adds slice implementation report and validation.

## Slice 3: Workspace Graph Builder

Candidate tasks:

- [x] Add workspace graph model and builder.
- [x] Scan `.chc/` Markdown files.
- [x] Parse frontmatter into lightweight nodes.
- [x] Classify supported document kinds.
- [x] Build case-rooted views.
- [x] Build project-wide views.
- [x] Track unresolved case references.
- [x] Track invalid frontmatter diagnostics.
- [x] Add graph fixture tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `0fa2ef0` adds workspace graph model and builder.
- T3.2: `0fa2ef0` adds `.chc/` Markdown scanning and document classification.
- T3.3: `0fa2ef0` adds case-rooted/project-wide views and diagnostics.
- T3.4: `86e47e8` adds slice implementation report and validation.

## Slice 4: Filesystem Refresh Semantics

Candidate tasks:

- [x] Define refresh/reload behavior in code contracts.
- [x] Implement graph rebuild on refresh.
- [x] Update graph after app writes.
- [x] Surface invalid Markdown/frontmatter diagnostics.
- [x] Remove database resync terminology from runtime paths.
- [x] Add runtime guardrails for active edits when source files changed.
- [x] Add tests for external edits and refresh.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `2bb78b2` adds workspace refresh executor contract.
- T4.2: `2bb78b2` rebuilds the runtime graph from filesystem refresh.
- T4.3: `2bb78b2` adds diagnostics/status reporting and edit guardrails.
- T4.4: `095e308` adds slice implementation report and validation.

## Slice 5: TUI Project And Cases Rewire

Candidate tasks:

- [x] Load Project from Markdown service.
- [x] Load Cases from workspace graph.
- [x] Show case documents and associated artifacts from graph views.
- [x] Show unresolved case references in Cases.
- [x] Preserve inline editing for Project and Cases.
- [x] Rebind refresh behavior to graph rebuild.
- [x] Remove index/database wording from Project and Cases UI.
- [x] Add TUI tests for graph-backed Project/Cases behavior.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `46e995b` adds graph-backed Project/Cases projection.
- T5.2: `46e995b` rebinds Project/Cases refresh and post-save graph refresh.
- T5.3: `46e995b` adds graph-backed TUI tests.
- T5.4: `46e995b` adds slice implementation report and validation.

## Slice 6: Context Assembly From Graph

Candidate tasks:

- [x] Add case-rooted context request model.
- [x] Add project-wide context request model.
- [x] Collect selected case document.
- [x] Collect case-scoped dumps, quizzes, specs, and plans.
- [x] Collect project-wide material when no case applies.
- [x] Read document bodies lazily during context assembly.
- [x] Report missing or unresolved references.
- [x] Add context assembly tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `9cd64b8` adds workspace context request and assembly model.
- T6.2: `9cd64b8` adds case-rooted and project-wide document selection.
- T6.3: `9cd64b8` adds lazy body reads and source reporting.
- T6.4: `9cd64b8` adds authoring quiz integration, tests, report, and validation.

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
- 2026-05-23: `go test ./internal/chc ./internal/chcapp ./internal/chctui ./cmd/chc` passed for Slice 3.
- 2026-05-23: `make verify` passed for Slice 3.
- 2026-05-23: `go test ./internal/chctui ./internal/chcapp ./internal/chc ./cmd/chc` passed for Slice 4.
- 2026-05-23: `make verify` passed for Slice 4.
- 2026-05-23: `go test ./internal/chctui ./internal/chcapp ./internal/chc ./cmd/chc` passed for Slice 5.
- 2026-05-23: `make verify` passed for Slice 5.
- 2026-05-23: `go test ./internal/chc ./internal/chcapp ./internal/chctui ./cmd/chc` passed for Slice 6.
- 2026-05-23: `make verify` passed for Slice 6.
