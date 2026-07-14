package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "short flag", args: []string{"-h"}},
		{name: "long flag", args: []string{"--help"}},
		{name: "command", args: []string{"help"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, stdout, stderr := runScript(test.args, "")

			if status != 0 {
				t.Fatalf("run() status = %d, want 0", status)
			}

			want := "Usage: tictactoe\n\n" +
				"Run a local two-player tic-tac-toe game in the terminal.\n"
			if stdout != want {
				t.Fatalf("stdout = %q, want %q", stdout, want)
			}

			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}
		})
	}
}

func TestRunXWin(t *testing.T) {
	status, stdout, stderr := runScript(nil, "1\n4\n2\n5\n3\n")

	if status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	finalBoard := " X | X | X \n" +
		"---+---+---\n" +
		" O | O | 6 \n" +
		"---+---+---\n" +
		" 7 | 8 | 9 "
	assertContains(t, stdout, finalBoard)
	assertOrdered(t, stdout, finalBoard, "Player X wins!")
}

func TestRunDraw(t *testing.T) {
	status, stdout, stderr := runScript(nil, "1\n2\n3\n5\n4\n6\n8\n7\n9\n")

	if status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	finalBoard := " X | O | X \n" +
		"---+---+---\n" +
		" X | O | O \n" +
		"---+---+---\n" +
		" O | X | X "
	assertContains(t, stdout, finalBoard)
	assertOrdered(t, stdout, finalBoard, "Game ended in a draw.")
}

func TestRunInvalidInputRecovery(t *testing.T) {
	status, stdout, stderr := runScript(nil, "center\n1\n4\n2\n5\n3\n")

	if status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	assertContains(t, stdout, "Invalid move: position must be a number from 1 to 9. Try again.")
	assertContains(t, stdout, "Player X wins!")
}

func TestRunOutOfRangeInputRecovery(t *testing.T) {
	status, stdout, stderr := runScript(nil, "10\n1\n4\n2\n5\n3\n")

	if status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	assertContains(t, stdout, "Invalid move: position must be between 1 and 9. Try again.")
	assertContains(t, stdout, "Player X wins!")
}

func TestRunOccupiedCellRecovery(t *testing.T) {
	status, stdout, stderr := runScript(nil, "1\n1\n4\n2\n5\n3\n")

	if status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	assertContains(t, stdout, "Invalid move: cell is already occupied. Try again.")
	assertContains(t, stdout, "Player X wins!")
}

func runScript(args []string, input string) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run(args, strings.NewReader(input), &stdout, &stderr)

	return status, stdout.String(), stderr.String()
}

func assertContains(t *testing.T, text string, want string) {
	t.Helper()

	if !strings.Contains(text, want) {
		t.Fatalf("output does not contain %q:\n%s", want, text)
	}
}

func assertOrdered(t *testing.T, text string, first string, second string) {
	t.Helper()

	firstIndex := strings.Index(text, first)
	if firstIndex == -1 {
		t.Fatalf("output does not contain %q:\n%s", first, text)
	}

	secondIndex := strings.Index(text, second)
	if secondIndex == -1 {
		t.Fatalf("output does not contain %q:\n%s", second, text)
	}

	if firstIndex > secondIndex {
		t.Fatalf("%q appears after %q:\n%s", first, second, text)
	}
}
