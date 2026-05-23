# Slice 1: Workspace Index Contract

Status: delivered
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-1-workspace-index-contract`
PR: #164

## Purpose

Define the workspace index, semantic cache, refresh, document association, and
draft removal contracts before runtime implementation starts.

## Delivered Behavior

- Documented the workspace index as a local graph/read-model over Markdown
  files.
- Documented structural index responsibilities and lazy body loading.
- Documented that dumps, quizzes, specs, and plans may be case-scoped by slug
  or project-wide when no case is declared.
- Documented discardable workspace cache behavior.
- Documented semantic digests as optional LLM-generated retrieval hints cached
  by source hash and related invalidation inputs.
- Documented that semantic digest generation must not block TUI startup.
- Documented that `Draft` is not a first-class document kind and draftness is a
  spec status.

## Implementation Notes

This slice is intentionally documentation-only. It builds on the delivered
Markdown-only workspace model and frames the next implementation slices as an
evolution of the existing `WorkspaceGraph`, not as a new durable database.

The TUI design notes now use `WorkspaceIndex` language for the richer
read-model direction while preserving the current Markdown source-of-truth
contract.

## Contracts Added or Changed

- `internal/chc` owns the workspace index model.
- Structural indexing is deterministic, local, and rebuildable from files.
- Semantic indexing is optional, cache-backed, and separate from structural
  indexing.
- Documents without a case slug are project-wide.
- Missing case slugs remain unresolved references.
- `Draft` is removed from the active first-class document model.

## Files of Interest

- `.agents/ARCHITECTURE.md`
- `docs/explanation/tui-interactive-design-notes.md`
- `docs/explanation/document-generation-model.md`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-1-workspace-index-contract.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Active command and reference docs still contain historical draft command
  examples. Runtime and documentation cleanup is intentionally deferred to Slice
  9.
