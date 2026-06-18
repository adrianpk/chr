package game

import "testing"

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
		{name: "rejects ten", input: "10", wantErr: true},
		{name: "rejects text", input: "center", wantErr: true},
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
