package game

import (
	"errors"
	"testing"
)

// Parsing must preserve Atoi forms and distinguish malformed input from range errors.
func TestParsePosition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr error
	}{
		{name: "first cell", input: "1", want: 0},
		{name: "second cell", input: "2", want: 1},
		{name: "third cell", input: "3", want: 2},
		{name: "fourth cell", input: "4", want: 3},
		{name: "center cell", input: "5", want: 4},
		{name: "sixth cell", input: "6", want: 5},
		{name: "seventh cell", input: "7", want: 6},
		{name: "eighth cell", input: "8", want: 7},
		{name: "last cell", input: "9", want: 8},
		{name: "trims whitespace", input: " \t5\r\n", want: 4},
		{name: "explicit plus", input: "+3", want: 2},
		{name: "leading zero", input: "03", want: 2},
		{name: "rejects blank", input: "", wantErr: ErrInvalidInput},
		{name: "rejects whitespace", input: " \t\n", wantErr: ErrInvalidInput},
		{name: "rejects text", input: "center", wantErr: ErrInvalidInput},
		{name: "rejects suffix", input: "3x", wantErr: ErrInvalidInput},
		{name: "rejects two numbers", input: "3 4", wantErr: ErrInvalidInput},
		{name: "rejects decimal", input: "3.0", wantErr: ErrInvalidInput},
		{name: "rejects hexadecimal", input: "0x3", wantErr: ErrInvalidInput},
		{name: "rejects overflow", input: "9999999999999999999999999999999999999999", wantErr: ErrInvalidInput},
		{name: "rejects zero", input: "0", wantErr: ErrPositionOutOfRange},
		{name: "rejects negative", input: "-1", wantErr: ErrPositionOutOfRange},
		{name: "rejects ten", input: "10", wantErr: ErrPositionOutOfRange},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParsePosition(test.input)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ParsePosition error = %v, want %v", err, test.wantErr)
			}

			if got != test.want {
				t.Fatalf("ParsePosition = %d, want %d", got, test.want)
			}
		})
	}
}

// Legal placement returns one changed cell without modifying the source board.
func TestBoardPlace(t *testing.T) {
	tests := []struct {
		name  string
		board Board
		index int
		mark  Mark
	}{
		{name: "first cell X", board: NewBoard(), index: 0, mark: X},
		{name: "center cell O", board: Board{cells: [9]Mark{X}}, index: 4, mark: O},
		{name: "last cell X", board: Board{cells: [9]Mark{4: O}}, index: 8, mark: X},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := test.board
			placed, err := test.board.Place(test.index, test.mark)
			if err != nil {
				t.Fatalf("Place returned error: %v", err)
			}
			if test.board != original {
				t.Fatal("Place changed the source board")
			}

			for index, want := range original.cells {
				if index == test.index {
					want = test.mark
				}
				got, err := placed.At(index)
				if err != nil {
					t.Fatalf("At(%d) returned error: %v", index, err)
				}
				if got != want {
					t.Fatalf("At(%d) = %q, want %q", index, got, want)
				}
			}
		})
	}
}

// Rejected placements retain both board values and expose a stable category.
func TestBoardPlaceRejections(t *testing.T) {
	tests := []struct {
		name    string
		index   int
		mark    Mark
		wantErr error
	}{
		{name: "occupied by X", index: 0, mark: O, wantErr: ErrCellOccupied},
		{name: "occupied by O", index: 4, mark: X, wantErr: ErrCellOccupied},
		{name: "same mark occupied", index: 0, mark: X, wantErr: ErrCellOccupied},
		{name: "negative index", index: -1, mark: X, wantErr: ErrIndexOutOfRange},
		{name: "past last index", index: 9, mark: O, wantErr: ErrIndexOutOfRange},
		{name: "empty mark", index: 1, mark: Empty, wantErr: ErrInvalidMark},
		{name: "lowercase mark", index: 1, mark: "x", wantErr: ErrInvalidMark},
		{name: "unknown mark", index: 1, mark: "Z", wantErr: ErrInvalidMark},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			board := Board{cells: [9]Mark{X, Empty, Empty, Empty, O}}
			original := board
			got, err := board.Place(test.index, test.mark)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Place error = %v, want %v", err, test.wantErr)
			}
			if board != original || got != original {
				t.Fatalf("rejected placement changed board: source %v, returned %v, want %v", board, got, original)
			}
		})
	}
}

