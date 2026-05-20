# MVP Agent Proving Ground Slicing Plan

Status: active

## Purpose

Plan the `MVP Agent Proving Ground` slicing.

This slicing opens a temporary direct execution window before returning to the
official post-cutoff roadmap.

The purpose is to prove that `chc` can operate as a working agent runner: given
a controlled project and a concrete problem, it can run Codex in the agent
sandbox, validate the result, iterate through a bounded Ralph Loop, and leave
the code working.

## Source

Primary sources:

- `.prj/specs/specs/ticket-execution-context-hardening.md`
- `.prj/specs/specs/ticket-resolution-lifecycle.md`
- `.prj/specs/specs/ralph-loop.md`
- `.prj/specs/specs/agent-sandbox-and-runtime.md`
- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`

## Current State

Delivered mechanics:

- project workspaces can be initialized;
- runtime configuration and agent Dockerfile paths exist;
- Codex worker context can be materialized;
- Docker harness commands can be rendered and invoked;
- execution results and validation artifacts can be captured.

Still missing for the proof:

- a controlled project fixture with a deliberate bug and a missing feature;
- a direct proving-ground objective independent from full ticket workflow;
- a sandbox runtime known to run Codex for the fixture;
- a bounded Ralph Loop that feeds validation failures back into Codex;
- an end-to-end proof that the fixture is fixed with compile, lint, and tests
  passing.

## Slicing Scope

This slicing includes:

- a todolist fixture project source and materialization path;
- a hand-authored agent sandbox Dockerfile for the fixture;
- direct execution objective files for bugfix and feature tasks;
- validation commands for compile, lint, and tests;
- a minimal Ralph Loop execution boundary;
- evidence artifacts proving the final state.

## Out of Scope

This slicing does not include:

- notification or remote command surfaces;
- multi-project loops;
- generalized ticket planning automation;
- Git clone orchestration inside the container;
- pushing commits from inside the container;
- replacing the official ticket/spec/plan workflow.

Hardcoding is allowed inside this slicing when it shortens the path to the
proof. Later slicings can generalize the proven behavior.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Todolist Fixture Project

Status: delivered.

Expected branch: `feat/slice-1-proving-ground-fixture`
Expected PR title: `feat(slice-1): proving ground fixture`
Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-1-proving-ground-fixture.md`

Goal:

Create a controlled todolist project fixture that can be materialized under
`_workspace/testing/projects/todolist`.

Candidate scope:

- add fixture source files;
- include a deliberate backend bug;
- include a missing "mark task completed" feature;
- include `.chc/` runtime config and hand-authored sandbox Dockerfile;
- include compile, lint, and test validation commands;
- include fixture materialization documentation or helper;
- add slice implementation report;
- run `make verify`.

Expected outcome:

- the proving ground has a stable project target before any agent execution
  automation is added.

### Slice 2: Direct Proving Objective Contract

Status: in review.

Expected branch: `feat/slice-2-proving-objective-contract`
Expected PR title: `feat(slice-2): proving objective contract`
Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-2-proving-objective-contract.md`

Goal:

Define a direct execution input that can drive Codex without depending on the
full ticket workflow.

Candidate scope:

- define objective files for the bugfix and feature tasks;
- define expected validation commands;
- define artifact paths;
- define success, failed, and blocked outcomes;
- keep the contract fixture-scoped and explicit;
- add focused tests or static validation where practical.

Expected outcome:

- `chc` can receive a concrete proving objective even if ticket orchestration is
  bypassed for this window.

### Slice 3: Sandbox Codex Runner

Status: planned.

Expected branch: `feat/slice-3-proving-sandbox-runner`
Expected PR title: `feat(slice-3): proving sandbox runner`
Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-3-proving-sandbox-runner.md`

Goal:

Run Codex inside the fixture's agent sandbox with the project workspace mounted
writable.

Candidate scope:

- build or reuse the fixture agent image;
- materialize execution context;
- run `codex exec` in the sandbox;
- capture logs and result artifacts;
- avoid clone and push behavior for now;
- add dry-run or command-rendering tests.

Expected outcome:

- the proving ground can launch Codex in the intended sandbox.

### Slice 4: Validation Runner

Status: planned.

Expected branch: `feat/slice-4-proving-validation-runner`
Expected PR title: `feat(slice-4): proving validation runner`
Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-4-proving-validation-runner.md`

Goal:

Run compile, lint, and tests after an agent attempt and capture evidence.

Candidate scope:

- run fixture validation commands;
- capture stdout, stderr, and exit codes;
- write validation artifacts;
- classify validation failure enough for Ralph Loop input;
- add tests for failure and success handling.

Expected outcome:

- the proving ground can tell whether the agent actually fixed the project.

### Slice 5: Minimal Ralph Loop

Status: planned.

Expected branch: `feat/slice-5-proving-ralph-loop`
Expected PR title: `feat(slice-5): proving ralph loop`
Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-5-proving-ralph-loop.md`

Goal:

Retry failed agent attempts by feeding validation output back into Codex.

Candidate scope:

- define max attempts;
- append validation failure context to the next prompt;
- stop on pass, repeated failure, blocked state, or max attempts;
- preserve attempt evidence;
- keep the loop fixture-scoped for this proof.

Expected outcome:

- the proving ground can fix compile, lint, or test failures through bounded
  retries.

### Slice 6: End-to-End Agent Proof

Status: planned.

Expected branch: `feat/slice-6-proving-e2e-agent-proof`
Expected PR title: `feat(slice-6): proving e2e agent proof`
Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-6-proving-e2e-agent-proof.md`

Goal:

Demonstrate the full proof on the todolist fixture.

Candidate scope:

- materialize the fixture under `_workspace/testing/projects/todolist`;
- run the bugfix objective;
- run the feature objective;
- verify compile, lint, and tests pass;
- capture final evidence;
- document limitations and follow-up generalization.

Expected outcome:

- `chc` has proven that it can operate a sandboxed Codex agent that resolves a
  non-trivial code task with validation.

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- This slicing is post-cutoff MVP hardening.
- It intentionally prioritizes proof over generality.
- Hardcoded fixture paths and hand-authored sandbox details are acceptable in
  this slicing.
