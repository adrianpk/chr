# Slice 4: Canonical queue scheduling

Status: reviewing
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-4-native-queue-scheduling`
PR: pending

## Purpose

Schedule independent admitted integration units through the existing repository coordinator, frozen task frontier and shared role-capacity owner.

## Delivered Behavior

T4.1 selects each queued unit's current admitted task in durable queue-sequence order. It uses existing frozen-input and prerequisite checks, including cumulative task validation and prior receipt barriers. Paused, human-blocked and currently owned units are excluded. Status selection does not replay a transaction or acquire a claim. Launch rechecks the selection under the common repository lock and delegates the actual claim to the existing task and capacity transaction.

T4.2 gives eligible aggregate validation, combined preparation, merger-purpose reservations and candidate review priority over new implementation. A candidate needing its last validation/review is ranked ahead of new preparation. Every launching workflow purpose reserves from the existing repository pool. Human approval, paused/blocked units, occupied roles, unknown target publication and preparations at their target bound do not retain runnable integration priority.

Preparation consumes a retained slot in the owning queue entry. The default target bound is two; configured values are `1..8`, further bounded by admitted queue capacity. Pausing, blocking or restarting keeps the retained preparation slot. Verified terminal transition releases it; replacing the same unit does not allocate another slot. Existing pre-native preparation records count against the same target bound.

T4.3 normalizes safe HTTPS repository spellings before native admission, queue/role filtering and target-key derivation. Host case, DNS root spelling, an explicit default HTTPS port and the existing Git suffix resolve to one public identity. Nondefault ports and repository path case remain distinct; credential-bearing URLs and private transport aliases remain rejected as public identities. Linked checkouts and namespaces use the existing common Git directory for both worker capacity and publication journals.

Mutation checks the host coordinator registration under the common lock before replaying any transition. Missing registration with retained authority, a copied registration or a foreign-host owner blocks before domain mutation or provider use. Consistent files remain readable after registration loss, but a disconnected read cannot replay a journal. A registered owner can still recover its own interrupted transition.

T4.4 certifies actual native queue launches in separate host subprocesses and isolated worker checkouts. Default configuration admits one agent; configured capacity two permits two live workers with distinct slots and excludes a third native or legacy launch. Lowering the limit while both workers are live preserves their owners and denies another launch until both finish. Every admitted worker commits its source and passes its real task gate; accepted source proof remains pending canonical integration.

Two existing publication owners reach provider IO concurrently for independent targets after retaining their own common-directory intents. An unknown outcome excludes a queued native unit on that target while the other publication advances its actual Git ref and obtains its verified receipt. A separately retained genuine native unknown dispatch also preserves its journal and owning namespace byte for byte while a linked alias reads status and completes an unrelated native task.

## Implementation Notes

The scheduler reads authoritative invocation entries and task records. A derived schedule or its revision/hash grants no execution ownership. Existing direct task claims also respect native entry pauses and blockers. There is no alternative admission list or receipt authority.

Role selection reuses existing frozen-input, cumulative task, source candidate, prepared Git and current review-policy owners. It performs no forge IO or transition replay. The scheduler owns capacity ordering; each launch owner still performs its existing exact-input/provider checks and owns its credentials. Merger-purpose reservation uses the same pool; the bounded merger payload/harness remains the separately recorded Slice 6 task.

The configuration and upgrade-snapshot owners accept the new prepared limit while preserving immutable snapshots containing either prior option set. Omitting a newly represented option cannot authorize a changed value.

T4.1 implementation and canonical report introduction: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/2051bb53de7b8e452f455809915d47c48352ba61). Acta recorded the task as `01M4Q27DZTATW05Z6KQWY9JTN5`.

T4.2 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/05d438047230b65833949c84e7c8b1ac99190f33). Acta recorded the task as `01M4Q5MBSA7CN05RM6356V2KE0`.

T4.3 implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/35d39f6b6df235d79848fb59054fd7d8b2ef0da3). Acta recorded the task as `01M4QAACBMDJ1MYQ53FZ1K83R9`.

## Contracts Added or Changed

- Bounded native schedule pages with current task and invocation/entry bindings, stable pagination and exclusion counts.
- Explicit next-task claim under existing repository exclusion and capacity ownership.
- Read-only frozen-input/frontier ownership checks using the existing no-replay preparation policy.
- Bounded observational integration-role pages and priority checks at the existing common reservation boundary.
- `work.integration.max_prepared_per_target` with default two and linked-worktree policy consistency.
- Retained `preparation_reserved` queue-entry capacity across pauses and restart; this capacity marker grants no candidate, execution or publication authority.
- Strict compatibility for prior upgrade-snapshot option sets and the expanded effective CLI configuration.
- Canonical HTTPS host/port normalization at the existing repository identity owner; observational filters accept the same safe spellings.
- Coordinator validation before mutation/replay, including retained native records and journals when a copied namespace marker is missing.
- Consistent reads remain available without granting a disconnected coordinator recovery authority.

## Files of Interest

- `internal/core/workflow_integration_schedule*.go`: authoritative task selection and guarded launch.
- `internal/core/workflow_claim.go`, `workflow_claim_inputs.go` and `workflow_coordinator_admission.go`: native pause and read-only ownership guards.
- `internal/core/workflow_claim_lock.go` and existing preparation/review/closing launch owners: one role-aware reservation path.
- `internal/core/config_workflow.go` and `workflow_upgrade_journal.go`: prepared configuration and retained policy compatibility.
- `internal/core/workflow_integration_entry*.go`: durable preparation capacity and guarded release.
- `internal/core/workflow_integration_schedule_roles_test.go`: real validation/review/implementation, target bounds and approval waits.

- `internal/core/workflow_repository_identity.go` and `workflow_integration_query.go`: one canonical public identity for admission, target keys and filters.
- `internal/core/workflow_coordinator_mutation.go`, `workflow_lock.go`, `workflow_store.go` and `workflow_invocation_read.go`: registered mutation and recovery boundary.
- `internal/core/workflow_integration_coordinator_test.go` and `workflow_repository_identity_test.go`: real remote aliases, linked checkout capacity, guarded target barriers and copied-journal preservation.

- `internal/core/workflow_integration_schedule_process_test.go`: actual native subprocess overlap, default serial configuration, lowered live capacity and retained source proof.
- `internal/core/workflow_integration_schedule_targets_test.go` and existing parallel fixtures: independent actual publication, native queue target exclusion and unknown native dispatch isolation.

## Validation

T4.1 checks passed:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeSchedule|^TestNativeEntry|^TestNativeQueue|^TestNativeBudget|^TestNativeSubmission|^TestWorkflowClaim|^TestWorkflowFrontier' -count=1`
- `git diff --check`

