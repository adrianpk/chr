# TUI Interactive Design Notes

Status: working notes

These notes capture product direction for the interactive TUI. They are not a
spec, ticket, or implementation plan.

## Operating Mode

- Work directly and interactively.
- Do not turn questions into code changes.
- Do not create PRs unless explicitly requested.
- Use notes to preserve design context before choosing whether to quiz, draft,
  or implement.

## Current Design Feedback

- The lower button bar is the right general direction.
- Do not duplicate the lower button bar with another navigation bar at the top.
- The top area should not show internal branding as a primary element.
- The top area should not show generated project IDs such as
  `project-60d1f32d...`.
- Do not assume the top area is for status.
- Avoid terminal-dashboard decoration that only stretches to fill space.
- The TUI should feel closer to operator tools such as Lazygit, k9s, or
  Lazydocker: reactive, focused, dense, keyboard-first, and workflow-oriented.

## Reusable Screen Geometry

- Editable tabs should use the same spatial grammar unless a concrete workflow
  requires a different layout.
- The upper-right area is the natural detail-editing surface in master/detail
  screens.
- Derived metadata should not compete with editable content in the upper-right
  area.
- For `Project`, `Cases`, and similar editable screens:
  - upper area: primary in-place editing when the screen represents a single
    record;
  - upper-left: primary selection when the screen has a real list;
  - upper-right: editable detail fields only when the screen has real
    master/detail behavior;
  - lower-left: long-form body/context or supporting content;
  - lower-right: derived, read-only metadata.
- Derived metadata includes artifact counts, source/index facts, and actionable
  diagnostics.
- Inline editing is the default editing model. Editing should happen where the
  field is displayed rather than moving the field into a separate temporary
  editing region.

## Project Identity

- User-facing project identity should be human-readable.
- Prefer the real project name when available.
- Fall back to the final path segment when no explicit name exists.
- Internal project IDs should remain hidden unless a diagnostic/debug view needs
  them.

## Project

- `Project` is not a dashboard.
- It should describe the project semantically for human orientation and agent
  context.
- The content should explain what the project does or intends to do.
- It should emphasize technical and behavioral aspects.
- It should not copy the README blindly.
- It should not become a wall of words.
- Initial fields:
  - `name`
  - `description`
- `name` can be inferred from the project path.
- `description` can be inferred from README or other project signals, but should
  be curated and compact.
- This content may later be used as context, so it should avoid irrelevant token
  cost.
- The normal screen should make it clear how the user changes `name` and
  `description`; showing inferred values without an edit path is incomplete.
- The editing model is unresolved:
  - direct Markdown editing may be too file-centric for a TUI field;
  - invoking the same application command path as the CLI would keep behavior
    consistent;
  - persistence may need both SQLite and a Markdown representation, consistent
    with the rest of project-local data.
- Open persistence question: for editable project metadata, should the command
  layer write SQLite first and then render/update Markdown, or should Markdown be
  the canonical source and SQLite be an index/cache?
- Current persistence direction:
  - project-local files under `.chc/` should remain human-editable and cannot be
    treated as a sealed vault;
  - SQLite should support fast querying/indexing for UI and command responses;
  - user edits to project-local files imply an indexing/reconciliation process
    that notices file changes and refreshes the database when the file remains
    parseable;
  - if a user breaks the file schema, the database should not blindly update from
    invalid content;
  - broken file parseability should surface as an actionable diagnostics state,
    not silently corrupt indexed state.
- Project metadata file-format direction:
  - project metadata should not use a dedicated YAML file;
  - it should use Markdown with YAML frontmatter;
  - `name` and `description` are frontmatter fields;
  - the Markdown body is a separate long-form project context field;
  - the body lets the project carry richer context without making
    `description` too long for frontmatter or UI summaries.
- Synchronization direction under discussion:
  - manual file edits are a first-class workflow for terminal/editor-oriented
    users;
  - file-to-database synchronization can be delegated to an explicit manual
    resync action at first, possibly exposed as a function-key command;
  - a daemon or watcher may come later, but should not be required for the first
    usable model;
  - CLI/TUI writes are easier to keep consistent because they can use the
    application/domain layer and persist database-to-file in one controlled
    operation;
  - database-to-file writes must account for user edits made outside the app, so
    blind overwrites need conflict detection or an explicit overwrite policy.

