# Slice 1: Markdown-Only Contract

Status: reviewing
Slicing: Markdown-Only Workspace Model
Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
Tracker: `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`
Branch: `feat/slice-1-markdown-only-contract`
PR: pending

## Purpose

Define the durable contract for moving Project and Cases away from duplicated
relational persistence.

This slice is intentionally documentation-only. It establishes the source of
truth and runtime graph model before implementation slices remove SQLite-backed
project/case indexes.

## Delivered Behavior

No runtime behavior changes are delivered in this slice.

The documented behavior is:

- `.chc/project.md` is the source of truth for Project.
- `.chc/cases/*.md` are the source of truth for Cases.
- SQLite is not authoritative for Project or Cases.
- Runtime navigation may use an ephemeral `WorkspaceGraph`.
- The graph stores lightweight metadata and relationships, not document bodies.
- Manual filesystem edits are valid and become visible after refresh.
- Missing case references are diagnostics, not implicit case creation.
- Context assembly should start from a selected case when available, or from
  project-wide material when no case applies.

## Implementation Notes

The Project and Cases specs now describe Markdown as the only durable
representation for those entities.

The previous file/index model is replaced with:

- Markdown file parsing;
- ephemeral graph metadata;
- refresh from filesystem;
- lazy body reads for detail and LLM context assembly.

The specs still allow runtime source hashes or modified times as guardrails for
active edit conflicts. Those values are process-local safety checks, not
persistent indexes.

## Contracts Added or Changed

- Project metadata contract changed from Markdown plus SQLite index to
  Markdown-only source of truth.
- Case metadata contract changed from Markdown plus SQLite index to
  Markdown-only source of truth.
- Manual resync language changed to refresh/rebuild-from-Markdown language.
- Case artifact relationship rules now include unresolved case references.
- Runtime graph terminology is introduced as the navigation and context
  selection model.

## Files of Interest

- `.prj/specs/specs/tui-project-tab.md`
- `.prj/specs/specs/tui-cases-screen.md`
- `docs/explanation/tui-interactive-design-notes.md`
- `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
- `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`

## Validation

- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Slice 2 must decide whether unreleased project/case index migrations are
  removed outright or preserved as inert compatibility migrations.
- Later slices must keep SQLite behavior for unrelated runtime domains such as
  tickets, executions, loops, notifications, and remote commands.
