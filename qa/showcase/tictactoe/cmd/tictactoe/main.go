package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"example.com/chr-qa-tictactoe/internal/game"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var err error
	if wantsHelp(args) {
		err = printUsage(stdout)
	} else {
		err = play(stdin, stdout)
	}
	if err != nil {
		// Diagnostics are best effort when a process stream has failed.
		_, _ = fmt.Fprintf(stderr, "tictactoe: %v\n", err)
		return 1
	}

	return 0
}

func play(stdin io.Reader, stdout io.Writer) error {
	board := game.NewBoard()
	player := game.X
	if err := printBoard(stdout, board); err != nil {
		return err
	}

	reader := bufio.NewReader(stdin)
	for {
		if _, err := fmt.Fprintf(stdout, "Player %s, choose a position (1-9):\n", player); err != nil {
			return err
		}

		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read input: %w", err)
		}
		if len(line) == 0 && errors.Is(err, io.EOF) {
			_, err := fmt.Fprintln(stdout, "Game ended.")
			return err
		}

		position, err := game.ParsePosition(line)
		if err == nil {
			board, err = board.Place(position, player)
		}

		var rejection string
		switch {
		case errors.Is(err, game.ErrInvalidInput):
			rejection = "Invalid input. Enter a number from 1 to 9."
		case errors.Is(err, game.ErrPositionOutOfRange):
			rejection = "Position out of range. Enter a number from 1 to 9."
		case errors.Is(err, game.ErrCellOccupied):
			rejection = "That position is occupied. Choose an empty cell."
		case err != nil:
			return err
		}
		if rejection != "" {
			if _, err := fmt.Fprintln(stdout, rejection); err != nil {
				return err
			}
			continue
		}

		if err := printBoard(stdout, board); err != nil {
			return err
		}
		if winner, won := board.Winner(); won {
			_, err := fmt.Fprintf(stdout, "Player %s wins!\n", winner)
			return err
		}
		if board.Full() {
			_, err := fmt.Fprintln(stdout, "It's a draw!")
			return err
		}

		if player == game.X {
			player = game.O
		} else {
			player = game.X
		}
	}
}

func printBoard(stdout io.Writer, board game.Board) error {
	var cells [9]string
	for index := range cells {
		mark, err := board.At(index)
		if err != nil {
			return err
		}
		cells[index] = strconv.Itoa(index + 1)
		if mark != game.Empty {
			cells[index] = string(mark)
		}
	}

	_, err := fmt.Fprintf(stdout,
		" %s | %s | %s \n---+---+---\n %s | %s | %s \n---+---+---\n %s | %s | %s \n\n",
		cells[0], cells[1], cells[2], cells[3], cells[4], cells[5], cells[6], cells[7], cells[8])
	return err
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

func printUsage(stdout io.Writer) error {
	_, err := fmt.Fprint(stdout, "Usage: tictactoe\n\nRun a local two-player tic-tac-toe game in the terminal.\n")
	return err
}
