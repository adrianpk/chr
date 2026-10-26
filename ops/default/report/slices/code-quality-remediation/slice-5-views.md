# Slice 5: TUI model ownership and Web view sections

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-5-quality-views`
PR: `#442`

## Purpose

Reduce measured terminal model copying and separate terminal/Web presentation
state while preserving interaction, form bindings and application ownership.

## Delivered Behavior

NewModel returns one *Model implementing Bubble Tea's interface. Its 85 fields
have 17 navigation, layout, interaction, background and per-view owners. Internal
update stages mutate that owner and return commands or named handled outcomes;
only the imposed Update method retains the external result signature. Commands
capture requests/executors before leaving the event loop, and completion guards
retain their existing ordering. Ownership tests cover queued dependency binding
and rendering/navigation while background IO waits for completion.

Spec and Plan catalogs share presentation and panel-layout rules while keeping
separate generation, feature, regeneration and tracking actions. Ordered modal
and per-view dispatch/hints replace repeated conditionals. All 32 frozen complete
frames retain empty/missing/generated/busy states, selected records and maximized
panes. The plan adapter shares logging/projection reload while preserving its
three distinct service requests and zero-result failure behavior.

Guided Web views group all 47 original fields into existing presentation owners.
Seven leaf partials receive only their form, retry, actions, status, recovery,
progress or result section. Original owner/native revisions, request/question
identities, pending content, visible controls and HTML/form spellings remain.
Frozen zero/empty/populated images and field manifests preserve exact encoded
bytes. Existing adapter-neutral WorkflowWorkStatus components and their frozen
public JSON contract are reused.

All 50 assigned inventory rows have dispositions: 47 fixed occurrences, two
cohesive five-input panel renderers and one exact external Update interface.
Global temporary debt drops from 472 to 425, with ceiling seven and the same
single external exception. Three unreachable compatibility wrappers and their
unused suppressions are removed; active legacy document paths remain live.

## Implementation Notes

Separate mechanical commits preserve all 193 model/shell declarations across
16 focused files (largest 434 lines), and all 53 Web entrypoint declarations
across seven HTTP owners (largest 209 lines). Later semantic changes and one
signature-header normalization have separate evidence. D11's actual baseline
ranges are the Spec/Plan layouts; the source inventory's ticket-column label
is corrected by those ranges. Compact ticket columns already shared a renderer;
their wide input now uses a cohesive display row.

The same 128-ticket benchmark fixture retains three samples per case before
receiver changes and after implementation. The model remains 49,832 bytes.
Median Update results are:

| Path | Before B/op | After B/op | Before ns/op | After ns/op |
| --- | ---: | ---: | ---: | ---: |
| Idle | 229,376 | 0 | 51,532 | 14.61 |
| Pane focus | 345,162 | 1,096 | 88,149 | 847.7 |
| Resize | 172,032 | 0 | 41,288 | 13.98 |

Median plain/shell View times change from 225,533/692,490 ns to
122,040/497,802 ns. Their 703/1,404 allocation counts remain unchanged. These
measurements establish the named paths; they do not justify blanket pointer
changes to unrelated large values. The sample ledger identifies both production
revisions and retains every observation.

## Contracts Added or Changed

The terminal API uses a pointer model. Named presentation inputs/results preserve
geometry, panel content, Markdown defaults, harness pagination/diagnostics and
ticket fragment/error selection. New reusable state/results store no operation
context. Web sections remain projections; service results import no TUI/Web
types. HTTP routes, access checks, login ownership and logging remain unchanged.

## Files of Interest

- `internal/tui/model.go`, `model_state.go`, `model_update.go`, `key_modes.go`:
  pointer ownership, state groups and ordered dispatch.
- `internal/tui/document_catalog.go`, `model_ownership_test.go`,
  `catalog_render_test.go`: shared rendering, command ownership and frozen frames.
- `internal/webui/guided_sections.go`, `guided_projection.go`, guided templates:
  narrow section inputs and unchanged operation bindings.
- `internal/webui/quality_guided_wire_test.go`, `testdata/quality-guided/`:
  original field manifest and exact byte images.
- [All 50 dispositions](../../../plan/code-quality-remediation-slice-5-dispositions.csv).
- [Model fields](../../../plan/code-quality-remediation-slice-5-model-fields.csv),
  [guided fields](../../../plan/code-quality-remediation-slice-5-guided-fields.csv),
  [Web sections](../../../plan/code-quality-remediation-slice-5-web-sections.csv).
- [Model/shell moves](../../../plan/code-quality-remediation-slice-5-moves.csv),
  [Web moves](../../../plan/code-quality-remediation-slice-5-web-moves.csv),
  [all benchmark samples](../../../plan/code-quality-remediation-slice-5-benchmarks.csv).

## Validation

- `go test ./internal/tui ./internal/webui ./internal/app -run '^Test(DocumentCatalogFrames|Model|Tab|Mouse|TUI|LoadDocument|GuidedViewWire|QualityRecordWire|GuidedUseCaseTUI|CaseWorkflow|TicketEditor)' -count=1 -timeout=3m`
  — passed, including real guided TUI/plan, HTTP, filesystem and retained bindings.
- `go test ./internal/webui -count=1 -timeout=3m` — passed after the mechanical HTTP split.
- `go test ./internal/tui -run '^$' -bench '^BenchmarkModel(Update|View)$' -benchmem -benchtime=200ms -count=3`
  — passed before receiver changes and on clean production `739cdf1b3e1b6734a68d4ba7a0c891f030074f9c`.
- `go run ./scripts/goquality` — passed with 425 exact debt entries and one exception.
- Pinned strict app/TUI/Web lint passed; all moved declaration texts were checked
  before semantic extraction and the signature-format check passed after normalization.

Final validation runs through the controller after the PR-reference commit:
make verify, full TUI/Web package checks, focused real application journeys and
status bytes, focused ownership race checks, the two required Chromium journeys,
then git diff --check. The controller retains the exact clean HEAD, ordered
commands, results and logs; that record and PR Validation own final results.

## Risks and Follow-ups

None recorded.
