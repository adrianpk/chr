# Slice 10: Read-only Plans Catalog

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-10-plan-catalog`
PR: `#407`

[PR #407](https://forge.adrianpk.com/adrianpk/chr/pulls/407) targets dev.

## Purpose

Deliver the approved [plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md) Slice 10. Slice 9
is verified integrated through PR #406 at
`78f115d0789263928130493460856e1a732c5e70`.

## Delivered Behavior

Plans lists concrete canonical documents across namespaces, including multiple
same-scope documents and retained archives. Cards show titles and bounded
opening body text using the fixed Specs geometry and pale inset preview.
Frontmatter, paths, status counters and missing-scope placeholders are absent.

Selecting a card opens its complete body through the existing escaped-text
renderer, with no manual Edit, Save or Append controls. Contextual generation
remains in the selected case. Trackers have no cards, detail route, editor or
creation action in the Web UI. Direct tracker references through plan and
generic artifact URLs are unavailable; internal files remain intact.

## Implementation Notes

T10.1 and T10.2 are complete. Concrete reads remain separate from latest-in-scope
lookup used by generation and existing adapters. The reader validates canonical
family/identity, frontmatter and envelope and uses existing coherent observation
and bounded file reads. Pending publication, corruption and symlink traversal
fail without recovery or migration. Missing or foreign detail references return
HTTP 404. The removed Web tracking action cannot be invoked by a retained form.

## Contracts Added or Changed

The central operation service exposes concrete plan-document listing. Reads are
bounded to 4096 inventory entries, 8 MiB aggregate and the existing 1 MiB per-file
limit. Plan preview selection strips managed frontmatter and a duplicate opening
title; complete detail content remains available. Generation, native approval,
internal map publication and CLI/TUI operations retain their existing owners.

## Files of Interest

- `internal/core/plan_document.go` and `plan_catalog.go`: exact, bounded plan reads.
- `internal/app/service_plan_catalog.go` and `web_plan.go`: shared listing and Web mapping.
- `internal/app/web_actions.go`: removal of the legacy Web tracking operation.
- `internal/webui/assets/templates/plan.html`: fixed cards and read-only detail.
- `internal/app/web_plan_catalog_test.go`, `web_plan_internal_test.go` and browser test: real filesystem/HTTP/browser boundaries.
- User Guide and Web action matrix: current browsing and internal-map boundaries.

## Validation

Task-level checks passed:

- `go test ./internal/core ./internal/app ./internal/webui -run 'PlanCatalog|WebPlanInternal|HandlerArtifactCatalogAndDetailRoutes|HandlerRendersGenerationAndWorkActionForms|WebPlan(Context|Bindings|Native)$' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui`
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_PLAN_CATALOG_SCREENSHOTS=../../ops/default/report/slices/web-ui-workflow-consolidation/files/slice-10 go test ./internal/app -run '^TestWebPlanCatalogBrowser$' -count=1 -v`
- `git diff --check`

Tests cover same-scope identities, namespaces, complete body selection, corrupt
and foreign references, symlinks, file bounds, escaped script text, rejected
legacy tracking forms, mutation-free browsing and unchanged contextual planning.
Owned browser/server fixtures finish with their tests; no live runner or model
is needed.

[Desktop catalog](files/slice-10/catalog-1280.png),
[320px catalog](files/slice-10/catalog-320.png),
[desktop detail](files/slice-10/detail-desktop.png) and
[narrow detail](files/slice-10/detail-narrow.png) have adjacent JSON measurements.
Additional captures cover 768px and 390px catalogs and unbroken detail text.
At DPR 1 and 100% browser zoom, short, long and unbroken cards remain 300×218px;
previews remain 278×132px with clipped overflow and no internal scrollbar.
Complete detail text exceeds the preview limit and has zero editing controls.
All measured views avoid document-width overflow.

Final clean-HEAD validation is pending at this report-reference commit. The canonical PR
and controller ledger record required `make verify`, affected focused checks,
plan browser regressions and whitespace validation after PR-reference follow-up.

## Risks and Follow-ups

None recorded.
