package main

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// One real binary checks process outcomes against test-only transcript snapshots.
func TestExecutable(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "tictactoe")
	buildContext, cancelBuild := context.WithTimeout(context.Background(), time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildContext, "go", "build", "-o", binary, ".")
	var buildStdout, buildStderr bytes.Buffer
	build.Stdout = &buildStdout
	build.Stderr = &buildStderr
	buildErr := build.Run()
	if buildContext.Err() != nil {
		t.Fatalf("build exceeded 60-second deadline: %v; stdout = %q, stderr = %q",
			buildContext.Err(), buildStdout.String(), buildStderr.String())
	}
	if buildErr != nil {
		t.Fatalf("build failed: %v; stdout = %q, stderr = %q",
			buildErr, buildStdout.String(), buildStderr.String())
	}

	// Fixed board states and text fixtures do not call the game rules or renderer.
	initial := emptyBoardText + "\n" + promptX
	tests := []struct {
		name  string
		args  []string
		input string
		want  string
	}{
		{
			name:  "X win stops gameplay",
			input: "1\n4\n2\n5\n3\n6\n7\n",
			want: initial +
				boardText("X23456789") + promptO +
				boardText("X23O56789") + promptX +
				boardText("XX3O56789") + promptO +
				boardText("XX3OO6789") + promptX +
				boardText("XXXOO6789") + "Player X wins!\n",
		},
		{
			name:  "O win stops gameplay",
			input: "1\n4\n2\n5\n9\n6\n3\n7\n",
			want: initial +
				boardText("X23456789") + promptO +
				boardText("X23O56789") + promptX +
				boardText("XX3O56789") + promptO +
				boardText("XX3OO6789") + promptX +
				boardText("XX3OO678X") + promptO +
				boardText("XX3OOO78X") + "Player O wins!\n",
		},
		{
			name:  "draw stops gameplay",
			input: "1\n2\n3\n5\n4\n6\n8\n7\n9\n1\n5\n",
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
			name:  "ninth move win precedes draw",
			input: "1\n2\n5\n3\n6\n4\n7\n8\n9\n1\n2\n",
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
		{
			name:  "rejections preserve turns and trim whitespace",
			input: "3x\n10\n \t1 \r\n1\n \t5 \n",
			want: initial + invalidInput + promptX + outOfRange + promptX +
				boardText("X23456789") + promptO + occupied + promptO +
				boardText("X234O6789") + promptX + gameEnded,
		},
		{
			name: "immediate EOF",
			want: initial + gameEnded,
		},
		{
			name:  "EOF after accepted moves",
			input: "1\n5\n",
			want: initial +
				boardText("X23456789") + promptO +
				boardText("X234O6789") + promptX + gameEnded,
		},
		{
			name:  "EOF after rejection",
			input: "3x\n",
			want:  initial + invalidInput + promptX + gameEnded,
		},
		{
			name:  "unterminated accepted move",
			input: "5",
			want:  initial + boardText("1234X6789") + promptO + gameEnded,
		},
		{
			name:  "unterminated rejected move",
			input: "3x",
			want:  initial + invalidInput + promptX + gameEnded,
		},
		{name: "short help", args: []string{"-h"}, input: "1\n", want: usageText},
		{name: "long help", args: []string{"--help"}, input: "1\n", want: usageText},
		{name: "word help", args: []string{"help"}, input: "1\n", want: usageText},
		{
			name:  "short help among arguments",
			args:  []string{"other", "-h", "more"},
			input: "1\n",
			want:  usageText,
		},
		{
			name:  "long help among arguments",
			args:  []string{"other", "--help", "more"},
			input: "1\n",
			want:  usageText,
		},
		{
			name:  "word help among arguments",
			args:  []string{"other", "help", "more"},
			input: "1\n",
			want:  usageText,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, test.args...)
			command.Stdin = strings.NewReader(test.input)
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			err := command.Run()
			if ctx.Err() != nil {
				t.Fatalf("execution exceeded 5-second timeout: %v; stdout = %q, stderr = %q",
					ctx.Err(), stdout.String(), stderr.String())
			}
			status := -1
			if command.ProcessState != nil {
				status = command.ProcessState.ExitCode()
			}
			if err != nil || status != 0 || stderr.Len() != 0 {
				t.Fatalf("status = %d, error = %v, stderr = %q; want status 0 and empty stderr; stdout = %q",
					status, err, stderr.String(), stdout.String())
			}
			if got := stdout.String(); got != test.want {
				t.Fatalf("stdout = %q\nwant   = %q", got, test.want)
			}
		})
	}
}
