# Changelog

Notable project changes are summarized here for users and maintainers.

This changelog groups changes by capability. It intentionally omits mechanical
slice bookkeeping, lint-only edits, and minor wording-only commits.

## Unreleased

Compared with `v0.1.0` (`c9008cb`, tagged 2026-05-23) through `741271f` on
`dev`.

### Added

- Create use cases directly from the Web Cases page. Enter a name, description
  and context; the new case is immediately available when capturing a dump.

- Projects can select Codex or Pi independently for workers and content
  generation. Web, TUI and confirmed Telegram controls offer named defaults,
  local readiness and revision-checked saves without a service restart. Prepared
  Pi workers run in Docker with private call state and explicit credentials. Admitted work
  keeps its selected runtime when project configuration changes.

- Protect the Web UI with one locally configured account. Log in once and
  navigate without URL tokens; sessions expire after 12 hours and can be
  revoked with Log out. Private terminal setup keeps passwords out of command
  arguments and project files.

- Web Dash now shows the pending queue, independent active ticket contexts and
  retained results together, with a stacked layout on narrow screens. Quiet
  validation stays visible, and confirmed delivery remains distinct from host
  validation or a successful harness run. Refresh preserves scroll and selection;
  temporary outages keep the last observation with a stale notice. Preview Work
  and Stop Daemon remain available while monitoring updates.
- Plans now offers Generate Plan directly. One submission prepares the plan,
  its task tickets and the internal order map; no separate tracking or ticket
  approval step is needed. Existing plans and ticket identities survive retries,
  and authoring does not start execution.
  Plans can also be generated or regenerated from a named case's saved context,
  with guidance when discovery is absent or insufficient. A useful dump or
  answered quiz can supply the context without requiring both. Plan readiness
  uses the existing five-minute authoring budget through the selected harness
  instead of stopping after one minute. Failed generation and intervening edits
  keep the existing plan available; spec approval and
  starting work remain separate actions. Every generated revision now retains
  its own internal order map. Approved native materialization binds task tickets
  once, and approved amendments retain completed work and earlier maps. Ticket
  progress leaves the map unchanged; missing or edited required maps must be
  repaired before workers can prepare their context. The Web catalog shows saved
  plans as fixed-size opening-text cards with complete read-only details; internal
  trackers stay out of Web browsing and creation.
- Specs now supports manual draft or direct-approved creation and editing of
  individual manual or generated documents from their cards. Named scope
  selection, one Body editor and a compact Draft control retain identity,
  provenance and native state. Stale saves keep the form available, and generated
  output cannot replace an edit made while it was being prepared.
  Generate now selects an existing named use case and combines its saved dumps
  and quiz answers with shared project context, with progress feedback during
  generation. Project-wide context is not offered as a generation target.
  A successful generation opens its saved document in the editor. Generation
  results and errors remain available until explicitly closed. Valid workspace
  protocol references survive publication without duplicating their namespace.
- Quizzes now provide a focused Web answer screen with one question and a
  spacious answer editor in a single panel, a Deep/Fast segmented selector and
  a compact action row. Opening a new quiz prepares its first question directly,
  with progress feedback and retry on failure. Answer history appears only after
  answers exist. Drafts
  survive refresh, mode changes and request errors; leaving or skipping with
  unsent text requires a discard decision. Shared actions reject stale questions
  and acknowledge retries once. Fast retains one generated batch across mode
  changes and restart; failed continuation keeps the confirmed answer available
  in history. Cancel and Done preserve earlier answers with distinct statuses.
  Deep discovery retains the full scoped answer history, honors completion
  recommendations and stops at its safety budget. Repeated answered questions
  are rejected without losing the latest answer.
  The catalog shows Project Quiz and answered case cards with fixed opening-text
  previews. New Quiz lists only unanswered named scopes, preserves unfinished
  sessions and disappears when every scope has answers.

- Project-wide context now grows in one shared dump across Web, CLI, TUI and
  authoring conversations. Web New offers Project-wide or existing case names;
  case captures remain independent. Retries preserve confirmed captures, edits
  retain raw material and prior versions, and explicit reconciliation preserves
  older documents and their references. Dump editing uses one Body field and
  Save, retains pending edits after failures, and requires reloading before
  replacing newer changes.
- The Board now provides one New/Edit ticket form with full descriptions,
  classification, ownership and delivery references. Maximize and restore retain
  unsaved input, and comments post independently. Status selection and dragging
  reject stale or unsupported moves with an explanation. Canonical, legacy and
  native tickets share six readable lanes, with paging for larger projects.
  Plan-bound task titles now display their ownership guard before a rename can
  be submitted; permitted metadata and independent comments remain editable.
