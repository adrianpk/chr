# Slice 5: Document And Metadata Actions

Status: delivered

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-5-document-and-metadata-actions`

Expected PR Title: `feat(slice-5): document and metadata actions`

PR: `#347`

## Summary

- Added the first Web UI write-action contract with stable result envelopes for
  success, validation failure, stale writes, and blocked writes.
- Added confirmed browser actions for case metadata update, dump creation, dump
  save, and dump append.
- Added revision tokens and stale-write rejection before editable document
  writes.
- Added form rendering, redirect refresh behavior, tests, and smoke evidence.

## Changes

- `internal/webui/views.go`
  - Adds write action request/response contracts, write statuses, page data,
    artifact IDs, structured case metadata fields, and revision tokens.
- `internal/app/web_actions.go`
  - Implements Web UI write actions through existing app services:
    `PatchCaseMetadata`, `CreateDump`, `SaveDump`, and `AppendDump`.
  - Adds confirmation enforcement, stale revision checks, validation envelopes,
    blocked envelopes, and success redirects.
- `internal/app/web_artifacts.go`
  - Adds revision tokens to artifact cards and details.
  - Adds structured case detail projection for safe metadata editing.
- `internal/webui/webui.go`
  - Adds `POST /actions`, form parsing, write result rendering, standard
    redirects, and htmx `HX-Redirect` refresh behavior.
- `internal/webui/assets/templates/shell.html`
  - Adds case metadata edit, dump create, dump save, dump append, and write
    result templates.
- `internal/webui/assets/static/chr.css`
  - Adds form and write-result styling.
- `internal/webui/webui_test.go`
  - Covers action redirects and blocked write result envelopes.
- `internal/app/cli_test.go`
  - Covers provider-backed case metadata update, dump creation, dump save, and
    stale dump save rejection.

## Validation

- `go test ./internal/webui ./internal/app -run 'WebServe|Handler|Catalog|Artifact|WriteAction'` (pass)
- `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- Temporary-project smoke with `go run ./cmd/chr web serve --dir /tmp/tmp.I8grD0CG1y --port 7335` plus:
  - `curl -fsS http://127.0.0.1:7335/dumps` found the dump create form (pass)
  - `curl -X POST http://127.0.0.1:7335/actions` for dump create returned `303 /dumps` (pass)
  - `curl -X POST http://127.0.0.1:7335/actions` for case metadata update returned `303 /cases/detail?id=blog` (pass)
  - `curl -X POST http://127.0.0.1:7335/actions` for dump save returned `303 /dumps/detail?...` (pass)
  - repeated dump save with the stale original revision returned `400` and a stale envelope (pass)
- `make verify` (pass)

## Notes

- The action set is intentionally limited to operations already present in the
  application service layer.
- This slice does not start agent work or generation flows.
- Successful writes refresh via redirect instead of keeping stale browser
  state.
