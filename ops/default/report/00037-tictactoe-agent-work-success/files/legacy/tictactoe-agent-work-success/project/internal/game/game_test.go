package game

import (
	"strconv"
	"strings"
	"testing"
	"testing/quick"
)

func TestParsePosition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "first cell", input: "1", want: 0},
		{name: "last cell", input: "9", want: 8},
		{name: "trims whitespace", input: " 5\n", want: 4},
		{name: "rejects zero", input: "0", wantErr: true},
		{name: "rejects negative", input: "-1", wantErr: true},
		{name: "rejects ten", input: "10", wantErr: true},
		{name: "rejects text", input: "center", wantErr: true},
		{name: "rejects empty input", input: "", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParsePosition(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatal("ParsePosition returned nil error")
				}

				return
			}

			if err != nil {
				t.Fatalf("ParsePosition returned error: %v", err)
			}

			if got != test.want {
				t.Fatalf("ParsePosition = %d, want %d", got, test.want)
			}
		})
	}
}

func TestParsePositionValidNumbersMapToIndexes(t *testing.T) {
	property := func(raw uint8) bool {
		value := int(raw%9) + 1

		got, err := ParsePosition(strconv.Itoa(value))
		return err == nil && got == value-1
	}

	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func FuzzParsePosition(f *testing.F) {
	seeds := []string{
		"",
		"1",
		"9",
		" 5\n",
		"0",
		"10",
		"-1",
		"center",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := ParsePosition(input)
		if err != nil {
			return
		}

		if got < 0 || got > 8 {
			t.Fatalf("ParsePosition(%q) = %d, want index from 0 through 8", input, got)
		}

		value, atoiErr := strconv.Atoi(strings.TrimSpace(input))
		if atoiErr != nil || value < 1 || value > 9 {
			t.Fatalf("ParsePosition(%q) accepted invalid input with index %d", input, got)
		}

		if got != value-1 {
			t.Fatalf("ParsePosition(%q) = %d, want %d", input, got, value-1)
		}
	})
}

func TestBoardPlace(t *testing.T) {
	board := NewBoard()

	board, err := board.Place(4, X)
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}

	got, err := board.At(4)
	if err != nil {
		t.Fatalf("At returned error: %v", err)
	}

	if got != X {
		t.Fatalf("At = %q, want %q", got, X)
	}

	_, err = board.Place(4, O)
	if err == nil {
		t.Fatal("Place returned nil error for occupied cell")
	}
}

func TestBoardAtRejectsOutOfRangeIndexes(t *testing.T) {
	for _, index := range []int{-1, 9} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			got, err := NewBoard().At(index)
			if err == nil {
				t.Fatal("At returned nil error")
			}

			if got != Empty {
				t.Fatalf("At = %q, want %q", got, Empty)
			}
		})
	}
}

func TestBoardPlaceRejectsInvalidMovesWithoutMutation(t *testing.T) {
	boardWithOccupiedCell := mustPlace(t, NewBoard(), 4, X)

	tests := []struct {
		name  string
		board Board
		index int
		mark  Mark
	}{
		{name: "negative index", board: NewBoard(), index: -1, mark: X},
		{name: "high index", board: NewBoard(), index: 9, mark: X},
		{name: "empty mark", board: NewBoard(), index: 0, mark: Empty},
		{name: "unknown mark", board: NewBoard(), index: 0, mark: Mark("Z")},
		{name: "occupied cell", board: boardWithOccupiedCell, index: 4, mark: O},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.board.Place(test.index, test.mark)
			if err == nil {
				t.Fatal("Place returned nil error")
			}

			if got != test.board {
				t.Fatalf("Place mutated board: got %#v, want %#v", got, test.board)
			}
		})
	}
}

func TestNewGameStartsWithX(t *testing.T) {
	game := NewGame()

	if game.CurrentPlayer() != X {
		t.Fatalf("CurrentPlayer = %q, want %q", game.CurrentPlayer(), X)
	}

	if game.Outcome() != OutcomeInProgress {
		t.Fatalf("Outcome = %v, want %v", game.Outcome(), OutcomeInProgress)
	}

	if game.Over() {
		t.Fatal("Over = true, want false")
	}

	if game.Draw() {
		t.Fatal("Draw = true, want false")
	}

	if winner, ok := game.Winner(); ok || winner != Empty {
		t.Fatalf("Winner = %q, %t; want %q, false", winner, ok, Empty)
	}
}

