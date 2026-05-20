# Slice 4: Tree and Search Context

Status: delivered
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-4-context-tree-search`
PR: #81

## Purpose

Add cheap repository orientation before semantic Go graph context is available.

## Delivered Behavior

The repository now has a tree/search context loader that can:

- capture a bounded, deterministic filesystem tree;
- skip noisy directories such as `.git`, `node_modules`, `vendor`,
  `_workspace`, and `.chc/runtime/runs`;
- derive search terms from ticket key, title, tags, use case, project area, and
  validation command;
- run bounded text search over source, tests, docs, and `.chc` files;
- render tree and search results into `code-context.md` content;
- record filesystem tree and text search source provenance.

## Implementation Notes

The search is intentionally simple and deterministic. It is not a semantic code
analysis layer and does not replace the planned `gograph` provider.

The loader returns in-memory context and rendered Markdown. It does not write
runtime artifacts yet; Docker-compatible materialization remains planned for
Slice 7.

## Contracts Added or Changed

New internal contracts:

- `TreeSearchContextOptions`
- `TreeSearchContext`
- `ContextTreeEntry`
- `ContextSearchHit`
- `LoadTreeSearchContext`
- `DeriveContextSearchTerms`
- `RenderCodeContext`

Source provenance uses the existing source inventory model from Slice 1.

## Files of Interest

- `internal/chc/context_tree_search.go`
- `internal/chc/context_tree_search_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestLoadTreeSearchContext|TestRenderCodeContext'`: pass
- `make verify`: pass

## Risks and Follow-ups

Text search is deliberately shallow. Symbol-aware selection, callers, callees,
tests, dependencies, and hotspots remain planned for the gograph slice.
