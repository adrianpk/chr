# Slice 6: Guarded automatic publication

Status: delivered

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-6-workflow-publication`

PR: [#369](https://forge.adrianpk.com/adrianpk/chr/pulls/369)

## Purpose

Drive a retained slice candidate through exact review, combined validation,
guarded publication and verified shared finalization with resumable stops.

## Delivered Behavior

Human review waits for approval of the displayed immutable binding. Agentic
review runs a distinct supervised role; trusted policy grants current validated
inputs. Candidate, task, target or policy changes invalidate old approval.
Automatic repair retains the unaffected prefix and consumes a persistent budget
of three repairs per unit/invocation.

The configured merge, rebase or squash result is prepared on a separate owned
lane. The frozen host slice gate validates that combined tree and ordered
history. Publication requires current review and provider capability to protect
source, validated target base and exact result in one server transaction.
Current native Forgejo/GitHub source-only merge APIs retain an explicit blocker.

Before provider IO, an atomic Git common-directory journal records stable
submission identity, frozen authority, consumed attempt and target exclusion.
Accepted responses remain uncertain. Restart, cancellation, lease expiry and
linked namespace/project aliases cannot release that exclusion. Canonical
PR/Git proof or a stable-ID unpublished seal resolves the operation. The default
submission budget is three; explicit frozen limits permit 1 through 10.

The central operation observes delivery before preparing, reviewing or
submitting again. Moved bases require new combined validation and review;
verified delivery finalizes the shared receipt without another task worker or
submission. External rewritten results remain distinct from exact guarded
intent. Pause/cancel intent survives observation and task finalization. Sealed
attempt consumption survives repaired candidates; late responses cannot replace
confirmed external delivery or its receipt.

## Implementation Notes

`WorkflowStore` owns immutable review/preparation evidence, fenced roles,
repository process exclusion, common-directory publication journals and shared
receipt/task/set transitions. The central application service delegates through
trusted provider, preparation and Git observation ports. Constructors perform
no IO. One advance call processes only the retained slice and performs at most
one combined gate, one review and one submission.

Git preparation and verification use owned temporary bare repositories and
actual canonical targets. Merge checks both parents and the combined tree;
rebase checks every ordered rewritten tree; squash checks one combined commit.
Canonical identity/result/target and retained source evidence are rechecked after
IO. Historical prepared lanes retain common process exclusion without copied
operational locks in their validated trees.

Review/preparation history is bounded to 128 entries, ordered Git history to
4096 commits, publication history to ten attempts and repository journals to
4096. Files retain the 1 MiB limit. Submission and observation use three-minute
deadlines; supervised roles retain fenced lease/capacity limits. Corruption,
missing projected journals and unsupported protections fail closed.

## Contracts Added or Changed

- Immutable review inputs and decisions bind current policy, canonical PR, target, candidate and prepared evidence; manual approval names their exact hash.
- `chr-integration-candidate/v1` retains configured combined-result proof, isolated lane and successful host gate.
- Trusted publication capability requires atomic source/base guards and the exact prepared result; no method or legacy merge fallback is permitted.
- `chr-integration-journal/v1` retains durable intent, uncertainty, exclusion and consumed attempts in the Git common directory; invocation bindings project its guarded image.
- `external_delivery` distinguishes verified external results from the exact guarded intent. Native receipts name the confirmed submission; external receipts retain canonical PR binding identity.
- `AdvanceWorkflowSliceIntegration` returns the retained invocation and verified shared receipt while preserving execution stops and the current slice frontier.

## Files of Interest

- [Exact review and approval](../../../../../internal/core/workflow_review.go).
- [Combined preparation and capability](../../../../../internal/core/workflow_preparation.go).
- [Durable submission](../../../../../internal/core/workflow_submission.go).
- [Common-directory journal and exclusion](../../../../../internal/core/workflow_submission_store.go).
- [Shared publication operation](../../../../../internal/core/workflow_publication.go).
- [Central service](../../../../../internal/app/service_workflow_publication.go).
- [Host-death finalization checks](../../../../../internal/core/workflow_publication_recovery_test.go).
- [Execution reference](../../../../../docs/reference/workflow-execution.md).

## Validation

All four task pre-commit `make verify` checks passed with the retained owned CLI
curation fixture documented in the [Slice 2 report](slice-2-workflow-materialization.md).
Strict lint reported zero issues. T6.4 uses a scoped 30-minute Go test timeout;
its final refreshed commit gate disables test-result caching. Successful focused
race checks:

- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowReview|TestServiceWorkflowReview)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowPreparation|TestWorkflowReview|TestServiceWorkflowPreparation|TestServiceWorkflowReview)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowSubmission|TestServiceWorkflowSubmission)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowPublication|TestServiceWorkflowPublication)' -count=1`.
- `git diff --check`.

Actual bare-Git server transactions verify source and base together for all three
methods. Real subprocess, linked-worktree, concurrency, corruption and crash
fixtures cover role ownership, exact approval, moved targets, unknown responses,
budget retention, repaired candidates, paused/cancelled external delivery,
late responses and journaled receipt fan-out. Central service tests exercise the
same operation. Controlled adapters establish the publication contract.

Final PR-head evidence is recorded through the controller for this ordered list:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

All four commands passed on clean head `22050846d7373395ae776b37bb98df76208dbdf7`,
recorded in the controller state directory as the matching `checks/<HEAD>.json`
and log. Canonical PR #369 merged that exact head through Rebase + Fast-forward
into dev; canonical and local dev match. Acta merge event:
`01M4F3YG8WEZMVZYGE4EZ7XM80`. The four task tickets are solved. These candidate
checks are distinct from the delivery set's later full gate.

## Risks and Follow-ups

Live guarded-provider support requires the designated disposable repositories
and evidence recorded for Slice 12. Current native source-only API paths remain
blocked. Ordered continuation, adapter controls and default activation retain
their approved later slice assignments.
