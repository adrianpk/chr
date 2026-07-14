# Tictactoe

Small Go CLI tic-tac-toe project used as a Chryso QA reference scenario.

The seed intentionally contains only a small, compilable base. It has enough
project context for Chryso to generate a gameplay specification, plan,
tracking document, and implementation tickets, but the gameplay loop itself is
not implemented yet.

From the parent `chr` repository:

```sh
make qa-reset QA_CASE=tictactoe QA_REF=003
make qa-test QA_CASE=tictactoe
```

After reset, open Chryso in the active project:

```sh
cd qa/project/tictactoe
chr
```

The intended exercise is to generate the spec, generate the plan, create plan
tracking, then run the generated ready tickets through the agent-work loop.

The reference also includes evidence strategy examples under:

```text
.chr/default/evidence-examples/
```

Those examples cover:

- documentation work where no tests or Gherkin are needed;
- CLI behavior work where active Gherkin is required;
- blocked validation that should prevent review handoff.

The acceptance feature at `.chr/default/acceptance/cli-gameplay.feature` makes
Gherkin active for observable CLI behavior during evidence strategy decisions.
