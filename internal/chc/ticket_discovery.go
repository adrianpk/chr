package chc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type TicketFileDiscoverer interface {
	DiscoverTicketFiles(ctx context.Context, opts DiscoverTicketFilesOptions) ([]DiscoveredTicketFile, error)
}

type DiscoverTicketFilesOptions struct {
	ProjectDir string
}

type DiscoveredTicketFile struct {
	Path         string
	RelativePath string
}

type ticketFileDiscoverer struct{}

func NewTicketFileDiscoverer() TicketFileDiscoverer {
	return &ticketFileDiscoverer{}
}

func (d *ticketFileDiscoverer) DiscoverTicketFiles(
	ctx context.Context,
	opts DiscoverTicketFilesOptions,
) ([]DiscoveredTicketFile, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	ticketsDir := ticketsDirectory(opts.ProjectDir)
	files := []DiscoveredTicketFile{}

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

		files = append(files, DiscoveredTicketFile{
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

func sortDiscoveredTicketFiles(files []DiscoveredTicketFile) []DiscoveredTicketFile {
	slices.SortFunc(files, func(left, right DiscoveredTicketFile) int {
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

func ticketsDirectory(projectDir string) string {
	return UserPath(projectDir, "tickets")
}
