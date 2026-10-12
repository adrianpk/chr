# Slice 4: Canonical queue scheduling

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-4-native-queue-scheduling`
PR: pending

## Purpose

Schedule independent admitted integration units through the existing repository coordinator, frozen task frontier and shared role-capacity owner.

## Delivered Behavior

T4.1 selects each queued unit's current admitted task in durable queue-sequence order. It uses existing frozen-input and prerequisite checks, including cumulative task validation and prior receipt barriers. Paused, human-blocked and currently owned units are excluded. Status selection does not replay a transaction or acquire a claim. Launch rechecks the selection under the common repository lock and delegates the actual claim to the existing task and capacity transaction.

## Implementation Notes

The scheduler reads authoritative invocation entries and task records. A derived schedule or its revision/hash grants no execution ownership. Existing direct task claims also respect native entry pauses and blockers. There is no alternative admission list or receipt authority.

T4.1 implementation and canonical report introduction: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/2051bb53de7b8e452f455809915d47c48352ba61). Acta recorded the task as `01M4Q27DZTATW05Z6KQWY9JTN5`.

## Contracts Added or Changed

- Bounded native schedule pages with current task and invocation/entry bindings, stable pagination and exclusion counts.
- Explicit next-task claim under existing repository exclusion and capacity ownership.
- Read-only frozen-input/frontier ownership checks using the existing no-replay preparation policy.

## Files of Interest

- `internal/core/workflow_integration_schedule*.go`: authoritative task selection and guarded launch.
- `internal/core/workflow_claim.go`, `workflow_claim_inputs.go` and `workflow_coordinator_admission.go`: native pause and read-only ownership guards.

## Validation

T4.1 checks passed:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeSchedule|^TestNativeEntry|^TestNativeQueue|^TestNativeBudget|^TestNativeSubmission|^TestWorkflowClaim|^TestWorkflowFrontier' -count=1`
- `git diff --check`

Strict lint reported zero issues. The full core suite passed in 1030.781 seconds and app tests in 82.622 seconds. The expanded focused race check passed in 166.087 seconds. Tests cover actual namespace sequence allocation, stable restart/pagination, owner-token exclusion, real independent task progression during paused/human-blocked work, ordered cumulative task execution, missing receipt barriers, stale task bytes and interrupted peer transitions without query replay.

The first `go test ./internal/core -run '^TestNativeSchedule' -count=1` run failed paused/blocked progression because launch looked up a selected peer in the caller namespace. Launch now resolves the selected entry's project/namespace/invocation owner; the initial native schedule/entry/queue race regression then passed in 54.315 seconds. Strict-lint development attempts found complexity and spacing issues; cohesive claim preparation and test helpers were split before the passing gate. No assertions or role limits were relaxed.

Execution uses Go 1.27.1, repository-pinned golangci-lint v2.14.0, the retained neutral Git/default-agent fixtures and the forty-five-minute package allowance established by Slice 3. Host settings and explicit positive executor fixtures remain intact. Complete-candidate ordered checks are recorded through Looper after all four task and PR-reference commits.

## Risks and Follow-ups

T4.2 role fairness and candidate bounds, T4.3 canonical coordination and T4.4 concurrency certification remain in this slice's approved task list. Final canonical PR and ordered exact-HEAD checks follow all four tasks and their reference commits.