Strict lint reported zero issues. The full core suite passed in 1030.781 seconds and app tests in 82.622 seconds. The expanded focused race check passed in 166.087 seconds. Tests cover actual namespace sequence allocation, stable restart/pagination, owner-token exclusion, real independent task progression during paused/human-blocked work, ordered cumulative task execution, missing receipt barriers, stale task bytes and interrupted peer transitions without query replay.

The first `go test ./internal/core -run '^TestNativeSchedule' -count=1` run failed paused/blocked progression because launch looked up a selected peer in the caller namespace. Launch now resolves the selected entry's project/namespace/invocation owner; the initial native schedule/entry/queue race regression then passed in 54.315 seconds. Strict-lint development attempts found complexity and spacing issues; cohesive claim preparation and test helpers were split before the passing gate. No assertions or role limits were relaxed.

T4.2 task checks passed:

- `make verify`
- `go test -race ./internal/core ./internal/app -run '^TestNativeRole|^TestNativePrepared|^TestNativeSchedule|^TestNativeEntry|^TestNativeQueue|^TestNativeSubmission|^TestNativeBudget|^TestNativeProfile|^TestNativeLegacy|^TestWorkflowCapacity|^TestWorkflowPreparation|^TestWorkflowReview|^TestWorkflowConfig|^TestWorkflowUpgrade|^TestInitWorkflow' -count=1`
- `git diff --check`

Core passed in 680.292 seconds and app in 8.638 seconds. Checks exercise real source validation, combined preparation, candidate review and independent task execution, serial reservations for every role purpose, target quota overflow/restart/pause/cancellation, human approval and blocked-role byte preservation, previous queue/budget/submission contracts, configuration bounds and immutable upgrade snapshots. The required full pre-commit gate passed with zero lint issues, core in 1195.940 seconds, app in 139.514 seconds and all other packages passing.

Development checks found and corrected a released preparation slot after pause and a fixture gate that required combined-target content during source validation. `go test ./internal/core -run '^TestNativeRole|^TestNativePrepared|^TestNativeSchedule|^TestWorkflowConfigBounds' -count=1` passed after both corrections. Later `make verify` and expanded focused race attempts exposed fixed-size configuration/capacity expectations and the upgrade owner rejecting the expanded option set. Those owners now verify the additional bound while retaining strict prior-snapshot and unknown-key checks. Strict lint complexity/spacing issues were corrected with cohesive helpers and formatting before the passing lint gate.

T4.3 task checks passed:

