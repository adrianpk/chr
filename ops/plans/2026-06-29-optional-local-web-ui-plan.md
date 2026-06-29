# Optional Local Web UI Plan

Status: in_progress

Active Slice: Slice 9: visual identity and layout parity

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Related Spec: `ops/specs/2026-06-29-execution-unit-runtime-refinements.md`

## Purpose

Add an optional local Web UI for the same managed-project operations currently
available through ChrysoCoord's TUI and CLI.

The Web UI is an additional operator surface. It must not become a separate
workflow, separate persistence model, hosted service, or replacement for the
TUI. It starts only when explicitly enabled or explicitly invoked.

The delivery keeps the visual and operational shape close to the selected
reference application during implementation, while using ChrysoCoord identity,
palette, commands, state, and domain language.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice updates its implementation report under
  `.prj/reports/slices/optional-local-web-ui/`.
- Each task gets one conventional commit.
- Plan and tracker must be committed before runtime code changes begin.
- Start the next slice only after the previous slice is merged into `dev`.
- Keep the Web UI disabled by default throughout the rollout.
- Do not start a background server from `chc`, `chc tui`, or normal work
  commands.
- Do not introduce a second source of truth for project state.
- Do not add new lifecycle concepts that are unavailable from the existing TUI
  or CLI.
- Do not expand public network exposure without explicit operator
  configuration.

## Architecture Decisions

- Runtime shape: a single ChrysoCoord process serves the Web UI through an
  explicit command.
- Default bind: `127.0.0.1`.
- Non-loopback bind: allowed only through explicit config or command flags and
  must print a warning.
- Persistence: reuse existing project-local artifacts, user indexes, runtime
  records, event streams, and generated reports.
- Server stack: use the standard Go HTTP stack, server-rendered
  `html/template` pages, htmx partial updates, and embedded templates/static
  assets.
- Helper strategy: copy and adapt only the needed HTML/htmx/UI primitives into
  ChrysoCoord-owned internal Web UI packages such as `internal/webui/htmx`,
  `internal/webui/render`, and `internal/webui/ui`. Do not add a module
  dependency on the source helper library.
- Package boundary: business, core, app, persistence, and TUI packages must not
  import `internal/webui` helper packages.
- Client state: browser state is cache and presentation only. It is never the
  authoritative state.
- Synchronization: the Web UI reloads from shared state after TUI, CLI, or web
  actions. Polling is acceptable at first; server-sent events may replace or
  augment it when the state boundary is stable.
- Security: bind-local by default, same-origin write actions, optional token
  protection, no remote-control exposure by accident.
- Visual identity: keep the selected reference application's layout density,
  navigation rhythm, board/document surfaces, and command ergonomics. Replace
  only identity tokens needed for ChrysoCoord: chrysocolla/Go-adjacent palette
  and a simple rounded bot/agent avatar.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | web config and serve command | `feat/slice-1-web-config-serve-command` | `feat(slice-1): web config and serve command` | `.prj/reports/slices/optional-local-web-ui/slice-1-web-config-serve-command.md` |
| Slice 2 | html htmx and helper foundation | `feat/slice-2-html-htmx-helper-foundation` | `feat(slice-2): html htmx helper foundation` | `.prj/reports/slices/optional-local-web-ui/slice-2-html-htmx-helper-foundation.md` |
| Slice 3 | read-only project board and dash | `feat/slice-3-readonly-project-board-dash` | `feat(slice-3): read-only project board and dash` | `.prj/reports/slices/optional-local-web-ui/slice-3-readonly-project-board-dash.md` |
| Slice 4 | read-only artifact catalogs | `feat/slice-4-readonly-artifact-catalogs` | `feat(slice-4): read-only artifact catalogs` | `.prj/reports/slices/optional-local-web-ui/slice-4-readonly-artifact-catalogs.md` |
| Slice 5 | document and metadata actions | `feat/slice-5-document-and-metadata-actions` | `feat(slice-5): document and metadata actions` | `.prj/reports/slices/optional-local-web-ui/slice-5-document-and-metadata-actions.md` |
| Slice 6 | work and generation actions | `feat/slice-6-work-and-generation-actions` | `feat(slice-6): work and generation actions` | `.prj/reports/slices/optional-local-web-ui/slice-6-work-and-generation-actions.md` |
| Slice 7 | live sync and conflict handling | `feat/slice-7-live-sync-and-conflicts` | `feat(slice-7): live sync and conflicts` | `.prj/reports/slices/optional-local-web-ui/slice-7-live-sync-and-conflicts.md` |
| Slice 8 | access control and exposure safety | `feat/slice-8-web-access-control` | `feat(slice-8): web access control` | `.prj/reports/slices/optional-local-web-ui/slice-8-web-access-control.md` |
| Slice 9 | visual identity and layout parity | `feat/slice-9-web-visual-identity` | `feat(slice-9): web visual identity` | `.prj/reports/slices/optional-local-web-ui/slice-9-web-visual-identity.md` |
| Slice 10 | docs qa and release readiness | `feat/slice-10-web-docs-qa` | `feat(slice-10): web docs and qa` | `.prj/reports/slices/optional-local-web-ui/slice-10-web-docs-qa.md` |

