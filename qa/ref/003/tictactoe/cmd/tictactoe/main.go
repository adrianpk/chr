package main

import (
	"fmt"
	"os"
)

func main() {
	if wantsHelp(os.Args[1:]) {
		printUsage()
		return
	}

	fmt.Fprintln(os.Stdout, "tic-tac-toe CLI seed")
	fmt.Fprintln(os.Stdout, "gameplay loop is not implemented yet")
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

func printUsage() {
	fmt.Fprintln(os.Stdout, "Usage: tictactoe")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Run a local two-player tic-tac-toe game in the terminal.")
}
