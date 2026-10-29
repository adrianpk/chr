# Slice 6: Spec catalog and manual editing

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-6-tui-specs`
PR: [#460](https://forge.adrianpk.com/adrianpk/chr/pulls/460)
Base: `c1d0497c1c638b69533d0aae771db71a0cbeaf39` on dev, after verified PR #459 integration and Slice 5 closure

## Purpose

Resolve G-023 and G-024 with the complete canonical specification catalog and
manual create/edit forms while retaining generation and source inspection.

## Delivered Behavior

- Specs reads the same bounded concrete-document catalog as Web. Manual,
  generated and project-wide documents all appear, including several in one
  case. Rows show case and document title; refresh reconciles selection by
  canonical path rather than selecting another document in the same scope.
- `c` collects scope, title, body and Draft/Approved state before one Create.
  `e` reads the selected canonical document into the same form owner, retaining
  its case binding while title, body and state remain editable. Palette Create
  and Edit route to those same methods. No generation prerequisite is added.
- Canonical validation failures retain all inputs. Esc confirms discard, and
  stale Save cannot overwrite a concurrent change. Cancelled reload keeps fields
  and files; confirmed reload reads the captured path and current revision.
  A changed case binding is rejected without silently rebinding the draft.
- Recovery retries retain the exact request identity and inputs. If creation
  committed but catalog refresh failed, its unchanged retry reconciles the
  receipt and refreshes without creating another document. Old asynchronous
  catalog replies cannot replace a newer form.
- Spec and acceptance-feature generation, saved-body inspection, source paths,
  external source opening and their existing domain guards remain available.
  Contextual Help and the real footer include manual Create/Edit actions.

## Implementation Notes

T6.1 is `b4cbad75cce9bf02c7dea91749a3d287766a4a32`; T6.2 is `0e2d48d231fe52f71f1677e47b6e867b04a76b6f`;
T6.3 is `ab927ea40c7e95f967554a573d1ca5cbe2bb0429`; T6.4 is `df73bed5809cc29d40a9dc9007411080e0015b47`.

The first regressions reproduced omitted manual records and refresh selecting a
different document in the same case. The TUI now calls ListSpecDocuments, while
ListSpecs retains its latest-in-scope CLI/generation purpose. The canonical
reader preserves its 4,096-entry and 8 MiB aggregate inventory limits.

One `openSpecForm` method builds both create/edit presentations, and the existing
text form owns fields, focus, rendering, confirmation and submission. This is the
Spec component owner for subsequent guided-stage binding in the recorded plan;
there is no parallel wizard markup or update handler in this slice.
Title/body limits remain 512 bytes and 1 MiB under canonical mutation validation.
Adapters map neutral operation outcomes and do not write document files directly.

G-024 is resolved by real manual and generated publication, project-wide records,
production TUI adapter reads, and actual HTTP catalog responses that contain all
of the same canonical references. The production PTY uses the root loader and
managed Bubble Tea host, owned project state and real concurrent writers; model
responses alone are deterministic. Its first run exposed a real missing footer
hint for reachable Create/Edit actions; the footer now advertises them. All
original content, revision, discard and resource assertions remain intact.

The older acceptance-feature fixture had body-only legacy inputs; the TUI test
now installs metadata for its two owned canonical seed documents before reading
the complete catalog. Generation and feature assertions remain unchanged.
Spec frames reflect the delivered catalog and controls. Plan frame values and
serialized baseline lines remain unchanged. The temporary frame producer was
removed before commit; validation treats the fixture as read-only.

## Contracts Added or Changed

TUI requests carry the captured path, composite revision, operation identity,
case, title, body and lifecycle. CreateSpec, SaveSpec and ReadSpecDocument remain
the document owners. Save preserves generation provenance and rejects a changed
revision or case. UI lifecycle choices map to canonical draft/approved values.
The shared form owns transient input only; no persistence format changes.

## Files of Interest

- `internal/app/tui_adapters_loaders.go`, `tui_spec_editor.go`: complete catalog,
  canonical body/metadata reads and shared mutation service mapping.
- `internal/tui/spec_form.go`, `spec_recovery.go`, `text_form.go`: single form
  owner, captured recovery and retained fields.
- `internal/tui/specs.go`, `spec_actions.go`, `document_catalog.go`,
  `operation_binding.go`: identity selection, catalog display and late replies.
- `internal/app/tui_spec_*test.go`, `internal/tui/spec_*test.go`: actual IO/HTTP,
  production terminal, retry and ownership regressions.
- User Guide, TUI reference, action matrix and changelog: usable controls and
  the preserved difference between CLI scoped inventory and Web/TUI documents.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including complete catalog frames.
- `go test ./internal/app -run '^Test(TUISpecCatalog|TUISpecAdapterGeneratesAcceptanceFeature)$' -count=1 -timeout=3m` — passed after canonical fixture metadata was restored.
- `go test ./internal/app -run '^TestTUISpecForm$' -count=1 -timeout=3m` — passed for project-wide and case-bound forms.
- `go test ./internal/tui -run '^Test(SpecFormOwnership|SpecReloadScope)$' -count=1` — passed.
- `go test ./internal/app -run '^TestTUISpecTerminal$' -count=1 -timeout=3m` — passed after correcting the real footer.
- `go test ./internal/app -run '^Test(TUISpec.*|WebSpec(Authoring|ReadOnly|NativeFailure))$' -count=1 -timeout=3m` — passed, including real HTTP reference comparison.
- `go test ./internal/app -run '^TestTUISpecRetry$' -count=1 -timeout=3m` — passed with exactly two documents after repairing the projection source.
- `go test ./internal/tui -run '^Test(SpecFormPalette|SpecGenerationCommandsOnlyAppearInSpecs|DocumentCatalogFrames)$' -count=1` — passed through real palette keys and unchanged generation assertions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed.

Final ordered validation is recorded by Looper on the clean PR head after the
PR-reference commit: make verify, focused terminal race checks, focused
application IO/HTTP/PTY race checks, then git diff --check. The canonical PR and
controller evidence retain the exact candidate and command list. These task
checks do not claim final-candidate validation or dev integration.

## Risks and Follow-ups

None recorded.
