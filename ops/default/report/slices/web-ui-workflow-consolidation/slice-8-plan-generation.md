# Slice 8: Contextual Safe Plan Generation

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-8-plan-generation`
PR: pending

## Purpose

Deliver contextual plan authoring under the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 7 is verified
integrated through PR #404 at `e000dfd64924cb9c798313cab12ed3b1c16ebf5f`.

## Delivered Behavior

A named case shows Generate Plan when its plan is absent and Regenerate Plan
for its displayed existing revision. The panel links discovery context, the
selected case's quiz and its current plan. Plans has no detached generation
forms. Errors retain the case form and its pending input; success replaces only
the planning panel. A stale operation keeps its original binding and requires
explicit reload. Native forms return to the same case.

The shared operation distinguishes absent discovery, insufficient information
and sufficient context. A complete dump alone, a short dump for a narrow case
or sufficient quiz answers can qualify. A manual spec alone does not waive
discovery. Readable missing-information reasons explain the next step. Spec
generation/approval and starting work remain separate operations.

## Implementation Notes

- T8.1: `c3958c22811ae517a606e249511b971f5353274d` — scoped immutable observations and structured readiness.
- T8.2: `b840a5cdccf2a4b5c64bf26656b6b5f995946193` — readiness enforcement, frozen model inputs and guarded recoverable publication.
- T8.3: contextual Web controls, native/HTTP/browser evidence and user-facing documentation; commit reference follows PR creation.

Model output is staged in memory and validated. The shared authoring owner
rechecks source identities/hashes and the target body/envelope before publishing
an archive and replacement together. Competing creates and stale operations
cannot overwrite current work. Archives retain the prior body, native metadata
and source bindings. Existing project-scoped API baseline behavior remains
available and cannot waive the shared discovery assessment.

## Contracts Added or Changed

`ObservePlanSources` includes selected project/case context, canonical general
and matching scoped dumps, actual retained quiz answers and applicable specs.
Two complete reads reject torn observations; root/scope-bound hashes include
source revisions and the selected inventory, including absent quiz state. Reads
allocate nothing and never recover writes. Context is limited to 96 KiB in
addition to the existing document/inventory bounds.

The existing configured structured provider assesses readiness within 60 seconds.
Absent meaningful discovery fails first. Strict output validation binds the
observation hash and rejects invalid, contradictory, failed or cancelled results.
There is no fixed length threshold or requirement for both discovery types.

Caller-owned target/source guards are excluded from model serialization.
Plan metadata retains exact source references/revisions and the observation hash,
advances its native revision and preserves unknown envelope fields. The existing
recovery journal publishes all images. Pending publication blocks source/target
observation until explicit mutation recovery. Readiness/model failures,
cancellation and stale inputs publish neither an archive nor a replacement.

## Files of Interest

- `internal/core/plan_sources*.go` and `plan_publication*.go`: scoped observations, target guards and recovery.
- `internal/app/service_plan*.go`: structured readiness, immutable stages and the shared publication boundary.
- `internal/app/web_plan*.go`, `internal/webui/plan.go` and `assets/templates/plan.html`: named case actions, retained input and native/HTMX results.
- `docs/how-to/user-guide/index.md` and `CHANGELOG.md`: current planning guidance and user-visible behavior.

## Validation

- `go test ./internal/core ./internal/app -run '^(TestPlanSources|TestPlanDiscoveryPresence|TestPlanReadiness|TestPlanReadinessProviderBinding)$' -count=1` — passed source observation and assessment contracts.
- `go test -race ./internal/core ./internal/app -run 'Plan' -count=1` — passed discovery variants, readiness/model/cancellation/stale failures, competing creates, interrupted journal recovery, affected TUI behavior and contextual HTTP contracts.
- `go test ./internal/app ./internal/webui -run '^(TestWebPlanContext|TestWebPlanBindings|TestWebPlanNative|TestHandlerRendersGenerationAndWorkActionForms)$' -count=1` — passed named/revision-bound Web actions, native results and removal of global generation forms.
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_PLAN_SCREENSHOTS="$PWD/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-8" go test ./internal/app -run '^(TestWebPlanBrowser|TestWebSpecBrowser)$' -count=1 -v` — passed owned planning journey/captures and affected shared spec-editor regression.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui` — zero issues.
- `git diff --check` — passed.

Model and semantic assessment boundaries are deterministic substitutes;
filesystem, HTTP, shared mutation/recovery and Chromium boundaries are real.
Owned fixtures stop after their checks. Final clean-HEAD `make verify`, focused
plan race checks, affected browser checks and `git diff --check` are recorded
through the controller after PR-reference commits. Its evidence binds the exact
HEAD and ordered command list. Exhaustive/live certification remains separate.

## Captured Views

- [Generate](files/slice-8/generate-desktop.png), [missing context](files/slice-8/missing-context.png) and [Regenerate](files/slice-8/regenerate-desktop.png): 1280 × 1000 CSS px.
- [Narrow planning](files/slice-8/regenerate-narrow.png): 390 × 844 CSS px, wrapped guidance/actions and retained case input without page overflow.

Captures use 100% browser zoom, DPR 1 and page scale 1. Adjacent JSON records
retain viewport, page width, panel geometry and current action.

## Risks and Follow-ups

None recorded.
