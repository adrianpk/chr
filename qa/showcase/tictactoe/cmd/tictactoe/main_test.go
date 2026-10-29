package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

const (
	emptyBoardText = " 1 | 2 | 3 \n---+---+---\n 4 | 5 | 6 \n---+---+---\n 7 | 8 | 9 \n"
	promptX        = "Player X, choose a position (1-9):\n"
	promptO        = "Player O, choose a position (1-9):\n"
	invalidInput   = "Invalid input. Enter a number from 1 to 9.\n"
	outOfRange     = "Position out of range. Enter a number from 1 to 9.\n"
	occupied       = "That position is occupied. Choose an empty cell.\n"
	gameEnded      = "Game ended.\n"
	usageText      = "Usage: tictactoe\n\nRun a local two-player tic-tac-toe game in the terminal.\n"
)

// Exact transcripts protect spacing, turn changes, retries, and terminal output.
func TestRunTranscripts(t *testing.T) {
	initial := emptyBoardText + "\n" + promptX
	tests := []struct {
		name          string
		input         string
		stopBeforeEOF bool
		want          string
	}{
		{
			name: "initial EOF",
			want: initial + gameEnded,
		},
		{
			name:  "alternate then EOF",
			input: "1\n5\n",
			want: initial +
				boardText("X23456789") + promptO +
				boardText("X234O6789") + promptX + gameEnded,
		},
		{
			name:  "unterminated move",
			input: "5",
			want:  initial + boardText("1234X6789") + promptO + gameEnded,
		},
		{
			name:  "unterminated rejection",
			input: "3x",
			want:  initial + invalidInput + promptX + gameEnded,
		},
		{
			name:  "retry without changing turn",
			input: "\n3x\n3 4\n3.0\n0\n-1\n10\n1\n1\n5\n5\n",
			want: initial + strings.Repeat(invalidInput+promptX, 4) +
				strings.Repeat(outOfRange+promptX, 3) +
				boardText("X23456789") + promptO + occupied + promptO +
				boardText("X234O6789") + promptX + occupied + promptX + gameEnded,
		},
		{
			name:  "trim and preserve Atoi forms",
			input: " \t+3 \r\n03\n",
			want:  initial + boardText("12X456789") + promptO + occupied + promptO + gameEnded,
		},
		{
			name:  "whole long line",
			input: strings.Repeat(" ", 70*1024) + "5\n",
			want:  initial + boardText("1234X6789") + promptO + gameEnded,
		},
		{
			name:          "X wins and ignores later moves",
			input:         "1\n4\n2\n5\n3\n6\n",
			stopBeforeEOF: true,
			want: initial +
				boardText("X23456789") + promptO +
				boardText("X23O56789") + promptX +
				boardText("XX3O56789") + promptO +
				boardText("XX3OO6789") + promptX +
				boardText("XXXOO6789") + "Player X wins!\n",
		},
		{
			name:          "O wins",
			input:         "1\n4\n2\n5\n9\n6\n",
			stopBeforeEOF: true,
			want: initial +
				boardText("X23456789") + promptO +
				boardText("X23O56789") + promptX +
				boardText("XX3O56789") + promptO +
				boardText("XX3OO6789") + promptX +
				boardText("XX3OO678X") + promptO +
				boardText("XX3OOO78X") + "Player O wins!\n",
		},
		{
			name:          "draw",
			input:         "1\n2\n3\n5\n4\n6\n8\n7\n9\n",
			stopBeforeEOF: true,
			want: initial +
				boardText("X23456789") + promptO +
				boardText("XO3456789") + promptX +
				boardText("XOX456789") + promptO +
				boardText("XOX4O6789") + promptX +
				boardText("XOXXO6789") + promptO +
				boardText("XOXXOO789") + promptX +
				boardText("XOXXOO7X9") + promptO +
				boardText("XOXXOOOX9") + promptX +
				boardText("XOXXOOOXX") + "It's a draw!\n",
		},
		{
			name:  "ninth move wins before draw",
			input: "1\n2\n5\n3\n6\n4\n7\n8\n9",
			want: initial +
				boardText("X23456789") + promptO +
				boardText("XO3456789") + promptX +
				boardText("XO34X6789") + promptO +
				boardText("XOO4X6789") + promptX +
				boardText("XOO4XX789") + promptO +
				boardText("XOOOXX789") + promptX +
				boardText("XOOOXXX89") + promptO +
				boardText("XOOOXXXO9") + promptX +
				boardText("XOOOXXXOX") + "Player X wins!\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdin io.Reader = strings.NewReader(test.input)
			if test.stopBeforeEOF {
				stdin = io.MultiReader(stdin, readFunc(func([]byte) (int, error) {
					t.Fatal("terminal game attempted to read more input")
					return 0, io.EOF
				}))
			}
			var stdout, stderr strings.Builder
			status := run(nil, stdin, &stdout, &stderr)
			if status != 0 || stderr.Len() != 0 {
				t.Fatalf("run status = %d, stderr = %q; want 0 and empty stderr", status, stderr.String())
			}
			if got := stdout.String(); got != test.want {
				t.Fatalf("stdout = %q\nwant   = %q", got, test.want)
			}
		})
	}
}

