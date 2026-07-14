# Optional Local Web UI

Status: draft

Date: 2026-06-29

Related:

- `docs/reference/tui.md`
- `docs/reference/commands.md`
- `docs/reference/project-workspace.md`
- `docs/reference/ticket-status.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/agent-work-observability.md`
- `docs/reference/cli-tui-action-matrix.md`

## Summary

Chryso should provide an optional local Web UI for the same managed-project
operations currently available through the TUI and CLI.

The Web UI is an additional operator surface. It is not a replacement for the
TUI, not a hosted service, and not a new state model.

The Web UI must:

- be disabled by default;
- run only when explicitly enabled or explicitly started;
- bind to localhost by default;
- allow non-local network interfaces only through explicit configuration or
  command flags;
- use the same project-local Markdown files and user-level indexes as the TUI;
- stay synchronized with TUI and CLI changes by reloading shared state;
- preserve the existing TUI board model and operational views;
- use a Chryso-specific visual palette and simple agent avatar;
- avoid introducing a separate product workflow.

## Motivation

The current TUI is convenient for local operation. It provides a compact board,
project views, command palette behavior, runtime evidence, and confirmation
flows.

A local Web UI would make the same operational surface available in a browser:

- wider boards and panes;
- easier copying of paths and evidence;
- easier inspection from another machine when the user deliberately exposes the
  local server through a private network or tunnel;
- a familiar interaction model for users who prefer browser-based project
  dashboards.

The goal is parity and accessibility, not visual reinvention.

## Design Principles

### Optional By Default

No server starts unless the user explicitly asks for it.

The TUI remains the default behavior for:

```sh
chr
chr tui
```

The Web UI starts only through an explicit command or explicit configuration.

### Local First

The default bind address is:

```text
127.0.0.1
```

Default URLs are local:

```text
http://127.0.0.1:<port>
http://localhost:<port>
```

The implementation must not expose the Web UI on all interfaces by accident.

### Same State, Same Operations

The Web UI must use the same backing state as the TUI:

- project-local Markdown artifacts under `.chr/`;
- `.chr/user/` active userspace;
- user-level SQLite indexes and records;
- runtime artifact files;
- execution and loop records;
- event streams;
- generated reports.

The Web UI must not introduce a browser-only state store that becomes a source
of truth.

### Preserve The Existing Interaction Model

The Web UI should follow the existing TUI structure:

- Project;
- Cases;
- Dumps;
- Quizzes;
- Specs;
- Plans;
- Board;
- Dash;
- Help and commands.

The Web UI has no existing screen layout to preserve. Its full screen
composition should match the reference application selected during
implementation, across Board, Dash, Dumps, Quizzes, Specs, Plans, Project,
Cases, Help, and command surfaces. Use that reference for layout, spacing, pane
structure, card density, navigation, editing surfaces, and operational
ergonomics. Preserve the Chryso TUI operational model and change only the
visual identity needed to make the UI belong to Chryso.

### No Workflow Expansion

The Web UI must not add new lifecycle concepts.

If an operation exists in TUI or CLI, the Web UI may expose it.

If an operation does not exist in TUI or CLI, the Web UI should not invent it in
the first implementation.

## User Model

### TUI-Only User

The user runs:

```sh
chr
```

No Web UI starts.

No port is opened.

No background web server is installed.

### Local Browser User

The user runs:

```sh
chr web serve
```

Chryso starts an HTTP server bound to `127.0.0.1` and prints:

```text
web ui: http://127.0.0.1:7331
project: /path/to/project
bind: 127.0.0.1
```

The user opens the local browser and operates the same project.

### Explicit Network User

The user runs:

```sh
chr web serve --host 0.0.0.0
```

or configures:

```yml
web:
  enabled: true
  host: 0.0.0.0
  port: 7331
```

Chryso prints an explicit warning:

```text
warning: web ui is bound to a non-loopback interface
```

The user is responsible for private-network access, firewall rules, reverse
proxy, private tunnel, or TLS termination.

Chryso should remain useful behind private overlay networks and local
tunnels, but it should not require or manage those systems.

