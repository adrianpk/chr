# Slice 2: Dump Pair Consumption And TUI

Status: reviewing

Branch: `feat/slice-2-dump-pair-consumption-tui`

Pull request: #232

## Objective

Make curated dumps the primary consumption artifact while retaining raw dump siblings as source metadata.

## Changes

- Workspace nodes now preserve dump pair frontmatter metadata: `type`, `source`, and `curated`.
- Raw dump siblings are classified with `chc.IsRawDumpNode` and curated dumps with `chc.IsCuratedDumpNode`.
- Workspace context, quiz supplemental context, and spec context skip raw dump siblings.
- `ListDumps` returns curated primary dumps and exposes the raw sibling path as metadata.
- F3 Dumps shows curated dumps and renders the raw source path in the selected dump detail when present.
- Project and use-case artifact counts exclude raw dump siblings from the displayed dump total.
- User-facing dump docs now describe curated/raw pair behavior.

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui` passed.

## Notes

- `alt+e` remains focused on the selected curated dump because raw siblings are no longer part of the primary dump list.
