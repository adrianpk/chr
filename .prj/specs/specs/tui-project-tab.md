# TUI Project Tab

Status: draft

Sources:

- `docs/explanation/tui-interactive-design-notes.md`
- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`

## Purpose

Define the `Project` screen for the ChrysoCoord TUI.

The `Project` tab is the place where the user sees and edits the managed
project's own identity and project-level context. It is not an agent dashboard,
not an application branding surface, and not a workspace diagnostics screen.

The tab must answer one question clearly:

> What is this project, and what compact context should ChrysoCoord keep about
> it?

## User Model

The primary user is a terminal/editor-oriented local operator. They may use the
TUI, the CLI, or a text editor such as Emacs against files under `.chc/`.

Manual file editing is a first-class workflow. The application must not assume
that all changes originate from the TUI or CLI.

## Product Rules

- The tab label is `Project`.
- The visible project identity must not expose internal generated IDs such as
  `project-...`.
- The visible project identity must not use application branding as the primary
  content.
- The primary editable fields are:
  - `name`;
  - `description`;
  - long-form project context stored in the Markdown body.
- `name` is short and human-readable.
- `description` is short or medium-length and technically relevant. It should
  describe the project's main case, capabilities, or behavior.
- The Markdown body is for richer project context that would make
  `description` too long.
- Artifact counts may appear as secondary information.
- The project path may appear as secondary information.
- Workspace implementation details must not appear as normal content:
  - no `.prj` status;
  - no `.chc` status when the TUI is already running normally;
  - no labels such as `scope`, `signals`, or `local state`;
  - no tracker count until trackers have a resolved product role.

## Canonical File

Project metadata is stored as Markdown with YAML frontmatter.

Proposed path:

```text
.chc/project.md
```

The file is human-editable and belongs to the managed project workspace.

### File Format

```markdown
---
schema_version: 1
name: chc-blog-work
description: CLI/blog workspace used to validate CHC authoring flows.
---

Long-form project context.

This section can explain technical behavior, constraints, main workflows,
domain assumptions, and anything useful as extended project context.
```

### Frontmatter Fields

`schema_version`

: Required integer. Defines the parser contract for the frontmatter schema.

`name`

: Required string. Human-readable project name shown in the TUI and indexed in
SQLite.

`description`

: Required string. Compact technical project description shown in the TUI and
indexed in SQLite.

### Markdown Body

The body is free-form Markdown.

The body should be usable as extended project context, but it is not a
relational schema. The application may derive summaries or search material from
it, but must not depend on arbitrary body prose to populate structured columns.

## Persistence Model

ChrysoCoord has two representations of project data:

- `.chc/project.md` as the human-editable project file;
- SQLite as the fast query/index layer used by the UI, CLI, and context
  assembly.

The file is the authoring surface. SQLite is the indexed operational view of
the last valid synchronized content.

```plantuml
@startuml
title Project metadata persistence model

rectangle ".chc/project.md\nMarkdown + YAML frontmatter" as File
component "Project metadata parser\nschema validation" as Parser
database "SQLite project metadata index" as DB
rectangle "Project screen" as UI
rectangle "CLI project commands" as CLI
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

This diagram describes the logical interaction between UI state, application
model, domain persistence, the relational index, and the project metadata file.
It intentionally avoids key-level interaction details.

```plantuml
@startuml
title Project metadata save interaction

actor User
participant "Project screen" as UI
participant "TUI model" as Model
participant "ProjectMetadataExecutor" as Executor
participant "ProjectMetadataService" as Service
database "SQLite index" as DB
collections ".chc/project.md" as File

User -> UI: edit project metadata fields
UI -> Model: update in-memory ProjectMetadata
Model -> Model: mark project metadata dirty

User -> UI: save
UI -> Executor: SaveProjectMetadata(metadata)
Executor -> Service: Save(metadata)

Service -> Service: validate schema and required fields
Service -> File: read current file hash

alt source hash changed outside CHC
    Service --> Executor: conflict error
    Executor --> UI: report resync required
else source hash matches or file is new
    Service -> File: write Markdown atomically
    Service --> Executor: saved metadata with new source_hash
    Executor -> DB: upsert indexed metadata
    DB --> Executor: indexed
    Executor --> UI: saved metadata
    UI -> Model: replace metadata and clear dirty state
end

@enduml
```

The service owns:

- validation;
- conflict detection;
- atomic file writes;
- database update;
- source hash tracking;
- diagnostics returned to the UI/CLI.

## Manual Edit And Resync Model

Manual file edits are supported, but they do not silently mutate the database.

The user can run an explicit resync action. The first implementation may expose
this through the command palette or a contextual screen action. A daemon or file
watcher may be added later, but is not required for the first complete model.

