---
id: TKT-20261017191106
title: "Use one editor and Save for dump content"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: dumps, editing, interaction
source: chat
reported_at: 2026-10-17T19:11:06Z
commits: [600f30bf1be349403ab38b6d98fa6dad6e5b5c93, 21176131f4e16869c910a5c40a48ae69072317b0]
branch: fix/slice-3-dump-editor
pr: "#400"
resolution: fixed
reviewed_at: 2026-10-18T16:35:59Z
closed_at: 2026-10-18T16:35:59Z
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Use one editor and Save for dump content

## Context

Finding from [the Web UI review](../in_progress/20261017105432-review-draft-web-ui-usability.md),
following a maintainer-provided screenshot of the dump editor. The maintainer
cannot understand why the same content appears in a yellow read-only area
and an editable area, or why another Append area is needed below the editor.
The maintainer explicitly rejects displaying the same dump body simultaneously
in a read-only panel and an editor, and rejects the separate Append form in
this manual editing interface. The requested interaction is one editor and
Save. Capture only; no interface or application changes are authorized.

Related behavior work:
[One evolving project-wide dump](../solved/20261017190508-unify-dump-capture-and-project-wide-ownership.md).
Keep the interaction finding distinct from canonical dump ownership.

## Confirmed Current Behavior

1. The top yellow area is a read-only `<pre>` displaying `Detail.Content`.
2. The Content textarea is initialized from that exact same `Detail.Content`.
   The two areas do not represent separate raw and curated versions, or a
   live preview of pending edits.
3. Save submits `dump.save` and replaces the stored dump body with the full
   contents of the Content textarea, after the existing revision check.
4. The separate Append note textarea submits `dump.append`. It adds only its
   own text to the end of the stored dump; it does not include unsaved edits
   in the Content textarea.
5. The screen presents both independent write operations and the duplicate
   read-only text together, without making their relationship clear.

## Desired Outcome

Use one editor for the dump body. Remove the permanent yellow read-only copy
of the same content above that editor. Remove the separate Append note
textarea and Append button from the dump editor.

The single editor contains the existing Markdown document. Users correct,
remove, reorganize or add text in that editor, then use Save. Saving writes
the full edited body to the same dump. Adding material at the end is ordinary
document editing, without another form or save path.

Preserve pending edits and existing stale-revision safeguards. Make the content
that Save will persist unambiguous.

## Design Boundary

The agreed manual editing interface is one body editor and one Save action.
Both the duplicate read-only panel and the separate Append form must be removed.
This ticket does not replace the shared append primitive used by other capture
paths, introduce a new editor framework or preview system, or authorize
implementation. Keep the interaction compatible with the selected canonical
dump destination model.

The current dump body is stored in a Markdown file. Append writes text to the
end of that file; it does not create a separate note entity or database row.

## Acceptance

1. A user can identify the stored content, editable content and pending changes.
2. The dump editor shows one editing surface for the body and no simultaneous
   read-only panel duplicating that complete text.
3. The manual dump editor has no separate Append note textarea, Append form or
   Append button. New text is entered in the same editor and saved with Save.
4. The edited body, including corrections and new text at the end, is persisted
   to the same dump without creating another artifact or losing pending edits.
5. Validate open, edit, add text, save and reopen journeys when implementation
   is authorized. Preserve the existing stale-revision and write safeguards.

## Implementation References

- `internal/webui/assets/templates/shell.html`: dump modal and shared edit forms.
- `internal/app/web_actions.go`: save/append actions and revision checks.
- `internal/app/service_dump.go`: full-body replacement and append behavior.

## Implementation Progress

T3.1 is implemented in `600f30bf1be349403ab38b6d98fa6dad6e5b5c93`.
Web reads expose only the editable body while preserving complete source images.
Save uses the existing canonical serialized owner, operation identity and source
revision. The real HTTP regression verifies same-document replacement,
frontmatter/raw retention and stale rejection.

T3.2 is implemented in `21176131f4e16869c910a5c40a48ae69072317b0`.
Modal and full-page views show one Body field and Save. Inline validation,
transport, stale and pending-projection results preserve the form and guards.
Unchanged retries retain identity; edited submitted arguments start a new intent.
Explicit reload confirms discard, catalog refresh preserves the modal and focus,
and the full-page editor does not poll over pending input. Capture adapters keep
the shared append primitive. User Guide, action matrix and Unreleased are updated.

Successful development checks:

- `go test -race ./internal/core ./internal/app -run '^(TestDumpMutation(Save|Append|.*Guard.*)|TestServiceDump.*|TestWebDumpSaveDocument)$' -count=1`
- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app ./internal/webui -run '^(TestWebDumpEditorBrowser|TestWebDumpCaptureBrowser|TestWebBoardBrowser|TestDump.*|TestHandlerDump.*|TestHandlerFullPagesConfigurePolling|TestHandlerWriteAction.*)$' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/app ./internal/webui/...`
- `git diff --check`

The [Slice 3 report](../../report/slices/web-ui-workflow-consolidation/slice-3-dump-editor.md)
records the initial failures, corrections and desktop/narrow journey evidence.
All tasks are implemented in canonical [PR #400](https://forge.adrianpk.com/adrianpk/chr/pulls/400).
Final clean-HEAD `make verify`, focused race/browser checks and whitespace evidence
are recorded through Looper, with the exact ordered commands in the PR. The
controller merged PR #400 into dev at `63b31c3ec83b706d501a79918a448a5fc3d82ce3`
on 2026-10-18 after all four ordered exact-HEAD checks passed. The ticket is solved.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).
