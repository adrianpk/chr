# Slice 10: Read-only Plans Catalog

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-10-plan-catalog`
PR: pending

## Purpose

Deliver the approved [plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md) Slice 10. Slice 9
is verified integrated through PR #406 at
`78f115d0789263928130493460856e1a732c5e70`.

## Delivered Behavior

Plans lists concrete canonical documents, including multiple documents in one
scope and retained archives, with fixed title/opening-text cards. Selecting a
card reads its complete body through the existing escaped-text renderer.
Frontmatter, paths, status counters and missing-scope placeholders are absent.

## Implementation Notes

T10.1 is complete. T10.2 follows to remove the remaining legacy tracking action
and verify unavailable direct tracker views. Shared generation and execution
owners remain authoritative.

## Contracts Added or Changed

Concrete plan reads separate catalog membership from latest-in-scope generation
lookup. Reads validate canonical family/identity, frontmatter and envelope,
reject symlinks and preserve bounded inventory/file reads without recovery.

## Files of Interest

- `internal/app/web_artifacts.go`: catalog and document detail mapping.
- `internal/webui/assets/templates/plan.html`: plan presentation.

## Validation

- `go test ./internal/core ./internal/app ./internal/webui -run 'PlanCatalog|HandlerArtifactCatalogAndDetailRoutes|WebPlan(Context|Bindings|Native)$' -count=1` passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui` passed with zero issues.
- `git diff --check` passed.

## Risks and Follow-ups

None recorded.
