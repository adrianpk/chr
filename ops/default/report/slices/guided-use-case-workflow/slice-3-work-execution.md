# Slice 3: Automatic Work Preparation and Execution

Status: reviewing
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-3-guided-work-execution`
PR: pending

## Purpose

Continue the guided approved specification into committed native preparation
and exact policy-bound execution, with truthful Web progress and recoverable
owner operations.

Plan: [Approved slice map](../../../plan/guided-use-case-workflow.md).
Tracker: [Tasks and evidence](../../../tracker/guided-use-case-workflow.md).

## Delivered Behavior

- Plan generation consumes the committed dump, sufficient answered quiz and
  approved specification. Existing plan publication and native materialization
  owners retain the ordered tasks, contracts, tracker and receipt. A preparation
  checkpoint seals their initial images and full committed dependency chain.
- Retained actor intent binds one exact native preview, materialization receipt,
  checkpoint digest, canonical code target and effective policy. The native
  owner rejects replaced contracts, changed policy and missing target ownership.
  Older, later and unrelated tickets are excluded; later-ticket adoption is
  disabled for the invocation.
- Eligible Web input advances through preparation and admission without another
  routine start prompt. The managed server activates existing native
  continuation independently of the HTTP request and keeps original finite
  budgets, claims, review requirements and stop intentions.
- Lost plan, materialization, stage-commit, admission and start acknowledgements
  reconcile retained identities. A real worker can consume its original attempt
  before start-response recovery without a replacement invocation or attempt.
- Web queries expose native state, validation and confirmed-delivery counts,
  remaining work, admitted worker limits, review actions and complete committed
  preparation documents. Missing host, worker, provider or explicitly required
  certified publisher prerequisites remain actionable blocks. Independent
  specification, plan, ticket and work operations retain their manual behavior.

## Implementation Notes

The shared application service composes canonical authoring, specification,
plan-publication, materialization and native workflow owners. It retains each
request before effects and uses existing journals, receipts and revision guards
for recovery. Stage commits use the scoped Git checkpoint mechanism and never
select the code integration base or include unrelated index/worktree changes.

A managed host schedules only referenced guided invocations through
`ResumeRetainedWork`. Its bounded active map prevents duplicate local dispatch;
native claims and fences remain the cross-process authority. Querying does not
start the host, advance stages or refill budgets. Native review waits remain
visible without implying delivery from a validated worker or open pull request.

## Contracts Added or Changed

- Optional guided work state retains canonical plan publication, materialization,
  exact initial task membership, native admission and invocation references.
  Absent optional native admission fields preserve manual request serialization.
- `guided-work/v1` records actor, run, operation, preparation checkpoint and
  materialization provenance. Native admission checks that provenance under the
  authoring and native owners, then validates original policy and membership.
- Plans are bounded to 64 KiB, generated contracts to 128 KiB and exact native
  membership to 128 tasks. Scoped checkpoint observations allow at most 1024
  owned paths and 8 MiB. Existing native execution budgets remain authoritative.
- Provider readiness checks local configuration, executable and credential
  availability without creating or probing a pull request. Explicit native
  publisher selection requires its existing private certified target profile.
- Namespaced plan/archive and native materialization paths preserve explicit
  scope; absent optional namespace inputs retain the standalone default layout.

## Files of Interest

- `internal/core/guided_preparation.go` and `guided_admission.go`: canonical
  preparation, immutable membership/provenance guards and native references.
- `internal/app/service_guided_preparation.go`, `service_guided_plan_context.go`
  and `service_guided_admission.go`: committed-source generation and composition.
- `internal/app/guided_work_host.go`, `service_guided_work_readiness.go` and
  `service_guided_work_status.go`: bounded native activation and pure inspection.
- `internal/app/web_guided_work.go`, `internal/webui/guided_use_case.go` and its
  template: actual execution counts, review action and retained worker limits.
- `docs/reference/authoring-sessions.md` and `web-ui-action-matrix.md`: guided
  automatic progression and preserved independent operations.

## Validation

Focused task checks passed:

- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCasePreparation|WorkflowMaterializationRecovery|PlanPublicationRecovery|WebPlanDelivery|ServicePlanOrderPublication)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCasePreparation' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCase(Preparation|Admission)|ServiceWorkNative|TelegramWorkAdmission)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestGuidedUseCaseAdmission' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app ./internal/webui -run '^Test(GuidedUseCase(AdmissionWeb|AdmissionPrerequisites|AdmissionPolicy|AdmissionProvider|SpecWeb)|WebPlanDelivery|ServicePlanOrderPublication)' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestGuidedUseCaseAdmission(Web|Policy|Prerequisites)$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/... ./internal/webui/...` — zero issues.
- `git diff --check`

Initial development probes exposed a Git-fault argument mismatch, incomplete
checkpoint/coordinator fixtures and an invalid provider configuration key.
Corrected fixtures pass; the tracker retains their context. These are local
real Git/filesystem/HTTP and owned-worker results with deterministic model and
forge boundary providers, not live authentication or publication certification.

The canonical PR and controller retain final clean-HEAD evidence after the PR
reference follow-up commit. Required commands, in order:

- `make verify`
- `go test ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Preparation|Admission)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app ./internal/webui -run '^TestGuidedUseCase(Preparation|Admission)' -count=1 -timeout=3m`
- `git diff --check`

## Risks and Follow-ups

Slice 4 owns managed restart/reconnect recovery, guided controls, later ideas,
final results and the complete browser journey. Local readiness does not prove
remote authentication or certified publication. Exhaustive verification remains
with the scheduled repository gate.
