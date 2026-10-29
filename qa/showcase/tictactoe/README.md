# Tictactoe showcase

A complete, local two-player tic-tac-toe CLI generated through chr's
**CLI gameplay** workflow. The Go files are preserved from published commit
`0f897632dde259a190c052451bb5a23d8bd6c833`; the incomplete
[reference seed](../../ref/003/tictactoe/) remains separate.

## Play

From this directory:

```sh
go run ./cmd/tictactoe
```

Two people share the terminal. X starts. Enter a position from 1 to 9 for each
move. Invalid input leaves the board and current player unchanged. The game
ends with a win, a draw or end-of-input. Use `--help` for command help.

The module declares Go 1.24 and uses only the standard library. This capture
was verified with Go 1.27.1. Playing the preserved game needs no chr service,
model credentials or forge connection.

## Verify

```sh
go test ./...
printf '%s\n' 1 4 2 5 3 | go run ./cmd/tictactoe
printf '%s\n' 1 2 3 5 4 6 8 7 9 | go run ./cmd/tictactoe
```

The scripts demonstrate an X win and a draw. Package tests cover the rules,
process I/O and exact executable transcripts. Executable tests build one
binary and enforce a five-second timeout per case.

## Follow the capture

- [Web gallery](../../../docs/explanation/gallery-web.md)
- [TUI gallery](../../../docs/explanation/gallery-tui.md)
- [Capture guide](../../../docs/how-to/capture-showcase-galleries.md)
- [Inputs](inputs/cli-gameplay.json), [quiz answers](inputs/quiz-answers.md)
- [Capture and validation record](capture.json)

The record distinguishes host validation, Git publication and native delivery
confirmation, and records the operator assistance used in this run.
