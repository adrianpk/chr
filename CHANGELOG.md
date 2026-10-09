# Changelog

Notable project changes are summarized here for users and maintainers.

This changelog groups changes by capability. It intentionally omits mechanical
slice bookkeeping, lint-only edits, and minor wording-only commits.

## Unreleased

Compared with `v0.1.0` (`c9008cb`, tagged 2026-05-23) through `741271f` on
`dev`.

### Added

- Ticket-native CLI, TUI and Web controls can preview and run exact tickets, slices,
  slice sets, use cases and loose work, then inspect or continue retained
  executions. TUI and Web confirmations retain the selected revision and show stale
  conflicts, capacity waits and validated versus integrated progress. Approval
  uses the displayed candidate binding; supervised CLI manual handoffs require
  host validation before progress advances. The daemon continues retained
  authorization with the same lifecycle and status; startup preserves operator
  stops and human approval waits without admitting the ready backlog.
- Workflow configuration can select versioned defaults, report effective values
  and their sources, and retain human approval restrictions for native work.
  Invalid policy blocks execution; admission freezes execution limits while
  publication rechecks current project policy. New workspaces in every profile
  select ticket authority, isolated worktrees, trusted automatic integration and
  one worker. Reinitialization preserves existing policy, including absent config.
  Existing projects can upgrade explicitly with a prior-policy snapshot and a
  deliberate authority choice. Upgrade preserves review policy, rejects active
  ownership and adopts only selected legacy work while retaining its history.
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
  actions, Chryso visual identity, documentation, and browser smoke
  tooling.
- Added Markdown-first authoring sessions and assisted authoring flows for
  generating and promoting project documentation artifacts.
- Added workspace graph/index work, semantic digest support, cache boundaries,
  and TUI consumers for project-local Markdown artifacts.
- Added extensive Diataxis documentation for commands, TUI behavior, runtime,
  ticket execution, reporting, review, merge, maintenance, notifications,
  acceptance generation, and QA workflows.

### Changed

- Updated the Web UI with editable ticket details and comments, shared case
  editing panels, and dump creation dialogs.
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
- Migrated CHR operational artifacts to the canonical `ops/<namespace>/...`
  layout, using `ops/default` for no-user workspaces, while preserving `.chr/`
  for CHR-private project configuration and state.

### Fixed

- Prevented outdated ticket forms and comments from changing a ticket after
  another edit. Reload the board before submitting again.
- Made failed Web UI form submissions display their error and reload actions
  inside the current dialog.
- Fixed acceptance feature generation to preserve feature-round answers and
  satisfy Codex feature schema expectations.
- Fixed spec generation context so generated specs stay neutral and do not
  accidentally produce feature artifacts.
- Fixed work readiness so Chryso-owned runtime artifacts can be repaired
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