## Slice 1: Web Config And Serve Command

Goal: add the disabled-by-default web configuration contract and an explicit
server command without delivering the full UI.

Behavior:

- `chc`, `chc tui`, and current work commands do not start a web server.
- `chc web serve` resolves the current project and web config.
- The default host is `127.0.0.1`.
- The default port is stable and documented.
- `--host`, `--port`, and config values use the same resolution path.
- Non-loopback host values print an explicit exposure warning.
- A dry-run mode reports the resolved project, host, port, token mode, and
  static asset mode without listening on a socket.
- The command can serve a minimal health response so later slices have a real
  integration point.

Tasks:

- T1.1 Add the `web` config section to the project/user config model with
  defaults.
- T1.2 Add config validation for host, port, refresh interval, token mode, and
  allowed origins.
- T1.3 Add CLI parsing for `chc web serve`.
- T1.4 Add `--host`, `--port`, `--open-browser`, `--dry-run`, and
  `--require-token` command flags.
- T1.5 Add resolved-config output for normal startup and dry-run mode.
- T1.6 Add non-loopback bind warning output.
- T1.7 Add a minimal HTTP server boundary with `/healthz`.
- T1.8 Add unit tests for config defaults, overrides, invalid values, and
  warning decisions.
- T1.9 Add CLI tests for dry-run and command output.
- T1.10 Add the slice implementation report and validation evidence.

## Slice 2: HTML HTMX And Helper Foundation

Goal: deliver a browser-loadable shell and embedded asset boundary without
exposing project mutation.

Behavior:

- `chc web serve` serves a real server-rendered HTML shell.
- Templates and static assets are embedded in the ChrysoCoord binary.
- Asset routes are stable and cacheable.
- htmx is available for partial updates.
- The first copied/adapted helper primitives live under `internal/webui`.
- The helper packages contain presentation primitives only.
- The shell has the top-level application frame, navigation, loading state,
  empty state, error state, and command-surface placeholder.
- The shell can render project bootstrap state without a JSON-only client app.
- The Web UI remains read-only.
- The server shuts down cleanly on process interruption.

Tasks:

- T2.1 Add an internal web package boundary for server construction.
- T2.2 Add `internal/webui/htmx` with the needed htmx attribute, request header,
  response header, and out-of-band swap helpers.
- T2.3 Add `internal/webui/render` with template rendering and function-map
  helpers.
- T2.4 Add `internal/webui/ui` with the first needed UI primitives for nav,
  buttons, chips, links, and page layout.
- T2.5 Add embedded templates and static assets for HTML, CSS, htmx, and the
  application shell.
- T2.6 Add root route handling for the application shell.
- T2.7 Add a bootstrap view model with resolved project identity, config,
  feature flags, and navigation model.
- T2.8 Add consistent HTML error page and htmx fragment error helpers.
- T2.9 Add graceful shutdown handling for the web command.
- T2.10 Add helper tests for escaping, htmx attributes, htmx headers, and
  rendering.
- T2.11 Add server tests for routes, content types, cache headers, and 404s.
- T2.12 Add browser or documented manual smoke evidence for the shell.
- T2.13 Add the slice implementation report and validation evidence.

## Slice 3: Read-Only Project Board And Dash

