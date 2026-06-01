# Testing and Engineering Doctrine

ChrysoCoord should help projects move toward stronger engineering practice
without assuming that every project shares the same stack, risk profile, or
operating discipline.

This document is a direction of travel. It describes the testing and engineering
approach ChrysoCoord should recommend, preserve, and eventually help enforce
where a project opts into that level of structure.

## Core Idea

The goal is not more tests for their own sake.

The goal is software that makes its assumptions explicit, verifies important
properties, fails within known boundaries, and leaves enough evidence for a
human to trust the result.

That means treating tests, assertions, limits, interfaces, and review evidence
as one system.

## Baseline Testing

Every project should keep a basic testing floor:

- example-based unit tests for expected behavior;
- table-driven tests for known matrices, edge cases, and regressions;
- integration tests where behavior depends on multiple packages or external
  boundaries;
- validation commands that can run before commit and before promotion.

These are the minimum controls. They prove that known behavior still works.
They do not prove that the system is resilient.

## Invariants and Boundaries

Stronger tests encode properties that should always hold.

Useful checks include:

- invariant tests for state that must never become invalid;
- property-based or fuzz tests for generated input spaces;
- fault injection for IO, persistence, network, timeout, retry, and partial
  failure behavior;
- boundedness checks for loops, queues, retries, buffers, workers, and
  concurrency;
- runtime assertions in critical paths where continuing after a violated
  assumption would make the result less trustworthy.

This approach is aligned with the spirit of TigerStyle: explicit limits,
defense in depth, logical interfaces, assertions, and systems that verify
themselves while running.

## Complexity Risk

Coverage alone is a weak signal.

The CRAP metric, Change Risk Anti-Patterns, is useful because it combines
cyclomatic complexity with test coverage. Complex code with weak coverage is
more dangerous than simple code with weak coverage.

Conceptually:

```text
CRAP(m) = CC^2 * (1 - coverage)^3 + CC
```

Where `CC` is cyclomatic complexity and `coverage` is a fraction between `0`
and `1`.

ChrysoCoord should treat this kind of signal as a risk radar, not as a
management target. The useful output is not a global score. The useful output
is a short list of complex, under-tested areas that deserve tests, refactoring,
or explicit acceptance of risk.

## Mutation and Test Strength

A test suite can execute code without proving that it would notice a meaningful
mistake.

Mutation testing helps evaluate whether tests actually detect broken behavior
by making small changes to code and checking whether tests fail. It is expensive
and should not be mandatory for every project or every slice, but it is valuable
for critical logic, high-risk refactors, and code that has already shown a
history of subtle regressions.

Property-based tests and mutation testing complement each other:

- property-based tests explore input space;
- mutation tests challenge test strength.

## Agent Guidance

When ChrysoCoord generates specs, plans, tickets, or review evidence, it should
prefer work that improves the project's ability to verify itself.

That means encouraging tickets and acceptance criteria that ask:

- what invariant does this feature preserve?
- what boundary can fail?
- what limit should be explicit?
- what evidence will prove the behavior changed correctly?
- does this add complexity that needs stronger tests?
- does this introduce a dependency or interface surface that changes the risk
  profile?

The agent should not blindly add heavyweight testing. It should choose the
smallest testing strategy that addresses the actual risk.

## Enforcement Direction

ChrysoCoord can start by recommending this doctrine and preserving it as project
context.

Future enforcement can be opt-in:

- project-level doctrine documents included in generation context;
- ticket templates that ask for invariants, boundaries, and validation evidence;
- review packets that call out complexity, weak tests, and scope drift;
- configurable gates for coverage, CRAP-like risk, mutation testing, fuzzing, or
  package-specific validation;
- warnings when an agent changes high-risk code without corresponding evidence.

The default should remain portable. A project may be a small library, a CLI, a
web service, or a safety-critical subsystem. The doctrine should shape the work
without pretending that one validation profile fits all projects.

The future `dev` to `main` enforcement direction is described in
[Dev to main gate direction](dev-main-gate-direction.md).
