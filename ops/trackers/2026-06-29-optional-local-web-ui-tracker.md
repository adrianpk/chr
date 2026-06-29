# Optional Local Web UI Tracker

Status: in_progress

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Related Spec: `ops/specs/2026-06-29-execution-unit-runtime-refinements.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | web config and serve command | delivered | `feat/slice-1-web-config-serve-command` | `#343` | `.prj/reports/slices/optional-local-web-ui/slice-1-web-config-serve-command.md` |
| Slice 2 | html htmx and helper foundation | delivered | `feat/slice-2-html-htmx-helper-foundation` | `#344` | `.prj/reports/slices/optional-local-web-ui/slice-2-html-htmx-helper-foundation.md` |
| Slice 3 | read-only project board and dash | delivered | `feat/slice-3-readonly-project-board-dash` | `#345` | `.prj/reports/slices/optional-local-web-ui/slice-3-readonly-project-board-dash.md` |
| Slice 4 | read-only artifact catalogs | delivered | `feat/slice-4-readonly-artifact-catalogs` | `#346` | `.prj/reports/slices/optional-local-web-ui/slice-4-readonly-artifact-catalogs.md` |
| Slice 5 | document and metadata actions | delivered | `feat/slice-5-document-and-metadata-actions` | `#347` | `.prj/reports/slices/optional-local-web-ui/slice-5-document-and-metadata-actions.md` |
| Slice 6 | work and generation actions | delivered | `feat/slice-6-work-and-generation-actions` | `#348` | `.prj/reports/slices/optional-local-web-ui/slice-6-work-and-generation-actions.md` |
| Slice 7 | live sync and conflict handling | delivered | `feat/slice-7-live-sync-and-conflicts` | `#349` | `.prj/reports/slices/optional-local-web-ui/slice-7-live-sync-and-conflicts.md` |
| Slice 8 | access control and exposure safety | reviewing | `feat/slice-8-web-access-control` | pending | `.prj/reports/slices/optional-local-web-ui/slice-8-web-access-control.md` |
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
| T2.1 | Add an internal web package boundary for server construction. | done | `f85f4ba`, `032ef29` | Added `internal/webui` handler boundary consumed by `internal/app`. |
| T2.2 | Add `internal/webui/htmx` with the needed htmx attribute, request header, response header, and out-of-band swap helpers. | done | `f85f4ba`, `032ef29` | Copied/adapted primitives, no external module dependency. |
| T2.3 | Add `internal/webui/render` with template rendering and function-map helpers. | done | `f85f4ba`, `032ef29` | Server-rendered HTML foundation. |
| T2.4 | Add `internal/webui/ui` with the first needed UI primitives for nav, buttons, chips, links, and page layout. | done | `f85f4ba`, `032ef29` | Presentation-only helpers. |
| T2.5 | Add embedded templates and static assets for HTML, CSS, htmx, and the application shell. | done | `f85f4ba` | Embedded `chc.css`, `htmx.min.js`, and shell templates. |
| T2.6 | Add root route handling for the application shell. | done | `f85f4ba` | `/` renders the browser-loadable shell. |
| T2.7 | Add a bootstrap view model with resolved project identity, config, feature flags, and navigation model. | done | `f85f4ba` | Shell data includes project, bind, token, refresh, and navigation state. |
| T2.8 | Add consistent HTML error page and htmx fragment error helpers. | done | `f85f4ba` | Initial render errors return stable HTTP errors; htmx fragment route added for bootstrap. |
| T2.9 | Add graceful shutdown handling for the web command. | done | `f85f4ba` | Reuses Slice 1 foreground server shutdown path. |
| T2.10 | Add helper tests for escaping, htmx attributes, htmx headers, and rendering. | done | `f85f4ba`, `032ef29` | Added tests under `internal/webui/...`. |
| T2.11 | Add server tests for routes, content types, cache headers, and 404s. | done | `f85f4ba`, `032ef29` | Added shell, asset, bootstrap, and health route tests. |
| T2.12 | Add browser or documented manual smoke evidence for the shell. | done | `f85f4ba` | Manual HTTP smoke covered `/`, CSS, htmx asset, and `/healthz`. |
| T2.13 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add read view models for project summary, board column, board card, board detail, dash summary, event summary, and report summary. | done | `bf49824`, `93611f7` | Added Web UI read contracts for Project, Board, Dash, readiness, queue, events, loop, and execution summaries. |
| T3.2 | Add app/service queries that reuse existing TUI or core read models. | done | `bf49824`, `93611f7` | Web provider reuses `service.Board()`, TUI readiness, and TUI observability projections. |
| T3.3 | Add Project, Board, board item detail, and Dash routes with full-page and htmx fragment responses. | done | `4573193`, `93611f7` | Added `/project`, `/board`, `/dash`, and fragment routes. |
| T3.4 | Render Project, Board, and Dash screens through server templates. | done | `4573193` | Shared shell renders read-only Project, Board, and Dash content. |
| T3.5 | Add board-column empty states and attention states. | done | `4573193` | Board columns render empty states and attention cards. |
| T3.6 | Add detail panes or detail pages for board items. | done | `4573193` | Ticket cards expose key metadata without mutation. |
| T3.7 | Add tests for view-model shape, status mapping, empty project state, and attention reasons. | done | `896331b`, `93611f7` | Added route tests and provider-backed Board smoke test. |
| T3.8 | Add screenshot or DOM smoke evidence for Project, Board, and Dash. | done | `896331b` | Manual DOM smoke covered full pages, fragments, and CSS. |
| T3.9 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add read view models for case, dump, quiz, spec, plan, and generic document previews. | done | `90f6962` | Shared `ArtifactCatalogView`, `ArtifactCardView`, and `ArtifactDetailView`. |
| T4.2 | Add service queries for artifact catalogs using existing discovery/index behavior. | done | `90f6962` | Reuses case metadata service and existing dump, quiz, spec, and plan loaders. |
| T4.3 | Add Cases, Dumps, Quizzes, Specs, and Plans list routes with full-page and htmx fragment responses. | done | `087a454` | Added family catalog routes and htmx fragments. |
| T4.4 | Add detail routes for each artifact family. | done | `087a454` | Detail routes use catalog IDs and path-confined reads. |
| T4.5 | Add safe Markdown rendering or pre-rendered HTML handling consistent with existing dependencies. | done | `087a454` | Markdown is rendered as escaped text in `<pre>`, with no raw HTML rendering. |
| T4.6 | Add catalog and detail screens for Cases, Dumps, Quizzes, Specs, and Plans. | done | `087a454` | Shared catalog cards and detail content templates. |
| T4.7 | Add tests for empty catalogs, malformed artifact metadata, missing files, and path confinement. | done | `100f63a` | Handler and app-provider tests cover empty/missing states, malformed case metadata, and confinement. |
| T4.8 | Add screenshot or DOM smoke evidence for all artifact catalog screens. | done | `100f63a` | DOM smoke covered all catalog screens, detail page, and htmx fragments. |
| T4.9 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Define the write-response envelope for success, validation failure, stale edit, and blocked operation. | done | `b35c318` | Added `WriteActionRequest` and `WriteActionResponse` with stable status values. |
| T5.2 | Add revision tokens to relevant read view models. | done | `b35c318` | Added revision fields for artifact cards/details; case metadata uses source hash and dump details use content hash. |
| T5.3 | Add same-service write endpoints for supported metadata updates. | done | `b35c318` | Case metadata updates call `PatchCaseMetadata`. |
| T5.4 | Add dump create, append, and edit endpoints if matching CLI/TUI operations exist. | done | `b35c318` | Dump actions call `CreateDump`, `SaveDump`, and `AppendDump`. |
| T5.5 | Add browser confirmation flows for destructive or overwrite-prone actions. | done | `668dccc` | Forms require browser confirmation before submit. |
| T5.6 | Add stale-write tests and path-confinement tests. | done | `44b9b30` | Provider tests cover stale dump saves and existing path confinement tests remain in place. |
| T5.7 | Add client refresh behavior after successful writes. | done | `668dccc` | Successful writes redirect to refreshed catalog/detail pages; htmx uses `HX-Redirect`. |
| T5.8 | Add manual or automated smoke evidence for create, edit, stale edit, and validation failure. | done | `44b9b30` | HTTP smoke covered dump create, case update, dump save, stale rejection, and blocked envelope tests. |
| T5.9 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Inventory existing TUI/CLI actions that are safe and in scope for the first Web UI action set. | done | `9cab5c4` | Inventory committed before endpoint expansion. |
| T6.2 | Add typed action request and response envelopes for form posts and htmx responses. | done | `d6dd4ac` | Extended write response with operation state. |
| T6.3 | Add routes for approved spec, plan, quiz, and work actions. | done | `d6dd4ac` | Added dispatch for spec, plan, quiz, work preview, and daemon stop actions through `/actions`. |
| T6.4 | Add operation-state reads for active and recent actions. | done | `d6dd4ac` | Action results include operation kind, status, id, and detail. |
| T6.5 | Add confirmation modals for work-starting and overwrite-prone actions. | done | `569e8b1` | All generation/work forms require browser confirmation before submit. |
| T6.6 | Add client command palette or command surface using the same action model as the visible screens. | done | `569e8b1` | Added command forms to Specs, Plans, Quizzes, and Dash. |
| T6.7 | Add tests proving unsupported actions are rejected. | done | `4424a37` | Unsupported work-start action returns blocked envelope. |
| T6.8 | Add tests proving actions call existing services and write expected evidence artifacts. | done | `4424a37` | Spec generation test uses service fake model and verifies generated artifact; work preview verifies no executions. |
| T6.9 | Add smoke evidence for one generation flow and one work-control flow. | done | `4424a37` | Targeted tests cover generation; HTTP smoke covers command surfaces, work preview, and unsupported action rejection. |
| T6.10 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

