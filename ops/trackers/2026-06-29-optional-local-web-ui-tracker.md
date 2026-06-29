# Optional Local Web UI Tracker

Status: in_progress

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Related Spec: `ops/specs/2026-06-29-execution-unit-runtime-refinements.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | web config and serve command | ready for PR | `feat/slice-1-web-config-serve-command` | pending | `.prj/reports/slices/optional-local-web-ui/slice-1-web-config-serve-command.md` |
| Slice 2 | html htmx and helper foundation | planned | `feat/slice-2-html-htmx-helper-foundation` | pending | `.prj/reports/slices/optional-local-web-ui/slice-2-html-htmx-helper-foundation.md` |
| Slice 3 | read-only project board and dash | planned | `feat/slice-3-readonly-project-board-dash` | pending | `.prj/reports/slices/optional-local-web-ui/slice-3-readonly-project-board-dash.md` |
| Slice 4 | read-only artifact catalogs | planned | `feat/slice-4-readonly-artifact-catalogs` | pending | `.prj/reports/slices/optional-local-web-ui/slice-4-readonly-artifact-catalogs.md` |
| Slice 5 | document and metadata actions | planned | `feat/slice-5-document-and-metadata-actions` | pending | `.prj/reports/slices/optional-local-web-ui/slice-5-document-and-metadata-actions.md` |
| Slice 6 | work and generation actions | planned | `feat/slice-6-work-and-generation-actions` | pending | `.prj/reports/slices/optional-local-web-ui/slice-6-work-and-generation-actions.md` |
| Slice 7 | live sync and conflict handling | planned | `feat/slice-7-live-sync-and-conflicts` | pending | `.prj/reports/slices/optional-local-web-ui/slice-7-live-sync-and-conflicts.md` |
| Slice 8 | access control and exposure safety | planned | `feat/slice-8-web-access-control` | pending | `.prj/reports/slices/optional-local-web-ui/slice-8-web-access-control.md` |
| Slice 9 | visual identity and layout parity | planned | `feat/slice-9-web-visual-identity` | pending | `.prj/reports/slices/optional-local-web-ui/slice-9-web-visual-identity.md` |
| Slice 10 | docs qa and release readiness | planned | `feat/slice-10-web-docs-qa` | pending | `.prj/reports/slices/optional-local-web-ui/slice-10-web-docs-qa.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add the `web` config section to the project/user config model with defaults. | done | `d628644`, `fd3498a` | Disabled by default with localhost bind defaults. |
| T1.2 | Add config validation for host, port, refresh interval, token mode, and allowed origins. | done | `d628644`, `fd3498a` | Added `ValidateWebConfig` and loopback host detection. |
| T1.3 | Add CLI parsing for `chc web serve`. | done | `4ba5578` | Registered the `web serve` command without changing default TUI behavior. |
| T1.4 | Add `--host`, `--port`, `--open-browser`, `--dry-run`, and `--require-token` command flags. | done | `4ba5578` | Flags override loaded web config. |
| T1.5 | Add resolved-config output for normal startup and dry-run mode. | done | `4ba5578` | Dry-run prints project, bind, port, token mode, and healthz-only static mode. |
| T1.6 | Add non-loopback bind warning output. | done | `4ba5578`, `fd3498a` | Warning prints for `0.0.0.0` and other non-loopback hosts. |
| T1.7 | Add a minimal HTTP server boundary with `/healthz`. | done | `4ba5578`, `fd3498a` | Added foreground HTTP server with `/healthz`. |
| T1.8 | Add unit tests for config defaults, overrides, invalid values, and warning decisions. | done | `d628644`, `4ba5578`, `fd3498a` | Covered config parsing/validation and loopback/non-loopback behavior. |
| T1.9 | Add CLI tests for dry-run and command output. | done | `4ba5578`, `fd3498a` | Covered default dry-run, flag overrides, non-loopback warning, invalid config, and `/healthz`. |
| T1.10 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add an internal web package boundary for server construction. | pending | pending | Keep CLI parsing separate from server behavior. |
| T2.2 | Add `internal/webui/htmx` with the needed htmx attribute, request header, response header, and out-of-band swap helpers. | pending | pending | Copied/adapted primitives, no external module dependency. |
| T2.3 | Add `internal/webui/render` with template rendering and function-map helpers. | pending | pending | Server-rendered HTML foundation. |
| T2.4 | Add `internal/webui/ui` with the first needed UI primitives for nav, buttons, chips, links, and page layout. | pending | pending | Presentation-only helpers. |
| T2.5 | Add embedded templates and static assets for HTML, CSS, htmx, and the application shell. | pending | pending | Prefer Go embedded assets. |
| T2.6 | Add root route handling for the application shell. | pending | pending | Browser-loadable UI shell. |
| T2.7 | Add a bootstrap view model with resolved project identity, config, feature flags, and navigation model. | pending | pending | Rendered by server templates. |
| T2.8 | Add consistent HTML error page and htmx fragment error helpers. | pending | pending | Stable presentation errors for later slices. |
| T2.9 | Add graceful shutdown handling for the web command. | pending | pending | Clean process interruption. |
| T2.10 | Add helper tests for escaping, htmx attributes, htmx headers, and rendering. | pending | pending | Protect copied/adapted primitives. |
| T2.11 | Add server tests for routes, content types, cache headers, and 404s. | pending | pending | HTTP-level coverage. |
| T2.12 | Add browser or documented manual smoke evidence for the shell. | pending | pending | Avoid inventing a heavy browser stack too early. |
| T2.13 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add read view models for project summary, board column, board card, board detail, dash summary, event summary, and report summary. | pending | pending | View models are web contracts, not persistence. |
| T3.2 | Add app/service queries that reuse existing TUI or core read models. | pending | pending | Avoid reparsing state in the web layer. |
| T3.3 | Add Project, Board, board item detail, and Dash routes with full-page and htmx fragment responses. | pending | pending | Read-only routes. |
| T3.4 | Render Project, Board, and Dash screens through server templates. | pending | pending | Primary operational surfaces. |
| T3.5 | Add board-column empty states and attention states. | pending | pending | Must match existing status semantics. |
| T3.6 | Add detail panes or detail pages for board items. | pending | pending | No mutation in this slice. |
| T3.7 | Add tests for view-model shape, status mapping, empty project state, and attention reasons. | pending | pending | Protect the read contract. |
| T3.8 | Add screenshot or DOM smoke evidence for Project, Board, and Dash. | pending | pending | Evidence can be automated or manual for this slice. |
| T3.9 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add read view models for case, dump, quiz, spec, plan, and generic document previews. | pending | pending | Shared document-card model where useful. |
| T4.2 | Add service queries for artifact catalogs using existing discovery/index behavior. | pending | pending | No browser-only artifact discovery. |
| T4.3 | Add Cases, Dumps, Quizzes, Specs, and Plans list routes with full-page and htmx fragment responses. | pending | pending | Catalog routes. |
| T4.4 | Add detail routes for each artifact family. | pending | pending | Path-confined document reads. |
| T4.5 | Add safe Markdown rendering or pre-rendered HTML handling consistent with existing dependencies. | pending | pending | No unsafe raw HTML by accident. |
| T4.6 | Add catalog and detail screens for Cases, Dumps, Quizzes, Specs, and Plans. | pending | pending | Whole operational app, not board only. |
| T4.7 | Add tests for empty catalogs, malformed artifact metadata, missing files, and path confinement. | pending | pending | Error behavior is part of the contract. |
| T4.8 | Add screenshot or DOM smoke evidence for all artifact catalog screens. | pending | pending | Capture layout coverage. |
| T4.9 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Define the write-response envelope for success, validation failure, stale edit, and blocked operation. | pending | pending | Stable client/server contract. |
| T5.2 | Add revision tokens to relevant read view models. | pending | pending | Required for stale-write rejection. |
| T5.3 | Add same-service write endpoints for supported metadata updates. | pending | pending | Do not duplicate write logic in web handlers. |
| T5.4 | Add dump create, append, and edit endpoints if matching CLI/TUI operations exist. | pending | pending | Keep action set limited to existing operations. |
| T5.5 | Add browser confirmation flows for destructive or overwrite-prone actions. | pending | pending | Confirm before write. |
| T5.6 | Add stale-write tests and path-confinement tests. | pending | pending | Protect shared-state integrity. |
| T5.7 | Add client refresh behavior after successful writes. | pending | pending | UI must show persisted state. |
| T5.8 | Add manual or automated smoke evidence for create, edit, stale edit, and validation failure. | pending | pending | Covers success and failure paths. |
| T5.9 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Inventory existing TUI/CLI actions that are safe and in scope for the first Web UI action set. | pending | pending | Action inventory must be committed before endpoints expand. |
| T6.2 | Add typed action request and response envelopes for form posts and htmx responses. | pending | pending | No arbitrary command execution. |
| T6.3 | Add routes for approved spec, plan, quiz, and work actions. | pending | pending | Existing operations only. |
| T6.4 | Add operation-state reads for active and recent actions. | pending | pending | Needed for long-running behavior. |
| T6.5 | Add confirmation modals for work-starting and overwrite-prone actions. | pending | pending | Same confirmation discipline as TUI. |
| T6.6 | Add client command palette or command surface using the same action model as the visible screens. | pending | pending | Browser command surface. |
| T6.7 | Add tests proving unsupported actions are rejected. | pending | pending | Negative path first. |
| T6.8 | Add tests proving actions call existing services and write expected evidence artifacts. | pending | pending | Evidence must match CLI/TUI behavior. |
| T6.9 | Add smoke evidence for one generation flow and one work-control flow. | pending | pending | Browser operation smoke. |
| T6.10 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 7 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T7.1 | Choose and document the initial sync transport: polling, server-sent events, or a hybrid. | pending | pending | Decision belongs in the slice report or docs. |
| T7.2 | Add state revision markers for project, board, dash, and artifact views. | pending | pending | Used by refresh and conflict detection. |
| T7.3 | Add `/api/events` or equivalent change-feed endpoint if needed. | pending | pending | Depends on selected sync transport. |
| T7.4 | Add client refresh scheduling with backoff and hidden-tab throttling. | pending | pending | Keep local server cheap. |
| T7.5 | Add conflict UI for stale writes with reload and discard choices. | pending | pending | No silent overwrite. |
| T7.6 | Add tests for state revision changes after CLI/service mutations. | pending | pending | Proves cross-surface sync. |
| T7.7 | Add tests for stale write rejection and successful retry. | pending | pending | Full conflict loop. |
| T7.8 | Add manual two-surface smoke evidence using one TUI/CLI change and one Web UI change. | pending | pending | Required sync evidence. |
| T7.9 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 8 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T8.1 | Implement `require_token` modes: `auto`, `true`, and `false`. | pending | pending | Match spec config. |
| T8.2 | Read token values from the configured environment variable. | pending | pending | Default token env is defined by spec. |
| T8.3 | Require token authentication when configured and for non-loopback auto-mode if that is the selected policy. | pending | pending | Final policy decision required. |
| T8.4 | Add origin checks for write requests. | pending | pending | Same-origin protection. |
| T8.5 | Add explicit allowed-origin handling. | pending | pending | Only configured origins allowed. |
| T8.6 | Add safe startup and request logging. | pending | pending | Do not leak tokens. |
| T8.7 | Add tests for loopback without token, non-loopback warning, token required, missing token, bad token, and allowed origins. | pending | pending | Security regression coverage. |
| T8.8 | Add docs for private-network and tunnel exposure boundaries without managing those systems. | pending | pending | Operator-owned exposure. |
| T8.9 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 9 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T9.1 | Define CSS tokens for color, type, spacing, borders, shadows, and status colors. | pending | pending | Chrysocolla/Go-adjacent identity. |
| T9.2 | Implement the ChrysoCoord avatar asset in the repo-native asset format chosen by the implementation. | pending | pending | Simple rounded bot/agent mark. |
| T9.3 | Apply reference-layout parity to Board and Dash. | pending | pending | Layout and density pass. |
| T9.4 | Apply reference-layout parity to Dumps, Quizzes, Specs, and Plans. | pending | pending | Document surfaces. |
| T9.5 | Apply reference-layout parity to Project, Cases, Help, and command surfaces. | pending | pending | Remaining app surfaces. |
| T9.6 | Add responsive constraints for board columns, artifact cards, detail panes, navigation, and action bars. | pending | pending | Prevent text and layout overflow. |
| T9.7 | Add accessibility checks for contrast, focus states, labels, and keyboard operation. | pending | pending | Usability gate. |
| T9.8 | Add screenshot evidence for desktop and mobile widths across the major screens. | pending | pending | Visual QA evidence. |
| T9.9 | Add the slice implementation report and validation evidence. | pending | pending | Report path is defined in the slice table. |

