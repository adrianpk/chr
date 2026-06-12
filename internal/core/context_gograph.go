package core

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultGoGraphMaxSymbols = 5
	defaultGoGraphTimeout    = 30 * time.Second
)

// GoGraphContextOptions describes a package boundary value.
type GoGraphContextOptions struct {
	ProjectDir  string
	Terms       []string
	Precise     bool
	MaxSymbols  int
	Timeout     time.Duration
	CommandPath string
	Runner      GoGraphCommandRunner
}

// GoGraphContext describes a package boundary value.
type GoGraphContext struct {
	ProjectDir string
	Available  bool
	Command    string
	Precise    bool
	Build      GoGraphCommandResult
	Queries    []GoGraphQueryResult
	Warnings   []string
	Sources    []ContextSourceInventoryEntry
}

// GoGraphCommandResult describes a package boundary value.
type GoGraphCommandResult struct {
	Command  []string
	Stdout   string
	Stderr   string
	ExitCode int
	Skipped  bool
}

// GoGraphQueryResult describes a package boundary value.
type GoGraphQueryResult struct {
	Name   string
	Symbol string
	Result GoGraphCommandResult
}

// GoGraphCommandRunner defines a package boundary interface.
type GoGraphCommandRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, dir, command string, args ...string) GoGraphCommandResult
}

// OSGoGraphCommandRunner describes a package boundary value.
type OSGoGraphCommandRunner struct{}

// LookPath executes the corresponding package operation.
func (OSGoGraphCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Run executes the configured operation.
func (OSGoGraphCommandRunner) Run(
	ctx context.Context,
	dir string,
	command string,
	args ...string,
) GoGraphCommandResult {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()

	result := GoGraphCommandResult{
		Command: append([]string{command}, args...),
		Stdout:  string(output),
	}
	if err == nil {
		return result
	}

	result.Stderr = err.Error()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
	}

	return result
}

// LoadGoGraphContext loads the corresponding package data.
func LoadGoGraphContext(opts GoGraphContextOptions) (GoGraphContext, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return GoGraphContext{}, err
	}

	runner := opts.Runner
	if runner == nil {
		runner = OSGoGraphCommandRunner{}
	}

	command := strings.TrimSpace(opts.CommandPath)
	if command == "" {
		command = "gograph"
	}

	resolvedCommand, err := runner.LookPath(command)
	if err != nil {
		return unavailableGoGraphContext(projectDir, command, opts.Precise), nil
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultGoGraphTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	build := runGoGraphBuild(ctx, runner, projectDir, resolvedCommand, opts.Precise)
	symbols := goGraphSymbols(opts.Terms, opts.MaxSymbols)
	queries := runGoGraphQueries(ctx, runner, projectDir, resolvedCommand, symbols)

	return GoGraphContext{
		ProjectDir: projectDir,
		Available:  true,
		Command:    resolvedCommand,
		Precise:    opts.Precise,
		Build:      build,
		Queries:    queries,
		Sources:    goGraphSources(true, "gograph CLI output."),
	}, nil
}

// RenderGoGraphContext renders the corresponding package data.
func RenderGoGraphContext(ctx GoGraphContext) string {
	var output strings.Builder

	output.WriteString("## Go Graph Context\n\n")
	writeMarkdownField(&output, "Available", markdownBool(ctx.Available))
	writeMarkdownField(&output, "Command", ctx.Command)
	writeMarkdownField(&output, "Precise Mode", markdownBool(ctx.Precise))

	if len(ctx.Warnings) > 0 {
		writeMarkdownList(&output, "Warnings", ctx.Warnings)
	}

	output.WriteString("### Build\n\n")
	writeGoGraphCommandResult(&output, ctx.Build)

	output.WriteString("### Queries\n\n")

	if len(ctx.Queries) == 0 {
		output.WriteString("No gograph queries were run.\n")

		return output.String()
	}

	for _, query := range ctx.Queries {
		output.WriteString("#### ")
		output.WriteString(query.Name)

		if query.Symbol != "" {
			output.WriteString(" ")
			output.WriteString(query.Symbol)
		}

		output.WriteString("\n\n")
		writeGoGraphCommandResult(&output, query.Result)
	}

	return output.String()
}