func TestGameApplyMoveAlternatesAfterValidNonTerminalMoves(t *testing.T) {
	tests := []struct {
		name              string
		index             int
		mark              Mark
		wantCurrentPlayer Mark
	}{
		{name: "X opens in top left", index: 0, mark: X, wantCurrentPlayer: O},
		{name: "O answers in center", index: 4, mark: O, wantCurrentPlayer: X},
		{name: "X takes bottom right", index: 8, mark: X, wantCurrentPlayer: O},
	}

	game := NewGame()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next, err := game.ApplyMove(test.index, test.mark)
			if err != nil {
				t.Fatalf("ApplyMove returned error: %v", err)
			}

			got, err := next.Board().At(test.index)
			if err != nil {
				t.Fatalf("At returned error: %v", err)
			}

			if got != test.mark {
				t.Fatalf("At = %q, want %q", got, test.mark)
			}

			if next.CurrentPlayer() != test.wantCurrentPlayer {
				t.Fatalf("CurrentPlayer = %q, want %q", next.CurrentPlayer(), test.wantCurrentPlayer)
			}

			if next.Outcome() != OutcomeInProgress {
				t.Fatalf("Outcome = %v, want %v", next.Outcome(), OutcomeInProgress)
			}

			game = next
		})
	}
}

func TestGamePlayUsesCurrentPlayer(t *testing.T) {
	game := NewGame()

	game, err := game.Play(4)
	if err != nil {
		t.Fatalf("Play returned error: %v", err)
	}

	got, err := game.Board().At(4)
	if err != nil {
		t.Fatalf("At returned error: %v", err)
	}

	if got != X {
		t.Fatalf("At = %q, want %q", got, X)
	}

	if game.CurrentPlayer() != O {
		t.Fatalf("CurrentPlayer = %q, want %q", game.CurrentPlayer(), O)
	}
}

func TestGameApplyMoveRejectsInvalidMovesWithoutMutation(t *testing.T) {
	gameWithOccupiedCell := mustApplyMoves(t, NewGame(), []move{
		{index: 0, mark: X},
	})

	tests := []struct {
		name  string
		game  Game
		index int
		mark  Mark
	}{
		{name: "negative index", game: NewGame(), index: -1, mark: X},
		{name: "high index", game: NewGame(), index: 9, mark: X},
		{name: "empty mark", game: NewGame(), index: 0, mark: Empty},
		{name: "unknown mark", game: NewGame(), index: 0, mark: Mark("Z")},
		{name: "mark out of turn", game: NewGame(), index: 0, mark: O},
		{name: "occupied cell", game: gameWithOccupiedCell, index: 0, mark: O},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.game.ApplyMove(test.index, test.mark)
			if err == nil {
				t.Fatal("ApplyMove returned nil error")
			}

			if got != test.game {
				t.Fatalf("ApplyMove mutated game state: got %#v, want %#v", got, test.game)
			}
		})
	}
}

func TestGameApplyMoveStopsAfterTerminalOutcomes(t *testing.T) {
	tests := []struct {
		name              string
		moves             []move
		wantOutcome       Outcome
		wantCurrentPlayer Mark
		wantWinner        Mark
	}{
		{
			name: "X win",
			moves: []move{
				{index: 0, mark: X},
				{index: 3, mark: O},
				{index: 1, mark: X},
				{index: 4, mark: O},
				{index: 2, mark: X},
			},
			wantOutcome:       OutcomeXWon,
			wantCurrentPlayer: X,
			wantWinner:        X,
		},
		{
			name: "O win",
			moves: []move{
				{index: 0, mark: X},
				{index: 3, mark: O},
				{index: 1, mark: X},
				{index: 4, mark: O},
				{index: 8, mark: X},
				{index: 5, mark: O},
			},
			wantOutcome:       OutcomeOWon,
			wantCurrentPlayer: O,
			wantWinner:        O,
		},
		{
			name: "draw",
			moves: []move{
				{index: 0, mark: X},
				{index: 1, mark: O},
				{index: 2, mark: X},
				{index: 4, mark: O},
				{index: 3, mark: X},
				{index: 5, mark: O},
				{index: 7, mark: X},
				{index: 6, mark: O},
				{index: 8, mark: X},
			},
			wantOutcome:       OutcomeDraw,
			wantCurrentPlayer: X,
			wantWinner:        Empty,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := mustApplyMoves(t, NewGame(), test.moves)

			if game.Outcome() != test.wantOutcome {
				t.Fatalf("Outcome = %v, want %v", game.Outcome(), test.wantOutcome)
			}

			if game.CurrentPlayer() != test.wantCurrentPlayer {
				t.Fatalf("CurrentPlayer = %q, want %q", game.CurrentPlayer(), test.wantCurrentPlayer)
			}

			if !game.Over() {
				t.Fatal("Over = false, want true")
			}

			if game.Draw() != (test.wantOutcome == OutcomeDraw) {
				t.Fatalf("Draw = %t, want %t", game.Draw(), test.wantOutcome == OutcomeDraw)
			}

			winner, ok := game.Winner()
			if test.wantWinner == Empty {
				if ok || winner != Empty {
					t.Fatalf("Winner = %q, %t; want %q, false", winner, ok, Empty)
				}
			} else if !ok || winner != test.wantWinner {
				t.Fatalf("Winner = %q, %t; want %q, true", winner, ok, test.wantWinner)
			}

			got, err := game.Play(0)
			if err == nil {
				t.Fatal("Play returned nil error after terminal outcome")
			}

			if got != game {
				t.Fatalf("Play after terminal outcome mutated game state: got %#v, want %#v", got, game)
			}
		})
	}
}

