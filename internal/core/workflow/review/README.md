# `workflow/review`

## Purpose

This package owns immutable review input and decision policy and the bounded
completion flow that consumes the host's guarded decision-acceptance port.

## Contract

`InputBytes`, `InputHash` and `CheckInputHash` retain the bounded JSON and exact
input binding. `ParseDecision` rejects unknown fields, trailing output and invalid
decisions before acceptance. `CheckPolicy` and `DecisionState` preserve the current
policy rules and stored state spellings.

`Prepare` selects human waiting, the exact trusted decision or validated agentic
execution without claiming or writing. `Restrict` applies the existing human-only
invocation override by value. The host reads configuration and applies admission
inside the original Store lock, retaining explicit approval, atomic decisions
and reservation ownership. Only trusted admission reads the retained binding.

`Complete` combines one role's observed output/failure with its immutable input,
observes the watched context's interruption and calls `Decisions.Accept` within
the five-second detached cleanup bound. `Result` retains the partial binding and
role-selected repair task key. Command supervision and subsequent repair use the
host-owned lifecycle and original operation context.

`CheckNativeInput` owns native candidate/generation, owner, Git and capability
binding policy. `NativeProofs` composes the existing host record and preparation
validators without importing another stage or exposing mutable state. Current
entry/source fences, capability re-observation and atomic decisions remain host
operations. The exact record/Git predicate has one neutral state owner.

The host uses named observations, retained records, claimed preparation, selected
roles and hashed evidence references while keeping every context explicit.
Partial failure values and reservation cleanup remain local.

The private workflow adapter binds claim, prepared input and provider. It retains
credential/execution/current-input checks, explicit human approval, repair budgets,
repository locks, journals, atomic writes and recovery. No mutable Store or claim
capability is exposed to this package.

## Validation

Existing real Git, filesystem and process fixtures cover policy changes, source
and provider fences, repairs, interruption and journal recovery. Focused races cover
concurrent review, IO, stop and native provider fences. The dependency checker
permits only shared `workflow/state` and `contenthash` module dependencies, and
rejects coordinator, configuration and mutation imports.
