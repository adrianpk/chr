# Slice 7: Individual Spec Authoring

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-7-spec-authoring`
PR: pending

## Purpose

Deliver individual manual/generated spec authoring under the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 6 is integrated
through PR #403 at `c9e0679951b63ed5257a4c40e3b9d3f3fcc69d40`.

## Delivered Behavior

Shared manual create/save operations now allocate independent spec artifacts, honor draft or approved directly, preserve generated identities and provenance, and reject stale body or metadata replacement.

## Implementation Notes

T7.1 uses the existing process-safe authoring owner, journal and bounded operation receipts. Creation does not call a model. T7.2 adds concrete listing across namespaces independently of the existing latest-in-scope inventory. Generation selects the canonical general dump and matching scoped discovery; unrelated cases and raw/archive material stay excluded. Both spec generation paths compare their observed target revision and publish through the same shared owner/journal. T7.3 Web authoring is pending.

## Contracts Added or Changed

`ReadSpecDocument` separates managed frontmatter and generation provenance from the editable body. `MutateSpec` binds a save to the exact root/document/body/metadata revision and replays a committed operation once. Approval aliases accepted/final normalize to approved on read; writes use draft/approved.

## Files of Interest

`internal/core/spec_document.go`, `spec_mutation*.go` and `internal/app/service_spec_authoring.go`.

## Validation

`go test -race ./internal/core -run '^TestSpec(Mutation|Lifecycle|Metadata)' -count=1` passed: distinct creates, direct approval, generated identity/provenance, lifecycle aliases, changed retry arguments, stale body/metadata, concurrent saves, rejected inputs and interruption after every journal image.

`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app` passed with zero issues.

`git diff --check` passed. Final controller validation follows the remaining task and PR-reference commits.

## Risks and Follow-ups

None recorded.

T7.2 validation: `go test -race ./internal/core ./internal/app -run '(Spec(Catalog|GenerationRevision|Mutation|Lifecycle|Metadata)|SpecDocumentCatalog|SpecGenerationScopes|ServiceGenerateSpec|QuizDraftSpec|Service.*DraftSpec)' -count=1` passed. This covers namespace enumeration, malformed/symlink/size rejection, stale generated targets, separate document/latest queries and matching optional discovery. Full affected lint passed with zero issues.

The initial context fixture inherited a default dump provider; its result was discarded. The fixture now sets the existing deterministic fallback provider explicitly. The corrected project/case context selection passed in 0.03 seconds without generation-provider access beyond the explicit test model.
