# Slice 8: Access Control And Exposure Safety

Status: delivered

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-8-web-access-control`

Expected PR Title: `feat(slice-8): web access control`

PR: `#350`

## Summary

- Added effective token policy resolution for `auto`, `true`, `false`, and
  `always`.
- Required a token automatically for non-loopback binds in `auto` mode.
- Read token values from the configured environment variable without printing
  token values.
- Added same-origin and configured-origin request checks.
- Added sanitized request logging that omits query strings.
- Added operator-facing docs for deliberate private-network or tunnel exposure.

## Changes

- `internal/app/web.go`
  - Resolves the effective Web UI access policy from config, bind host, and
    environment.
  - Prints `token_required` and `allowed_origins` in dry-run/startup output.
  - Warns when token protection is disabled on a non-loopback bind.
  - Passes sanitized request logging into the Web UI handler.
- `internal/webui/webui.go`
  - Adds token enforcement for non-static Web UI routes.
  - Supports `X-CHC-Web-Token`, `Authorization: Bearer`, and `?token=`.
  - Exempts `/healthz` and embedded static assets from token checks.
  - Enforces same-origin or explicitly allowed origins.
  - Supports allowed-origin CORS preflight responses.
  - Logs method, path, status, and duration without query strings.
- `internal/app/cli_test.go`
  - Covers non-loopback `auto` requiring a token.
  - Covers missing required token env rejection.
  - Updates dry-run output assertions for access policy fields.
- `internal/webui/webui_test.go`
  - Covers token enforcement, origin denial, allowed-origin preflight,
    wildcard-bind same-origin behavior, and sanitized logs.
- `docs/how-to/expose-local-web-ui.md`
  - Documents local-only use, private network or tunnel exposure, token use,
    allowed origins, and operator-owned infrastructure boundaries.

## Validation

- `go test ./internal/webui` (pass)
- `go test ./internal/app -run 'WebServe|RootCommand'` (pass)
- Temporary-project smoke with
  `CHC_WEB_TOKEN=... go run ./cmd/chc web serve --dir /tmp/... --host 0.0.0.0 --port 7338` plus:
  - `/healthz` without token returned `200` (pass)
  - `/` without token returned `401` (pass)
  - `/` with `X-CHC-Web-Token` returned `200` (pass)
  - allowed-origin `OPTIONS /actions` returned `204` (pass)
  - disallowed-origin request returned `403` (pass)
  - same-origin request through `127.0.0.1` while bound to `0.0.0.0`
    returned `200` (pass)
  - server log contained `token_required: true` and configured
    `allowed_origins` (pass)
  - server log did not contain token values or query strings (pass)
- `make verify` (pass)

## Notes

- This slice hardens the local Web UI boundary before treating non-loopback
  exposure as ready.
- The implementation must keep loopback/local usage low-friction while making
  explicit exposure obvious and protected.
