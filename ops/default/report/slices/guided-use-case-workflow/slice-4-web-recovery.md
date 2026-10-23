# Slice 4: Web Recovery, Controls and Results

Status: reviewing
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-4-guided-web-recovery`
PR: `#433`

## Purpose

Complete the resumable Web journey with native controls, deferred ideas and
receipt-backed results. Task commits: recovery and controls
`0ea3ff90e3d4bc5e35b36b7ad188cde3c8cb0274`, deferred ideas and results
`c6fd84c0197842011a2ae6cdcececaf739894c56`, complete Web acceptance and guidance
`0fa7cec16dc3e6b02f3db1d593e3cf4c0081da69`.

Plan: [Approved slice map](../../../plan/guided-use-case-workflow.md).
Tracker: [Tasks and evidence](../../../tracker/guided-use-case-workflow.md).

## Delivered Behavior

- Provider-bound input is retained before dispatch on the managed host lifetime.
  Client cancellation does not discard an accepted operation. Host startup and
  bounded inventory sweeps reconcile original pending owner operations across
  namespaces; inspection and polling remain read-only.
- Fresh pause/resume/cancel controls bind current authoring and native revisions.
  They preserve pending stage input, the original invocation, frozen membership
  and finite budgets. Native acceptance and actual quiescence are recorded
  separately. Uncertain responses recover the original receipt before mutable
  prerequisites; stopping uncertain admission only targets an existing exact
  request.
- Real dead child owners and graceful worker interruption recover through native
  ownership. Consumed attempts remain spent; operator stops prevent new dispatch.
  Native resume clears an interrupted task/validation marker while retaining
  set-closure recovery. Human review still reports unconfirmed delivery.
- Web control forms use current control revisions independently of an older
  pending stage. Paused pages suppress stage input and retries. Stale and foreign
  controls cannot replace retained work. Failed provider calls are not repeated
  by every inventory sweep.

- Later ideas are separately journaled with actor, time and original request.
  They remain pending separate review, including while paused and with results,
  without changing checkpoints, materialized membership or native budgets.
- Native result inspection verifies original admission, frozen task names,
  hashed validation/delivery/closing evidence and confirmed code locations.
  Only owned lifecycle reconciliation finalizes the authoring run. Unknown
  publication, an open PR, worker success or corrupt evidence cannot complete it.
- The complete HTTP journey preserves committed input after working-copy edits
  and reconciles original admission and publication responses without repeating
  work. Actual Chromium observes rendered input, controls, native approval and
  completed results. The Web User Guide and changelog describe the usable flow.

## Implementation Notes

Canonical authoring exclusion serializes stage/control mutations and recovers
its journal. Native admission/control/execution owners retain their own journals,
receipts, claims and budgets. Native activation checks guided authority under
its acceptance lock using a pure canonical read, avoiding lock inversion.

The managed host bounds stage jobs and native activations independently to 16,
uses bounded cross-namespace inventory pages and finite per-job deadlines, and
joins only its owned jobs on shutdown. Failure fingerprints suppress repeated
provider calls until an explicit retry or a new host can reconcile the request.

Ideas are bounded to 64 entries of 8192 bytes, in separate runtime files bound
by hashed session references. Native result reads have a ten-second deadline,
bounded task inventory and aggregate task-evidence bytes; they do not call
providers, recover journals or return worker output/credentials. Result state
retains native evidence references rather than introducing another receipt owner.

## Contracts Added or Changed

- Optional guided control state retains independent pending control, stop intent,
  native request/receipt, acknowledged revision and up to 128 completed controls.
- Fresh controls require displayed authoring/native revisions; original receipt
  replay precedes current revision and prerequisite checks. Only a verified
  rejection without an accepted native receipt releases an uncertain reservation.
- Native control requests optionally carry guided operation provenance. Absent
  optional fields preserve existing manual request serialization and behavior.
- Managed lifecycle recovery is separate from query/render/poll operations.
  Original native policy and operator stop intent remain authoritative.

