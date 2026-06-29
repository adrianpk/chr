# Slice 1: Web Config And Serve Command

Status: ready for PR

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-1-web-config-serve-command`

Expected PR Title: `feat(slice-1): web config and serve command`

## Summary

- Added the disabled-by-default Web UI config contract with localhost defaults.
- Added validation for web host, port, token mode, token env, refresh interval,
  and allowed origins.
- Added `chc web serve` with dry-run output, host/port overrides, token-mode
  override, open-browser flag parsing, non-loopback warning, and a minimal
  `/healthz` HTTP endpoint.

## Changes

- `internal/core/config.go`
  - Adds `WebConfig`, `WebTokenMode`, defaults, config parsing, and string-list
    parsing for `allowed_origins`.
- `internal/core/web_config.go`
  - Adds `ValidateWebConfig` and `WebHostIsLoopback`.
- `internal/core/config_test.go`
  - Covers Web config defaults, parsing, partial override defaults, validation
    failures, and loopback detection.
- `internal/app/app.go`
  - Registers the new `web` top-level command without changing default TUI
    behavior.
- `internal/app/web.go`
  - Adds `web serve`, flag override resolution, dry-run/start output, warning
    output, foreground HTTP server startup, graceful shutdown, and `/healthz`.
- `internal/app/cli_test.go`
  - Covers command registration, dry-run output, flag overrides, non-loopback
    warning, invalid port rejection, and the health endpoint.
- `ops/plans/2026-06-29-optional-local-web-ui-plan.md`
  - Marks Slice 1 as the active slice.
- `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`
  - Records Slice 1 task completion, commits, and validation evidence.

## Validation

- `go test ./internal/core -run 'WebConfig|WorkspaceConfig'` (pass)
- `go test ./internal/core -run 'WebConfig|WorkspaceConfig' && go test ./internal/app -run 'RootCommand|WebServe'` (pass)
- `go test ./internal/core ./internal/app` (pass)

## Notes

- The Web UI remains disabled by default.
- Plain `chc` and `chc tui` continue to use the TUI path.
- `--open-browser` is parsed and reported in Slice 1, but browser opening is
  not executed yet.
- Remote verification and PR creation are currently blocked by SSH access to
  `origin`.
