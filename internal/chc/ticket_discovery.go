package chc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type TicketFileDiscoverer interface {
	DiscoverTicketFiles(ctx context.Context, opts DiscoverTicketFilesOptions) ([]DiscoveredTicketFile, error)
}

type DiscoverTicketFilesOptions struct {
	ProjectDir string
}

type DiscoveredTicketFile struct {
	Path string
}

type ticketFileDiscoverer struct{}

func NewTicketFileDiscoverer() TicketFileDiscoverer {
	return &ticketFileDiscoverer{}
}

func (d *ticketFileDiscoverer) DiscoverTicketFiles(
	ctx context.Context,
	opts DiscoverTicketFilesOptions,
) ([]DiscoveredTicketFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ticketsDir := filepath.Join(opts.ProjectDir, DirectoryName, "tickets")
	files := []DiscoveredTicketFile{}

	err := filepath.WalkDir(ticketsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		files = append(files, DiscoveredTicketFile{Path: path})

		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("ticket directory is required: %s", ticketsDir)
	}

	if err != nil {
		return nil, fmt.Errorf("discover ticket files: %w", err)
	}

	return files, nil
}
