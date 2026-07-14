# Dev to Main Gate Direction

Chryso does not currently implement this gate.

This document captures the need and direction for a future `dev` to `main`
promotion gate. The goal is to make promotion evidence-driven without turning
agent review into a linguistic judgment about whether code "looks good".

An extensive pre-spec draft for this direction is available in
[Dev to main tooling gate spec draft](dev-main-tooling-gate-spec-draft.md).

## Principle

Promotion should be based on facts produced by tools.

The preferred model is:

```text
Go tools + git + Chryso artifacts produce evidence.
Optional AI may summarize evidence.
The human decides promotion.
```

The gate should not depend on a model approving code. A model may help organize
or explain facts, but the facts should come from deterministic tooling wherever
possible.

## Evidence Sources

A future gate can collect evidence from:

- `git log` and `git diff` between `main` and `dev`;
- changed files, packages, and public surfaces;
- ticket, plan, and execution artifacts under `.chr/default/`;
- validation commands such as `go test ./...` and `golangci-lint`;
- coverage profiles and package coverage summaries;
- cyclomatic or cognitive complexity tooling;
- CRAP-like risk signals when both complexity and coverage are available;
- `govulncheck`, `go vet`, `staticcheck`, or equivalent configured tools;
- optional race tests, fuzz targets, mutation tests, or package-specific checks;
- documentation and reference updates when contracts change.

The output should include both machine-readable data and human-readable
Markdown.

## Report Shape

A future command could produce:

```text
ops/default/report/<id>/ (Markdown in artifact.md, JSON in evidence/gate-report.json)YYYYMMDD-dev-main.json
ops/default/report/<id>/ (Markdown in artifact.md, JSON in evidence/gate-report.json)YYYYMMDD-dev-main.md
```

The report should record:

- base and head SHAs;
- commits included;
- tickets and slices included;
- files and packages changed;
- tests and checks executed;
- coverage and complexity signals when available;
- artifacts or evidence linked from completed tickets;
- changed contracts, runtime behavior, persistence, configuration, or docs;
- warnings and hard failures according to project policy.

The Markdown report exists for review. The JSON report exists for automation.

## Deterministic Rules

Rules should be configurable and start conservative.

Useful early checks include:

- fail when the report does not match the current base and head SHAs;
- fail when completed tickets have no execution evidence;
- warn when Go logic changes without nearby tests or an explicit rationale;
- warn or fail when runtime, persistence, or configuration changes lack a
  compatibility note;
- warn when public contracts change without reference documentation updates;
- warn when complexity increases in weakly tested code;
- warn when high-risk paths change without extra evidence.

These checks should identify risk. They should not pretend that one global
score can decide whether a release is safe.

## Relation to Review

This gate is the enforcement side of the
[agent review model](agent-review-model.md).

Agent review should focus on accumulated coherence and evidence. The gate makes
that practical by collecting facts first. A reviewer, human or model-assisted,
can then reason from the same evidence instead of free-form code impressions.

## Relation to Testing Doctrine

This gate is also the promotion point for the
[testing and engineering doctrine](testing-and-engineering-doctrine.md).

The doctrine recommends invariants, boundaries, complexity risk, mutation
testing, property testing, and boundedness checks. The gate is where those ideas
can gradually become project-specific policy.

The default should remain portable. Stronger gates should be opt-in and tied to
project risk, not imposed as ceremony.
