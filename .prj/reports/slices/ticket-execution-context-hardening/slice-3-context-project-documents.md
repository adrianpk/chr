# Slice 3: Project and Use-Case Document Context

Status: reviewing
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-3-context-project-documents`
PR: pending

## Purpose

Collect bounded project and ticket-related document context before later slices
add repository tree/search context, semantic Go graph context, Ralph attempt
context, and Docker-compatible materialization.

## Delivered Behavior

The repository now has a project/document context loader that can:

- summarize project directory, `.chc` workspace directory, selected ticket, and
  validation command;
- parse the selected ticket as the primary fresh document;
- follow explicit ticket references to `.prj/`, `docs/`, and `.chc/` files;
- render `project.md`;
- render `documents.md`;
- render source inventory rows for project, ticket, included documents, and
  missing referenced documents;
- explicitly exclude implicit conversation history from Ralph Loop context.

## Implementation Notes

Document loading is reference-driven and bounded. The loader does not scan the
whole repository for project documents. It includes only the selected ticket and
project document paths explicitly mentioned by that ticket.

References that do not exist are retained as skipped source inventory entries
with a missing-file reason. This keeps context readiness and review evidence
auditable without guessing.

## Contracts Added or Changed

New internal contracts:

- `ProjectDocumentContextOptions`
- `ProjectDocumentContext`
- `LoadProjectDocumentContext`
- `RenderProjectContext`
- `RenderDocumentsContext`
- `RenderSourceInventory`

The source inventory renderer is generic and can be reused by later slices when
additional context sources are added.

## Files of Interest

- `internal/chc/context_project_documents.go`
- `internal/chc/context_project_documents_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestLoadProjectDocumentContext|TestRenderSourceInventory'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice does not infer related documents through search or semantic analysis.
Only explicit ticket references are followed. Broader discovery belongs to the
tree/search and gograph slices.