## Slice 7 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T7.1 | Choose and document the initial sync transport: polling, server-sent events, or a hybrid. | done | `0d4f68f` | htmx polling selected; SSE and `/api/events` deferred. |
| T7.2 | Add state revision markers for project, board, dash, and artifact views. | done | `133be32` | Full pages and fragments carry stable revision markers. |
| T7.3 | Add `/api/events` or equivalent change-feed endpoint if needed. | done | `0d4f68f` | Not added because polling was selected. |
| T7.4 | Add client refresh scheduling with backoff and hidden-tab throttling. | done | `133be32` | htmx polling uses the configured refresh interval and a visible-tab trigger guard. |
| T7.5 | Add conflict UI for stale writes with reload and discard choices. | done | `133be32` | Stale responses include reload and discard links. |
| T7.6 | Add tests for state revision changes after CLI/service mutations. | done | `133be32` | Provider tests mutate the workspace outside the Web UI and assert revision changes. |
| T7.7 | Add tests for stale write rejection and successful retry. | done | `133be32`, `c7cb387` | Tests cover stale rejection, reload, and retry with the refreshed revision. |
| T7.8 | Add manual two-surface smoke evidence using one TUI/CLI change and one Web UI change. | done | `89c5997` | Temporary-project smoke covered external file mutation and Web UI save. |
| T7.9 | Add the slice implementation report and validation evidence. | done | `89c5997`, `51fd872` | Report path is defined in the slice table. |

