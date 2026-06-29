# Slice 10: Docs QA And Release Readiness

Status: reviewing

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-10-web-docs-qa`

Expected PR Title: `feat(slice-10): web docs and qa`

PR: `#352`

## Summary

- Completed user-facing reference docs, how-to docs, Web UI parity tracking,
  release notes, and repeatable QA smoke coverage for the optional local Web
  UI.
- Added a browser-capable smoke script that creates a temporary managed
  project, serves the Web UI, validates pages and htmx fragments, performs a
  write action, and captures screenshots when Chromium or Chrome is available.

## Changes

- `docs/reference/commands.md`
  - Documents `chc web serve`, dry-run mode, local bind overrides, token policy
    overrides, and Web UI reference links.
- `docs/reference/project-workspace.md`
  - Documents the `web:` config section, defaults, local-only example, and
    private-network or tunnel example.
- `docs/reference/web-ui-action-matrix.md`
  - Adds the Web UI equivalent matrix for navigation, views, write actions, and
    rollout boundaries.
- `docs/how-to/run-local-web-ui.md`
  - Adds the localhost workflow, dry-run preview, serve commands, local use
    model, and smoke validation command.
- `docs/how-to/expose-local-web-ui.md`
  - Adds dry-run exposure preview and cross-links the local-only how-to.
- `CHANGELOG.md`
  - Adds the optional local Web UI release note under Unreleased.
- `scripts/qa/web-ui-smoke.sh`
  - Adds repeatable QA coverage for Web UI pages, htmx fragments,
    `dump.create`, and browser screenshots.
- `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`
  - Records final slice task status and validation evidence.

## Validation

- `CHC_WEB_SMOKE_PORT=7340 scripts/qa/web-ui-smoke.sh` (pass)
  - Pages, htmx fragments, and `dump.create` action passed.
  - Browser screenshots passed.
  - Screenshot directory: `/tmp/chc-web-ui-smoke-326889`.
- `make verify` (pass)
  - `0 issues.`
  - `go test` passed for `cmd/chc`, `internal/app`, `internal/core`,
    `internal/tui`, `internal/webui`, `internal/webui/htmx`,
    `internal/webui/render`, and `internal/webui/ui`.

## Notes

- The Web UI remains explicit and disabled as a background behavior. It starts
  only through `chc web serve`.
- Screenshot files are not committed; the report records the capture path and
  validation result to avoid storing local-path UI captures in the repo.