```plantuml
@startuml
title Manual project metadata resync

actor User
participant Editor
collections ".chc/project.md" as File
participant "Project screen" as UI
participant "Project metadata indexer" as Indexer
database "SQLite index" as DB

User -> Editor: edit project.md manually
Editor -> File: save
User -> UI: request resync
UI -> Indexer: reindex project metadata
Indexer -> File: read file
Indexer -> Indexer: parse frontmatter and validate schema

alt valid
    Indexer -> DB: update indexed metadata and source hash
    Indexer --> UI: resync complete
else invalid
    Indexer -> DB: keep last valid indexed metadata
    Indexer -> DB: record diagnostics
    Indexer --> UI: show parse/schema error
end

@enduml
```

## Staleness And Conflict Rules

Every indexed project metadata record should store enough source information to
detect external changes:

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

When invalid content is found, SQLite must not be overwritten from invalid file
data.

When a CLI/TUI write detects external file changes, it must not blindly
overwrite them. It should require either:

- resync first; or
- an explicit overwrite action.

## UI Layout

The current spatial direction is acceptable:

- the primary project fields occupy the main left area;
- secondary facts occupy compact right-side panels;
- the bottom navigation/action bar remains the primary control surface.

The main panel should not repeat `Project` as an extra heading when the active
screen is already visible in navigation.

```text
┌──────────────────────────────────────────────┬──────────────────────────────┐
│ Name                                         │ Artifacts                    │
│ [ chc-blog-work                            ] │ tickets        12            │
│                                              │ dumps           3            │
│ Description                                  │ quizzes         1            │
│ [ CLI/blog workspace used to validate...   ] │ specs           4            │
│                                              │ plans           2            │
│ Context                                      ├──────────────────────────────┤
│ [ Long-form Markdown project context...    ] │ Location                     │
│ [                                          ] │ /home/.../chc-blog-work      │
└──────────────────────────────────────────────┴──────────────────────────────┘

Project  Cases  Dumps  Quizzes  Specs  Plans  Board  Dash  Help  Exit
```

## Display Fields

Primary editable fields:

- `Name`;
- `Description`;
- `Context`.

Secondary read-only fields:

- project path;
- ticket count;
- dump count;
- quiz count;
- spec count;
- plan count.

Excluded from normal display:

- internal project ID;
- `.chc` existence when normal TUI startup already proves it;
- `.prj` status;
- implementation labels such as `signals`, `scope`, or `local state`;
- tracker count until the product model decides whether trackers remain
  separate from tickets.

## Interaction Model

Initial complete interaction:

- The primary navigation opens `Project`.
- Arrow keys move between editable fields when edit mode exists.
- `Enter` or an explicit edit action enters field editing.
- `Esc` cancels the current field edit.
- A save action persists through `ProjectMetadataService`.
- A resync action reindexes `.chc/project.md` into SQLite.
- If the file changed outside the app, save shows a conflict instead of
  overwriting.
- If the file is invalid, the tab shows the last valid indexed values plus an
  actionable diagnostic.

The exact save/resync bindings are not finalized. They should be visible in the
bottom action area or contextual help before implementation.

## Context Assembly Rules

Project metadata contributes to agent context as follows:

- `name` identifies the project compactly;
- `description` provides short technical orientation;
- body Markdown provides extended context when the context budget allows;
- artifact counts are UI metadata and should not normally be sent as project
  context unless a context builder explicitly needs them.

The app should avoid treating README content as canonical project metadata once
`.chc/project.md` exists. README inference may be used only as bootstrap or
fallback.

## Adjacent Artifact Workflow

Project context feeds the broader authoring workflow, but the `Project` screen
does not introduce a separate draft artifact category.

Preferred artifact progression:

```text
dump / quiz -> spec proposed -> spec accepted -> plan -> board
```

Draft-like work should be represented as a specification state when it belongs
to the specification path. `Drafts` should not become a first-class project
screen or durable artifact bucket unless a separate product role is established.

## Bootstrap Rules

When `.chc/project.md` does not exist:

- the TUI may infer `name` from the project path;
- the TUI may infer `description` from the first useful README paragraph;
- inferred values should be visibly unsaved or presented as bootstrap defaults;
- saving should create `.chc/project.md` through the shared domain service.

## Acceptance Criteria

- The tab shows only project-relevant information.
- The user can identify and edit project `name`, `description`, and long-form
  context.
- The tab does not show internal generated project IDs.
- The tab does not show irrelevant workspace implementation diagnostics.
- The canonical project file is Markdown with YAML frontmatter.
- SQLite indexes validated structured metadata and supports fast queries.
- Manual file edits are supported through explicit resync.
- Invalid manual edits do not corrupt the SQLite index.
- CLI and TUI writes share one domain/application operation.
- External file changes are detected before CLI/TUI writes overwrite files.