## Slice 10 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T10.1 | Update `docs/reference/commands.md` for `chc web serve`. | pending | pending | Command contract. |
| T10.2 | Update `docs/reference/project-workspace.md` for `web` config. | pending | pending | Config contract. |
| T10.3 | Update `docs/reference/cli-tui-action-matrix.md` or add the Web UI equivalent matrix if the existing matrix cannot represent a third surface. | pending | pending | Parity tracking. |
| T10.4 | Add a how-to for running the Web UI locally. | pending | pending | Local browser use. |
| T10.5 | Add a how-to for explicit private-network or tunnel exposure. | pending | pending | Exposure is operator-owned. |
| T10.6 | Add QA fixtures or smoke scripts for the main Web UI flows. | pending | pending | Repeatable verification. |
| T10.7 | Run full repository validation. | pending | pending | Expected final validation gate. |
| T10.8 | Run browser smoke validation and capture evidence. | pending | pending | Browser-level confidence. |
| T10.9 | Update changelog or release notes according to the repository's existing release documentation pattern. | pending | pending | Optional Web UI release note. |
| T10.10 | Add the slice implementation report and final rollout evidence. | pending | pending | Final rollout closeout. |

## Validation Log

- Slice 1 T1.1-T1.2: `go test ./internal/core -run 'WebConfig|WorkspaceConfig'` (pass)
- Slice 1 T1.3-T1.9: `go test ./internal/core -run 'WebConfig|WorkspaceConfig' && go test ./internal/app -run 'RootCommand|WebServe'` (pass)
- Slice 1: `go test ./internal/core ./internal/app` (pass)
- Slice 1: `make verify` (pass)

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

- This tracker starts with planning only. Runtime implementation begins after
  the plan and tracker are committed to `dev`.
- The Web UI must stay disabled by default until the rollout explicitly changes
  that behavior.
- Every slice should preserve TUI and CLI behavior.
