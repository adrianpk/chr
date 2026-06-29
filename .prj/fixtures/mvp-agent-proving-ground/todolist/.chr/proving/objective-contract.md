# Direct Proving Objective Contract

## Purpose

A direct proving objective is the temporary execution input used by the MVP
Agent Proving Ground.

It lets `chr` run Codex against a concrete objective without requiring the full
ticket orchestration path to be generalized first.

## File Format

Objectives are Markdown files with YAML frontmatter.

Required frontmatter fields:

- `objective_id`: stable objective identifier.
- `ticket_key`: related fixture ticket key.
- `kind`: `bugfix` or `feature`.
- `artifact_root`: project-relative artifact root for this objective.
- `context_path`: project-relative materialized context path.
- `result_path`: project-relative worker result path.
- `log_path`: project-relative agent log path.
- `validation_path`: project-relative validation log path.
- `max_attempts`: maximum Ralph Loop attempts for this objective.

Required sections:

- `Objective`
- `Allowed Scope`
- `Validation Commands`
- `Success Outcome`
- `Failed Outcome`
- `Blocked Outcome`

## Outcome Contract

The proving runner should classify an objective as:

- `succeeded`: the agent changed the fixture and all validation commands pass.
- `failed`: validation still fails after allowed attempts.
- `blocked`: the agent cannot continue because required context, tools,
  credentials, or permissions are unavailable.

The worker result artifact should be JSON with at least:

```json
{
  "status": "succeeded",
  "summary": "",
  "changed_files": [],
  "validation": [],
  "blocked_reason": ""
}
```

## Validation Policy

The proving objective must list the validation commands to run after each agent
attempt. For this fixture the command set is:

```sh
make compile
make lint
make test
make verify
```

`make verify` is the final acceptance gate.

## Scope Rule

The agent may edit fixture application files and tests. It must not delete tests
to make validation pass, and it must not edit `.chr/runtime/runs/` except for
runtime artifacts created by the proving runner.
