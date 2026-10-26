// Package main assembles the chr command-line application.
package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"forge.adrianpk.com/chr/internal/app"
)

//go:embed assets
var assetsFS embed.FS

func main() {
	err := app.Run(context.Background(), os.Args[1:], app.IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}, assetsFS)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(app.ExitCode(err))
	}
}