- Telegram conversations can explain project context, collect requirements,
  answer clarification questions, generate and refine dumps, quizzes, draft
  specifications and plans, and prepare native backlog tickets. Conversations
  retain their project and namespace across restart. Development authority binds
  the prepared scope; changed requirements need a new explicit decision. Confirmed
  native work continues through its configured review and integration policy
  without another message. Progress shows checked work, confirmed receipts and
  remaining tickets, with the same invocation available in CLI, TUI and Web.
- Ticket-native CLI, TUI and Web controls can preview and run exact tickets, slices,
  slice sets, use cases and loose work, then inspect or continue retained
  executions. TUI and Web confirmations retain the selected revision and show stale
  conflicts, capacity waits and validated versus integrated progress. Approval
  uses the displayed candidate binding and native invocation, entry, budget and
  candidate guards. Native status shows each target queue, actual occupied role,
  remaining attempts, token accounting and precise blocker or question. Work and
  daemon commands, Web forms and manual service callers can extend a retained
  allowance or explicitly resolve/defer a bound question; stale actions fail
  without refilling consumed work or granting publication. Supervised CLI manual handoffs require
  host validation before progress advances. The daemon continues retained
  authorization with the same lifecycle and status; startup preserves operator
  stops and human approval waits without admitting the ready backlog.
- Workflow configuration can select versioned defaults, report effective values
  and their sources, and retain human approval restrictions for native work.
  Invalid policy blocks execution; admission freezes execution limits while
  publication rechecks current project policy. New workspaces in every profile
  select ticket authority, isolated worktrees, trusted automatic review and
  one worker. Native source rewriting requires explicit `git_atomic_dual_ref`
  opt-in and current publisher/policy support; initialization and upgrade leave
  an absent profile disabled. Reinitialization preserves existing policy, including absent config.
  Existing projects can upgrade explicitly with a prior-policy snapshot and a
  deliberate authority choice. Upgrade preserves review policy, rejects active
  ownership and adopts only selected legacy work while retaining its history.
  Optional trackers export native progress with source and receipt provenance.
  Failed or stale views preserve progress; repair keeps operator edits unless
  replacement is explicitly authorized for their observed bytes.
- Ticket-native work can run up to eight independent lanes in parallel when
  configured, with one worker by default. Ordered work and target publication
  remain serialized, and capacity waits retain progress for retry.
- Added a Docker-backed ticket execution path with execution records, runtime
  context artifacts, result envelopes, validation output, changed-file
  detection, result application, and execution reports.
- Added isolated ticket worktree execution so agent runs can prepare
  ticket-specific Git worktrees, mount the isolated workspace into Docker,
  preserve execution workspace metadata, and keep the user's active checkout
  separate from agent edits.
- Added foreground ticket loops, loop run records, loop reports, filtered
  ticket selection, and ready-queue ordering.
- Added Dash agent-work controls in the TUI, including selected-ticket work,
  ready-loop execution, run-all-ready execution, stop requests, queue filters,
  and runtime observability.
- Added agent work event streams, human-readable Dash event rendering, CLI event
  readback with degraded-stream handling, and execution log lookup.
- Added configurable work readiness with project-local branch policy, work
  diagnostics, safe repair planning, safe repair execution, and Dash repair
  guidance.
- Added ticket review and merge workflows, including pull request metadata,
  review mode configuration, provider diagnostics, remote review adapters,
  review result application, repair handoff prioritization, merge preflight,
  provider merge execution, and merge metadata.
- Added a serialized ticket integration loop that completes development,
  review, bounded repair, and merge for the current ticket before selecting
  later ready work.
- Added serialized integration-loop observability for cycle rehydration,
  review, repair, and merge phases, plus Dash phase display and loop stop
  reasons.
- Added Gherkin acceptance feature generation as a separate workflow from spec
  generation, including feature-round prompts, feature artifact checks, CLI and
  TUI actions, and documentation.
- Added QA reference projects and reset/test/smoke tooling for blog, dialer, and
  tictactoe agent-work scenarios.
- Added the tictactoe QA reference and preserved a successful completed agent
  implementation as versioned evidence.
- Added dev-to-main gate tooling, gate reports, policy configuration, and
  Woodpecker workflow generation.
