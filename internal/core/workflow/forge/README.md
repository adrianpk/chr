# `workflow/forge`

## Purpose

This package owns forge submission dispatch and the immutable policy that relates
provider replies and PR observations to retained submission identity.

## Contract

`Submitter` is the narrow guarded provider IO capability used by `Submit`.
`Submit` receives an already journaled intent, dispatches once under the watched
context, and invokes `Responses.Retain` within the original five-second detached
cleanup bound. It preserves the joined provider and retention errors.

The host binds `Responses` to exact invocation/submission journal identity and
retains intent exclusion, current authorization, stop supervision, definitive
observation, receipt transactions and recovery. A late response cannot replace a
concurrent confirmed observation. Acceptance alone cannot establish delivery or
authorize retrying an uncertain dispatch.

`ResponseState`, `ValidSubmissionState`, `MatchesPullRequest` and
`CheckUnpublishedProof` preserve the existing response/state spellings and exact
legacy or native publication identity predicates. The native receipt-bound PR
rules remain distinct from legacy source-head equality and ticket-native work.

## Validation

Existing real local Git/provider fixtures cover guarded submission, uncertain and
lost responses, budgets/policy, atomic exclusion, stop, journals/corruption and
native recovery/proof. Focused races cover stop, exclusion and lost responses.
Mechanical token checks preserve all six moved declarations. The dependency
checker permits only shared `workflow/state` within the module and rejects
coordinator, configuration and mutation imports.
