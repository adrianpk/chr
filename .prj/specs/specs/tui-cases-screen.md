# TUI Cases Screen

Status: draft

Sources:

- `docs/explanation/tui-interactive-design-notes.md`
- `.prj/specs/specs/tui-project-tab.md`
- `.prj/specs/specs/tui-product-surface.md`

## Purpose

Define the `Cases` screen for the ChrysoCoord TUI.

The `Cases` screen is where the user lists, creates, edits, searches, and
orients work around project-local use cases. A case is a lightweight container
for related project artifacts such as dumps, quizzes, specs, plans, and board
work.

The screen must answer two questions clearly:

> Which case am I working in?

> What compact context should ChrysoCoord keep for that case?

## User Model

The primary user is one terminal/editor-oriented local operator working in one
managed project. The user may interact through the TUI, the CLI, or direct
editing of files under `.chc/`.

Manual file editing is a supported workflow. The application must preserve the
same file/index model used by `Project`: human-editable Markdown files are the
authoring surface, and SQLite is the fast indexed operational view of the last
valid synchronized content.

## Product Rules

- The tab label is `Cases`.
- A case belongs to the current project by containment under the project's
  `.chc/` directory.
- Do not store owner fields. ChrysoCoord is a local single-user app in this
  model.
- Do not store `project_id` in case frontmatter for normal operation.
- The primary editable fields are:
  - `name`;
  - `description`;
  - long-form case context stored in the Markdown body.
- `name` is required before a new case can be saved.
- `description` is desirable but may be empty.
- Tags are not part of the first case schema. They may be reconsidered when
  there is a concrete retrieval or grouping workflow.
- Artifact counts are derived metadata, not manually edited case fields.
- Case status may exist below the UI surface when archive behavior exists, but
  the first editable surface should not add a status field without a workflow.

## Canonical Files

Cases are stored as Markdown with YAML frontmatter.

Proposed path pattern:

```text
.chc/cases/<case-slug>.md
```

The slug is stable file identity. It may be derived from the initial name, but
renaming the visible `name` should not force a file rename unless an explicit
rename workflow is implemented.

### File Format

```markdown
---
schema_version: 1
name: Blog publishing
description: Capture the authoring workflow for producing and validating blog posts.
---

Long-form case context.

This section can explain behavior, constraints, assumptions, examples, and
additional context that would make the description too long.
```

### Frontmatter Fields

`schema_version`

: Required integer. Defines the parser contract for the frontmatter schema. It
is not shown as an editable field in the TUI.

`name`

: Required string. Human-readable case name shown in the TUI and indexed in
SQLite.

`description`

: Optional string. Compact technical case description shown in the TUI and
indexed in SQLite.

### Markdown Body

The body is free-form Markdown.

The body carries extended case context. It is not a relational schema and should
not be parsed for structured columns.

## Persistence Model

ChrysoCoord has two representations of case data:

- `.chc/cases/<case-slug>.md` as the human-editable case file;
- SQLite as the fast query/index layer used by the UI, CLI, and context
  assembly.

The file is the authoring surface. SQLite is the indexed operational view of
the last valid synchronized content.

```plantuml
@startuml
title Case persistence model

rectangle ".chc/cases/<slug>.md\nMarkdown + YAML frontmatter" as File
component "Case parser\nschema validation" as Parser
database "SQLite case index" as DB
rectangle "Cases screen" as UI
rectangle "CLI case commands" as CLI
rectangle "Context assembly" as Context

File --> Parser
Parser --> DB
DB --> UI
DB --> CLI
DB --> Context

@enduml
```

## Write Model

TUI and CLI writes must go through a shared domain/application operation.

The TUI should not manually edit Markdown or write SQLite directly. The CLI
should not carry separate persistence logic for the same behavior.

### Save Interaction

```plantuml
@startuml
title Case save interaction

actor User
participant "Cases screen" as UI
participant "TUI model" as Model
participant "CaseExecutor" as Executor
participant "CaseService" as Service
database "SQLite case index" as DB
collections ".chc/cases/<slug>.md" as File

User -> UI: edit case fields inline
UI -> Model: update in-memory CaseMetadata
Model -> Model: mark case dirty

User -> UI: save case
UI -> Executor: SaveCase(metadata)
Executor -> Service: Save(metadata)

Service -> Service: validate required fields and schema
Service -> File: read current file hash

alt source hash changed outside CHC
    Service --> Executor: conflict error
    Executor --> UI: report resync required
else source hash matches or case is new
    Service -> File: write Markdown atomically
    Service --> Executor: saved case with new source_hash
    Executor -> DB: upsert indexed case
    DB --> Executor: indexed
    Executor --> UI: saved case
    UI -> Model: replace case and clear dirty state
end

@enduml
```

## Manual Edit And Resync Model

Manual file edits are supported, but they do not silently mutate the database.

```plantuml
@startuml
title Manual case resync

actor User
participant Editor
collections ".chc/cases/<slug>.md" as File
participant "Cases screen" as UI
participant "Case indexer" as Indexer
database "SQLite case index" as DB

User -> Editor: edit case file manually
Editor -> File: save
User -> UI: request resync
UI -> Indexer: reindex cases
Indexer -> File: read case files
Indexer -> Indexer: parse frontmatter and validate schema

alt valid case file
    Indexer -> DB: update indexed case and source hash
    Indexer --> UI: resync complete
else invalid case file
    Indexer -> DB: keep last valid indexed case
    Indexer -> DB: record diagnostics
    Indexer --> UI: show parse/schema error
end

@enduml
```

