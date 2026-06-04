# Testing Strategy

Use this playbook when implementing or reviewing behavior changes.

## Objective

Choose and record the test strategy for each implementation without repeating
the full testing policy in every slice report.

## Baseline

- Use Go's standard `testing` package by default.
- Use table-driven tests as the baseline for deterministic input/output
  behavior, especially business rules, policy decisions, state transitions, and
  scoring/classification logic.
- Cover expected failures, edge cases, package boundary contracts, and input
  immutability when relevant.
- Do not skip tests for behavior changes.

## Desirable Techniques

The following techniques are desirable tools, not mandatory ceremony:

- property-based tests;
- fuzz tests;
- mutation testing;
- CRAP-style risk review;
- race tests;
- benchmark tests.

For business logic, explicitly decide whether each technique is useful for the
slice. Omitted techniques must be defensible from the behavior, risk, and
available repository tooling.

## When To Use

- Use property-based tests when the behavior has broad invariants over many
  generated inputs.
- Use fuzz tests when parsing, decoding, tokenizing, or handling untrusted
  external input.
- Use mutation testing when repository tooling exists and branch/condition
  confidence is otherwise weak.
- Use CRAP-style risk review for complex business logic, branching,
  calculations, or high-impact decisions.
- Use race tests when concurrency or shared mutable state changes.
- Use benchmarks when the changed code is on a measured performance path.

## Reporting

Slice reports should record the concrete commands run and outcomes in
`Validation`.

Do not paste this whole policy into each implementation report. Instead, include
a concise note only when a non-obvious testing decision affects review risk.