// Indexed reads reject values outside the board without changing its contents.
func TestBoardAtBounds(t *testing.T) {
	tests := []struct {
		name  string
		index int
	}{
		{name: "negative index", index: -1},
		{name: "past last index", index: 9},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			board := Board{cells: [9]Mark{X, O}}
			original := board
			got, err := board.At(test.index)
			if !errors.Is(err, ErrIndexOutOfRange) || got != Empty {
				t.Fatalf("At(%d) = %q, %v; want empty mark and index error", test.index, got, err)
			}
			if board != original {
				t.Fatal("At changed the board")
			}
		})
	}
}

// Each of the eight lines must win for either mark, including lines after empty rows.
func TestBoardWinner(t *testing.T) {
	lines := []struct {
		name  string
		cells [3]int
	}{
		{name: "top row", cells: [3]int{0, 1, 2}},
		{name: "middle row", cells: [3]int{3, 4, 5}},
		{name: "bottom row", cells: [3]int{6, 7, 8}},
		{name: "left column", cells: [3]int{0, 3, 6}},
		{name: "middle column", cells: [3]int{1, 4, 7}},
		{name: "right column", cells: [3]int{2, 5, 8}},
		{name: "main diagonal", cells: [3]int{0, 4, 8}},
		{name: "other diagonal", cells: [3]int{2, 4, 6}},
	}

	for _, mark := range []Mark{X, O} {
		for _, line := range lines {
			t.Run(string(mark)+"/"+line.name, func(t *testing.T) {
				board := NewBoard()
				for _, index := range line.cells {
					board.cells[index] = mark
				}
				got, won := board.Winner()
				if !won || got != mark {
					t.Fatalf("Winner = %q, %t; want %q, true", got, won, mark)
				}
			})
		}
	}
}

// Fullness is independent of winning so a full winning board can take precedence over a draw.
func TestBoardOutcomes(t *testing.T) {
	tests := []struct {
		name       string
		cells      [9]Mark
		wantWinner Mark
		wantFull   bool
	}{
		{name: "empty"},
		{name: "one mark", cells: [9]Mark{X}},
		{name: "partial row", cells: [9]Mark{X, X}},
		{name: "partial column", cells: [9]Mark{0: O, 3: O}},
		{name: "partial diagonal", cells: [9]Mark{0: X, 4: X}},
		{name: "mixed row", cells: [9]Mark{X, O, X}},
		{name: "first cell empty", cells: [9]Mark{Empty, O, X, X, O, O, O, X, X}},
		{name: "center cell empty", cells: [9]Mark{X, O, X, X, Empty, O, O, X, X}},
		{name: "one empty cell", cells: [9]Mark{X, O, X, X, O, O, O, X}},
		{name: "full draw", cells: [9]Mark{X, O, X, X, O, O, O, X, X}, wantFull: true},
		{name: "full win", cells: [9]Mark{X, O, O, O, X, X, X, O, X}, wantWinner: X, wantFull: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			board := Board{cells: test.cells}
			winner, won := board.Winner()
			if winner != test.wantWinner || won != (test.wantWinner != Empty) {
				t.Fatalf("Winner = %q, %t; want %q, %t", winner, won, test.wantWinner, test.wantWinner != Empty)
			}
			if got := board.Full(); got != test.wantFull {
				t.Fatalf("Full = %t, want %t", got, test.wantFull)
			}
		})
	}
}
