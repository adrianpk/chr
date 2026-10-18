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

T7.1 uses the existing process-safe authoring owner, journal and bounded operation receipts. Creation does not call a model. T7.2 concrete listing/context and T7.3 Web authoring remain pending.

## Contracts Added or Changed

`ReadSpecDocument` separates managed frontmatter and generation provenance from the editable body. `MutateSpec` binds a save to the exact root/document/body/metadata revision and replays a committed operation once. Approval aliases accepted/final normalize to approved on read; writes use draft/approved.

## Files of Interest

`internal/core/spec_document.go`, `spec_mutation*.go` and `internal/app/service_spec_authoring.go`.

## Validation

`go test -race ./internal/core -run '^TestSpec(Mutation|Lifecycle|Metadata)' -count=1` passed: distinct creates, direct approval, generated identity/provenance, lifecycle aliases, changed retry arguments, stale body/metadata, concurrent saves, rejected inputs and interruption after every journal image.

`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app` passed with zero issues.

`git diff --check` passed. Final controller validation follows the remaining tasks and PR-reference commits.

## Risks and Follow-ups

None recorded.
