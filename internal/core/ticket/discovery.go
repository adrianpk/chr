package ticket

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Discoverer defines a package boundary interface.
type Discoverer interface {
	Discover(ctx context.Context, opts DiscoverOptions) ([]DiscoveredFile, error)
}

// DiscoverOptions describes a package boundary value.
type DiscoverOptions struct {
	ProjectDir string
}

// DiscoveredFile describes a package boundary value.
type DiscoveredFile struct {
	Path         string
	RelativePath string
}

type discoverer struct{}

// NewDiscoverer constructs the corresponding package service or value.
func NewDiscoverer() Discoverer {
	return &discoverer{}
}

// Discover executes the corresponding package operation.
func (d *discoverer) Discover(ctx context.Context, opts DiscoverOptions) ([]DiscoveredFile, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	ticketsDir := Directory(opts.ProjectDir)
	files := []DiscoveredFile{}

	err = filepath.WalkDir(ticketsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		err := ctx.Err()
		if err != nil {
			return err
		}

		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		relativePath, err := filepath.Rel(ticketsDir, path)
		if err != nil {
			return fmt.Errorf("resolve relative ticket path: %w", err)
		}

		files = append(files, DiscoveredFile{
			Path:         path,
			RelativePath: filepath.ToSlash(relativePath),
		})

		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("ticket directory is required: %s", ticketsDir)
	}

	if err != nil {
		return nil, fmt.Errorf("discover ticket files: %w", err)
	}

	return sortDiscoveredTicketFiles(files), nil
}

func sortDiscoveredTicketFiles(files []DiscoveredFile) []DiscoveredFile {
	slices.SortFunc(files, func(left, right DiscoveredFile) int {
		if left.RelativePath < right.RelativePath {
			return -1
		}

		if left.RelativePath > right.RelativePath {
			return 1
		}

		return 0
	})

	return files
}

// Directory resolves the legacy ticket document directory.
func Directory(projectDir string) string {
	return workspacecore.UserPath(projectDir, "tickets")
}
