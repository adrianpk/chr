package main

import (
	"errors"
	"testing"

	"example.com/chr-qa-tictactoe/internal/game"
)

func TestRenderBoard(t *testing.T) {
	tests := []struct {
		name  string
		board game.Board
		want  string
	}{
		{
			name:  "empty board",
			board: game.NewBoard(),
			want: " 1 | 2 | 3 \n" +
				"---+---+---\n" +
				" 4 | 5 | 6 \n" +
				"---+---+---\n" +
				" 7 | 8 | 9 ",
		},
		{
			name: "partial board",
			board: mustPlaceMarks(t, []renderMove{
				{index: 0, mark: game.X},
				{index: 4, mark: game.O},
				{index: 8, mark: game.X},
			}),
			want: " X | 2 | 3 \n" +
				"---+---+---\n" +
				" 4 | O | 6 \n" +
				"---+---+---\n" +
				" 7 | 8 | X ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderBoard(test.board)

			if got != test.want {
				t.Fatalf("renderBoard() =\n%q\nwant\n%q", got, test.want)
			}
		})
	}
}

func TestRenderPromptIncludesCurrentPlayer(t *testing.T) {
	got := renderPrompt(game.O)
	want := "Player O, choose a position (1-9): "

	if got != want {
		t.Fatalf("renderPrompt() = %q, want %q", got, want)
	}
}

func TestRenderRecoverableError(t *testing.T) {
	got := renderRecoverableError(errors.New("cell is already occupied"))
	want := "Invalid move: cell is already occupied. Try again."

	if got != want {
		t.Fatalf("renderRecoverableError() = %q, want %q", got, want)
	}
}

func TestRenderResult(t *testing.T) {
	tests := []struct {
		name string
		game game.Game
		want string
	}{
		{
			name: "X winner",
			game: mustApplyGameMoves(t, []renderMove{
				{index: 0, mark: game.X},
				{index: 3, mark: game.O},
				{index: 1, mark: game.X},
				{index: 4, mark: game.O},
				{index: 2, mark: game.X},
			}),
			want: "Player X wins!",
		},
		{
			name: "O winner",
			game: mustApplyGameMoves(t, []renderMove{
				{index: 0, mark: game.X},
				{index: 3, mark: game.O},
				{index: 1, mark: game.X},
				{index: 4, mark: game.O},
				{index: 8, mark: game.X},
				{index: 5, mark: game.O},
			}),
			want: "Player O wins!",
		},
		{
			name: "draw",
			game: mustApplyGameMoves(t, []renderMove{
				{index: 0, mark: game.X},
				{index: 1, mark: game.O},
				{index: 2, mark: game.X},
				{index: 4, mark: game.O},
				{index: 3, mark: game.X},
				{index: 5, mark: game.O},
				{index: 7, mark: game.X},
				{index: 6, mark: game.O},
				{index: 8, mark: game.X},
			}),
			want: "Game ended in a draw.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderResult(test.game)

			if got != test.want {
				t.Fatalf("renderResult() = %q, want %q", got, test.want)
			}
		})
	}
}

type renderMove struct {
	index int
	mark  game.Mark
}

func mustPlaceMarks(t *testing.T, moves []renderMove) game.Board {
	t.Helper()

	board := game.NewBoard()
	for _, move := range moves {
		next, err := board.Place(move.index, move.mark)
		if err != nil {
			t.Fatalf("Place(%d, %q) returned error: %v", move.index, move.mark, err)
		}

		board = next
	}

	return board
}

func mustApplyGameMoves(t *testing.T, moves []renderMove) game.Game {
	t.Helper()

	state := game.NewGame()
	for _, move := range moves {
		next, err := state.ApplyMove(move.index, move.mark)
		if err != nil {
			t.Fatalf("ApplyMove(%d, %q) returned error: %v", move.index, move.mark, err)
		}

		state = next
	}

	return state
}
