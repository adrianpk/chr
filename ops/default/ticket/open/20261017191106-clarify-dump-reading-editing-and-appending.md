---
id: TKT-20261017191106
title: "Use one editor and Save for dump content"
status: open
kind: task
severity: unclassified
priority: unclassified
scope: ui
tags: dumps, editing, interaction
source: chat
reported_at: 2026-10-17T19:11:06Z
commits: []
---

# Use one editor and Save for dump content

## Context

Finding from [the Web UI review](20261017105432-review-draft-web-ui-usability.md),
following a maintainer-provided screenshot of the dump editor. The maintainer
cannot understand why the same content appears in a yellow read-only area
and an editable area, or why another Append area is needed below the editor.
The maintainer explicitly rejects displaying the same dump body simultaneously
in a read-only panel and an editor, and rejects the separate Append form in
this manual editing interface. The requested interaction is one editor and
Save. Capture only; no interface or application changes are authorized.

Related behavior work:
[One evolving project-wide dump](20261017190508-unify-dump-capture-and-project-wide-ownership.md).
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