func TestBoardWinner(t *testing.T) {
	tests := []struct {
		name    string
		board   Board
		want    Mark
		wantWin bool
	}{
		{
			name:    "top row",
			board:   Board{cells: [9]Mark{X, X, X, Empty, O, Empty, O, Empty, Empty}},
			want:    X,
			wantWin: true,
		},
		{
			name:    "middle row",
			board:   Board{cells: [9]Mark{O, Empty, X, X, X, X, Empty, O, Empty}},
			want:    X,
			wantWin: true,
		},
		{
			name:    "bottom row",
			board:   Board{cells: [9]Mark{O, Empty, Empty, X, O, Empty, X, X, X}},
			want:    X,
			wantWin: true,
		},
		{
			name:    "left column",
			board:   Board{cells: [9]Mark{O, X, Empty, O, X, Empty, O, Empty, X}},
			want:    O,
			wantWin: true,
		},
		{
			name:    "middle column",
			board:   Board{cells: [9]Mark{X, O, Empty, Empty, O, X, X, O, Empty}},
			want:    O,
			wantWin: true,
		},
		{
			name:    "right column",
			board:   Board{cells: [9]Mark{X, Empty, O, X, Empty, O, Empty, X, O}},
			want:    O,
			wantWin: true,
		},
		{
			name:    "descending diagonal",
			board:   Board{cells: [9]Mark{X, O, Empty, Empty, X, O, Empty, Empty, X}},
			want:    X,
			wantWin: true,
		},
		{
			name:    "ascending diagonal",
			board:   Board{cells: [9]Mark{O, Empty, X, Empty, X, O, X, Empty, Empty}},
			want:    X,
			wantWin: true,
		},
		{
			name:    "full board with no winner",
			board:   Board{cells: [9]Mark{X, O, X, X, O, O, O, X, X}},
			want:    Empty,
			wantWin: false,
		},
		{
			name:    "incomplete board with no winner",
			board:   Board{cells: [9]Mark{X, O, Empty, Empty, X, O, O, Empty, Empty}},
			want:    Empty,
			wantWin: false,
		},
		{
			name:    "mixed line does not win",
			board:   Board{cells: [9]Mark{X, O, X, Empty, Empty, Empty, Empty, Empty, Empty}},
			want:    Empty,
			wantWin: false,
		},
		{
			name:    "empty board has no winner",
			board:   NewBoard(),
			want:    Empty,
			wantWin: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotWin := test.board.Winner()

			if got != test.want {
				t.Fatalf("Winner mark = %q, want %q", got, test.want)
			}

			if gotWin != test.wantWin {
				t.Fatalf("Winner ok = %t, want %t", gotWin, test.wantWin)
			}
		})
	}
}

type move struct {
	index int
	mark  Mark
}

func mustApplyMoves(t *testing.T, game Game, moves []move) Game {
	t.Helper()

	for _, move := range moves {
		next, err := game.ApplyMove(move.index, move.mark)
		if err != nil {
			t.Fatalf("ApplyMove(%d, %q) returned error: %v", move.index, move.mark, err)
		}

		game = next
	}

	return game
}

func mustPlace(t *testing.T, board Board, index int, mark Mark) Board {
	t.Helper()

	next, err := board.Place(index, mark)
	if err != nil {
		t.Fatalf("Place(%d, %q) returned error: %v", index, mark, err)
	}

	return next
}

func TestBoardFull(t *testing.T) {
	tests := []struct {
		name  string
		board Board
		want  bool
	}{
		{
			name:  "full board",
			board: Board{cells: [9]Mark{X, O, X, X, O, O, O, X, X}},
			want:  true,
		},
		{
			name:  "one empty cell",
			board: Board{cells: [9]Mark{X, O, X, X, Empty, O, O, X, X}},
			want:  false,
		},
		{
			name:  "empty board",
			board: NewBoard(),
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.board.Full()

			if got != test.want {
				t.Fatalf("Full = %t, want %t", got, test.want)
			}
		})
	}
}