## Slice 8 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T8.1 | Implement `require_token` modes: `auto`, `true`, and `false`. | done | `dea4109` | `always` remains accepted for compatibility with existing validation. |
| T8.2 | Read token values from the configured environment variable. | done | `dea4109` | Required tokens are read from `token_env` and missing values block startup. |
| T8.3 | Require token authentication when configured and for non-loopback auto-mode if that is the selected policy. | done | `dea4109` | `auto` requires token on non-loopback and stays token-free on loopback. |
| T8.4 | Add origin checks for write requests. | done | `dea4109` | Origin checks apply to browser requests before route handling. |
| T8.5 | Add explicit allowed-origin handling. | done | `dea4109` | Exact configured origins are allowed and receive CORS preflight headers. |
| T8.6 | Add safe startup and request logging. | done | `dea4109` | Startup prints token requirement and origins; request logs omit query strings and token values. |
| T8.7 | Add tests for loopback without token, non-loopback warning, token required, missing token, bad token, and allowed origins. | done | `dea4109` | App and handler tests cover token/origin policies and sanitized logs. |
| T8.8 | Add docs for private-network and tunnel exposure boundaries without managing those systems. | done | pending | Added deliberate exposure how-to. |
| T8.9 | Add the slice implementation report and validation evidence. | done | this commit | Report path is defined in the slice table. |

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
- Slice 2 T2.1-T2.12: `go test ./internal/webui/... ./internal/app -run 'WebServe|Handler|Attrs|Button|Chip|Nav|Renderer'` (pass)
- Slice 2: `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Slice 2 smoke: `go run ./cmd/chc web serve --dir . --port 7332`; `curl -fsS /`, `/assets/chc.css`, `/assets/htmx.min.js`, `/healthz` (pass)
- Slice 2: `make verify` (pass)
- Slice 3 T3.1-T3.8: `go test ./internal/webui ./internal/app -run 'WebServe|Handler|Project|Board|Dash'` (pass)
- Slice 3: `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Slice 3 smoke: `go run ./cmd/chc web serve --dir . --port 7333`; `curl -fsS /`, `/project`, `/board`, `/dash`, `/project/fragment`, `/board/fragment`, `/dash/fragment`, `/assets/chc.css` (pass)
- Slice 3: `make verify` (pass)
- Slice 4 T4.1-T4.7: `go test ./internal/webui ./internal/app -run 'WebServe|Handler|Catalog|Artifact'` (pass)
- Slice 4: `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Slice 4 smoke: `go run ./cmd/chc web serve --dir . --port 7334`; `curl -fsS /cases`, `/dumps`, `/quizzes`, `/specs`, `/plans`, one `/dumps/detail`, catalog fragments, and one detail fragment (pass)
- Slice 4: `make verify` (pass)
- Slice 5 T5.1-T5.8: `go test ./internal/webui ./internal/app -run 'WebServe|Handler|Catalog|Artifact|WriteAction'` (pass)
- Slice 5: `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Slice 5 smoke: temporary project served with `go run ./cmd/chc web serve --dir /tmp/tmp.I8grD0CG1y --port 7335`; `curl` covered dump create, case update, dump save, and stale dump rejection (pass)
- Slice 5: `make verify` (pass)
- Slice 6 T6.1-T6.9: `go test ./internal/webui ./internal/app -run 'WebServe|Handler|WriteAction|Generation|WorkAction'` (pass)
- Slice 6: `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Slice 6 smoke: temporary project served with `go run ./cmd/chc web serve --dir /tmp/tmp.IJQmMJO7zb --port 7336`; `curl` covered command forms, `work.preview`, and unsupported work-start rejection (pass)
- Slice 6: `make verify` (pass)
- Slice 7 T7.1-T7.7: `go test ./internal/webui && go test ./internal/app -run 'WebServe.*(Artifact|WriteAction|Revision|Handler|Board|Catalog)'` (pass)
- Slice 7 smoke: temporary project served with `go run ./cmd/chc web serve --dir /tmp/... --port 7337`; `curl` covered healthz, detail revision marker, external file mutation changing htmx fragment revision, and Web UI `dump.save` redirect/write (pass)
- Slice 7: `make verify` (pass)
- Slice 8 T8.1-T8.7: `go test ./internal/webui && go test ./internal/app -run 'WebServe|RootCommand'` (pass)
- Slice 8 smoke: temporary project served with `CHC_WEB_TOKEN=... go run ./cmd/chc web serve --dir /tmp/... --host 0.0.0.0 --port 7338`; `curl` covered token-required access, allowed-origin preflight, disallowed-origin rejection, wildcard-bind same-origin access, startup policy output, and sanitized logs (pass)
- Slice 8: `make verify` (pass)

## Open Decisions

- Whether the first live-sync transport is polling, server-sent events, or a
  hybrid.
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