## Configuration

Project-local configuration may define:

```yml
web:
  enabled: false
  host: 127.0.0.1
  port: 7331
  open_browser: false
  require_token: auto
  token_env: CHR_WEB_TOKEN
  allowed_origins: []
  refresh_interval_seconds: 2
```

### `enabled`

Type: boolean.

Default:

```yml
enabled: false
```

Meaning:

- `false`: no automatic web server behavior;
- `true`: commands that start the web surface may use the configured defaults.

This flag must not make plain `chr` start the Web UI. Plain `chr` remains TUI.

### `host`

Type: string.

Default:

```yml
host: 127.0.0.1
```

Allowed examples:

```text
127.0.0.1
localhost
0.0.0.0
100.x.y.z
```

Non-loopback hosts require explicit configuration or explicit command flags.

### `port`

Type: integer.

Default:

```yml
port: 7331
```

If the configured port is unavailable, the command may either fail clearly or
select an alternate port only when the user passed an auto-port flag.

Silent port changes are not allowed.

### `open_browser`

Type: boolean.

Default:

```yml
open_browser: false
```

When true, `chr web serve` may open the local browser after the server is ready.

### `require_token`

Type: string.

Allowed values:

```text
auto
always
never
```

Default:

```yml
require_token: auto
```

Meaning:

- `auto`: token required when binding to non-loopback, optional on loopback;
- `always`: token required for every request except static assets needed for
  the login page;
- `never`: no token requirement.

`never` with a non-loopback host should print a warning.

### `token_env`

Type: string.

Default:

```yml
token_env: CHR_WEB_TOKEN
```

The token value is read from the environment. It is not written to project
configuration.

### `allowed_origins`

Type: list of strings.

Default:

```yml
allowed_origins: []
```

Empty means same-origin only.

Cross-origin browser access is denied unless explicitly configured.

### `refresh_interval_seconds`

Type: integer.

Default:

```yml
refresh_interval_seconds: 2
```

The frontend uses this as a fallback polling interval when server-sent events or
another live update channel is unavailable.

## Commands

### Serve

```sh
chr web serve
chr web serve --dir /path/to/project
chr web serve --host 127.0.0.1 --port 7331
chr web serve --host 0.0.0.0 --port 7331
chr web serve --open
```

Behavior:

1. Resolve the project.
2. Load project config.
3. Resolve effective web config.
4. Validate bind safety.
5. Start HTTP server.
6. Print URL and project path.
7. Block until interrupted.

The command should exit cleanly on `ctrl+c`.

### Preview

```sh
chr web serve --dry-run
```

Prints:

- resolved project;
- host;
- port;
- token requirement;
- allowed origins;
- static asset mode;
- whether the bind address is loopback or non-loopback.

Does not start a server.

### Open

Optional command:

```sh
chr web open
```

If a local server is already running and discoverable, open it.

If no local server is running, print the command needed to start one.

This command is optional for the first implementation.

## Server Scope

The first server is a single-process local HTTP server.

Recommended implementation:

- Go `net/http`;
- server-rendered HTML using `html/template`;
- htmx attributes and response headers for partial updates;
- embedded templates and static assets;
- HTML fragment endpoints for interactive surfaces;
- JSON endpoints only where they are the simpler contract for bootstrap,
  health, events, or machine-readable state;
- server-sent events or polling for refresh;
- no database server;
- no external frontend build service at runtime.

The server runs in the foreground of `chr web serve`.

No daemon is required for the first version.

## State Model

The Web UI reads and writes through the same application services used by CLI
and TUI where possible.

The Web UI must not duplicate write logic in HTTP handlers.

Preferred layering:

```text
HTTP route
  request validation
  service command/query call
  view model
  HTML page or htmx fragment
```

Avoid:

```text
HTTP route
  parse Markdown directly
  mutate files directly
  rebuild indexes inconsistently
```

## Internal Web UI Helpers

The Web UI may copy and adapt small local helper primitives for HTML and htmx
work into Chryso-owned internal packages.

