# Tictactoe

Small Go CLI tic-tac-toe project used as a ChrysoCoord QA reference scenario.

The seed intentionally contains only a small, compilable base. It has enough
project context for ChrysoCoord to generate a gameplay specification, plan,
tracking document, and implementation tickets, but the gameplay loop itself is
not implemented yet.

From the parent `chc` repository:

```sh
make qa-reset QA_CASE=tictactoe QA_REF=003
make qa-test QA_CASE=tictactoe
```

After reset, open ChrysoCoord in the active project:

```sh
cd qa/project/tictactoe
chc
```

The intended exercise is to generate the spec, generate the plan, create plan
tracking, then run the generated ready tickets through the agent-work loop.
