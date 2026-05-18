package workspace

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const DirectoryName = ".chc"

//go:embed templates/config.yml
var defaultConfig string

//go:embed templates/manifest.md
var defaultManifest string

type InitOptions struct {
	ProjectDir string
	DryRun     bool
}

type InitResult struct {
	Root    string
	Created []string
	Skipped []string
}

type entry struct {
	path    string
	content string
}

var directories = []string{
	"agents",
	"docs",
	"specs/drafts",
	"specs/accepted",
	"specs/archived",
	"tickets/open",
	"tickets/ready",
	"tickets/in_progress",
	"tickets/reviewing",
	"tickets/blocked",
	"tickets/solved",
	"tickets/closed",
	"comments/tickets",
	"context",
	"dumps/raw",
	"dumps/curated",
	"quizzes/raw",
	"quizzes/curated",
	"playbooks",
	"reports/runs",
	"reports/audits",
	"indexes",
	"runtime/runs",
	"runtime/locks",
	"runtime/traces",
	"notifications",
}

var files = []entry{
	{path: "manifest.md", content: defaultManifest},
	{path: "config.yml", content: defaultConfig},
}

func Init(ctx context.Context, opts InitOptions) (InitResult, error) {
	if err := ctx.Err(); err != nil {
		return InitResult{}, err
	}

	absProjectDir, err := resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return InitResult{}, err
	}

	result := InitResult{Root: filepath.Join(absProjectDir, DirectoryName)}

	if err := initDirectories(ctx, absProjectDir, opts.DryRun, &result); err != nil {
		return InitResult{}, err
	}

	if err := initFiles(ctx, absProjectDir, opts.DryRun, &result); err != nil {
		return InitResult{}, err
	}

	return result, nil
}

func resolveProjectDir(projectDir string) (string, error) {
	if projectDir == "" {
		projectDir = "."
	}

	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}

	info, err := os.Stat(absProjectDir)
	if err != nil {
		return "", fmt.Errorf("inspect project directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("project directory is not a directory: %s", absProjectDir)
	}

	return absProjectDir, nil
}

func initDirectories(ctx context.Context, projectDir string, dryRun bool, result *InitResult) error {
	for _, dir := range directories {
		if err := ctx.Err(); err != nil {
			return err
		}

		rel := filepath.Join(DirectoryName, filepath.FromSlash(dir))
		target := filepath.Join(projectDir, rel)

		created, err := ensureDir(target, dryRun)
		if err != nil {
			return err
		}

		recordPath(result, rel, created)
	}

	return nil
}

func initFiles(ctx context.Context, projectDir string, dryRun bool, result *InitResult) error {
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}

		rel := filepath.Join(DirectoryName, filepath.FromSlash(file.path))
		target := filepath.Join(projectDir, rel)

		created, err := createFile(target, file.content, dryRun)
		if err != nil {
			return err
		}

		recordPath(result, rel, created)
	}

	return nil
}

func recordPath(result *InitResult, rel string, created bool) {
	path := filepath.ToSlash(rel)
	if created {
		result.Created = append(result.Created, path)

		return
	}

	result.Skipped = append(result.Skipped, path)
}

func ensureDir(path string, dryRun bool) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return false, fmt.Errorf("path exists and is not a directory: %s", path)
		}

		return false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect directory: %w", err)
	}

	if dryRun {
		return true, nil
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, fmt.Errorf("create directory: %w", err)
	}

	return true, nil
}

func createFile(path, content string, dryRun bool) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("path exists and is not a file: %s", path)
		}

		return false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect file: %w", err)
	}

	if dryRun {
		return true, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create parent directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("create file: %w", err)
	}

	return true, nil
}
