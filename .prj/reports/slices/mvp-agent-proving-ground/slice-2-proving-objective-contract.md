# Slice 2: Direct Proving Objective Contract

Status: reviewing
Slicing: MVP Agent Proving Ground
Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
Branch: `feat/slice-2-proving-objective-contract`
PR: pending

## Purpose

Define the direct execution input for the proving ground before the full ticket
workflow is generalized.

## Delivered Behavior

The todolist fixture now includes a direct proving objective contract under:

```text
.chc/proving/objective-contract.md
```

It also includes two concrete objectives:

- `TODO-1-fix-task-id-reuse.md`;
- `TODO-2-complete-task.md`.

Each objective defines:

- objective ID;
- related ticket key;
- objective kind;
- artifact paths;
- max attempts;
- validation commands;
- success outcome;
- failed outcome;
- blocked outcome.

## Implementation Notes

The objective files are Markdown with YAML frontmatter. This keeps them readable
while leaving enough structure for later runner slices.

The contract is intentionally fixture-scoped. It is not the final generalized
ticket execution input.

## Contracts Added or Changed

New fixture contract:

- direct proving objectives live under `.chc/proving/objectives/`;
- each objective owns an artifact root under `.chc/runtime/runs/proving/`;
- `make compile`, `make lint`, `make test`, and `make verify` are the expected
  validation commands;
- outcomes are `succeeded`, `failed`, and `blocked`;
- `max_attempts` is declared per objective.

New validation helper:

```sh
./.prj/fixtures/mvp-agent-proving-ground/validate-objectives.sh
```

## Files of Interest

- `.prj/fixtures/mvp-agent-proving-ground/validate-objectives.sh`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chc/proving/objective-contract.md`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chc/proving/objectives/TODO-1-fix-task-id-reuse.md`
- `.prj/fixtures/mvp-agent-proving-ground/todolist/.chc/proving/objectives/TODO-2-complete-task.md`

## Validation

- `./.prj/fixtures/mvp-agent-proving-ground/validate-objectives.sh`: pass
- `./.prj/fixtures/mvp-agent-proving-ground/materialize-todolist.sh`: pass
- `make compile` in `_workspace/testing/projects/todolist`: pass
- `make lint` in `_workspace/testing/projects/todolist`: pass
- `make verify` in `_workspace/testing/projects/todolist`: expected failure
  matching the two proving objectives
- `make verify`: pass

## Risks and Follow-ups

The validation helper only checks objective structure. Later slices must parse
or consume the objective contract when running Codex and Ralph Loop attempts.