Goal: render the primary operational project surfaces in read-only mode using
existing state.

Behavior:

- The Project view shows the resolved project, userspace, config summary, and
  important paths.
- The Board view uses the existing board status model and attention state.
- Board columns match the existing TUI board model.
- Board cards expose status, title, source metadata, slice metadata, checklist
  counts, comment counts, evidence summary, and attention reason.
- Board detail opens without changing state.
- The Dash view shows work readiness, loop state, active execution state,
  recent events, reports, and recent outcomes.
- Data shown in the browser can be traced to existing TUI/CLI state sources.

Tasks:

- T3.1 Add read view models for project summary, board column, board card, board
  detail, dash summary, event summary, and report summary.
- T3.2 Add app/service queries that reuse existing TUI or core read models
  instead of reparsing state in the web layer.
- T3.3 Add Project, Board, board item detail, and Dash routes with full-page
  and htmx fragment responses.
- T3.4 Render Project, Board, and Dash screens through server templates.
- T3.5 Add board-column empty states and attention states.
- T3.6 Add detail panes or detail pages for board items.
- T3.7 Add tests for view-model shape, status mapping, empty project state, and
  attention reasons.
- T3.8 Add screenshot or DOM smoke evidence for Project, Board, and Dash.
- T3.9 Add the slice implementation report and validation evidence.

## Slice 4: Read-Only Artifact Catalogs

Goal: expose the operational artifact views in read-only mode before adding
write actions.

Behavior:

- Cases, Dumps, Quizzes, Specs, and Plans have first-class navigation entries.
- Each artifact family has a catalog view matching the selected reference
  application's document/catalog density.
- Catalog cards show title, state, source path, timestamps when available,
  origin metadata, and short preview.
- Artifact detail views render Markdown content safely.
- Missing artifact directories produce useful empty states.
- The Web UI does not create, edit, regenerate, or move artifacts in this
  slice.

Tasks:

- T4.1 Add read view models for case, dump, quiz, spec, plan, and generic document
  previews.
- T4.2 Add service queries for artifact catalogs using existing ChrysoCoord
  discovery/index behavior.
- T4.3 Add Cases, Dumps, Quizzes, Specs, and Plans list routes with full-page
  and htmx fragment responses.
- T4.4 Add detail routes for each artifact family.
- T4.5 Add safe Markdown rendering or pre-rendered HTML handling consistent
  with existing project dependencies.
- T4.6 Add catalog and detail screens for Cases, Dumps, Quizzes, Specs, and
  Plans.
- T4.7 Add tests for empty catalogs, malformed artifact metadata, missing
  files, and path confinement.
- T4.8 Add screenshot or DOM smoke evidence for all artifact catalog screens.
- T4.9 Add the slice implementation report and validation evidence.

## Slice 5: Document And Metadata Actions

Goal: add confirmed write actions for project metadata and editable documents
without starting work execution.

Behavior:

- Web writes go through the same application services as TUI/CLI writes.
- Write actions require an explicit browser confirmation.
- The server rejects stale writes using a revision, mtime, content hash, or
  equivalent concurrency token.
- Successful writes refresh the affected read models.
- Failed writes return stable error envelopes.
- The Web UI can create or update dumps and edit supported artifact metadata
  when those operations already exist in ChrysoCoord.
- The Web UI still does not start agent work or generation flows in this slice.

Tasks:

- T5.1 Define the write-response envelope for success, validation failure,
  stale edit, and blocked operation.
- T5.2 Add revision tokens to relevant read view models.
- T5.3 Add same-service write endpoints for supported metadata updates.
- T5.4 Add dump create, append, and edit endpoints if matching CLI/TUI
  operations exist.
- T5.5 Add browser confirmation flows for destructive or overwrite-prone
  actions.
- T5.6 Add stale-write tests and path-confinement tests.
- T5.7 Add client refresh behavior after successful writes.
- T5.8 Add manual or automated smoke evidence for create, edit, stale edit,
  and validation failure.
- T5.9 Add the slice implementation report and validation evidence.

## Slice 6: Work And Generation Actions

Goal: expose existing operational commands from the browser with the same
confirmation and evidence discipline as the TUI.

Behavior:

