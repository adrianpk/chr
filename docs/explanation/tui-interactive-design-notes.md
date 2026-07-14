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
- Derived metadata includes artifact counts, source file facts, and actionable
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
- Markdown-only persistence direction:
  - project-local files under `.chr/` should remain human-editable and cannot be
    treated as a sealed vault;
  - Markdown is the source of truth for project and case documents;
  - the TUI/CLI may keep a lightweight in-memory workspace index for
    navigation, suggestions, diagnostics, and context selection;
  - the index stores metadata, relationships, parse status, paths, source
    hashes, and optional semantic digest references, not full body context;
  - user edits through Vim or another editor are valid and become visible after
    refresh;
  - broken file parseability should surface as an actionable diagnostics state.
- Project metadata file-format direction:
  - project metadata should not use a dedicated YAML file;
  - it should use Markdown with YAML frontmatter;
  - `name` and `description` are frontmatter fields;
  - the Markdown body is a separate long-form project context field;
  - the body lets the project carry richer context without making
    `description` too long for frontmatter or UI summaries.
- Refresh direction:
  - manual file edits are a first-class workflow for terminal/editor-oriented
    users;
  - filesystem-to-memory refresh can be delegated to an explicit manual refresh
    action at first, possibly exposed as a function-key command;
  - a daemon or watcher may come later, but should not be required for the first
    usable model;
  - CLI/TUI writes should use the application/domain layer and write Markdown
    files atomically;
  - app writes must account for user edits made outside the app, so blind
    overwrites need runtime conflict detection or an explicit overwrite policy.

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
- Use cases are Markdown documents under `.chr/default/cases/`.
- Useful persistence metadata can exist below the UI surface:
  - stable `slug` or ID derived from the name;
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
  - persist a human-editable Markdown file with YAML frontmatter.
- A use case is conceptually a project-local container for related artifacts
  such as dumps, quizzes, specs, plans, and board work.
- A use case belongs to the current project by containment under `.chr/`; a
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
  - artifact counts, source file facts, and diagnostics below that document
    panel.
- A case reference from another artifact should point to a canonical case slug.
- A missing case slug should appear as an unresolved reference, not silently
  create a new case.
- The normal context assembly path starts from a selected case when available,
  and falls back to project-wide material when no case applies.
- Documents such as dumps, quizzes, specs, and plans are case-scoped when they
  declare a case slug. If they do not declare a case slug, they are
  project-wide.

## Workspace Index And Semantic Cache

- The workspace index is the local, in-memory read model for the TUI.
- It is a poor-man's graph over Markdown files, not a database.
- The filesystem remains the durable source of truth.
- The structural index should be rebuildable from files and should contain:
  - nodes;
  - edges;
  - source paths;
  - source hashes;
  - modified times;
  - lookup maps;
  - diagnostics.
- Full document bodies should be read lazily rather than stored in the normal
  structural index.
- A discardable cache can speed up startup, but cache contents must never become
  authoritative.
- Semantic digests are optional dense retrieval hints for agents and search.
- Semantic digests may be generated by an LLM, but only as an incremental,
  cache-backed worker path.
- Restarting the TUI should reuse valid semantic digest cache entries instead
  of calling an LLM again.
- Missing or stale semantic digests should be visible as index health, not block
  normal operation.

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
- Quiz actions should not generate specs; spec generation belongs in `F5 Specs`.

## F5 Specs

- `Drafts` should not remain a first-class artifact category or TUI surface.
- Rough capture material belongs in dumps and quizzes.
- Formal behavior belongs in specs.
- A spec can have `status: draft` when it needs review, but that is a spec
  state, not a separate document family.
- The Specs surface owns generation of specs from completed quiz material.
- Accepted workflow direction:
  - dump or quiz;
  - spec with `status: draft`;
  - spec accepted or finalized;
  - plan;
  - board.

## Board

- `Board` is the Kanban-like representation of boards under `.chr/`.
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
