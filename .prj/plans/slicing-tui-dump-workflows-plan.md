# TUI Dump Workflows Slicing Plan

Status: active

## Purpose

Plan the TUI dump workflow so project-wide and use-case-scoped dumps can be
created, viewed, edited, and refreshed from the TUI without moving runtime
logic into the TUI layer.

Dumps are rough Markdown knowledge captures. A dump with no associated use case
belongs directly to the project. A dump with a case slug belongs to that use
case. Both kinds may grow over time and remain directly editable as files.

## Source

Primary sources:

- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.agents/SLICES.md`
- `docs/explanation/document-generation-model.md`
- `docs/reference/project-workspace.md`
- `docs/reference/authoring-sessions.md`
- `internal/chrapp` operation service boundary
- `internal/chrtui` Dumps, Project, Cases, command palette, and artifact action
  surfaces

Conversation decisions:

- Dump handling must go through the application service.
- The TUI must not own path construction, frontmatter rules, scope resolution,
  Markdown writes, document reload, or index refresh.
- The service may satisfy multiple interfaces while remaining one concrete
  implementation.
- Empty use-case association means project-wide dump.
- Non-empty use-case association means use-case-scoped dump.
- Dumps are Markdown documents and the primary update model is full-document
  editing. Append behavior may remain as service compatibility, but it is not
  the main user-facing editing model.
- Direct user edits to dump files remain valid and must be visible after
  refresh.

## Current State

The TUI currently exposes a Dumps view through the generic document projection.
It can list dump-family documents when they are loaded, but it does not provide
a dedicated dump workflow.

Important limitations:

- The TUI read model is generic `DocumentView`, not dump-specific.
- Dump creation is not available as a first-class TUI operation.
- Append and edit flows are not available from the Dumps screen.
- The TUI document loader still contains layout knowledge and currently scans
  `.prj/dumps` for dumps, while product documentation and authoring artifact
  layout describe `.chr/dumps` as the project-local generated artifact path.
- The existing artifact action executor is generic and does not expose
  dump-specific requests.
- Reload behavior after dump operations is not centralized around a dump service
  result.

## Target Model

The target model is service-centered:

- `internal/chrapp.service` owns dump operations.
- The service resolves project-wide vs use-case-scoped dump requests.
- The service owns canonical dump paths, frontmatter, filename generation,
  appends, saves, opens, and reload/index refresh behavior.
- The service returns dump read models and operation results to adapters.
- TUI adapters translate TUI requests into service calls.
- The TUI renders service-provided state and emits user intent only.
- CLI, TUI, and future Telegram/Email adapters can reuse the same dump logic.

The target TUI behavior is:

- `F3 Dumps` shows project-wide dumps and use-case-scoped dumps.
- When a case is selected, dump creation can default to that case.
- When no case is selected or the user chooses project scope, creation produces
  a project-wide dump.
- Editing opens the selected dump as a complete Markdown document through the
  service boundary and refreshes visible state when control returns.
- Refresh reflects direct filesystem edits.

## Service Boundary

The service should expose dump operations through a focused interface consumed
by TUI adapters. The concrete implementation remains the existing application
service.

Candidate request and result shapes:

```go
type DumpScope struct {
    CaseSlug string
}

type DumpOperationRequest struct {
    Path     string
    Title    string
    Scope    DumpScope
    Content  string
    OpenMode string
}