func unavailableGoGraphContext(projectDir, command string, precise bool) GoGraphContext {
	return GoGraphContext{
		ProjectDir: projectDir,
		Available:  false,
		Command:    command,
		Precise:    precise,
		Warnings:   []string{"gograph CLI was not found; semantic Go graph context was skipped."},
		Build: GoGraphCommandResult{
			Command: []string{command, "build", "."},
			Skipped: true,
		},
		Sources: goGraphSources(false, "gograph CLI was not found"),
	}
}

func runGoGraphBuild(
	ctx context.Context,
	runner GoGraphCommandRunner,
	projectDir string,
	command string,
	precise bool,
) GoGraphCommandResult {
	args := []string{"build", "."}
	if precise {
		args = append(args, "--precise")
	}

	return runner.Run(ctx, projectDir, command, args...)
}

func runGoGraphQueries(
	ctx context.Context,
	runner GoGraphCommandRunner,
	projectDir string,
	command string,
	symbols []string,
) []GoGraphQueryResult {
	queries := []GoGraphQueryResult{
		{
			Name:   "changes",
			Result: runner.Run(ctx, projectDir, command, "changes"),
		},
		{
			Name:   "hotspot",
			Result: runner.Run(ctx, projectDir, command, "hotspot", "--top", "20"),
		},
	}

	for _, symbol := range symbols {
		queries = append(queries, symbolGoGraphQueries(ctx, runner, projectDir, command, symbol)...)
	}

	return queries
}

func symbolGoGraphQueries(
	ctx context.Context,
	runner GoGraphCommandRunner,
	projectDir string,
	command string,
	symbol string,
) []GoGraphQueryResult {
	return []GoGraphQueryResult{
		{Name: "context", Symbol: symbol, Result: runner.Run(ctx, projectDir, command, "context", symbol)},
		{Name: "source", Symbol: symbol, Result: runner.Run(ctx, projectDir, command, "source", symbol)},
		{Name: "callers", Symbol: symbol, Result: runner.Run(ctx, projectDir, command, "callers", symbol)},
		{Name: "callees", Symbol: symbol, Result: runner.Run(ctx, projectDir, command, "callees", symbol)},
		{Name: "tests", Symbol: symbol, Result: runner.Run(ctx, projectDir, command, "tests", symbol)},
		{Name: "deps", Symbol: symbol, Result: runner.Run(ctx, projectDir, command, "deps", symbol)},
	}
}

func goGraphSymbols(terms []string, maxSymbols int) []string {
	limit := maxSymbols
	if limit <= 0 {
		limit = defaultGoGraphMaxSymbols
	}

	seen := map[string]struct{}{}
	symbols := make([]string, 0, limit)

	for _, term := range terms {
		symbol := strings.TrimSpace(term)
		if len(symbol) < 3 {
			continue
		}

		_, ok := seen[symbol]
		if ok {
			continue
		}

		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)

		if len(symbols) >= limit {
			break
		}
	}

	sort.Strings(symbols)

	return symbols
}

func goGraphSources(included bool, reason string) []ContextSourceInventoryEntry {
	source := ContextSourceInventoryEntry{
		ID:          "gograph",
		Kind:        ContextSourceKindGoGraph,
		Class:       ContextSourceClassFresh,
		Path:        ".",
		Description: "Selected gograph CLI semantic code context.",
		Required:    false,
		Included:    included,
	}
	if !included {
		source.Reason = reason
	}

	return []ContextSourceInventoryEntry{source}
}

func writeGoGraphCommandResult(output *strings.Builder, result GoGraphCommandResult) {
	if result.Skipped {
		output.WriteString("Skipped.\n\n")

		return
	}

	if len(result.Command) > 0 {
		output.WriteString("Command: `")
		output.WriteString(strings.Join(commandBaseNames(result.Command), " "))
		output.WriteString("`\n\n")
	}

	if strings.TrimSpace(result.Stdout) != "" {
		output.WriteString("```text\n")
		output.WriteString(strings.TrimSpace(result.Stdout))
		output.WriteString("\n```\n\n")
	}

	if strings.TrimSpace(result.Stderr) != "" {
		output.WriteString("Error: ")
		output.WriteString(result.Stderr)
		output.WriteString("\n\n")
	}
}

func commandBaseNames(command []string) []string {
	names := make([]string, 0, len(command))

	for index, value := range command {
		if index == 0 {
			names = append(names, filepath.Base(value))

			continue
		}

		names = append(names, value)
	}

	return names
}
