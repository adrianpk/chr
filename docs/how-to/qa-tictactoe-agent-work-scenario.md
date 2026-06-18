# QA Tictactoe Agent Work Scenario

This runbook exercises the tictactoe QA reference project from reset through
generated specs, generated tickets, and Codex-backed agent execution.

## Reset The Scenario

From the ChrysoCoord repository:

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

## Open CHC

Open CHC against the active scenario:

```sh
cd qa/project/tictactoe
chc
```

The project contains seeded project context, a `cli-gameplay` use case, raw and
curated dumps, and raw and curated quiz notes. It intentionally does not contain
generated specs, plans, plan tracking, or tickets for the primary gameplay flow.

## Generate Work

From the active project, generate artifacts for the gameplay use case:

```sh
chc spec generate --use-case cli-gameplay
chc plan generate --use-case cli-gameplay
chc plan tracking create --use-case cli-gameplay
```

Preview generated ready work:

```sh
chc work run-all --dry-run --use-case cli-gameplay --order key
```

Run the generated ticket loop:

```sh
chc work run-all --use-case cli-gameplay --order key
```

In Dash, blocked work readiness appears in the Current panel. Use `x` to apply
a safe repair when one is available, then retry the work action.

Use this scenario to observe whether ChrysoCoord can preserve preseeded human
context into generated artifacts, launch Codex worker executions for generated
tickets, run project validation, and move tickets to `reviewing` or `blocked`.