type DumpOperationResult struct {
    Message string
    Dumps   []DumpView
    Index   chr.WorkspaceIndex
}
```

The exact type names may change during implementation, but the ownership rule
must not: service owns operational logic; TUI owns presentation and input state.

## Canonical Layout Decision

The active product model describes generated authoring artifacts under
`.chr/`, including:

```text
.chr/dumps/
```

This slicing should align dump operations with that product contract. Any
legacy `.prj/dumps` handling should be explicit compatibility behavior or
removed from active dump workflows if tests and docs confirm it is no longer
needed.

## Slice Map

Slicing name: TUI Dump Workflows

Active slice number: Slice 7

Active slice short name: dump editing ux polish

Expected branch name: `feat/slice-7-dump-editing-ux-polish`

Expected PR title: `feat(slice-7): dump editing ux polish`

Expected report path:
`.prj/reports/slices/tui-dump-workflows/slice-7-dump-editing-ux-polish.md`

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
- Slice 6 full dump Markdown editing through PR #180.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | dump service contract | `feat/slice-1-dump-service-contract` | `feat(slice-1): dump service contract` | `.prj/reports/slices/tui-dump-workflows/slice-1-dump-service-contract.md` |
| Slice 2 | scoped dump read model | `feat/slice-2-scoped-dump-read-model` | `feat(slice-2): scoped dump read model` | `.prj/reports/slices/tui-dump-workflows/slice-2-scoped-dump-read-model.md` |
| Slice 3 | create dumps from tui | `feat/slice-3-create-dumps-from-tui` | `feat(slice-3): create dumps from tui` | `.prj/reports/slices/tui-dump-workflows/slice-3-create-dumps-from-tui.md` |
| Slice 4 | append and edit tui dumps | `feat/slice-4-append-edit-tui-dumps` | `feat(slice-4): append and edit tui dumps` | `.prj/reports/slices/tui-dump-workflows/slice-4-append-edit-tui-dumps.md` |
| Slice 5 | polish dump workflows | `feat/slice-5-polish-dump-workflows` | `feat(slice-5): polish dump workflows` | `.prj/reports/slices/tui-dump-workflows/slice-5-polish-dump-workflows.md` |
| Slice 6 | full dump markdown editing | `feat/slice-6-full-dump-markdown-editing` | `feat(slice-6): full dump markdown editing` | `.prj/reports/slices/tui-dump-workflows/slice-6-full-dump-markdown-editing.md` |
| Slice 7 | dump editing ux polish | `feat/slice-7-dump-editing-ux-polish` | `feat(slice-7): dump editing ux polish` | `.prj/reports/slices/tui-dump-workflows/slice-7-dump-editing-ux-polish.md` |

## Slice 1: Dump Service Contract

Add the service-owned dump operation boundary and align dump filesystem layout
ownership.

Expected behavior:

- Service exposes list/create/update-oriented dump operation contracts.
- TUI adapter can depend on a dump executor interface without implementing dump
  logic.
- Canonical dump layout is resolved outside the TUI.
- Existing behavior remains compatible while the new path is introduced.

Out of scope:

- Full TUI create flow.
- Full append/edit interaction.
- Command palette polish.

## Slice 2: Scoped Dump Read Model

Add a dump-specific read model that distinguishes project-wide dumps from
use-case-scoped dumps.

Expected behavior:

- Dumps can be grouped or filtered by scope.
- The read model carries title, path, status, updated time, scope, and preview.
- The TUI Dumps screen can render this model without parsing Markdown.

Out of scope:

- Creating or editing dumps from the TUI.

## Slice 3: Create Dumps From TUI

Add the TUI creation workflow over the service boundary.

Expected behavior:

- User can start a create-dump action from Dumps and relevant case/project
  contexts.
- The action confirms scope before writing.
- The service writes the dump and returns refreshed dump state.
- Empty case slug creates a project-wide dump.
- Selected case slug creates a use-case-scoped dump.

Out of scope:

- Appending to existing dumps.
- External editor flow.

## Slice 4: Append And Edit TUI Dumps

Add growth and edit workflows for existing dumps.

Expected behavior:

- User can append content to the selected dump through a service request.
- User can open or edit the selected dump through a service-owned operation.
- The service preserves scope/frontmatter and refreshes state after changes.
- Direct filesystem edits are visible after refresh.

Out of scope:

- LLM-assisted dump curation.
- Promotion from dump to spec.

## Slice 5: Polish Dump Workflows

Finish navigation, commands, and user-facing consistency.

Expected behavior:

- Command palette exposes dump operations consistently.
- Empty states distinguish no project dumps, no use-case dumps, and no selected
  case where relevant.
- Status messages report created, appended, edited, opened, refreshed, and
  failed operations clearly.
- Tests cover the main TUI flows and service-adapter boundaries.

Out of scope:

- Quiz workflow changes.
- Spec promotion workflows.
- Semantic digest generation changes.

## Slice 6: Full Dump Markdown Editing

Convert the visible dump update flow from append-oriented behavior to
full-document Markdown editing.

Expected behavior:

- `Edit dump` is the primary update action for the selected dump.
- The TUI resolves the selected dump and scope through the service boundary.
- The service opens the complete dump Markdown file through an editor launcher
  abstraction owned outside the TUI presentation layer.
- Returning from the editor refreshes dump state and preserves the selected
  dump when possible.
- User-facing commands and status messages describe editing, not appending, as
  the main flow.

Out of scope:

- Adding an embedded TUI text editor.
- LLM-assisted dump rewriting.
- Spec promotion workflows.
- Removing service-level append compatibility if existing tests or callers
  still rely on it.

## Slice 7: Dump Editing UX Polish

Polish the dump editing surface now that full-document editing exists.

Expected behavior:

- The dump detail panel clearly distinguishes read mode from edit mode.
- Internal edit mode uses the available panel width and height instead of fixed
  editor sizing.
- Edit status, save hint, cancel hint, and external-editor hint are visible in
  the detail panel.
- The selected and focused dump remains visually consistent after edit, save,
  cancel, external edit, and refresh.
- Tests cover the edited detail rendering and keyboard behavior.

Out of scope:

- Rich Markdown syntax highlighting.
- Multi-file dump editing.
- LLM-assisted dump rewriting.

## Validation Strategy

Each slice must run focused tests for touched packages and `make verify` before
opening its PR unless a documented blocker prevents it.

Expected recurring checks:

```sh
go test ./internal/chr ./internal/chrapp ./internal/chrtui
make verify
```

## Out Of Scope

This slicing does not include:

- implementing LLM-assisted dump curation;
- promoting dumps to specs;
- adding an embedded TUI text editor;
- adding Telegram or Email adapters;
- changing unrelated quiz/spec/plan workflows;
- changing the Markdown-first durable source-of-truth model;
- making `.chr/tmp/` cache files authoritative.
