# Slice 11: Shared operation surfaces

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-11-native-integration-surfaces`
PR: pending

## Purpose

Expose retained native integration progress and guarded operator controls through
the existing shared application service and its execution surfaces.

## Delivered Behavior

T11.1 is committed after passing development validation. Read-only service status now
includes the current integration unit, canonical target queue position, phase,
original source and prepared result, actual role occupancy, retained publication
state and bounded budget accounting. Native questions remain tied to their
displayed revision and hash. Claim generations, leases, process identities,
capability fingerprints and harness inputs are excluded from the public model.

Budget grants and decision answers use the existing shared control operation.
The caller supplies displayed invocation, entry, candidate and budget revisions;
the host derives the private claim fence and delegates mutation to the existing
domain owner. Exact review approval retains its existing binding and revision
guard, with the displayed integration fence checked when supplied.

T11.1 implementation: [98915f97987104ae0cb3bc67ddcf3f8f91dbab0d](https://forge.adrianpk.com/adrianpk/chr/commit/98915f97987104ae0cb3bc67ddcf3f8f91dbab0d).
Acta commit activity: `01M4Y0R3R8REK566SDYW7AMC0M`. T11.2 is committed after passing development validation. CLI and TUI displays consume
the same retained fields: canonical target queue and position, integration
phase, original source/base/prepared result, actual occupied agent slots,
remaining allowances, token accounting and precise blocker actions. Decision
questions retain their displayed revision, hash, options and authority. No view
infers idle lanes or publication success from approval.

T11.2 implementation: [18b307141f84534b78d38429db1f79a380417c1a](https://forge.adrianpk.com/adrianpk/chr/commit/18b307141f84534b78d38429db1f79a380417c1a).
Acta commit activity: `01M4Y4616JTEX8TRQ2J10X1GKA`. T11.3 is committed after passing development validation. Web, work/daemon commands
and manual service callers submit public invocation, entry, budget, candidate
and question fences to the same control operation. Native command and Web approval require
the displayed guards. Grants preserve consumption; explicit resolve/defer
answers neither approve publication nor consume an agent call. Web questions
retain their choice, authority and optional original-step or finding payload.
Legacy review controls without native accounting retain their prior behavior.

T11.3 implementation: [e37e275b49c8a16f1349c916be9f5045b4be811f](https://forge.adrianpk.com/adrianpk/chr/commit/e37e275b49c8a16f1349c916be9f5045b4be811f).
Acta commit activity: `01M4Y9ACQZTX6W5SSREHCWNX35`. T11.4 setup/reference guidance and final ordered
clean-HEAD validation are pending.

## Implementation Notes

Native inspection reads the current invocation, queue, budget, role capacity,
candidate and decision evidence under repository exclusion. It contacts no
provider and performs no recovery. The decision reader now uses the same
read-only exclusion instead of the mutation path that can replay a journal.

Public status and control results omit process credentials and lease fields.
Review waits and approval remain separate from the retained queue phase; an
adapter cannot project approval into publication readiness. Legacy invocations
without retained integration accounting keep their existing status projection.

## Contracts Added or Changed

- `WorkflowInspection` includes retained native integration observations.
- `WorkflowWorkStatus` includes an optional public integration projection.
- `WorkflowControlRequest` supports revision-bound budget grants and explicit
  decision answers through `ControlWorkExtent`.
- `ReadWorkflowIntegrationDecisions` reports a pending transition without
  replaying it.
- CLI and TUI presentation shares package-local fields over the public service
  projection; rendering cannot allocate roles or advance integration.
- `work` and `daemon workflow` expose guarded `grant` and `answer-decision`
  commands. Web forms carry the same public control and question fences.
- Transport input maps only explicit operator fields; the service retains
  private claim generations and the existing domain owner validates meaning.

## Files of Interest

- `internal/core/workflow_integration_inspection.go`: coherent retained domain
  observations without provider IO or transition recovery.
- `internal/app/service_workflow_integration_status.go` and
  `service_workflow_integration_budget.go`: public observations and accounting.
- `internal/app/service_workflow_integration_control.go`: displayed action
  fences and delegation to existing domain mutations.
- `internal/core/workflow_integration_inspection_test.go` and
  `internal/app/service_workflow_integration_*test.go`: actual native preparation,
  interrupted decisions, owned capacity and guarded operator actions.

- `internal/app/work_workflow_integration_view.go`: shared retained presentation
  fields for CLI/TUI and their adapter mappings.
- `internal/app/work_workflow_integration_*test.go`: real prepared-candidate
  text/JSON reads, actual serial ownership and bounded display cases.

- `internal/app/work_workflow_integration_cli.go` and
  `work_workflow_integration_input.go`: bounded command and Web input mapping.
- `internal/app/web_workflow_integration.go` and
  `internal/webui/assets/templates/workflow.html`: shared observations and
  displayed-guard forms, including explicit question responses.
- `internal/app/work_workflow_integration_decision_*test.go` and
  `web_workflow_integration_test.go`: actual sealed questions, grants, native
  approval and stale operator-action rejection through each entrypoint.

## Validation

The first actual native service probe failed because retained integration facts
were absent from status. It failed in 4.645 s. After adding the shared projection,
the same probe passed in 6.530 s.

- `go test ./internal/app -run '^TestServiceIntegrationStatus$' -count=1` — passed
  before extending the probe with exact phase, identity and accounting checks.
- `go test ./internal/core ./internal/app -run '^(TestNativeInspection.*|TestNativeDecisionInspection|TestServiceIntegration.*|TestServiceNativeReview|TestServiceWorkflowGrant)$' -count=1` — the first run passed core in 5.113 s and failed app in 23.578 s on two incorrect phase assertions. The corrected checks retain `reviewing` in the queue while separately verifying the invocation's review wait and approval.

The corrected focused checks passed in 12.360 s for core and 32.820 s for app.
Strict lint initially identified two complexity and twenty-two formatting issues;
focused helpers and spacing corrections preserve the configured bounds.

- `make lint-strict` — passed with zero issues.
- `go test ./internal/core ./internal/app -run '^(TestNativeInspection.*|TestNativeDecision.*|TestNativeQueue.*|TestNativeGrant.*|TestServiceIntegration.*|TestServiceNativeReview|TestServiceWorkflowGrant|TestWorkflowCLIControls|TestWorkflowCLIApproval|TestServiceWorkflowStatus|TestWorkflowStatus.*)$' -count=1` — passed; core 69.682 s and app 33.102 s.

Acta focused activity: `01M4XXSWZ91P9JV7JBWDQ20S78`. The 11-file development
source is frozen at digest
`8ddc53dce797218763bebd341c36bd8578efd8b5e4e46bfa598a2caf727428ad`.
Owned unit `chr-native-slice11-t111-validation-8ddc53dc.service` executes these
pre-commit checks, with source, command, process and result evidence under
`/tmp/chr-native-slice11-t11-1-*`:

- `go test -race ./internal/core ./internal/app -run '^(TestNativeInspection.*|TestNativeDecision.*|TestNativeQueue.*|TestNativeGrant.*|TestServiceIntegration.*|TestServiceNativeReview|TestServiceWorkflowGrant|TestWorkflowCLIControls|TestWorkflowCLIApproval|TestServiceWorkflowStatus|TestWorkflowStatus.*)$' -count=1` — passed; core 115.109 s and app 53.292 s. Acta `01M4XY9J5W70A20HWQ4D5VXTXM`.
- `make verify` — passed; zero lint issues, core 2754.086 s, app 126.105 s, and all TUI/Web packages passed. Acta `01M4Y0NX6YKJTA466DP5TVQ8DA`.

The T11.1 validation service exited successfully, and all 11 source hashes
remained unchanged. Its task traceability is recorded above.

T11.2 first reproduced the missing native phase in both actual CLI and TUI
entrypoints; the probe failed in 5.161 s and passed in 4.540 s after mapping the
shared public projection. The owned-role fixture initially assumed a lane before
admission, then supplied a ticket instead of the grouped ticket's owning slice.
The fixture now uses its frozen slice identity; admission guards remain intact.
The corrected display checks passed in 4.777 s. Ten strict-lint spacing issues
were corrected without changing the configured quality bounds.

- `make lint-strict` — passed with zero issues.
- `go test ./internal/app ./internal/tui -run '^(TestWorkflowIntegration.*|TestWorkflowCLI.*|TestTUIWorkflow.*|TestNativeWork.*|TestServiceIntegration.*|TestServiceWorkflowStatus)$' -count=1` — passed; app 33.812 s. The TUI package compiled with no tests matching that subset.

Acta T11.2 focused activity: `01M4Y19DFC1ADK6N1T1ZWWYEH6`. The five-file development
source is frozen at `3d98c77273f1916314042ea00886b32695cb45b5068c69dbf01cff119069f44a`.
Owned unit `chr-native-slice11-t112-validation-3d98c772.service` executes these pre-commit checks
with command, source, process and result evidence under
`/tmp/chr-native-slice11-t11-2-*`:

- `go test -race ./internal/app ./internal/tui -run '^(TestWorkflowIntegration.*|TestWorkflowCLI.*|TestTUIWorkflow.*|TestModelWorkflow.*|TestServiceIntegration.*|TestServiceWorkflowStatus)$' -count=1` — passed; app 64.400 s and TUI workflow model regressions 1.023 s. Acta `01M4Y1EB7NM6TFQ16NCK6CEYN3`.
- `make verify` — passed; zero lint issues, core 2809.193 s, app 128.702 s and all TUI/Web packages passed. Acta `01M4Y44X498KK6Z58E062WY9MX`.

The owned T11.2 validation service exited successfully; all five source hashes
remained unchanged. Its task traceability is recorded above.

T11.3 first reproduced the missing shared native phase in an actual Web
projection; the probe failed in 5.113 s. After adding the public mapping, native
Web status and existing CLI/Web review controls passed in 12.573 s. Actual
work/daemon/manual grants and native CLI approval passed in 22.116 s; rendered
Web grants, approval and stale guards passed in 17.925 s.

The sealed-question fixture reads the actual set-owned slice contract, checks
its hash and retains the real source commit manifest. Its initial compile check
incorrectly treated slices as independent record fields; it now reads the
existing set-owned descriptors without changing production contracts. Actual
question answers across commands/manual/HTTP and Web resolve/defer checks
passed in 20.211 s. Replay ownership/finding payload and bounded command-input
checks passed in 0.008 s. Control and test complexity helpers and spacing
corrections preserve all configured strict-lint limits.

- `make lint-strict` — passed with zero issues.
- `go test ./internal/app ./internal/tui ./internal/webui -run '^(TestWorkflowIntegration.*|TestWebIntegration.*|TestWorkflowCLI.*|TestTUIWorkflow.*|TestWebWorkflow.*|TestDaemonWorkflow.*|TestModelWorkflow.*|TestHandler.*|TestServiceIntegration.*|TestServiceWorkflowStatus)$' -count=1` — passed; app 107.195 s, TUI 0.007 s and Web handlers 0.112 s.

Acta T11.3 focused activity: `01M4Y68YY6QKNVF97KGNP8BEW5`. The 15-file source, including
the embedded Web template, is frozen at
`ff3a974203742015969c6f47c4e35dfd1809c55d2dd87a211ad532b2238461e7`. Owned unit
`chr-native-slice11-t113-validation-ff3a9742.service` executes these pre-commit checks
with command, source, process and result evidence under
`/tmp/chr-native-slice11-t11-3-*`:

- `go test -race ./internal/app ./internal/tui ./internal/webui -run '^(TestWorkflowIntegration.*|TestWebIntegration.*|TestWorkflowCLI.*|TestTUIWorkflow.*|TestWebWorkflow.*|TestDaemonWorkflow.*|TestModelWorkflow.*|TestHandler.*|TestServiceIntegration.*|TestServiceWorkflowStatus)$' -count=1` — passed; app 216.665 s, TUI 1.032 s and Web handlers 1.218 s. Acta `01M4Y6HYY5E4VEBPX4MXV2CRE7`.
- `make verify` — passed; zero lint issues, core 2855.905 s, app 198.637 s, TUI 0.073 s and Web 0.094 s; all remaining Web packages passed. Acta `01M4Y9980W7F4RHA53HB2PC1WF`.

The T11.3 validation service exited successfully and all 15 source/template hashes
remained unchanged. Its task traceability is recorded above. Final ordered
clean-HEAD checks belong to the controller after all task and PR-reference
commits.

## Risks and Follow-ups

None recorded.