- The Web UI can trigger only operations that already exist in TUI or CLI.
- Work-starting actions require explicit confirmation.
- Long-running actions return an operation identifier and update visible state.
- The UI shows active, completed, blocked, and failed operation states.
- Generated specs, plans, quiz outcomes, and work reports are visible after the
  operation completes.
- The server never accepts arbitrary shell commands from the browser.
- Command surfaces mirror ChrysoCoord's existing command palette behavior and
  labels.

Tasks:

- T6.1 Inventory existing TUI/CLI actions that are safe and in scope for the
  first Web UI action set.
- T6.2 Add typed action request and response envelopes for form posts and htmx
  responses.
- T6.3 Add routes for approved spec, plan, quiz, and work actions.
- T6.4 Add operation-state reads for active and recent actions.
- T6.5 Add confirmation modals for work-starting and overwrite-prone actions.
- T6.6 Add client command palette or command surface using the same action
  model as the visible screens.
- T6.7 Add tests proving unsupported actions are rejected.
- T6.8 Add tests proving actions call existing services and write expected
  evidence artifacts.
- T6.9 Add smoke evidence for one generation flow and one work-control flow.
- T6.10 Add the slice implementation report and validation evidence.

## Slice 7: Live Sync And Conflict Handling

Goal: make TUI, CLI, and Web UI changes converge without manual browser reloads.

Behavior:

- Browser views refresh after web actions.
- Browser views notice state changes made by TUI or CLI.
- Read models include revision markers sufficient to detect stale displays.
- Stale edits are blocked with a clear conflict response.
- Active operation and event data update without full-page reload.
- Hidden or background tabs reduce refresh pressure.
- A reconnect path handles server restarts cleanly.

Tasks:

- T7.1 Choose and document the initial sync transport: polling, server-sent
  events, or a hybrid.
- T7.2 Add state revision markers for project, board, dash, and artifact views.
- T7.3 Add `/api/events` or equivalent change-feed endpoint if needed.
- T7.4 Add client refresh scheduling with backoff and hidden-tab throttling.
- T7.5 Add conflict UI for stale writes with reload and discard choices.
- T7.6 Add tests for state revision changes after CLI/service mutations.
- T7.7 Add tests for stale write rejection and successful retry.
- T7.8 Add manual two-surface smoke evidence using one TUI/CLI change and one
  Web UI change.
- T7.9 Add the slice implementation report and validation evidence.

## Slice 8: Access Control And Exposure Safety

Goal: harden the local web surface before any non-loopback use is considered
ready.

Behavior:

- Loopback remains the default.
- Non-loopback bind requires explicit host configuration or command flags.
- Startup output makes exposure state obvious.
- Token mode works for private overlay or tunnel use.
- Browser write actions use same-origin protections.
- Cross-origin access is denied unless explicitly configured.
- Logs avoid printing sensitive token values.
- Security-sensitive failures are tested.

Tasks:

- T8.1 Implement `require_token` modes: `auto`, `true`, and `false`.
- T8.2 Read token values from the configured environment variable.
- T8.3 Require token authentication when configured and for non-loopback
  auto-mode if that is the selected policy.
- T8.4 Add origin checks for write requests.
- T8.5 Add explicit allowed-origin handling.
- T8.6 Add safe startup and request logging.
- T8.7 Add tests for loopback without token, non-loopback warning, token
  required, missing token, bad token, and allowed origins.
- T8.8 Add docs for private-network and tunnel exposure boundaries without
  managing those systems.
- T8.9 Add the slice implementation report and validation evidence.

## Slice 9: Visual Identity And Layout Parity

Goal: finish the visual layer after the functional surfaces are known, keeping
layout parity with the selected reference application and ChrysoCoord identity.

Behavior:

- Board, Dash, Dumps, Quizzes, Specs, Plans, Project, Cases, Help, and command
  surfaces follow the selected reference application's layout patterns.
- The color system uses a chrysocolla/Go-adjacent palette without becoming a
  one-hue interface.
- The central avatar is a simple rounded bot/agent mark with egg-like
  proportions and ChrysoCoord colors.
- The avatar does not copy another project's mascot.
- Controls are compact, stable, and usable on desktop and mobile widths.
- Text does not overflow buttons, cards, sidebars, columns, or panels.
- Empty, loading, error, blocked, and active-operation states are visually
  complete.

