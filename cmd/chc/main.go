package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"forge.adrianpk.com/chc/internal/workspace"
)

const usage = `chc coordinates project-local agent development workflows.

Usage:
  chc init [--dir PATH]
  chc help

Commands:
  init    Create a .chc workspace in a project directory
  help    Show this help text
`

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	case "init":
		return runInit(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runInit(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var dir string
	fs.StringVar(&dir, "dir", ".", "project directory")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("init does not accept positional arguments")
	}

	result, err := workspace.Init(ctx, workspace.InitOptions{ProjectDir: dir})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "workspace: %s\n", result.Root)
	for _, created := range result.Created {
		fmt.Fprintf(stdout, "create %s\n", created)
	}
	for _, skipped := range result.Skipped {
		fmt.Fprintf(stdout, "skip existing %s\n", skipped)
	}
	fmt.Fprintln(stdout, "done")
	return nil
}