## Staleness And Conflict Rules

Every indexed case record should store enough source information to detect
external changes:

- `source_path`;
- `source_hash`;
- `indexed_at`;
- `schema_version`;
- parse or validation diagnostics when invalid.

Expected sync states:

- `synced`: the file and database index match.
- `stale`: the file changed since the last valid index.
- `invalid`: the file was read but frontmatter/schema validation failed.
- `conflict`: the app attempted to write while the file had changed outside the
  indexed version.

Invalid file content must not overwrite the last valid indexed case.

## UI Layout

The `Cases` screen uses the reusable editable-screen geometry shared with
`Project`.

```text
┌────────────────────────────────────┬───────────────────────────────────────┐
│ Cases                              │ Case                                  │
│ / filter cases                     │ Name                                  │
│                                    │ [ Blog publishing                   ] │
│ > Blog publishing                  │                                       │
│   CLI validation                   │ Description                           │
│   Runtime observation              │ [ Capture the authoring workflow... ] │
│                                    │                                       │
│                                    │ Context                               │
│                                    │ [ Long-form Markdown case context    │
│                                    │   remains editable inline here. ]    │
│                                    ├───────────────────────────────────────┤
│                                    │ Artifacts                             │
│                                    │ dumps           3                     │
│                                    │ quizzes         1                     │
│                                    │ specs           2                     │
│                                    │ plans           1                     │
│                                    │ board items      4                    │
│                                    │ Source                                │
│                                    │ .chc/cases/blog-publishing.md         │
└────────────────────────────────────┴───────────────────────────────────────┘

Project  Cases  Dumps  Quizzes  Specs  Plans  Board  Dash  Help  Quit
```

Layout rules:

- left: case list, selection, and filter query;
- right top: one editable document panel containing frontmatter fields and
  long-form context;
- right bottom: derived read-only metadata and diagnostics.

Derived metadata must stay in the same visual region across `Project`, `Cases`,
and later editable artifact screens unless a concrete workflow requires a
different layout.

## Display Fields

Primary editable fields:

- `Name`;
- `Description`;
- `Context`.

Primary navigation and search fields:

- case list;
- current filter query;
- selected case marker.

Secondary read-only fields:

- source path;
- dump count;
- quiz count;
- spec count;
- plan count;
- board item count;
- index diagnostics when present.

Excluded from normal display:

- owner;
- internal generated project ID;
- `project_id` frontmatter;
- schema version as an editable field;
- artifact counts stored as editable case fields.

## Interaction Model

Initial complete interaction:

- The primary navigation opens `Cases`.
- Arrow keys move through cases in the list when list focus is active.
- `/` opens text filtering over case `name` and `description`.
- `Esc` clears filter or cancels the current edit, depending on focus.
- `Enter` on an existing case opens or focuses editable detail.
- A new-case action creates an unsaved in-memory case and moves focus to
  `Name`.
- A new case cannot be saved until `Name` is non-empty.
- Field editing happens inline where the field is displayed.
- A save action persists through the shared case service.
- A resync action reindexes `.chc/cases/*.md` into SQLite.
- If a case file changed outside the app, save shows a conflict instead of
  overwriting.
- If a case file is invalid, the screen shows the last valid indexed values plus
  an actionable diagnostic.

Exact key bindings for save, resync, and new-case creation should be visible in
the bottom action area or contextual help before implementation.

## Artifact Relationship Rules

Artifacts may be project-wide or case-scoped.

- An artifact with no case association belongs to the project-wide set.
- An artifact associated with a case belongs to that case.
- Case artifact counts are derived from artifact indexes or queries.
- The case file must not maintain counters manually.

Initial artifact relationship direction:

```plantuml
@startuml
title Case artifact relationship

rectangle "Project" as Project
rectangle "Case" as Case
rectangle "Project-wide artifact\nno case association" as ProjectArtifact
rectangle "Case-scoped artifact\ncase association present" as CaseArtifact

Project --> Case
Project --> ProjectArtifact
Case --> CaseArtifact

@enduml
```

Preferred authoring progression remains:

```text
dump / quiz -> spec proposed -> spec accepted -> plan -> board
```

## Context Assembly Rules

Case metadata contributes to agent context as follows:

- `name` identifies the case compactly;
- `description` provides short technical orientation;
- body Markdown provides extended case context when the context budget allows;
- derived artifact counts are UI metadata and should not normally be sent as
  case context unless a context builder explicitly needs them.

## Bootstrap And Empty State

When no cases exist:

- the list should show an empty state, not fake rows;
- the detail region should explain that no case is selected through concise UI
  copy;
- the new-case action should create an unsaved editable case in memory;
- saving should create `.chc/cases/<slug>.md` through the shared domain service.

## Acceptance Criteria

- The screen lists existing cases.
- The user can search/filter cases by `name` and `description`.
- The user can create a new case.
- The user can edit case `name`, `description`, and long-form context inline.
- A new case cannot be saved without `name`.
- The canonical case file is Markdown with YAML frontmatter.
- SQLite indexes validated structured case metadata and supports fast queries.
- Manual file edits are supported through explicit resync.
- Invalid manual edits do not corrupt the SQLite index.
- CLI and TUI writes share one domain/application operation.
- External file changes are detected before CLI/TUI writes overwrite files.
- Derived metadata appears in the lower-right region, consistent with
  `Project`.
