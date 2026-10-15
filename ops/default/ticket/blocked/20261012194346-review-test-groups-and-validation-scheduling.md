---
id: TKT-20261012194346
title: "Review test groups and validation scheduling"
status: blocked
kind: task
severity: unclassified
priority: normal
scope: infra
tags: testing, test-selection, validation
source: chat
reported_at: 2026-10-12T19:43:46Z
commits: []
ready_at: 2026-10-15T19:07:29Z
started_at: 2026-10-15T19:07:29Z
branch: dev
---

# Review test groups and validation scheduling

## Problem

Repository test commands run a mixture of isolated logic checks, filesystem and
Git integration, subprocess lifecycle and concurrency checks, and explicitly
enabled live-forge certification. The inspected `internal/core` and
`internal/app` suites have no build-tag or `testing.Short()` separation.

During native integration Slice 6, four task gates accumulated about 85 minutes
of `core` test execution. Final candidate validation accumulated another
85 minutes by running `make verify`, an additional uncached normal core/app
test command, and the race-instrumented core/app suite. `make verify` already
includes normal repository tests. These are package-level measurements;
per-test timings and category proportions have not been established.

The current local commit policy also requires `make verify` for shared-contract
changes. Test organization and validation policy need to be reviewed together.

## Requested Work

1. Inventory the test suite and classify actual dependencies and boundaries:
   isolated unit tests, filesystem/persistence integration, real Git integration,
   subprocess and cross-process coordination, workflow acceptance, and external
   provider certification. Do not infer the category from the filename alone.
2. Measure duration by test and group. Identify the dominant slow cases,
   deliberate waits, repeated setup, flakiness, and redundant suite executions.
   Prefer retained evidence before scheduling additional expensive runs.
3. Define useful group labels and select a supported execution mechanism.
   Evaluate build tags, short mode, package separation, and explicit runner
   selection before choosing. Treat race detection and fuzz campaigns as
   additional validation modes rather than mutually exclusive test categories.
4. Specify when each group runs: during local iteration, before a task commit,
   before a slice PR merge, after the complete delivery set, in nightly CI, and
   during explicitly authorized live certification. Include change/risk triggers,
   environment prerequisites, and when changed candidates invalidate evidence.
5. Define and apply the agreed grouping or tags when implementation is
   authorized. Align repository commands, testing documentation, local commit
   rules, and future runner configurations with the selected policy. Remove
   duplicate executions while preserving the required coverage and exact-commit
   evidence.

## Expected Outcome

- A measured inventory with category, dependencies, duration, and group membership.
- A reviewed grouping decision with exact commands, selection rules, and clear
  default inclusion/exclusion behavior.
- An execution matrix mapping delivery stages and change risks to test groups.
- A bounded implementation scope for tagging/grouping tests and updating their
  commands and documentation.
- Verification that the union of the defined groups covers the existing suite,
  including concurrency, recovery, failure paths, and external certification.
- Evidence that representative selections run the intended cases and that the
  complete required validation remains available without duplicate commands.

## Execution Boundary

This ticket captures follow-up work only. Keep it outside the active native
integration delivery set. Capturing it does not authorize test changes, new
expensive validation runs, reduced merge requirements, or changes to an active
controller or its frozen configuration. Triage and authorize execution before
applying grouping or policy changes. Coordinate live tests through the existing
explicit fixture authorization.

## Validation

For capture, validate ticket metadata, filename/id/timestamp consistency,
status-directory alignment, and whitespace. During authorized execution, retain
per-test timing evidence, demonstrate the selected group commands, check that
the full suite remains covered, and report measured changes in repeated work.

## Authorized Execution

On 2026-10-15, the maintainer authorized fast development checks and scheduled
Forgejo nightly execution for all exhaustive checks, staggered against other
projects. Preserve the active delivery controller and its frozen final gate.
The nightly validates an immutable `dev` candidate without promoting branches
or publishing a release.

## Implementation and Measured Selection

The development gate now runs strict lint, short Go tests, the scheduler script
regressions and the application build. Standard `testing.Short()` guards at
real Git fixture boundaries defer expensive workflow integration. Browser and
live-forge fixtures have explicit short-mode guards. Tests and assertions remain
present; full normal and race commands have no short-mode selection.

The named-test inventory contains 1,865 tests plus one fuzz seed entry point.
The measured uncached short run passed 1,449 named tests and the fuzz seed entry
point. It skipped 408 real Git workflow tests, seven explicitly opted-in live
certification tests, and one browser fixture. Actual execution establishes this
membership; filenames alone were not used as category labels. The short run
reported core 4.083 seconds and app 11.145 seconds. Complete per-test integration
timings will be retained by the first scheduled nightly rather than another
expensive interactive measurement.

The mechanism keeps existing package ownership and uses Go's supported short
mode. Build tags would hide test compilation from the normal gate. Moving
packages would require a broader ownership refactor. A large name-based local
allowlist would drift. Shared real-dependency fixture guards keep the selection
explicit while leaving cheap parser, policy and filesystem tests enabled.

| Stage | Required selection |
| --- | --- |
| Local iteration and task commit | Changed-behavior regressions and formatting |
| Slice integration and local set closure | `make verify`; affected integration regressions without `-short` |
| Concurrency change | Focused affected tests with `-race` |
| Daily Forgejo nightly | Lint, build, script checks, one uncached full local pass, one full race pass, browser actions and live certification |
| Explicit diagnosis | Existing full, coverage, acceptance and provider commands remain available |

The full normal/race union retains every package test, including manifest-bound
acceptance and fault cases. Browser and live tests run in distinct passes with
their opt-in enabled; the normal/race passes disable those opt-ins. Named browser
and live evidence must show actual passes, so a renamed or skipped selection
cannot produce a green gate. The acceptance suite is not repeated as a third
local pass. Race instrumentation remains a separate mode.

## Validation Results

- `make verify`: passed, strict lint zero issues, short tests, six script
  regressions and build; 37.788 seconds with a warm toolchain cache and some
  cached package results.
- Focused non-short `TestWorkflowRoundTrip`, `TestWorkflowGitAtomic` with
  merge/rebase/squash cases, and `TestServiceWorkflowTask`: passed; core 0.472
  seconds, app 1.048 seconds. The integration fixtures remain executable.
- The same focused selection with `-race`: passed; core 1.429 seconds,
  app 2.376 seconds.
- Short mode with live opt-in set and all seven live roots selected: skipped
  before fixture execution, passed in 0.005 seconds without provider actions.
- `make test-ci`: six regression tests passed, including failure propagation,
  exact candidate identity, complete selected-test evidence, local/live
  isolation, missing authorization and daily admission.
- Python AST and Forgejo YAML structure checks: passed; schedule-only trigger,
  Warsaw timezone, immutable captured dev checkout and bounded job duration.
- `git diff --check`: passed.

## Activation Blocker

The implementation is ready on `dev`; scheduled execution is not activated.
Forgejo discovers schedules only on the default branch, currently `main`.
Publishing there requires the explicit authorization in `.agents/VERSIONING.md`.
Only the workflow needs bootstrap publication; no application promotion or
release is implied.

The existing `arch` runner is active, but its account configuration and private
fixture state are not readable from the development account. Verify scoped
fixture authentication, Chromium, Go/C toolchain and activity recording under
that runner account before activation. Missing live prerequisites fail early.
No credentials, runner services, default branches or release state were changed.
The original controller's frozen final gate remains unchanged.