// Every existing help trigger takes precedence over other arguments and stdin.
func TestRunHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "short", args: []string{"-h"}},
		{name: "long", args: []string{"--help"}},
		{name: "word", args: []string{"help"}},
		{name: "short among arguments", args: []string{"other", "-h", "more"}},
		{name: "long among arguments", args: []string{"other", "--help", "more"}},
		{name: "word among arguments", args: []string{"other", "help", "more"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdin := readFunc(func([]byte) (int, error) {
				t.Fatal("help attempted to read input")
				return 0, io.EOF
			})
			var stdout, stderr strings.Builder
			status := run(test.args, stdin, &stdout, &stderr)
			if status != 0 || stdout.String() != usageText || stderr.Len() != 0 {
				t.Fatalf("run = %d, stdout = %q, stderr = %q; want 0, %q, empty stderr",
					status, stdout.String(), stderr.String(), usageText)
			}
		})
	}
}

// Returned read errors must not become EOF, invalid moves, or partial placements.
func TestRunInputFailures(t *testing.T) {
	initial := emptyBoardText + "\n" + promptX
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "before first move", want: initial},
		{name: "partial line", prefix: "1", want: initial},
		{name: "after move", prefix: "1\n", want: initial + boardText("X23456789") + promptO},
		{name: "after rejection", prefix: "3x\n", want: initial + invalidInput + promptX},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputErr := errors.New("injected read failure")
			stdin := io.MultiReader(strings.NewReader(test.prefix), readFunc(func([]byte) (int, error) {
				return 0, inputErr
			}))
			var stdout, stderr strings.Builder
			status := run(nil, stdin, &stdout, &stderr)
			if status == 0 || !strings.Contains(stderr.String(), inputErr.Error()) {
				t.Fatalf("run status = %d, stderr = %q; want failure and diagnostic", status, stderr.String())
			}
			if got := stdout.String(); got != test.want {
				t.Fatalf("stdout = %q, want %q", got, test.want)
			}
		})
	}
}

// Byte-limited writers inject failures at each output stage without OS pipe behavior.
func TestRunOutputFailures(t *testing.T) {
	boardSize := len(emptyBoardText) + 1
	initialSize := boardSize + len(promptX)
	tests := []struct {
		name  string
		args  []string
		input string
		limit int
	}{
		{name: "initial board"},
		{name: "partial board", limit: 8},
		{name: "prompt", limit: boardSize + 4},
		{name: "EOF message", limit: initialSize + 3},
		{name: "accepted board", input: "1\n", limit: initialSize + 8},
		{name: "invalid input", input: "3x\n", limit: initialSize + 3},
		{name: "out of range", input: "10\n", limit: initialSize + 3},
		{name: "retry prompt", input: "3x\n", limit: initialSize + len(invalidInput) + 4},
		{name: "occupied cell", input: "1\n1\n", limit: initialSize + boardSize + len(promptO) + 3},
		{name: "X result", input: "1\n4\n2\n5\n3\n", limit: 6*boardSize + 5*len(promptX) + 3},
		{name: "O result", input: "1\n4\n2\n5\n9\n6\n", limit: 7*boardSize + 6*len(promptX) + 3},
		{name: "draw result", input: "1\n2\n3\n5\n4\n6\n8\n7\n9\n", limit: 10*boardSize + 9*len(promptX) + 3},
		{name: "help", args: []string{"--help"}, limit: 8},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputErr := errors.New("injected write failure")
			remaining := test.limit
			failed := false
			stdout := writeFunc(func(data []byte) (int, error) {
				if failed {
					t.Error("stdout write attempted after failure")
				}
				n := min(len(data), remaining)
				remaining -= n
				if n < len(data) {
					failed = true
					return n, outputErr
				}
				return n, nil
			})
			var stderr strings.Builder
			status := run(test.args, strings.NewReader(test.input), stdout, &stderr)
			if status == 0 || !failed || !strings.Contains(stderr.String(), outputErr.Error()) {
				t.Fatalf("run status = %d, failed write = %t, stderr = %q; want failure and diagnostic",
					status, failed, stderr.String())
			}
		})
	}
}

// A failed diagnostic must leave the process unsuccessful and must not cause retries.
func TestRunDiagnosticFailure(t *testing.T) {
	stdin := readFunc(func([]byte) (int, error) {
		return 0, errors.New("injected read failure")
	})
	attempts := 0
	stderr := writeFunc(func([]byte) (int, error) {
		attempts++
		return 0, errors.New("injected diagnostic failure")
	})
	status := run(nil, stdin, io.Discard, stderr)
	if status == 0 || attempts != 1 {
		t.Fatalf("run status = %d, diagnostic attempts = %d; want failure and one attempt", status, attempts)
	}
}

// Snapshots supply visible labels independently of the production board and renderer.
func boardText(cells string) string {
	text := emptyBoardText
	for index, cell := range cells {
		text = strings.ReplaceAll(text, string(rune('1'+index)), string(cell))
	}
	return text + "\n"
}

type readFunc func([]byte) (int, error)

func (read readFunc) Read(data []byte) (int, error) {
	return read(data)
}

type writeFunc func([]byte) (int, error)

func (write writeFunc) Write(data []byte) (int, error) {
	return write(data)
}