- Optional `later_ideas` and `result` state preserves compatibility with existing
  manual authoring sessions. Idea identities cannot alias stage/control identities.
- Native result inspection binds the exact guided admission and frozen member
  contracts. Completion requires every admitted member and required set closure;
  pending stage/control operations and explicit stops retain priority.

## Files of Interest

- `internal/core/guided_control.go`, `guided_native_control.go` and
  `guided_control_admission.go`: retained controls, native authority and exact
  uncertain-admission lookup.
- `internal/core/guided_dispatch.go`, `guided_recovery.go` and `workflow_stop.go`:
  durable asynchronous input, bounded recovery inventory and native resume.
- `internal/app/guided_host_recovery.go`, `guided_progress_host.go` and
  `service_guided_recovery.go`: managed lifecycle and shared recovery operations.
- `internal/app/service_guided_control.go`, `web_guided_control.go` and the guided
  Web template: native owner composition and rendered revision guards.
- `docs/reference/authoring-sessions.md` and `web-ui-action-matrix.md`: recovery,
  control state and Web contract boundaries.

- `internal/core/guided_ideas.go`, `guided_result.go`, `workflow_result.go` and
  `workflow_result_validation.go`: separate idea ownership and pure native evidence.
- `internal/app/web_guided_result.go`, `web_guided_journey_test.go` and
  `web_guided_browser_test.go`: result presentation and full HTTP/browser evidence.
- `docs/how-to/user-guide/index.md` and `CHANGELOG.md`: usable Web journey guidance.

## Validation

T4.1 focused checks passed on the implementation content in
`0ea3ff90e3d4bc5e35b36b7ad188cde3c8cb0274`:

- `go test ./internal/core ./internal/app ./internal/webui -run '^Test(GuidedUseCase(Recovery|Control|Admission(Web|Host|StartedAck)|QuizWeb|SpecWeb)|Workflow(Stop|ClosingRecovery|ClosingStop|ControlCAS))' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^Test(GuidedUseCase(Recovery|Control|Admission(Web|Host|StartedAck)|QuizWeb|SpecWeb)|Workflow(Stop|ClosingRecovery|ClosingStop|ControlCAS))' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.
- `git diff --check`

Normal timings: core 13.670s, app 34.983s. Race timings: core 28.059s,
app 85.131s. App HTTP fixtures exercise real Web rendering; the separate Web UI
package has no matching T4.1 test in this selection. These local fixture checks
do not establish browser observation or the final Slice 4 gate. The tracker
retains the corrected fixture and graceful-restart failure context.

T4.2 focused idea/result checks passed normally and with race instrumentation:

- `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Ideas|Result)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Ideas|Result)' -count=1 -timeout=3m`

T4.3 focused checks passed:

- `go test ./internal/core ./internal/app -run '^TestGuidedUseCase(CheckpointRecovery|PreparationCommit|WebJourney)$' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestGuidedUseCaseWebJourney$' -count=1 -timeout=3m`
- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app -run '^TestGuidedUseCaseWebBrowser$' -count=1 -timeout=5m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.

HTTP evidence covers later edits, original admission retry, actual applied Git
publication with a lost response, observation without resubmission, controls,
separate ideas and receipt-backed final rendering. Real lost Git ref-update
regressions also pass. Separately, owned Chromium submitted rendered forms and
native approval, observing complete results and pending ideas at desktop and
narrow widths. The browser check ran, rather than skipping. Its provider and Git
fixtures are local and temporary. Owned resources finish through test cleanup.

The canonical PR and Looper checks ledger carry final clean-HEAD evidence in
this required order: `make verify`, the configured normal recovery/control/
ideas/result/WebJourney selection, its race selection, and `git diff --check`.
The final batch runs after the PR-reference follow-up commit and push.

## Risks and Follow-ups

Telegram and TUI adapters remain the next recorded slices after canonical
Slice 4 integration. Browser evidence here establishes the local Web journey;
external-forge certification remains owned by the scheduled acceptance checks.
