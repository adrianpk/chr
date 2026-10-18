# Slice 9: Revision-Bound Order-Only Trackers

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-9-plan-order-tracker`
PR: `#406`

[PR #406](https://forge.adrianpk.com/adrianpk/chr/pulls/406) targets dev.

## Purpose

Deliver internal plan-revision order maps under the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 8 is verified
integrated through PR #405 at `e48bcd8e44991c7c89a7628928471d8039938f42`.

## Delivered Behavior

Generation retains an unbound companion for each committed plan revision.
Fresh generations retain separate revisions even when rendered body bytes match;
explicit operation retries return the committed revision without another model
call. Earlier bodies, maps, provenance and ticket history remain available.

Approved native materialization binds canonical task references once and
publishes the set, tickets, maps and registration together. Missing or changed
approval inputs fail before key allocation. A later draft remains independent
of active work. Guarded amendments retain existing identities, contracts,
completed tickets and receipts while binding approved future work to its new
revision. Direct record writes cannot replace the bound structural order.

Workers receive the current set-specific order map. Required missing or edited
images block preparing input until repair. Claims, validated work and real
receipt-finalized integration leave map bytes, structural hash and generation
time unchanged. Edited output requires exact observed-hash replacement. Legacy
v1 exports retain their rendering, checksum shape and optional role.

## Implementation Notes

T9.1, T9.2 and T9.3 are complete. Structured single-pass output and staged slice
details supply identities, titles and order without parsing generated Markdown
as execution authority. Generation allocates no tickets. The legacy tracking
action reports the retained companion and requires native approval for allocation.

The existing authoring journal publishes generation pairs and retry receipts.
Approved binding takes authoring exclusion before workflow exclusion and stages
all images through the existing native journal. Pending binding prevents a
competing plan publication. Recovery reuses the same allocation. Canonical
manifests retain revision files; archives reference the original binding and
source manifest without claiming nonexistent local copies.

V2 source discovery reads native sets/tasks without receipt payloads or mutable
processing fields. Immutable planned images remain separate from bound images.
Retained checksummed projection metadata supports registration recovery without
changing known generation time. Native repair restores required companions
without executing work or importing edits.

## Contracts Added or Changed

`PlanRevisionBinding`, structured intended order and caller-owned publication
operation IDs extend plan generation. Explicit approved workflow input supplies
the binding, immutable body reference and observed body hash. Approval validates
scope and retained intended order before allocation. Required v2 images retain
blank legacy progress cells. Existing v1 and body-only legacy reads do not
migrate artifacts. Legacy adoption remains explicit and cannot bypass required
revision binding.

The accepted ticket-native specification and owning references now distinguish
required v2 order maps from optional v1 progress views. The User Guide and
Unreleased notes describe revision retention, approval, amendment and repair.

## Files of Interest

- `internal/core/plan_order_publication.go`: retained revisions and current output guards.
- `internal/core/plan_publication_receipt.go`: committed-operation retry receipts.
- `internal/core/workflow_plan_order_approval.go`: approval and mutation ownership.
- `internal/core/workflow_plan_order_publish.go`: atomic native ticket/map binding.
- `internal/core/workflow_plan_order_context.go`: required worker context and repair.
- `internal/core/workflow_tracker_order.go`: structural v2 rendering and freshness.
- Plan-order tests: real filesystem/Git, interruption, history, native progress and integration.

## Validation

Task-level checks passed:

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `go test -race ./internal/core ./internal/app -run 'Plan|WorkflowTracker|WorkflowOrder|WorkflowAdoption' -skip 'Browser' -count=1`
- `git diff --check`

Fixtures own their filesystem, Git repositories, subprocesses and provider
boundaries. The focused selection includes approved binding/retry, interrupted
publication, old-revision retry after regeneration, structural guards, active
amendment rejection, required-context repair, actual Git integration and retained
completed-task receipts. Existing v1 and legacy adoption regressions also pass.

Final clean-HEAD validation is pending at this report-reference commit. The
canonical PR and controller ledger record the required `make verify`, affected
focused race and browser checks, and whitespace check after this follow-up.

## Risks and Follow-ups

None recorded.
