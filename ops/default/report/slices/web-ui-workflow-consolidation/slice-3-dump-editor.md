# Slice 3: Single Dump Editor

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `fix/slice-3-dump-editor`
PR: pending

## Purpose

Deliver the single dump editor and guarded Save defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
Slice 2 is integrated. T3.1 is implemented; T3.2 removes duplicate editing
surfaces and preserves pending input.

## Delivered Behavior

Web Save replaces the observed canonical body under the existing serialized
mutation and rejects older revisions. The editor receives only the body; the
complete stored image, frontmatter and raw provenance remain owned by the
shared document writer. The single-surface UI is in progress.

## Implementation Notes

Slice 2 already routes Web Save through the canonical serialized mutation with
operation identity and observed document revision. Reuse that owner rather
than introducing another writer or changing capture append behavior.

## Contracts Added or Changed

`DumpDocumentObservation.Body` exposes the editable body while `Content`
continues to retain the exact stored document image.

## Files of Interest

- `internal/app/web_actions.go` and `service_dump.go`: shared Save binding.
- `internal/webui/assets/templates/shell.html`, `chr.js` and `webui.go`: editor and write-result retention.

## Validation

T3.1's HTTP regression first failed because frontmatter appeared in the editable
content. Separating the body corrected that failure.

- `go test -race ./internal/core ./internal/app -run '^(TestDumpMutation(Save|Append|.*Guard.*)|TestServiceDump.*|TestWebDumpSaveDocument)$' -count=1` — passed; canonical replacement, stale writes, raw provenance and shared owner regressions.

- `go test ./internal/app -run '^TestWebDumpSaveDocument$' -count=1` — passed after extracting the detail assertion to meet test complexity limits.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app` — passed with zero issues.

Final browser and exact clean-HEAD controller evidence remain pending.

## Risks and Follow-ups

The controller owns integration into dev. No live project data is changed.
