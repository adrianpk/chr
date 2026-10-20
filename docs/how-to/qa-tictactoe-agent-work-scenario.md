# QA Tictactoe Agent Work Scenario

This runbook exercises the tictactoe QA reference project from reset through
generated specs, generated tickets, and Codex-backed agent execution.

## Reset The Scenario

From the chr repository:

```sh
make qa-reset-tictactoe
```

This recreates `qa/project/tictactoe` from `qa/ref/003/tictactoe`, moves any
previous active copy to `qa/tmp/tictactoe`, and initializes an independent Git
repository in the active project.

Run the baseline tests:

```sh
make qa-test-tictactoe
```

The baseline should pass. The project is intentionally incomplete, but it is
compilable and has enough seed tests to verify the reset.

Run the smoke check before opening the TUI manually:

```sh
make qa-smoke-tictactoe
```

The smoke check resets the scenario, verifies the independent Git repository,
confirms work readiness is ready, and runs the baseline tests.

## Open chr

Open chr against the active scenario:

```sh
cd qa/project/tictactoe
chr
```

The project contains seeded project context, a `cli-gameplay` use case, raw and
curated dumps, and raw and curated quiz notes. It intentionally does not contain
generated specs, plans, plan tracking, or tickets for the primary gameplay flow.

## Generate Work

From the active project, generate artifacts for the gameplay use case:

```sh
chr spec generate --use-case cli-gameplay
chr plan generate --use-case cli-gameplay
chr plan tracking create --use-case cli-gameplay
```

Preview generated ready work:

```sh
chr work run-all --dry-run --use-case cli-gameplay --order key
```

Run the generated ticket loop:

```sh
chr work run-all --use-case cli-gameplay --order key
```

In Dash, blocked work readiness appears in the Current panel. Use `x` to apply
a safe repair when one is available, then retry the work action.

Use this scenario to observe whether chr can preserve preseeded human
context into generated artifacts, launch Codex worker executions for generated
tickets, run project validation, and move tickets to `reviewing` or `blocked`.

## Serialized Loop Barrier Fixture

The resettable reference also includes a two-ticket fixture under:

```text
ops/default/ticket/<id>/artifact.md
```

Use it when you want to exercise the loop barrier without regenerating tickets:

```sh
cd qa/project/tictactoe
copy the relevant `ops/default/ticket/<id>/artifact.md` fixtures into the active ticket workflow if needed.
chr ticket board
chr work run-all --use-case cli-gameplay --order key
```

Expected behavior:

- CHR-1 is selected before CHR-2.
- CHR-2 does not start while CHR-1 is waiting for review, repair, or merge.
- Dash shows the current phase and latest event from the persisted event
  stream.
- `chr report show --loop <loop-id>` includes the loop stop reason.

For a full `agentic_review` or `trusted_auto_merge` merge barrier test, attach
real pull request metadata to the fixture tickets and configure the review
provider before running `chr work run-all`.
