# Slice 5: Case and Dump forms

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-5-tui-case-dump`
PR: [#459](https://forge.adrianpk.com/adrianpk/chr/pulls/459)
Base: `bcabdf02513ae95fce41b6a8640b6cae2b4bd4c0` on dev, after verified PR #458 integration and Slice 4 closure

## Purpose

Resolve G-015, G-017, G-018 and G-019 with complete terminal forms, captured
append targets and consistent external editing without losing unsaved input.

## Delivered Behavior

- Cases collects Name, Description and Context before one Create. Validation
  retains every field; confirmed cancellation creates no blank case or source.
  Existing metadata editing, artifact counts and unresolved references remain.
- Dumps collects a named case or Project-wide scope, title and body. Canonical
  capture preserves raw material and selects the returned document. Body editing
  shows retained title/scope metadata outside the textarea.
- `a` and the append palette action collect free text for the captured selected
  path and revision. No fixed capture or append content is submitted.
- `alt+e` within the body editor round-trips the current unsaved text through a
  private temporary file. Returning retains edits until explicit Ctrl+S; process
  errors retain the previous draft. Outside the editor, the existing canonical
  document editor remains available. Displayed hints match both bindings.
- Stale Save retains the draft. Cancelled reload keeps input and canonical
  bytes; confirmed reload replaces it with the current body and revision before
  another Save. Old external-editor replies cannot replace a newer interaction.

## Implementation Notes

T5.1 is `8a5691065226ff3b8499290bc7ec39ce36145073`; T5.2 is
`d6f933abf9b328a3befc4837bd5744733108c4cc`; T5.3 is
`cf8c5e4b3d68ffb2d61349075f1cd6d7507a9b55`. T5.4 is `fbd0b6d5a6dc7be995713fbbd8cbbb059ae9e6f2`.

Forms own their input mode and use existing terminal styles and Action outcomes.
Tab/Shift+Tab select fields, Ctrl+S submits, and Enter submits only from the
button or advances a single-line field. Multiline Enter remains text. Changing
capture or append input rotates its operation identity; unchanged retries keep
it. Source and operation bindings exclude unrelated queue observations.

The draft-editor service uses the configured launcher with a private directory
and 0600 file. Its completion reads a regular file within the existing 1 MiB
body limit and removes the exact owned directory after success or failure.
The TUI adapter owns process execution and maps neutral outcomes. The source
revision remains captured until explicit canonical Save.

G-018 runtime uncertainty is resolved by actual production PTY input plus real
concurrent filesystem writes and Web HTTP reference fixtures. The terminal
fixture retains the root loader and managed host; only model responses are
deterministic. Initial driver failures came from consuming a lower-screen status
before asserting earlier text and sending two Escape keys without observing the
intermediate panel. Corrected observation order and state-based waits preserve
all content assertions and existing PTY deadlines; no acceptance or limits were
weakened. Cancelled creation and failed/cancelled stale recovery compare actual
canonical file images. Children and external-editor files are joined and cleaned.

## Contracts Added or Changed

Canonical metadata and dump capture/read/save/append remain the only document
writers. Body Save retains stored frontmatter, title and scope, matching Web.
The shared local form state adds no persistence format or domain orchestration.
`OpenDumpDraftEditor` prepares an owned unsaved copy and returns bounded cleanup;
its TUI completion is bound to the original operation and interaction.

## Files of Interest

- `internal/tui/text_form.go`, `case_form.go`, `dump_form.go`: complete form
  inputs, captured scope/revision and retained submission errors.
- `internal/tui/dump_external_edit.go`, `dump_actions.go`, `dumps.go`: external
  round trip, retained body editor and accurate controls.
- `internal/app/service_dump_editor.go`, `tui_dump_editor.go`: private draft
  lifecycle and thin terminal adapter.
- `internal/app/tui_*form*_test.go`, `tui_dump_append_test.go`,
  `service_dump_editor_test.go`: real canonical IO, production PTY and cleanup.
- User Guide, TUI reference, CLI/TUI action matrix and changelog: usable controls.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed.
- `go test ./internal/app -run '^Test(ServiceDumpEditor|TUIDumpAppend|TUIDumpForm|TUICaseForm|TUIOperationDraft)$' -count=1 -timeout=3m` — passed.
- `go test ./internal/app -run '^TestTUIForms(Terminal|StaleTerminal)$' -count=1 -timeout=3m` — passed after driver sequencing corrections.
- `go test ./internal/app -run '^TestTUIFormsTerminal$' -count=1 -timeout=3m` — passed with cancelled creation and unchanged canonical files.
- `go test ./internal/app -run '^Test(WebCaseCreate|WebDumpCapture|WebDumpSaveDocument|TUIOperationDraft|TUIOperationTerminal)$' -count=1 -timeout=3m` — passed without short mode.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed after whitespace corrections.

Final ordered validation will be recorded by Looper on the clean PR head after
PR-reference commits: `make verify`, focused terminal race tests, focused
application IO/HTTP/PTY race tests, then `git diff --check`. The canonical PR and
controller evidence retain the exact command list and candidate; these earlier
task checks do not claim final-candidate validation or dev integration.

## Risks and Follow-ups

None recorded.