- Added maintenance and recovery commands for diagnostic bundles, SQLite
  migration state, project registry drift, registry repair, execution artifact
  drift, and runtime artifact cleanup previews.
- Added notification and remote command intake support for Telegram-backed
  setup checks, notification audit records, loop/ticket notifications, and
  fixture-based remote command handoff parsing.
- Added a first usable iteration of the optional local Web UI served by
  `chr web serve`, including localhost-only defaults, token and origin
  controls, server-rendered pages, htmx refresh behavior, document and work
  actions, chr visual identity, documentation, and browser smoke
  tooling.
- Added Markdown-first authoring sessions and assisted authoring flows for
  generating and promoting project documentation artifacts.
- Added workspace graph/index work, semantic digest support, cache boundaries,
  and TUI consumers for project-local Markdown artifacts.
- Added extensive Diataxis documentation for commands, TUI behavior, runtime,
  ticket execution, reporting, review, merge, maintenance, notifications,
  acceptance generation, and QA workflows.

### Changed

- The product name is chr. Commands continue to use `chr`.

- chr is now licensed under Apache License 2.0.
- Web document catalogs share the Specs and Plans layout, with consistent
  margins, fixed-size cards and opening-text previews across screen sizes.
- Web UI page headings show the section or artifact title, with the project
  name kept in the shared header.
- The Project page is available by direct URL and shows only the project name,
  root directory and diagnostic ID. Removed its navigation links, internal
  paths, server settings and automatic refresh.
- The Web UI supports direct, token-free access from another computer on a
  trusted internal network with `chr web serve --allow-lan`. This mode is
  disabled by default and keeps navigation, forms and live updates usable
  without token-bearing links.
- Telegram `/help` groups all available commands by category, with complete
  arguments on separate readable lines and a plain-text fallback when Telegram
  rejects the formatting.
- Updated the Web UI with editable ticket details and comments, shared case
  editing panels, and dialogs to create, edit and append to dumps. Successful
  dump saves close the dialog and keep the catalog open.
- Moved the application code into clearer `internal/app`, `internal/core`, and
  `internal/tui` package boundaries.
- Reworked generated specs so spec generation remains behavior-focused and no
  longer embeds Gherkin feature output.
- Reworked project and case metadata toward Markdown-first project-local
  records instead of treating SQLite indexes as authoritative.
- Reworked branch/readiness checks so agent work and merge preflight use
  configured work policy instead of hardcoded branch names.
- Reworked ready-ticket loop execution so later tickets wait for the current
  ticket to pass review policy and merge, or block after the finite repair
  limit.
- Expanded the TUI from project/case editing into a broader operational console
  for documents, specs, plans, Board, Dash, commands, and runtime state.
- Expanded command coverage so CLI and TUI work actions route through shared
  service paths where both surfaces exist.
- Updated workspace initialization profiles, protocol assets, runtime
  Dockerfiles, and generated project files to match the current workflow.
- Migrated chr operational artifacts to the canonical `ops/<namespace>/...`
  layout, using `ops/default` for no-user workspaces, while preserving `.chr/`
  for CHR-private project configuration and state.

### Fixed

- Plan generation recognizes saved spec approval even when older text still
  calls the document a draft. Newly generated specs keep review state in
  metadata instead of embedding approval disclaimers in the body.

- The Dumps catalog fits a third card column at normal desktop zoom. Cards
  keep the same size for short and long documents, with clipped opening text
  in a pale green inset preview. Narrow windows reduce the column count.

- Prevented outdated ticket forms and comments from changing a ticket after
  another edit. Reload the board before submitting again.
- Made failed Web UI form submissions display their error and reload actions
  inside the current dialog.
- Fixed acceptance feature generation to preserve feature-round answers and
  satisfy Codex feature schema expectations.
- Fixed spec generation context so generated specs stay neutral and do not
  accidentally produce feature artifacts.
- Fixed work readiness so chr-owned runtime artifacts can be repaired
  safely without treating them as user-owned dirty state.
- Fixed work readiness diagnostics so missing branches, missing remote refs,
  behind-remote branches, invalid policy, dirty worktrees, and Git in-progress
  states produce actionable reports instead of raw Git failures.
- Fixed merge preflight defaults so the expected pull request base comes from
  configured work policy.
- Fixed multiple TUI and acceptance workflow lint issues while keeping local
  strict lint rules intact.
- Fixed pre-release lint gate blockers in ops artifact allocation, runtime
  planning, report listing, proving logs, and ticket result application paths.

## v0.1.0 - 2026-05-23

Initial tagged release.
