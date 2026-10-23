# Slice 4: Web Recovery, Controls and Results

Status: drafting
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-4-guided-web-recovery`
PR: pending

## Purpose

Complete the resumable Web journey with native controls, deferred ideas and
receipt-backed results. T4.1 is implemented in `0ea3ff90e3d4bc5e35b36b7ad188cde3c8cb0274`;
T4.2 and T4.3 remain pending.

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

## Implementation Notes

Canonical authoring exclusion serializes stage/control mutations and recovers
its journal. Native admission/control/execution owners retain their own journals,
receipts, claims and budgets. Native activation checks guided authority under
its acceptance lock using a pure canonical read, avoiding lock inversion.

The managed host bounds stage jobs and native activations independently to 16,
uses bounded cross-namespace inventory pages and finite per-job deadlines, and
joins only its owned jobs on shutdown. Failure fingerprints suppress repeated
provider calls until an explicit retry or a new host can reconcile the request.

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

## Risks and Follow-ups

T4.2 deferred ideas and receipt-backed final results remain pending. T4.3 owns
the complete Web journey, publication-response fault variants, targeted browser
evidence, User Guide and changelog. Final controller validation must cover the
completed PR head after its reference follow-up commit.