The copied helpers must live under an internal Web UI boundary, such as:

```text
internal/webui/htmx
internal/webui/render
internal/webui/ui
```

Allowed helper categories:

- htmx attribute builders;
- htmx request and response header helpers;
- out-of-band swap helpers;
- template function maps;
- template rendering helpers;
- small UI primitives such as buttons, chips, links, nav, tables, panels, and
  forms.

Rules:

- copy or adapt only the primitives needed by the current slice;
- keep helper package names and APIs local to Chryso;
- do not add a module dependency on the source helper library;
- do not import Web UI helpers from business, core, app, persistence, or TUI
  packages;
- do not put Chryso business behavior in helper packages;
- add local tests for copied or adapted primitives;
- preserve clear escaping boundaries by using `html/template` types only where
  the rendered HTML is intentionally trusted.

## Synchronization With TUI And CLI

The TUI, CLI, and Web UI must remain synchronized by sharing the same durable
state.

### Read Synchronization

The Web UI refreshes from:

- project-local Markdown files;
- ticket index;
- runtime records;
- event streams;
- report files.

The Web UI should support:

- manual refresh;
- automatic refresh while the page is visible;
- pause or reduced refresh while the tab is hidden.

### Write Synchronization

When the Web UI writes, it uses existing service operations.

After each successful write, the Web UI reloads the affected view from the
server response or from a fresh query.

The TUI should observe the change on its next refresh cycle.

### Concurrent Edits

The Web UI must detect stale edits for text-like artifacts.

For editable documents, server responses should include:

- project-relative path;
- content;
- content hash;
- updated timestamp when available.

Save requests should include the content hash the user edited.

If the file changed since it was loaded, the server rejects the write with a
conflict response and returns:

- current content hash;
- current content;
- submitted content;
- path.

The first implementation may show a simple conflict screen rather than a merge
editor.

### Running Operations

Spec generation, plan generation, quiz actions, and agent work can take time.

The Web UI should represent long-running operations as operation records or
transient server-side jobs only if the underlying service already supports that
state.

For the first version, the server may keep the request open until the operation
finishes for operations that are already foreground in CLI/TUI.

The UI must show busy state and prevent duplicate submissions for the same
operation.

## View Parity

The Web UI should map to the TUI views.

### Project

Required:

- show project name;
- show project description;
- show project context;
- show project path;
- show artifact counts;
- edit and save project metadata;
- refresh.

Equivalent command:

```sh
chr metadata project set
```

### Cases

Required:

- list use cases;
- show selected use-case name, description, and context;
- filter use cases;
- create use case;
- edit and save use-case metadata;
- refresh.

Equivalent command:

```sh
chr metadata case set
```

### Dumps

Required:

- list curated dumps;
- show selected dump metadata;
- create dump;
- edit dump content;
- append to dump when supported by services;
- open or copy path;
- refresh.

Equivalent commands:

```sh
chr dump list
chr dump new
chr dump show
chr dump save
chr dump append
```

### Quizzes

Required:

- list quiz artifacts;
- show active question;
- start or resume quiz;
- answer question;
- skip question;
- complete quiz;
- cancel quiz;
- switch mode when supported;
- refresh.

Equivalent commands:

```sh
chr quiz list
chr quiz start
chr quiz resume
chr quiz answer
chr quiz skip
chr quiz complete
chr quiz cancel
chr quiz mode
```

### Specs

Required:

- list use cases and spec state;
- generate missing spec;
- show spec path;
- show generated spec content or preview;
- refresh.

Equivalent commands:

```sh
chr spec list
chr spec generate
```

### Plans

Required:

- list use cases and plan state;
- generate missing plan;
- regenerate plan after confirmation;
- create plan tracking after confirmation;
- show plan path;
- show plan content or preview;
- refresh.

Equivalent commands:

```sh
chr plan list
chr plan generate
chr plan regenerate
chr plan tracking create
```

### Board

Required:

- show normal columns:
  - Backlog;
  - Ready For Dev;
  - In Progress;
  - Ready For Review;
  - Ready For Merge;
  - Done;