Tasks:

- T9.1 Define CSS tokens for color, type, spacing, borders, shadows, and
  status colors.
- T9.2 Implement the ChrysoCoord avatar asset in the repo-native asset format
  chosen by the implementation.
- T9.3 Apply reference-layout parity to Board and Dash.
- T9.4 Apply reference-layout parity to Dumps, Quizzes, Specs, and Plans.
- T9.5 Apply reference-layout parity to Project, Cases, Help, and command
  surfaces.
- T9.6 Add responsive constraints for board columns, artifact cards, detail
  panes, navigation, and action bars.
- T9.7 Add accessibility checks for contrast, focus states, labels, and
  keyboard operation.
- T9.8 Add screenshot evidence for desktop and mobile widths across the major
  screens.
- T9.9 Add the slice implementation report and validation evidence.

## Slice 10: Docs QA And Release Readiness

Goal: complete user-facing documentation, operator guidance, and full validation
for the optional Web UI.

Behavior:

- Reference docs describe config, command flags, bind behavior, token behavior,
  routes at a high level, and action boundaries.
- How-to docs explain local browser use and explicit private-network exposure.
- TUI-only usage remains documented as the default.
- The action matrix covers Web UI parity against TUI/CLI operations.
- QA covers disabled default behavior, local serve, non-loopback warning,
  read-only views, write actions, work actions, sync, conflicts, and security.
- The release note explains the Web UI as optional.

Tasks:

- T10.1 Update `docs/reference/commands.md` for `chc web serve`.
- T10.2 Update `docs/reference/project-workspace.md` for `web` config.
- T10.3 Update `docs/reference/cli-tui-action-matrix.md` or add the Web UI
  equivalent matrix if the existing matrix cannot represent a third surface.
- T10.4 Add a how-to for running the Web UI locally.
- T10.5 Add a how-to for explicit private-network or tunnel exposure.
- T10.6 Add QA fixtures or smoke scripts for the main Web UI flows.
- T10.7 Run full repository validation.
- T10.8 Run browser smoke validation and capture evidence.
- T10.9 Update changelog or release notes according to the repository's
  existing release documentation pattern.
- T10.10 Add the slice implementation report and final rollout evidence.

## Validation Strategy

- Unit tests for config defaults, validation, view-model mapping, action envelopes,
  stale-write handling, token policy, and origin policy.
- Service tests proving web endpoints call existing application services rather
  than duplicating state mutation.
- CLI tests for `chc web serve --dry-run`, startup output, host/port override,
  invalid config, and non-loopback warnings.
- HTTP tests for route behavior, content types, HTML fragments, htmx response
  headers, static assets, cache headers, authentication, and authorization
  failures.
- Browser or DOM smoke tests for Project, Board, Dash, Dumps, Quizzes, Specs,
  Plans, Cases, Help, and command surfaces.
- Manual two-surface sync smoke: run the Web UI, mutate state through TUI or
  CLI, and verify browser refresh without server restart.
- Manual write conflict smoke: open a document, change it outside the browser,
  attempt a stale browser write, and verify conflict behavior.
- Security smoke: prove default loopback bind, explicit non-loopback warning,
  token enforcement when enabled, and no token leakage in logs.
- Final `make verify` or the repository's current full validation command.

## Open Decisions

- Whether the first live-sync transport is polling, server-sent events, or a
  hybrid.
- Whether `require_token: auto` means token-required for all non-loopback binds
  or only when a token environment variable is present.
- Whether the web server supports only the resolved current project or later
  gains a multi-project launcher.
- Which existing action-matrix document is the best durable location for
  Web/TUI/CLI parity.
- Whether visual screenshot validation is implemented through an automated
  browser runner or documented manual evidence in early slices.
- Which copied helper primitives are included in Slice 2 versus deferred until
  a later screen needs them.

## Notes

- The Web UI should be boring operational software: fast to start, local by
  default, clear about exposure, and faithful to existing ChrysoCoord state.
- The selected reference application is a design and ergonomics source, not a
  runtime dependency.
- Visual identity is intentionally late in the slicing because functional
  surfaces must exist before layout parity can be judged across the whole app.
