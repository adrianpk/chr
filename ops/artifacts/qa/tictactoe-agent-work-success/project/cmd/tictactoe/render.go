package main

import (
	"fmt"
	"strconv"
	"strings"

	"example.com/chr-qa-tictactoe/internal/game"
)

func renderBoard(board game.Board) string {
	var builder strings.Builder

	for row := 0; row < 3; row++ {
		if row > 0 {
			builder.WriteString("\n---+---+---\n")
		}

		for col := 0; col < 3; col++ {
			if col > 0 {
				builder.WriteString("|")
			}

			index := row*3 + col
			builder.WriteString(" ")
			builder.WriteString(renderCell(board, index))
			builder.WriteString(" ")
		}
	}

	return builder.String()
}

func renderCell(board game.Board, index int) string {
	mark, err := board.At(index)
	if err != nil {
		panic(fmt.Sprintf("render cell %d: %v", index, err))
	}

	if mark == game.Empty {
		return strconv.Itoa(index + 1)
	}

	return string(mark)
}

func renderPrompt(player game.Mark) string {
	return fmt.Sprintf("Player %s, choose a position (1-9): ", player)
}

func renderRecoverableError(err error) string {
	if err == nil {
		return ""
	}

	return fmt.Sprintf("Invalid move: %v. Try again.", err)
}

func renderResult(state game.Game) string {
	if winner, ok := state.Winner(); ok {
		return fmt.Sprintf("Player %s wins!", winner)
	}

	if state.Draw() {
		return "Game ended in a draw."
	}

	return ""
}