- show attention view;
- filter visible items;
- select an item;
- show detail pane;
- show synopsis;
- show source metadata;
- show slice metadata;
- show checklist counts;
- show comment counts;
- show body context;
- show checklist preview;
- show evidence strategy summary;
- show recent comment headings;
- show attention reasons;
- show source path;
- update status through supported actions;
- start work through supported action;
- run ready loop through supported action;
- refresh.

The Web Board should preserve the current TUI board mental model. It should not
introduce a different kanban taxonomy.

Equivalent commands:

```sh
chr ticket board
chr ticket list
chr ticket show
chr ticket status
chr ticket start
chr ticket loop
chr work ticket
chr work run-all
```

### Dash

Required:

- show work readiness status;
- show repair guidance;
- show ready queue;
- show current loop or execution state;
- show recent event messages;
- show execution report path;
- show runtime log path;
- show validation log path;
- show recent execution outcomes;
- start selected work after confirmation;
- run visible ready work after confirmation;
- apply safe readiness repair after confirmation;
- request stop after confirmation;
- refresh automatically.

Equivalent TUI behavior should remain the source of truth for semantics.

### Help And Commands

Required:

- show keyboard shortcuts and primary actions;
- show command palette equivalent;
- expose only actions available in the current view;
- show unavailable actions as disabled with reason.

## Action Confirmation

The Web UI must require confirmation for actions that:

- write files;
- change status;
- start agent work;
- run generation;
- regenerate artifacts;
- apply safe repairs;
- request stop.

Confirmation dialogs must state:

- action name;
- target project;
- target artifact or key;
- expected write paths when known;
- whether the action starts a long-running operation.

## HTTP Route Shape

The HTTP surface is private to the local Web UI in the first implementation,
but it should still be explicit and testable.

The primary browser contract is server-rendered HTML:

- full-page routes for direct navigation;
- htmx fragment routes for partial replacement;
- form posts for write actions;
- htmx response headers for redirects, refreshes, retargeting, and out-of-band
  updates.

JSON endpoints are allowed for health, bootstrap, event feeds, and
machine-readable state when HTML is not the simpler contract.

Suggested route families:

```text
GET  /
GET  /project
POST /project

GET  /cases
GET  /cases/{slug}
POST /cases
POST /cases/{slug}

GET  /dumps
GET  /dumps/{path}
POST /dumps
POST /dumps/{path}
POST /dumps/{path}/append

GET  /quizzes
POST /quizzes/{scope}/start
POST /quizzes/{scope}/answer
POST /quizzes/{scope}/skip
POST /quizzes/{scope}/complete
POST /quizzes/{scope}/cancel

GET  /specs
POST /specs/{scope}/generate

GET  /plans
POST /plans/{scope}/generate
POST /plans/{scope}/regenerate
POST /plans/{scope}/tracking

GET  /board
GET  /board/items/{key}
POST /board/items/{key}/status
POST /board/items/{key}/start

GET  /dash
POST /work/run-all
POST /work/repair
POST /work/stop

GET  /api/events
```

Route names may change during implementation if they map more cleanly to the
existing command/service names.

All write routes should return an updated full page, an htmx fragment, or
enough invalidation metadata for the browser to refresh the affected surface.

## View Models

The server should construct view-specific models rather than rendering raw
internal structs.

Reasons:

- preserve internal package boundaries;
- avoid accidental exposure of filesystem internals;
- make template rendering stable;
- allow TUI and Web UI to evolve independently while sharing services.

### Board View Model

Minimum shape:

```json
{
  "project": {
    "name": "Example",
    "path": "/path/to/project"
  },
  "columns": [
    {
      "id": "ready_for_dev",
      "title": "Ready For Dev",
      "items": []
    }
  ],
  "attention": [],
  "selected_key": "CHR-1",
  "detail": {
    "key": "CHR-1",
    "title": "Add runtime check",
    "status": "ready_for_dev",
    "synopsis": "Add runtime check.",
    "source_path": ".chr/user/tickets/CHR-1.md",
    "evidence": {
      "state": "complete",
      "summary": "complete"
    }
  }
}
```

