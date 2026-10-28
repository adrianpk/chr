# `workflow/forge`

## Purpose

This package owns forge submission dispatch, provider PR protocol and the immutable
policy that relates replies and observations to retained submission identity.

## Contract

`Client` owns all-state PR pagination, structured creation/read/branch requests,
bounded response decoding and provider-specific CLI arguments. `Endpoint` is the
resolved repository/PR and enabled command. `Commands` is the direct consumer IO
port; the host supplies its existing deadline/output-bounded runner. A nil request
body selects output-only IO; an explicit body selects structured input without a
separate boolean flag. GitHub hostname and Forgejo origin/port remain explicit.

Host adapter methods retain enabled configuration and canonical URL/repository
resolution before IO, including the original error order and local prerequisite
diagnostics. These are permanent host responsibilities. The leaf never loads
workspace configuration, receives Store, retries uncertain creation or confirms
receipt-backed delivery. Shared PR query/request/content values live in state;
original field and JSON tag order are preserved.

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
Mechanical token checks preserve the six original pure declarations and 11 new
PR decoder/content/input declarations after owner naming. Existing adapter tests
retain exact command/payload/pagination assertions through the host boundary;
real local Git/provider and focused concurrent PR checks pass. The dependency
checker permits only shared `workflow/state` within the module and rejects
coordinator, configuration and mutation imports.
