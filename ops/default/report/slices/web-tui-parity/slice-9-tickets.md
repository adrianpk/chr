# Slice 9: Ticket editing, comments and references

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-9-tui-tickets`
PR: pending

## Purpose

Resolve G-027, G-029, G-030 and the ticket portion of I-004 through complete,
revision-bound terminal ticket interactions using existing canonical owners.

## Delivered Behavior

- Board and Tickets offer Create, Edit, Comment, Attach review reference and
  Refresh with `c`, `e`, `a`, `v` and `r`, plus equivalent palette commands.
  Create collects entered title, description, case and metadata. Edit retains
  current values and displays phase, slice and canonical read-only reasons.
- Tab navigation keeps later metadata fields visible in smaller terminals.
  Failed and stale submissions retain the complete draft. Reload requires
  confirmation and refuses to silently change a captured case binding.
- Comments accept free text as a separate operation. Saved discussion remains
  readable and scrollable; empty/stale submissions write nothing. Metadata/body
  edits preserve discussion, and multiple entered comments remain distinct.
- Review attachment collects type, provider, HTTP(S) URL, optional ID, status and
  branch. It preserves existing references and uses canonical revision, active
  claim and verified-delivery restrictions. Pending managed metadata can change
  without replacing its approved execution description.
- External source editing opens canonical `ops/` paths correctly. Checklist
  display and existing dump append stay reachable. A committed write followed
  by a failed refresh cannot be resubmitted from its retained form.

## Implementation Notes

T9.1: `fc8cd8e06e07568045c79e8d08b49370d5493cc9`.
T9.2: `f2ddfdfd17490910683a2a0473a9cda53b09395d`.
T9.3: `29095befa1a96578d8eb9a1f81435f524097ce7d`.
T9.4: `ad327372411f98e7fc5854ed3de1c56ac949dbd7`.

Ticket editor choices and restrictions now have one transport-neutral application
representation. Web and TUI map it independently. Terminal writes call
CreateBoardTicket or SaveTicketDocument with the observed revision and operation
identity. The old fixed artifact-creation route and fixed comment dispatch are
removed from the terminal.

Review attachment reuses the existing review-reference formatter inside the
canonical ticket edit transaction. A regression exposed stored legacy references
missing from the Board projection; the reader now retains them. Native plain-body
tickets retain references in existing envelope metadata, and later edits render
those references without changing the approved description or workflow identity.
Creation also returns its committed revision separately from a later read result,
matching the existing edit result contract.

The production PTY journey uses the root loader and managed terminal host with
real files and input. It creates and edits a ticket, adds two different comments,
scrolls both retained entries, attaches user-entered review metadata and invokes
an owned external editor on the canonical source. The editor fixture changes the
description explicitly, rather than appending inside the final Comments section.
All assertions and the existing 45-second fixture ownership limit remain enabled.
The fixture stops only its owned child after checks; no provider is needed.

## Contracts Added or Changed

Ticket edit requests can carry an existing review-reference value, guarded by the
same source image and workflow owner as field edits. No new persistence schema or
domain workflow is introduced. Verified delivery references remain read-only.
Comments cannot be mixed with body, field or reference edits.

Terminal forms carry captured ticket/case/revision/operation identity in feedback
bindings. Reload retains the form mode and rejects a changed case association.
Committed partial results offer Close and refresh guidance instead of a duplicate
write. All service files remain transport-neutral.

## Files of Interest

- `internal/app/service_ticket_editor.go`, `ticket_editor_fields.go`:
  neutral canonical editor metadata and restrictions.
- `internal/app/tui_ticket_editor.go`, `tui_adapters_spec_plan_ticket.go`:
  canonical reads/writes and terminal result mapping.
- `internal/tui/ticket_form.go`, `ticket_comment_form.go`,
  `ticket_review_form.go`, `ticket_recovery.go`, `text_form.go`:
  complete input, retained history, visible focus and guarded reload.
- `internal/core/workflow/ticket_edit_review.go`, `ticket_edit.go`,
  `ticket_native_display.go`, `internal/core/ticket/board_document.go`:
  reference transaction, native retention and canonical read projection.
- `internal/app/tui_ticket_*test.go`, `internal/tui/ticket_form_test.go`:
  actual authority/terminal IO and draft/ownership coverage.
- User Guide, TUI reference, action matrix and changelog: delivered controls.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including unchanged complete catalog frames and source-editor keyboard compatibility.
- `go test ./internal/app -run '^Test(TicketEditorOptions|TicketRepositoryOptions|WebTicketEditor)$' -count=1 -timeout=3m` — passed for shared options and independent Web mapping.
- `go test ./internal/app -run '^Test(TUITicket.*|TUIBoard.*|TUIForms(Terminal|StaleTerminal)|WebTicketEditor)$' -count=1 -timeout=3m` — passed for legacy/native forms, separate comments, stale drafts, reference guards, production terminal and retained dump append.
- `go test ./internal/core/workflow ./internal/core/ticket -run '^Test(Ticket(Document.*|Native.*|ReviewReceiptGuard)|Board.*)$' -count=1 -timeout=3m` — passed for canonical creation/editing, native ownership, verified references and Board projections.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app ./internal/core/ticket ./internal/core/workflow` — passed.
- `git diff --check` — passed before task commits.

Final ordered validation is recorded by the controller on the clean PR head after
the PR-reference follow-up. Its canonical commands and immutable report link
belong to the PR; integration evidence will close this report after merge.

## Risks and Follow-ups

Canonical PR integration remains pending. Checklist mutation retains its existing
unavailable-writer guard; checklist display remains covered.
