# Slice 8: Plan and ticket preparation

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-8-tui-plans`
PR: pending
Base: `882a81c7e90027879d481b485a144e02f71e49b7` on dev, after verified PR #461 integration and Slice 7 closure

## Purpose

Resolve G-004 and G-025 through one terminal plan/ticket preparation action,
retained recovery and case-preserving navigation to actual results.

## Delivered Behavior

- Plans `g` calls `GeneratePlanDelivery` with the observed case, target revision
  and a captured operation ID. It saves the plan and prepares native tickets
  together. Existing plans reconcile without implicit replacement or execution.
- Partial preparation retains its saved plan and exposes `y` Retry the original
  operation. The identical request resumes canonical publication/materialization;
  duplicate execution and reopening preparation allocate no duplicate tickets.
- `G` remains a confirmed, guarded replacement through `RegeneratePlan`.
  Cancelling writes nothing; changed sources or target revisions preserve the
  usable plan. The scoped catalog prefers the canonical generation target to
  newer retained native copies.
- The separate terminal tracking shortcut and palette command are removed.
  Independent CLI generation and tracking keep their existing service semantics.
- Successful Spec/Plan generation offers `o` Open document in the existing
  Action overlay. Combined plan preparation also offers `t` View tickets.
  Actions retain the case, open a populated Board column and perform no write.
  An outdated interaction cannot use result recovery or navigation.

## Implementation Notes

T8.1 is `446998e8544333e541aa067af244e4a0d5f4dc23`.
T8.2 is `0f98d041c40005ff7ce88728e13eb603089d8f00`.
T8.3 is `77f84ecbda38bac787f65d35c993cfbfbdbd6ade`.
T8.4 is `1523b132f98fdf8c8022c0a1572fdce6bf09341e`.

The first adapter regression reproduced a saved plan with zero native records.
Routing through the existing shared delivery owner corrected it. The terminal
reads resulting tickets through the canonical Board reader, not the legacy
index path. A failed-provider regression then reproduced missing original retry;
the adapter now preserves committed partial results and the UI captures their
request. Real retries, receipt replay and reopened preparation retain one plan
and one intended native ticket.

A catalog assertion exposed a newer materialized copy hiding the actual target;
the service now prefers the canonical generation document while retaining scoped
legacy fallback. Existing plan-only terminal tests were updated to the combined
contract or read-only refresh. Independent tracking-service regressions remain.
The result-navigation test exposed an empty inherited Board column; explicit
View tickets now selects the case's populated column.

One real PTY journey uses the production root loader and managed host. Only
provider/readiness responses and runtime binding are deterministic. It generates
and opens a Spec, fails ticket-contract generation after saving a plan, retries,
opens the plan and tickets, and cancels Regenerate. Native files, allocator,
receipts, terminal input and child cleanup are real. Final tasks remain pending
with no active attempt. The existing 45-second fixture ownership limit and all
assertions remain enabled.

## Contracts Added or Changed

Terminal requests carry explicit operation identity and observed target revision;
unbound writes fail before provider work. Queued commands retain their original
case/revision, and late results cannot replace a newer observation. Service files
remain transport-neutral. Existing publication, source guards, native ticket
allocation, receipts and finite execution controls remain their current owners;
no persistence schema or domain workflow is added.

Result actions reuse the existing operation overlay and document/Board shell.
They are offered only after successful generation has retained the relevant
output. Complete catalog-frame comparisons remain enabled; the existing
`documentCatalogFrames` producer updated 13 Plan frames for the approved controls,
while all Spec frames and the covered states/layouts remained unchanged.

## Files of Interest

- `internal/app/tui_plan_delivery.go`, `tui_adapters_spec_plan_ticket.go`:
  combined delivery, captured requests and retained partial results.
- `internal/app/service_plan_view.go`, `tui_adapters_loaders.go`: canonical
  target preference and observed revisions.
- `internal/tui/plan_actions.go`, `plan_recovery.go`, `operation_binding.go`:
  guarded original retry, replacement confirmation and late-result ownership.
- `internal/tui/generation_results.go`, `operation_feedback.go`,
  `spec_actions.go`: context-bound result actions in the existing shell.
- `internal/app/tui_plan_*test.go`, `tui_generation_result_test.go`,
  `internal/tui/plan_*test.go`: production IO/HTTP/PTY and binding evidence.
- User Guide, TUI reference, action matrix and changelog: current combined
  preparation controls and independent CLI behavior.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including complete catalog frames.
- `go test ./internal/app -run '^Test(TUIPlanDelivery|WebPlanDelivery|WebPlanDeliveryRecovery|ServiceListPlansReturnsScopedReadModels)$' -count=1 -timeout=3m` — passed after reproducing missing native records and using the canonical Board reader.
- `go test ./internal/app -run '^Test(TUIPlan(Retry|Delivery)|WebPlanDelivery(Recovery)?)$' -count=1 -timeout=3m` — passed after the original-retry correction.
- `go test ./internal/tui -run '^Test(PlanBinding|Model.*Plan.*|Operation.*)$' -count=1` — passed.
- `go test ./internal/app -run '^Test(TUIPlan.*|PlanPublicationFailures|ServiceMaterializePlanTracking.*|ServiceListPlansReturnsScopes)$' -count=1 -timeout=3m` — passed with captured writes, canonical catalog, replacement source/target conflicts and independent tracking protection.
- `go test ./internal/app -run '^Test(TUIPlanTerminal|TUIGenerationResult)$' -count=1 -timeout=3m` — passed after selecting the populated Board column.
- `go test ./internal/app -run '^Test(TUIPlan.*|TUIGenerationResult|WebPlanDelivery(Recovery)?|TUISpec.*)$' -count=1 -timeout=3m` — passed with the full terminal journey and shared Spec regressions.
- `go test ./internal/tui -run '^Test(Operation.*|Plan.*|Model.*Plan.*)$' -count=1` — passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed.

Final ordered validation will be recorded by Looper on the clean PR head after
PR-reference commits: make verify, focused terminal race checks, focused
application IO/HTTP/PTY race checks and git diff --check. Canonical PR/controller
evidence retains the exact candidate and ordered commands. These task checks do
not claim final-candidate validation or dev integration.

## Risks and Follow-ups

None recorded.
