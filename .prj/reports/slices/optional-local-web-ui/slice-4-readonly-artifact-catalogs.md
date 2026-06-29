# Slice 4: Read-Only Artifact Catalogs

Status: delivered

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-4-readonly-artifact-catalogs`

Expected PR Title: `feat(slice-4): read-only artifact catalogs`

PR: `#346`

## Summary

- Added read-only Web UI catalog and detail contracts for Cases, Dumps,
  Quizzes, Specs, and Plans.
- Added an application-layer provider that reuses existing case metadata and
  artifact discovery/index reads.
- Added full-page and htmx fragment routes for artifact catalogs and details.
- Added templates, styles, tests, path confinement, and DOM smoke evidence for
  the read-only artifact surfaces.

## Changes

- `internal/webui/views.go`
  - Adds artifact catalog, card, detail, and request view models.
- `internal/app/web_artifacts.go`
  - Maps existing case metadata, dump, quiz, spec, and plan reads into Web UI
    catalog cards and detail views.
  - Adds project-path confinement before reading artifact detail files.
- `internal/webui/webui.go`
  - Adds `/cases`, `/dumps`, `/quizzes`, `/specs`, `/plans`, family fragment
    routes, and family detail routes.
- `internal/app/web.go`
  - Adds `Cases` to the Web UI navigation.
- `internal/webui/assets/templates/shell.html`
  - Renders catalog cards and escaped text detail content inside the shared
    shell.
- `internal/webui/assets/static/chc.css`
  - Adds styles for artifact grids, cards, detail actions, and preformatted
    content.
- `internal/webui/webui_test.go`
  - Covers catalog and detail rendering, missing detail state, empty catalog
    state, and htmx fragments with a fake provider.
- `internal/app/cli_test.go`
  - Covers provider-backed case and dump catalogs, detail reads, malformed case
    metadata, and path confinement.

## Validation

- `go test ./internal/webui ./internal/app -run 'WebServe|Handler|Catalog|Artifact'` (pass)
- `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- `go run ./cmd/chc web serve --dir . --port 7334` plus:
  - `curl -fsS http://127.0.0.1:7334/cases` (pass)
  - `curl -fsS http://127.0.0.1:7334/dumps` (pass)
  - `curl -fsS http://127.0.0.1:7334/quizzes` (pass)
  - `curl -fsS http://127.0.0.1:7334/specs` (pass)
  - `curl -fsS http://127.0.0.1:7334/plans` (pass)
  - `curl -fsS http://127.0.0.1:7334/dumps/detail?id=.chc%2Fuser%2Fdumps%2Fcurated%2F2026-06-02-merger-agent-role.md` (pass)
  - `/usr/bin/curl -fsS -H 'HX-Request: true'` for `/cases/fragment`,
    `/dumps/fragment`, `/quizzes/fragment`, `/specs/fragment`, and
    `/plans/fragment` (pass)
  - `curl -fsS -H 'HX-Request: true'` for one dump detail fragment (pass)
- `make verify` (pass)

## Notes

- The Web UI remains read-only in this slice.
- Markdown content is rendered as escaped text in `<pre>`; the browser does not
  receive trusted raw HTML for artifact bodies.
- Detail reads require IDs from the catalog and resolve artifact paths only
  after project-directory confinement.
