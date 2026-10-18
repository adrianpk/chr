# Slice 7: Individual Spec Authoring

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-7-spec-authoring`
PR: `#404`

## Purpose

Deliver independent manual/generated specs under the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 6 is verified
integrated through PR #403 at `c9e0679951b63ed5257a4c40e3b9d3f3fcc69d40`.

## Delivered Behavior

Specs shows one title/opening-body card per actual document, including several
specs in one scope. Cards remain 300 × 218px with fixed 132px pale-green inset
previews. Responsive grids change column count; long titles, bodies and unbroken
text cannot enlarge cards or introduce preview scrollbars. There are no missing
scope placeholders, artifact counters, technical badges or detached slug forms.

One New action opens a vertical form with Title, a named Project-wide/use-case
selector, one large Body editor, a compact Draft checkbox and Create. Draft is
the default; turning it off creates approved directly without a precursor draft
or a model request. Both manual and generated cards open that exact spec with
prefilled content and Save. Editing retains its scope, native state, creation
time, unknown fields and generation provenance. Approval aliases normalize at
the shared reader.

Failed or stale writes keep the form intact. Explicit reload asks before
replacing pending modal input. Identical retries acknowledge once, including
a lost successful response. Background catalog refresh leaves an open editor
and its input intact. Keyboard focus stays in the modal and returns to its
catalog selection.

## Implementation Notes

- T7.1: `e130f1a27117359c27e8b53d4060752e285b0e5e` — shared manual create/save/state and guarded retained images.
- T7.2: `88c6b43cd5eb79506862897a785663af4883c3d0` — concrete listing, scoped context and recoverable generation publication.
- T7.3: `273fcfde69ef4b317bd441421847a4a39913e903` — Web cards, one editor, native state controls and owned visual/HTTP evidence.

Concrete listing is independent of the existing latest-in-scope inventory.
Spec generation receives project context, canonical general dump and matching
optional discovery while excluding unrelated cases, raw/archive material and
unrelated spec/plan documents. It observes the generated target or its absence
before model preparation and compares it under the shared writer. Both direct
and quiz-to-spec generation preserve allocated identity and reject an intervening
edit or changed scope.

## Contracts Added or Changed

`ReadSpecDocument`, `MutateSpec`, `ListSpecDocuments` and `SpecTargetRevision`
use the existing ops envelope, process-safe authoring owner and recovery journal.
Saves compare root/document/body/metadata revision; operation IDs bind immutable
arguments and replay one committed result. Each image remains at most 1 MiB,
receipt inventories at most 4096 entries, and catalog reads at most 4096 inventory
entries/8 MiB. Pending publications fail closed on read and recover only through
an explicit mutation. Target revisions remain caller-owned and are excluded
from model serialization. Source changes invalidate dependent observations
without rewriting historical plans or authorizing execution.

## Files of Interest

- `internal/core/spec_document.go`, `spec_mutation*.go`, `spec_catalog.go` and `spec_generation_publication.go`: native state, observations, publication and recovery.
- `internal/app/service_spec*.go`, `service_quiz.go` and `web_spec*.go`: shared operations, concrete catalog, context and adapters.
- `internal/webui/spec.go`, `assets/templates/spec.html`, `assets/static/spec.js`, `chr.js` and `chr.css`: minimal editor, retry/error retention and fixed card composition.
- `docs/reference/spec-authoring.md`, `docs/how-to/user-guide/index.md` and `CHANGELOG.md`: durable contract and usage.

## Validation

Successful development checks:

- `go test -race ./internal/core -run '^TestSpec(Mutation|Lifecycle|Metadata)' -count=1` — independent draft/approved creates, changed retries, exact generated identity/provenance, aliases, stale body/metadata, competing saves and interruption after every journal image.
- `go test -race ./internal/core ./internal/app -run '(Spec(Catalog|GenerationRevision|Mutation|Lifecycle|Metadata)|SpecDocumentCatalog|SpecGenerationScopes|ServiceGenerateSpec|QuizDraftSpec|Service.*DraftSpec)' -count=1` — concrete namespaces, malformed/symlink/size rejection, distinct latest queries, optional scoped context and generated target guards.
- `go test ./internal/app ./internal/webui -run '(WebSpec|Spec|HandlerArtifactCatalogAndDetailRoutes|HandlerRendersGenerationAndWorkActionForms)' -count=1` — real form creation, native state, retries, exact manual/generated edits, stale writes and non-HTMX error retention.
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_SPEC_SCREENSHOTS="$PWD/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-7" go test ./internal/app -run '^(TestWebSpecBrowser|TestWebDumpEditorBrowser|TestWebQuizCatalogBrowser)$' -count=1 -v` — owned spec journey/captures and affected shared modal/catalog regressions.
- `go test -race ./internal/core ./internal/app -run 'Spec' -count=1` — passed after adding generated scope protection.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui` — zero issues.
- `git diff --check` — passed.

The initial context fixture inherited a default dump provider; its result was
discarded. It now selects the deterministic fallback explicitly. The corrected
project/case context test passed in 0.03 seconds. All retained provider-dependent
evidence uses deterministic model boundaries. Browser tests use real HTTP,
filesystem state and owned Chromium, and stop their own fixtures after checks.

Final clean-HEAD `make verify`, focused spec race checks, affected browser checks
and `git diff --check` are recorded through the controller after PR-reference
commits. Its check record binds the exact HEAD and ordered command list.

## Captured Views

- [Catalog](files/slice-7/catalog-desktop.png): 1280 × 1000 CSS px, short/long/unbroken and generated documents.
- [New draft](files/slice-7/new-draft.png) and [New approved](files/slice-7/new-approved.png): same viewport, Project-wide/named scope, one Body and compact state control.
- [Generated editor](files/slice-7/generated-editor.png) and [Stale editor](files/slice-7/stale-editor.png): exact generated text and retained pending input on conflict.
- [Tablet catalog](files/slice-7/catalog-768.png): 768 × 844 CSS px, two columns.
- [Narrow catalog](files/slice-7/catalog-390.png) and [small catalog](files/slice-7/catalog-320.png): 390/320 × 844 CSS px, fixed cards without page overflow.
- [Narrow editor](files/slice-7/editor-narrow.png): 390 × 844 CSS px, one Body with visible Draft and Save controls.

All captures use 100% browser zoom, page scale 1 and DPR 1. Adjacent JSON records
retain viewport, page width, card/preview/editor rectangles and body-editor count.

## Risks and Follow-ups

None recorded.