- `make verify`
- `go test -race ./internal/core -run '^TestWorkflowClaimForeignClone$|^TestNativeRepositoryAliases$|^TestNativeTarget|^TestNativeCoordinator|^TestWorkflowCoordinator|^TestWorkflowEntryRepository$' -count=1`
- `go test -race ./internal/core ./internal/app -run '^TestNativeRepository|^TestNativeTarget|^TestNativeCoordinator|^TestNativeSchedule|^TestNativeRole|^TestNativePrepared|^TestNativeEntry|^TestNativeQueue|^TestNativeSubmission|^TestWorkflowCoordinator|^TestWorkflowCapacity|^TestWorkflowEntry|^TestWorkflowStore|^TestWorkflowTransition|^TestWorkflowAdopt|^TestWorkflowUpgrade|^TestWorkflowMaterial' -count=1`
- `git diff --check`

The full pre-commit gate passed with zero lint issues, core in 1279.578 seconds, app in 113.484 seconds and all other packages passing. The final focused race check passed in 24.308 seconds. The expanded final-tree race regression passed for core in 276.815 seconds and app in 3.579 seconds. Tests exercise real configured Git remote aliases, canonical normalization stability, actual task admission and worker execution from a linked checkout, namespace-wide worker and preparation limits, a real retained native submission barrier, independent eligible implementation, and exact-byte preservation after rejected clone mutation. Missing registration, copied registration, foreign-host ownership and removal of a copied namespace marker all reject before journal replay. Registered restart recovery and read-only access after registration loss remain covered.

Development regression checks reproduced distinct default-port repository identities and transition replay before rejecting an unregistered clone: `go test ./internal/core -run '^TestNativeTargetAliases$|^TestNativeCoordinatorReplay$' -count=1` failed before the guards were implemented. `go test ./internal/core -run '^TestNativeCoordinatorReplay$/stripped_marker' -count=1` exposed the same replay gap when copied namespace registration was absent; retained native records and journals now block unregistered mutation as well. `go test ./internal/core -run '^TestNativeRepositoryAliases$/ipv6' -count=1` caught unstable IPv4-mapped IPv6 normalization during development; address normalization now preserves the mapped form. An early focused normal run also had an incorrect fixture expectation that uncertain publication stopped implementation; the test now checks the specified target integration barrier while proving eligible implementation remains observable. Development `make lint-strict` and `make verify` attempts found function-complexity and spacing issues; cohesive helpers and formatting corrected them before the passing lint gate.

A full development `make verify` run failed only `TestWorkflowClaimForeignClone`, which expected a generic conflict after inspecting copied ownership. Mutation now rejects missing coordinator registration earlier. The regression requires the specific coordinator error and also verifies unchanged cloned record bytes; the expanded focused check above passed after that correction.

T4.4 task checks passed:

- `make lint-strict`
- `go test -race ./internal/core -run '^TestNativeProcess|^TestWorkflowParallelExecution$|^TestWorkflowParallelBlocker$|^TestWorkflowParallelPublication$|^TestNativeTargetBarrier$|^TestNativeSubmissionUnknown$|^TestNativePreparedBound$' -count=1`
- `go test ./internal/core -run '^TestNativeProcessTargets$' -count=1`
- `git diff --check`

Strict lint reported zero issues. The focused race regression passed in 132.880 seconds; the final target-fixture normal check passed in 6.538 seconds. New subprocess tests cover AC-001, AC-002, AC-007, AC-009, AC-011 and AC-015, with the retained same-target publication regression covering AC-008. Earlier task tests cover ordering, receipt barriers, coordination, fairness and human waits. These checks exercise actual reservations, supervised workers, isolated Git commits, source validation, retained common target journals and canonical local-fixture publication/receipt evidence. Owned subprocesses finish through their release barriers and are reaped.

Development checks corrected an admission fixture that incorrectly assumed a worker lane existed before the first claim, and an independent-target PR fixture that requested the default base despite observing another base. A combined fixture attempt then failed because it tried to apply legacy submission accounting to a native entry; the final tests use each existing publication owner with its own record contract and verify their shared target barrier through an admitted native sibling. No submission owner or persistence contract was changed for that test. Strict lint found spacing and test-function complexity; formatting and a cohesive publication-result assertion helper resolved both.

Execution uses Go 1.27.1, repository-pinned golangci-lint v2.14.0, the retained neutral Git/default-agent fixtures and the forty-five-minute package allowance established by Slice 3. Host settings and explicit positive executor fixtures remain intact. Complete-candidate ordered checks are recorded through the controller after all four task and PR-reference commits.

## Risks and Follow-ups

All four task checks are complete. Canonical PR references and required ordered candidate checks remain before handoff. The controller records those checks against the final clean PR HEAD; task-level checks above do not substitute for that evidence.