### Dash DTO

Minimum shape:

```json
{
  "readiness": {
    "status": "ok",
    "detail": ""
  },
  "ready_queue": [],
  "current_execution": null,
  "recent_events": [],
  "recent_outcomes": []
}
```

## Visual Direction

The Web UI should use a stable visual reference selected during implementation.
The reference applies to the whole application, not only to the Board view.

Rules:

- preserve the selected reference application's screen layout closely;
- preserve the selected reference application's board, dump, quiz, spec, plan,
  dash, help, and command interaction patterns closely;
- preserve dense operational panes;
- preserve browser-first operational ergonomics while mapping actions to
  Chryso semantics;
- avoid marketing-page composition;
- avoid decorative hero sections;
- avoid redesigning the information architecture;
- change only the identity layer needed for Chryso.

### Palette

Use a Chrysocolla-inspired palette that remains close enough to Go-adjacent
colors to feel natural for a Go software tool.

Suggested tokens:

```css
--chr-bg: #0f1718;
--chr-surface: #162224;
--chr-surface-2: #1d2e31;
--chr-border: #2f4b4e;
--chr-text: #e6f1ef;
--chr-muted: #9db7b5;
--chr-cyan: #00add8;
--chr-turquoise: #2bb8a7;
--chr-green: #75b86b;
--chr-stone: #7f8f87;
--chr-warn: #d7b75f;
--chr-danger: #d86f5f;
```

The final palette should be checked against:

- readable contrast;
- board column density;
- code/log readability;
- light text on dark surfaces;
- attention and failure state visibility.

### Avatar

The central avatar should be a simple Chryso agent mark.

Required traits:

- compact;
- rounded;
- egg-like silhouette;
- simple bot or agent cues;
- no detailed character illustration;
- no dependence on any existing mascot;
- usable at small sizes;
- usable as favicon or app mark later.

The avatar should use the Chryso palette, especially cyan, turquoise, and
stone tones.

It must not become the center of the product experience. The board and
operational state remain primary.

## Frontend Implementation Constraints

The first frontend should be conservative.

Recommended:

- server-rendered HTML;
- htmx for partial updates and command interactions;
- plain CSS or a minimal design-token layer;
- no complex client-side state framework unless needed;
- no build pipeline required at runtime;
- static assets embedded into the Go binary when practical.

Acceptable if implementation needs it:

- a small TypeScript build;
- generated static assets committed or embedded;
- one frontend package directory with clear commands.

Avoid:

- unrelated component libraries;
- marketing templates;
- global state stores that duplicate backend state;
- client-side routing that hides project path or operation failures;
- a JSON-only single page app when server-rendered HTML is sufficient;
- browser-only persistence for operational state.

## Security

### Binding

Default bind:

```text
127.0.0.1
```

Binding to `0.0.0.0`, a LAN address, or an overlay-network address requires
explicit user action.

The startup output must show:

- effective host;
- effective port;
- whether the host is loopback;
- whether token auth is enabled;
- project path.

### Authentication

For loopback-only use, token auth may be optional.

For non-loopback use, token auth should be required by default.

The token is read from the configured environment variable.

The token must not be written into:

- `.chr/user/config.yml`;
- logs;
- browser-visible boot data;
- static files.

### CSRF

When token auth or cookies are used, write requests must include CSRF
protection or an equivalent same-origin write guard.

Same-origin restrictions should be strict by default.

### CORS

Default:

```text
no cross-origin API access
```

Configured origins should be exact origins, not wildcard patterns, unless a
future spec explicitly justifies wildcards.

### Logs

Server access logs should avoid leaking:

- tokens;
- full request bodies;
- environment variables;
- raw model prompts;
- large file contents.

## Error Handling

Errors should be shown in the relevant view and include:

- operation name;
- short human-readable message;
- affected path or item key when safe;
- diagnostic path when available;
- suggested next command when useful.

Examples:

