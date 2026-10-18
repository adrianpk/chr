# Slice 9: Revision-Bound Order-Only Trackers

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-9-plan-order-tracker`
PR: pending

## Purpose

Deliver internal plan-revision order maps under the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 8 is verified
integrated through PR #405 at `e48bcd8e44991c7c89a7628928471d8039938f42`.

## Delivered Behavior

Plan generation and regeneration now retain an unbound order companion for
every committed revision. The current tracker keeps its existing scoped path.
A new generation retains a new revision even when plan bytes match; a committed
operation retry returns its retained body without another model call.

The native tracker owner now supports v2 structural freshness. Real claims and
host-validated results leave current map bytes, metadata, source hash and
generation time unchanged. Structural changes stale the map; exact observed-hash
permission protects edited output. Existing v1 files retain their rendering and
checksum shape.

## Implementation Notes

T9.1 is complete. Structured single-pass output and staged slice details supply
stable slice/task IDs, positive order and titles without reading Markdown as
execution authority. Generated entries cannot supply allocated ticket references.

The existing authoring journal publishes the plan, manifest, current companion,
retained body/order/tracker/binding images and optional operation receipt as one
recoverable mutation. Revision files remain under the canonical plan's `files/`
area; archive metadata keeps the source binding and manifest without claiming
local copies. Explicit observed-hash permission protects edited current output.
Read-only pair inspection rejects missing or changed companions.

T9.2 is complete. The v2 source reader uses the existing set/task owners under
repository exclusion without opening receipt payloads. It retains immutable
planned-image hashes independently of the later bound images. Lost registration
can be reconstructed using the native revision binding and observed-output
permission without adding receipt dependencies.

T9.3 is active: approved materialization/amendment and execution-context binding.

## Contracts Added or Changed

`PlanRevisionBinding`, structured planned order and caller-owned publication
operation IDs extend the shared plan writer. New service generation requires
its companion; historical body-only payload decoding and low-level legacy
publication remain supported. The v2 journal accepts only bounded revision files
and existing tracker envelopes. The earlier journal path contract is unchanged.
Native sets and tracker metadata may retain the explicit revision binding. V2
hashes and renders order, titles and canonical references without mutable status,
phase, claim, attempt, review, receipt or closure values. Optional v1 exports
retain their earlier state projection and checksum compatibility.

## Files of Interest

- `internal/core/plan_order_publication.go`: retained revision images and current tracker guards.
- `internal/core/plan_publication_receipt.go`: root-bound committed-operation retry receipts.
- `internal/core/plan_revision_read.go`: non-mutating pair inspection.
- `internal/app/service_plan.go`: required companions for both generation backends.
- `internal/core/workflow_tracker_order.go`: versioned structural images and retained source integrity.
- `internal/core/workflow_tracker_sources.go`: receipt-independent v2 source observation.
- `docs/reference/workflow-trackers.md`: v1/v2 export and freshness contracts.
- Plan-order and workflow-tracker tests: real filesystem/Git, progress, legacy and interruption regressions.

## Validation

Task-level checks passed:

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui`
- `go test -race ./internal/core ./internal/app -run 'Plan' -skip 'Browser' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core`
- `go test -race ./internal/core ./internal/app -run 'Plan|WorkflowTracker|WorkflowOrder' -skip 'Browser' -count=1`
- `git diff --check`

The focused race selection covers plan publication/recovery/retry, both service
generation paths, native v2 progress, source/registration integrity and existing
v1 tracker recovery/receipt regressions. Fixtures substitute the model
boundary and own their filesystem/Git inputs. Final exact-HEAD development-gate
evidence will be recorded through the controller after T9.3.

## Risks and Follow-ups

Slice 9 is incomplete until T9.3 finishes. No PR is ready for integration.
