package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"example.com/chc-qa-tictactoe/internal/game"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, in io.Reader, out io.Writer, errOut io.Writer) int {
	if wantsHelp(args) {
		printUsage(out)
		return 0
	}

	state := game.NewGame()
	scanner := bufio.NewScanner(in)

	for !state.Over() {
		fmt.Fprintln(out, renderBoard(state.Board()))
		fmt.Fprint(out, renderPrompt(state.CurrentPlayer()))

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(errOut, "Error reading input: %v\n", err)
			} else {
				fmt.Fprintln(errOut, "Input ended before game finished.")
			}

			return 1
		}

		position, err := game.ParsePosition(scanner.Text())
		if err != nil {
			fmt.Fprintln(out, renderRecoverableError(err))
			continue
		}

		next, err := state.Play(position)
		if err != nil {
			fmt.Fprintln(out, renderRecoverableError(err))
			continue
		}

		state = next
	}

	fmt.Fprintln(out, renderBoard(state.Board()))
	fmt.Fprintln(out, renderResult(state))

	return 0
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help":
			return true
		}
	}

	return false
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage: tictactoe")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Run a local two-player tic-tac-toe game in the terminal.")
}
