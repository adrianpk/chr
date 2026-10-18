# Slice 8: Contextual Safe Plan Generation

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-8-plan-generation`
PR: pending

## Purpose

Deliver readiness-bound contextual plan generation under the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 7 is verified
integrated through PR #404 at `e000dfd64924cb9c798313cab12ed3b1c16ebf5f`.

## Delivered Behavior

The shared generation operation now requires sufficient discovery, keeps every model stage on captured inputs and compares sources and the target plan before publication. Contextual Web controls remain pending.

## Implementation Notes

T8.1: `c3958c22811a` — root/scope-bound source observations and absent/insufficient/sufficient assessment.

T8.2 stages model output in memory, validates it, rechecks the observation under the shared authoring owner, and publishes the archive/body/metadata as one recoverable journal operation. Competing creates and stale source/target bindings fail without replacing current work. Original archive metadata and source bindings are retained. Existing project-scoped baseline/API behavior remains available; it cannot waive the discovery assessment. T8.3 contextual controls are next.

## Contracts Added or Changed

`ObservePlanSources` retains selected project/case context, canonical general and matching scoped dumps, actual retained quiz answers and applicable specs. Two complete reads reject torn observations. Hashes bind selected source identities and retained revisions, including absent quiz state and newly added matching inputs. Reads never recover or allocate state. Assembled context is limited to 96 KiB; each owner retains its document/inventory bounds.

Readiness uses the existing configured structured provider and 60-second bound. Absent meaningful discovery fails before model evaluation. Strict output validation binds the assessment hash and rejects missing, contradictory, invalid, failed or cancelled results. It imposes neither a minimum length nor a requirement for both discovery types.

Model output cannot supply publication guards. Exact source references/revisions and the observation hash are retained in plan metadata. Replacements advance native metadata revision and preserve unknown envelope fields. A pending journal blocks source/target observation until explicit mutation recovery completes. Readiness, provider/model errors, cancellation and stale inputs publish neither an archive nor a replacement.

## Files of Interest

- `internal/core/plan_sources*.go`: bounded source observation and discovery presence.
- `internal/app/service_plan_readiness*.go`: shared structured assessment and readable failure reasons.
- `internal/app/service_plan.go`, `service_plan_context.go` and `internal/core/plan_publication*.go`: immutable generation inputs, target guards and recoverable publication.

## Validation

- `go test ./internal/core ./internal/app -run '^(TestPlanSources|TestPlanDiscoveryPresence|TestPlanReadiness|TestPlanReadinessProviderBinding)$' -count=1` — passed. Real filesystem observations verify mutation-free reads, scope filtering and source invalidation; deterministic assessment substitutes verify absent, insufficient, sufficient and invalid/provider-failure states.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app` — zero issues.
- `go test -race ./internal/core ./internal/app -run 'Plan' -count=1` — passed, including shared discovery scenarios, model/readiness/cancellation/stale failures, competing creates, journal interruption/recovery and affected TUI generation/materialization behavior. Assessment/model boundaries are deterministic; persistence and concurrency are real.
- `git diff --check` — passed.

## Risks and Follow-ups

None recorded.
