# Serialized Loop Ticket Fixture

This fixture contains two parseable ticket files for manually exercising the
serialized integration barrier in the tic-tac-toe QA project.

The files are intentionally not stored in `.chr/default/tickets` so the default
reset remains a pre-ticket reference exercise.

To use the fixture after `make qa-reset-tictactoe`:

```sh
cd qa/project/tictactoe
cp .chr/default/fixtures/serialized-loop-tickets/CHR-*.md .chr/default/tickets/
chr ticket board
chr work run-all --use-case cli-gameplay --order key
```

The expected barrier is that CHR-2 must not start while CHR-1 is still waiting
for review, repair, or merge completion.