```text
Could not start work: readiness is blocked.
Run `chr work doctor --dir <project>` or open Dash for repair guidance.
```

```text
Save conflict: this dump changed after you opened it.
Review the current content before overwriting.
```

## Accessibility

The Web UI should support:

- keyboard navigation for board columns and items;
- visible focus states;
- sufficient color contrast;
- non-color status labels;
- copyable paths;
- readable log text;
- responsive layout for laptop-width screens.

Mobile support is not a primary requirement for the first version, but the UI
should not become unusable on narrow browser windows.

## Performance

Initial targets:

- load a typical project board in under one second on localhost;
- avoid re-reading large files unnecessarily;
- debounce refresh while operations are running;
- cap rendered Markdown previews in dense board details;
- lazy-load large logs and reports.

The Web UI should not poll every large artifact on every refresh.

## Testing Requirements

Unit tests:

- web config defaults;
- host validation;
- token requirement resolution;
- view-model rendering;
- conflict detection;
- action request validation.

HTTP tests:

- `GET /` serves the shell;
- `GET /project` renders project state;
- `GET /board` renders normal columns and attention view;
- htmx fragment requests return fragment content and appropriate headers;
- write routes reject stale hashes;
- write routes require confirmation fields when applicable;
- non-loopback bind with missing token reports warning or error according to
  config;
- CORS defaults deny cross-origin writes.

Service integration tests:

- Web project metadata update uses the same service path as CLI/TUI.
- Web board status update writes the same ticket metadata as CLI/TUI.
- Web work start uses the same execution path as CLI/TUI.
- Web refresh observes file changes made by CLI.

Frontend smoke tests:

- board renders columns;
- item selection updates detail pane;
- command/action confirmation opens;
- disabled actions show reason;
- Dash renders readiness and recent events;
- conflict response is visible to the user.

## Documentation Requirements

Add or update:

- `docs/how-to/use-web-ui.md`;
- `docs/reference/web-ui.md`;
- `docs/reference/project-workspace.md` web config section;
- `docs/reference/commands.md` command entries;
- `docs/reference/cli-tui-action-matrix.md` with Web UI parity if the matrix is
  extended beyond CLI/TUI.

Docs must clearly state:

- Web UI is disabled by default;
- TUI remains the default;
- localhost binding is the default;
- exposing beyond localhost is user-controlled;
- Web UI uses the same project state as TUI and CLI.

## Suggested Implementation Slices

| Slice | Name | Purpose |
| --- | --- | --- |
| Slice 1 | Web config and serve command | Add config parsing, `chr web serve`, bind validation, dry-run output, and a minimal health endpoint. |
| Slice 2 | HTML, htmx, and helper foundation | Add server-rendered templates, embedded assets, htmx helpers, UI primitives, and a browser-loadable shell. |
| Slice 3 | Read-only project and board views | Add project, board, and Dash read routes with page and fragment rendering without write actions. |
| Slice 4 | View parity foundations | Add Cases, Dumps, Quizzes, Specs, and Plans read views with shared view models and refresh behavior. |
| Slice 5 | Confirmed write actions | Add project metadata, case metadata, dump editing, status transitions, and confirmation flow. |
| Slice 6 | Work actions and Dash parity | Add start-work, run-ready, repair, stop request, busy states, and event refresh. |
| Slice 7 | Visual identity and reference layout pass | Apply the selected reference layout, Chryso palette, avatar, accessibility checks, and screenshot-based verification. |
| Slice 8 | Security and exposure hardening | Add token mode, CORS policy, CSRF or same-origin write guard, non-loopback warnings, and docs. |

## Open Questions

- Should `web.enabled: true` allow an automatic web server in a future daemon,
  or should it only provide defaults for `chr web serve`?
- Should token auth be required even on loopback for consistency?
- Should the first implementation use server-sent events or polling?
- Should the selected visual reference be copied structurally into native
  templates or adapted through a small frontend app?
- Should the Web UI support multiple registered projects in one server, or only
  the resolved project for the first version?
- Should non-loopback binding require a typed confirmation on every start, or is
  explicit config enough?
