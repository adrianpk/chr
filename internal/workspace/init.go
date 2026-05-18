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

	projectDir := opts.ProjectDir
	if projectDir == "" {
		projectDir = "."
	}

	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return InitResult{}, fmt.Errorf("resolve project directory: %w", err)
	}

	info, err := os.Stat(absProjectDir)
	if err != nil {
		return InitResult{}, fmt.Errorf("inspect project directory: %w", err)
	}
	if !info.IsDir() {
		return InitResult{}, fmt.Errorf("project directory is not a directory: %s", absProjectDir)
	}

	root := filepath.Join(absProjectDir, DirectoryName)
	result := InitResult{Root: root}

	for _, dir := range directories {
		if err := ctx.Err(); err != nil {
			return InitResult{}, err
		}
		rel := filepath.Join(DirectoryName, filepath.FromSlash(dir))
		target := filepath.Join(absProjectDir, rel)
		created, err := ensureDir(target)
		if err != nil {
			return InitResult{}, err
		}
		if created {
			result.Created = append(result.Created, filepath.ToSlash(rel))
		} else {
			result.Skipped = append(result.Skipped, filepath.ToSlash(rel))
		}
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return InitResult{}, err
		}
		rel := filepath.Join(DirectoryName, filepath.FromSlash(file.path))
		target := filepath.Join(absProjectDir, rel)
		created, err := createFile(target, file.content)
		if err != nil {
			return InitResult{}, err
		}
		if created {
			result.Created = append(result.Created, filepath.ToSlash(rel))
		} else {
			result.Skipped = append(result.Skipped, filepath.ToSlash(rel))
		}
	}

	return result, nil
}

func ensureDir(path string) (bool, error) {
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
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, fmt.Errorf("create directory: %w", err)
	}
	return true, nil
}

func createFile(path string, content string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create parent directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("create file: %w", err)
	}
	return true, nil
}
