# Slice 3: Single Dump Editor

Status: delivered
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `fix/slice-3-dump-editor`
PR: `#400`

## Purpose

Deliver the single dump editor and guarded Save defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
T3.1 and T3.2 were integrated through [PR #400](https://forge.adrianpk.com/adrianpk/chr/pulls/400)
at `63b31c3ec83b706d501a79918a448a5fc3d82ce3` on 2026-10-18.

## Delivered Behavior

Opening a dump shows one Body editor and Save in the catalog modal or full page.
Corrections, removal and added text replace the body of the same canonical dump.
The duplicate read-only body and manual Append form are removed. Frontmatter
stays outside the editor, and shared writes preserve metadata, raw provenance
and prior body/metadata history. Capture adapters retain their append primitive.

Validation, transport and stale-revision failures render inside the live form
without replacing its body or guards. Unchanged retries keep their operation
identity; changed submitted arguments receive a new identity. Pending projection
results state that Save committed and permit an unchanged refresh retry. Controls
are disabled while a request is active, preventing unsubmitted changes during Save.

Concurrent writes remain fenced by the serialized shared owner. Reload stored
content asks before discarding pending edits and loads the current revision.
Catalog refreshes preserve an open modal. Full-page editing has no automatic
polling. Successful modal Save restores focus to the selected card across the
catalog refresh; desktop and narrow views use the same editor.

## Implementation Notes

- T3.1: `600f30bf1be349403ab38b6d98fa6dad6e5b5c93` — body-only Web reads, unchanged shared Save ownership and real HTTP replacement/stale regression.
- T3.2: `21176131f4e16869c910a5c40a48ae69072317b0` — single-surface templates, inline results, retained Save arguments, explicit reload and browser journeys.

The existing canonical mutation remains the only writer. Its source revision
includes identity, complete body and metadata under the serialized owner. The
new read field separates editable body from the exact stored image; other
adapters keep their existing complete-document reads.

## Contracts Added or Changed

`DumpDocumentObservation.Body` exposes the editable body while `Content` retains
the exact stored image. Web detail uses Body with the same canonical ID/revision.
Save fragments target the result region; a confirmed modal write retargets the
modal root, while confirmed full-page Save reloads the stored document. Pending
projection acknowledgements retain the editor and receipt binding.

Save identities use the existing per-form value until submitted arguments
change. Fingerprints bind document ID, observed revision and body. No provider,
new persistence owner or background runtime is introduced.

## Files of Interest

- `internal/core/dump_mutation.go` and `internal/app/web_artifacts.go`: body observation and Web mapping.
- `internal/app/web_dump_editor_test.go`: real HTTP Save and stale replacement over canonical fixtures.
- `internal/webui/assets/templates/shell.html`, `chr.js` and `webui.go`: one editor, inline outcomes, refresh/focus preservation and reload.
- `internal/app/web_dump_editor_browser_test.go`: owned desktop/narrow Chromium journey with lost acknowledgement and concurrent edit boundaries.
- `docs/how-to/user-guide/index.md`, `docs/reference/web-ui-action-matrix.md` and `CHANGELOG.md`: editing and retry behavior.

## Validation

Task checks passed before their corresponding commits:

- `go test -race ./internal/core ./internal/app -run '^(TestDumpMutation(Save|Append|.*Guard.*)|TestServiceDump.*|TestWebDumpSaveDocument)$' -count=1` — canonical full-body replacement, stale writes, raw provenance and shared owner regressions passed. Existing Save history assertions verify original body/metadata images.
- `go test ./internal/app -run '^TestWebDumpSaveDocument$' -count=1` — passed after extracting the detail assertion to meet test complexity limits.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app` — T3.1 lint passed with zero issues.
- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app ./internal/webui -run '^(TestWebDumpEditorBrowser|TestWebDumpCaptureBrowser|TestWebBoardBrowser|TestDump.*|TestHandlerDump.*|TestHandlerFullPagesConfigurePolling|TestHandlerWriteAction.*)$' -count=1` — editor, capture and Board browser regressions plus affected HTTP rendering checks passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/app ./internal/webui/...` — T3.2 lint passed with zero issues after whitespace corrections in the browser fixture.
- `git diff --check` — passed.

The new HTTP regression first reproduced editable frontmatter. Rendering checks
reproduced the duplicate body/Append surfaces and replacement of failed forms.
The browser journey caught focus loss after catalog replacement and inherited
HTMX disable selectors blocking explicit reload. The passing checks retain all
assertions after those corrections.

The browser covers edit/add/Save/reopen, lost acknowledgement and unchanged retry,
changed submitted intent, concurrent stale rejection, dirty modal refresh,
explicit reload, focus restoration and full-page persistence. CSS viewports are
1440 × 1000 and 390 × 844, DPR 1, at 100% zoom. Fixtures own and stop their
Chromium/server processes and use isolated stores with deterministic curation.

Final clean-HEAD validation passed through Looper for
`63b31c3ec83b706d501a79918a448a5fc3d82ce3` in canonical order:
`make verify`, focused race regressions, the owned editor/capture/Board browser
checks and `git diff --check`. The canonical PR lists the exact successful final
commands; Looper retains the tested HEAD and ordered evidence. Development
checks do not establish scheduled exhaustive/nightly certification.

## Risks and Follow-ups

The controller owns integration into dev. No live project data was changed.