## Button Bar

- The bottom button bar is the primary navigation/control surface.
- It should occupy the full horizontal bottom area.
- `F1` is `Project`.
- `F2` should be `Cases` or `Use cases`; current preference is likely `Cases`
  because it is shorter.
- The rest of the final button mapping is not yet fixed.

## Cases

- `Cases` is the space for listing and administering use cases.
- It should show the list of existing use cases.
- It should support creating new use cases.
- A use case is intentionally minimal:
  - `name`
  - `description`
- No additional fields are assumed for now.
- Use cases should have SQLite persistence support, consistent with the rest of
  the app.
- Use cases should also have a Markdown representation under `.chc/`.
- Useful persistence metadata can exist below the UI surface:
  - stable `slug` or ID derived from the name;
  - `status`, likely only `active` or `archived`;
  - `created_at` and `updated_at`;
  - `doc_path` for the Markdown representation.
- Do not add ownership fields.
- The app is single-user: each user has their own workspace and owns what they
  create.
- Ticket counts or related artifacts should be derived by queries, not manually
  maintained as use case fields.
- Reuse the `Project` screen implementation lessons:
  - use Bubble Tea/Bubbles components for text editing rather than custom
    low-level input handling;
  - keep a shared domain/application service between CLI and TUI;
  - persist a human-editable Markdown file with YAML frontmatter;
  - maintain SQLite as a query/index layer with source path, source hash,
    schema version, and indexed timestamp.
- A use case is conceptually a project-local container for related artifacts
  such as dumps, quizzes, specs, plans, and board work.
- A use case belongs to the current project by containment under `.chc/`; a
  `project_id` frontmatter field is not currently necessary for normal
  operation.
- A `project_id` frontmatter field could help standalone file rehydration, but
  it can also duplicate path context and create drift.
- Current leaning: do not include project identity in use-case frontmatter unless
  a concrete import/export workflow requires it.
- The Cases screen likely needs:
  - a navigable list;
  - quick search/autocomplete;
  - creation of new use cases;
  - an editable detail panel for `name` and `description`;
  - derived artifact counts by case.
- The Cases screen should follow the reusable screen geometry:
  - list/search in the left region;
  - editable `name`, `description`, and long-form case context in one right-side
    document panel;
  - artifact counts, source/index facts, and diagnostics below that document
    panel.

## F3 Dumps

- `F3 Dumps` is the space for free-form project knowledge dumps.
- Dumps can be attached to a use case.
- Dumps without a use case are project-wide.
- In other words, no use case means the dump belongs to the general project
  dump set.

## F4 Quizzes

- `F4 Quizzes` follows the same scope model as dumps.
- Quizzes can be attached to a use case.
- Quizzes without a use case are project-wide.
- The area should support quiz administration.
- It also needs an interactive GUI flow where the agent asks questions and the
  user answers.
- This should feel like a focused question/answer workspace, not a static list
  only.

## F5 Drafts

- Drafts and dumps likely represent the same conceptual input material.
- `Drafts` sounds more polished, but this is an agentic system.
- Agents should generally receive specs, not drafts.
- Drafts imply more deliberate editorial work, while dumps can remain rough
  capture material.
- Current product direction: avoid a separate drafts surface unless a clearer
  distinction emerges.
- A likely flow is:
  - dump;
  - quiz;
  - spec.
- Accepted workflow direction:
  - dump or quiz;
  - spec proposed;
  - spec accepted;
  - plan;
  - board.
- `Drafts` should not remain a first-class artifact category or TUI surface.
- A not-yet-accepted specification should be represented as a spec state rather
  than as a separate draft artifact.

## Board

- `Board` is the Kanban-like representation of boards under `.chc/`.
- It should reflect the board model used by the project, not invent a separate
  ticket-only abstraction.
- `~/Projects/redwright/` is a useful reference for how this kind of board
  interaction is handled elsewhere.

## Dash

- `Dash` is the dashboard.
- It is where the user sees what the agent is doing in real time.
- The exact visual representation is not decided yet.
- The core purpose is real-time agent activity, not static project metadata.

## Help And Exit

- `Help` is self-explanatory.
- `Exit` is self-explanatory.
