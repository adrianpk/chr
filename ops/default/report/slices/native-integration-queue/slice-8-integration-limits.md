# Slice 8: Persistent limits and decisions

Status: delivered
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-8-native-integration-limits`
PR: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382)

## Purpose

Retain finite integration and invocation budgets, share optional whole-call
reservations across agent roles, and require explicit current operator decisions
or bounded extensions before exhausted work can continue.

## Delivered Behavior

T8.1 adds admission-time merger and drift policy limits. Omitted values preserve
three merger launches and five drift revisions; explicit values are bounded from
one through ten. Existing worker, review, closing and submission limits remain
part of the retained invocation ledger. Every read verifies that admission
ceilings still agree with the frozen execution policy; positive grants retain
separate history and do not replace that admission image.

Call reservations require a future deadline within the retained worker timeout.
This applies to metered executor, reviewer, closing and merger reservations and
the existing unmetered merger mode. Invalid deadlines fail before consumption
changes. Retrying a reservation cannot replace its original deadline. Restart
and a new queue generation retain all previously charged counters and identities.

T8.2 freezes an optional positive invocation token ceiling before the first
worker can launch. Certified executor and reviewer calls reserve bounded input
and output allowances in the same ledger used by merger and closing-role
reservations. Missing whole-call enforcement, reporting, compatible units or
role certification rejects admission before process birth. Raw executor,
manual-worker and agentic-review commands cannot bypass a finite ceiling.

The host seals each certified call to its immutable task input, owned attempt
and generation. Process birth and the launched charge are retained before the
supervisor receives permission. Successful driver accounting settles actual
input plus output once; model text supplies no usage authority. Missing,
partial, stale, overflowing or truncated accounting retains the complete grant.
Host task and aggregate validation commands remain outside model accounting.
A finite ceiling cannot be introduced after an unmetered call was prepared;
the host rechecks the current ledger before granting a raw process permission.

T8.3 adds immutable candidate-bound questions under the existing integration
entry. Asking or deferring a question consumes no agent call and preserves a
waiting state across restart. Answers require the displayed record hash/revision,
current invocation, budget, entry and generation guards, and fresh source/base
observations. Each explicit response retains actor, required authority, reason,
prior revision and timestamp. All questions must be resolved before merger work
resumes; silence and partial answers do not resume it.

Resolved interpretations are sealed into a later merger input. The host verifies
the immutable operator record, original scope, requirements, conflict context
and unchanged validation contract. Typed replay-owner answers must name an
original source commit; explicit integration fixes require an actionable finding.
Neither response authorizes a scope change, an undeclared binary resolution,
a replacement gate or publication. A stopped repair checkpoint is independently
retained at the fresh entry revision before the resumed merger call.

Positive extensions append to the same invocation ledger without changing its
admission policy or any consumed calls, unknown allowances or historical grants.
They work before queue admission, including the first extension of an unmetered
worker invocation; the existing consumed history is migrated into the ledger.
The extension retains finite token mode when configured and preserves unmetered
mode otherwise. Existing absolute persistence and role-owner bounds still apply.

T8.4 verifies real driver crashes, timeouts, missing usage and host loss at the
process-birth and permission boundaries. Recovery retains the launched charge
and immutable process ownership. A surviving driver prevents replacement from
another process or linked worktree; its exit releases the serial slot while
unknown usage retains the complete grant. Restart cannot turn that allowance
into an uncharged replacement.

## Implementation Notes

The new optional policy fields use omitted zero values, so existing default
policy encodings remain compatible. Effective defaults are recorded in the
budget at admission. Entry retries retain explicit merger/drift policy values;
changed requested limits cannot silently replace frozen invocation authority.
The existing ledger, mutation exclusion, counter arithmetic and grant history
remain the owners of accounting.

The certified role driver uses `chr-agent-call/v1`. Its retained certificate
binds command arguments, executable hash, role capabilities, whole-call
limits/reporting and `input-output-tokens/v1` units. Cached input stays in input;
reasoning stays in output. The driver response carries explicit usage and a
separate model-result field. The existing merger broker now requires the same
accounting-unit contract for metered calls. No agent service or second ledger
is introduced.

Executor reservations precede queue admission, so their ownership is verified
against the sealed attempt input rather than a later queue generation. Existing
worker/reviewer attempt counters and reservations count the same owned call once.
A separate reservation for an already charged role/attempt is rejected.

Question revisions are immutable hash-named files under the owning queue entry.
The current entry points at the sealed record; each answer retains its previous
revision. The common store lock and existing mutation journal commit question,
entry and invocation images together. A merger's structured `needs_decision`
proposal is retained atomically with its stopped attempt and charged budget.
The adapter-neutral operation service delegates grant and decision actions to
these same owners. Exact retries retain the existing result without reapproving
or granting twice.

## Contracts Added or Changed

- `WorkflowExecutionPolicy` accepts `max_merger_attempts` and `max_drift_revisions` with finite automatic bounds.
- Budget admission limits must match the invocation's frozen execution policy.
- Agent-call reservation deadlines cannot exceed the retained worker timeout or be extended by retry identity.
- `WorkflowExecutionPolicy.agent_token_budget` freezes an optional positive invocation ceiling; omitted values preserve existing policy encodings.
- `ExecuteWorkflowTaskWithHarness` and `ReviewWorkflowSliceWithHarness` require a host-certified whole-call driver and finite per-call allowances.
- Budget reservations can retain an immutable attempt-input reference before queue admission; driver usage is retained independently of model text.

- `RecordWorkflowIntegrationDecisions`, `ReadWorkflowIntegrationDecisions` and `AnswerWorkflowIntegrationDecision` retain bounded questions and explicit current answers.
- Optional merger input decision evidence binds resolved interpretations and typed replay ownership to the original scope and contract.
- `GrantWorkflowIntegration` permits positive guarded pre-queue extensions and preserves consumed history and absolute bounds.
- Shared operation service methods delegate operator actions to the existing invocation owner.

## Files of Interest

- `internal/core/workflow_agent_harness.go`, `workflow_agent_reserve.go` and `workflow_agent_command.go`: certified role admission, owned launch binding and trusted accounting.
- `internal/core/workflow_agent_driver_test.go` and `workflow_agent_review_test.go`: real certified-driver execution and shared executor/reviewer usage.
- `internal/core/workflow_integration_limits.go`: default and bounded admission policy.
- `internal/core/workflow_integration_limits_test.go`: persisted policy, exhaustion and deadline guards.
- `internal/core/workflow_integration_budget_store.go`: frozen-policy authority checks.
- `internal/core/workflow_integration_budget_reserve.go`: pre-charge deadline admission.
- `internal/core/workflow_invocation.go` and `internal/core/workflow_entry.go`: policy bounds and compatible entry restoration.

- `internal/core/workflow_integration_decision*.go`: sealed questions, guarded answers, real merger persistence and typed repair choices.
- `internal/core/workflow_integration_grant.go` and `workflow_integration_operator_test.go`: early extensions, retained consumption and policy compatibility.
- `internal/app/service_workflow_operator.go`: adapter-neutral decision and grant ports.

- `internal/core/workflow_agent_recovery_test.go`: real driver fault, launch-boundary crash, orphan ownership and retained-budget recovery fixtures.

## Validation

The initial deadline regression failed for every role before correction:
`go test ./internal/core -run '^TestNativeBudgetDeadline$' -count=1` accepted
one-minute reservations under a five-second frozen worker timeout. The corrected
admission refuses those calls without changing retained budget state.

Development checks passed with Go 1.27.1:

- `make verify` — passed with zero lint issues; core 1853.274 s, app 99.190 s and all other packages passed.
- `go test ./internal/core -run '^(TestNativeBudgetDeadline|TestNativeLimitPolicy|TestNativeLimitExhaustion|TestNativeDeadlineRetry|TestWorkflowStartBounds|TestWorkflowEntry)$' -count=1` — passed in 2.308 s.
- `go test -race ./internal/core -run '^(TestNativeBudget|TestNativeLimit|TestNativeDeadline|TestNativeMergerUnmetered|TestWorkflowStartBounds|TestWorkflowEntry)' -count=1` — passed in 14.199 s after the test helper correction.
- `make lint-strict` — passed with zero issues.
- `git diff --check` — passed.

Initial linting required a smaller exhaustion test helper and explicit whitespace;
the corrected source retains the same assertions. The repository pre-commit
`make verify` check passed with all seven changed Go files unchanged throughout
validation. Owned checks exited naturally. Final exact-head slice validation is retained below.

T8.1 implementation: [94ea7e86745e9ab30d55bc9fa931fd23a91826bc](https://forge.adrianpk.com/adrianpk/chr/commit/94ea7e86745e9ab30d55bc9fa931fd23a91826bc).
Acta commit activity: `01M4T0JW66X49EPFWK9J7TQV4Q`.

T8.2's admission regression first failed with an uncertified worker launching
under a finite ceiling. A subsequent restart test exposed the old requirement
for a queue entry before any reservation; sealed attempt-input ownership now
covers executor calls without weakening queue-owner checks. A separate isolated
late-ceiling probe then failed because admission forgot an earlier unmetered
call; the corrected admission and process-permission guards prevent that gap.

T8.2 development checks passed with Go 1.27.1:

- `make verify` — passed with zero lint issues; core 2182.690 s, app 125.596 s and all other packages passed.
- `go test ./internal/core -run '^(TestNativeAgent.*|TestNativeBudget.*|TestNativeMergerSettlement|TestNativeEntry.*|TestWorkflow(StartBounds|Entry|TaskExact|ReviewModes).*)$' -count=1` — passed in 47.155 s after the late-admission correction.
- `go test ./internal/core -run '^TestNativeAgentPermitBudget$' -count=1` — passed in 0.347 s.
- `go test -race ./internal/core -run '^(TestNativeAgent.*|TestNativeBudget.*|TestNativeMergerSettlement)$' -count=1` — passed in 50.583 s after the correction.
- `make lint-strict` — passed with zero issues after function extraction and formatting corrections.
- `git diff --check` — passed.

The corrected-source pre-commit gate passed with all 24 changed Go files
unchanged throughout validation. The earlier source check also exited normally
but was superseded by the late-admission correction. Both owned check services
finished naturally. The initial admission regression command was
`go test ./internal/core -run '^TestNativeAgentAdmission$' -count=1`; its
uncertified worker was accepted before correction. The isolated late-ceiling
probe used `go test -overlay=/tmp/chr-native-agent-late-overlay.json ./internal/core -run '^TestNativeAgentLateCeiling$' -count=1`
and failed before the admission guard was added.

The focused fixtures include concurrent executor/merger/reviewer/closing
reservations, complete and unknown accounting after restart, incompatible meter
rejection, remaining-ceiling enforcement, and three executor calls plus a
separate reviewer without double-counting attempts. These are development checks; final exact-head controller validation is retained below.

T8.2 implementation: [e1e574d4c031a25452de66da886be0e1df7271a9](https://forge.adrianpk.com/adrianpk/chr/commit/e1e574d4c031a25452de66da886be0e1df7271a9).
Acta commit activity: `01M4TBXAE0MJ2MGP0TE9CH7CXF`.

T8.3's unmetered retry regression first failed with `workflow revision or image
conflict`: no budget ledger existed to receive a pre-queue worker extension.
The corrected first grant preserves the original consumed worker attempt and
policy while authorizing one bounded retry. The regression command was
`go test ./internal/core -run '^TestNativeGrantUnmetered$' -count=1`.

T8.3 development checks passed with Go 1.27.1:

- `make verify` — passed with zero lint issues; core 2136.587 s, app 115.808 s and all other packages passed.
- `go test ./internal/core ./internal/app -run '^(TestNativeDecision.*|TestNativeGrant.*|TestNativeBudget.*|TestNativeMergerExecution|TestServiceWorkflowGrant)$' -count=1` — passed; core 64.877 s, app 0.163 s.
- `go test -race ./internal/core ./internal/app -run '^(TestNativeDecision.*|TestNativeGrant.*|TestNativeBudget.*|TestNativeMergerExecution|TestServiceWorkflowGrant)$' -count=1` — passed; core 97.641 s, app 1.991 s.
- `make lint-strict` — passed with zero issues.
- `git diff --check` — passed.

The pre-commit repository check passed with all 27 changed Go files unchanged
throughout validation. All owned checks exited naturally. The real answered
repair retains its private checkpoint at the fresh revision before execution;
host acceptance remains pending validation and cannot replace review or publish
a result. Final exact-head controller validation is retained below.

T8.3 implementation: [cb24ddffc886af7db8fe2f746fdd0d9982bbcdd8](https://forge.adrianpk.com/adrianpk/chr/commit/cb24ddffc886af7db8fe2f746fdd0d9982bbcdd8).
Acta commit activity: `01M4TH7SCDF346JDG9Y5J8GKJQ`.

T8.4 development checks passed with Go 1.27.1:

- `go test ./internal/core -run '^(TestNativeAgent.*|TestNativeBudget.*|TestNativeMerger(Descendants|Timeout|Stop))$' -count=1` — passed in 38.816 s.
- `go test -race ./internal/core -run '^(TestNativeAgent.*|TestNativeBudget.*|TestNativeMerger(Descendants|Timeout|Stop))$' -count=1` — passed in 74.448 s.
- `make lint-strict` — passed with zero issues.
- `git diff --check` — passed.

The first birth-crash run missed its injection because the fixture hook was
attached to the caller store. Binding it to the claim's owning store corrected
the fixture; the isolated birth-crash test then passed in 0.318 s. Linting also
required extracting the stopped-process evidence assertion. Production code
and assertions were unchanged. Both test files were unchanged throughout the
successful normal and race checks. All owned fixtures and checks finished.
Final exact-head controller validation is retained below.

T8.4 implementation: [b0574cca109c65d46a00bbbfb0b426f35fe751d1](https://forge.adrianpk.com/adrianpk/chr/commit/b0574cca109c65d46a00bbbfb0b426f35fe751d1).
Acta commit activity: `01M4TJKYTBQ2KTKQWSNH7NNCGN`.

Final controller validation passed on clean HEAD `0096f6e960ecc25e7aae72a8a97fdc53c689418c` in configured order:

- `make verify` — zero lint issues; core 2114.878 s, app 113.772 s and all other packages passed.
- `go test ./internal/core ./internal/app -count=1` — core 2123.657 s, app 110.777 s; passed.
- `go test -race ./internal/core ./internal/app -count=1` — core 3652.947 s, app 300.915 s; passed.
- `git diff --check` — passed.

The controller retained evidence for the exact head and ordered command list.
Its owned validation service finished successfully; no fixture remains owned by
that check. Canonical integration: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382) at `0096f6e960ecc25e7aae72a8a97fdc53c689418c`
on 2026-10-13T18:09:40Z. Controller Acta merge: `01M4V79DVNVSXZG5D9TMZ0NMSX`.
All four task tickets are solved.

## Risks and Follow-ups

None recorded.
