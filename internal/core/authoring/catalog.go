package authoring

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type authoringCatalogDocument[T any] struct {
	value     T
	reference string
	size      int
}

type authoringCatalogReader[T any] func(context.Context, string, string) (authoringCatalogDocument[T], error)

type authoringCatalog[T any] struct {
	repository *fileAuthoringSessionRepository
	kind       string
	read       authoringCatalogReader[T]
	size       int
}

func listAuthoringDocuments[T any](ctx context.Context, projectDir, kind string, read authoringCatalogReader[T]) ([]T, error) {
	catalog := authoringCatalog[T]{repository: &fileAuthoringSessionRepository{projectDir: projectDir}, kind: kind, read: read}

	return catalog.list(ctx)
}

func (c *authoringCatalog[T]) list(ctx context.Context) ([]T, error) {
	namespaces, err := c.repository.authoringDirectoryEntries(opscore.DirectoryName)
	if err != nil {
		return nil, err
	}

	var observed []authoringCatalogDocument[T]

	count := len(namespaces)
	for _, namespace := range namespaces {
		if !namespace.IsDir() || !opsnamespace.Valid(namespace.Name()) {
			continue
		}

		root := filepath.Join(opscore.DirectoryName, namespace.Name(), c.kind)

		entries, err := c.repository.authoringDirectoryEntries(root)
		if err != nil {
			return nil, err
		}

		count += len(entries)
		if count > workflowstate.RecordLimit {
			return nil, fmt.Errorf("%s inventory exceeds 4096 entries", c.kind)
		}

		batch, err := c.readDirectory(ctx, root, entries)
		if err != nil {
			return nil, err
		}

		observed = append(observed, batch...)
	}

	sort.Slice(observed, func(i, j int) bool { return observed[i].reference < observed[j].reference })

	var documents []T
	for _, document := range observed {
		documents = append(documents, document.value)
	}

	return documents, c.repository.checkProjectDumpInspection(ctx)
}

func (c *authoringCatalog[T]) readDirectory(ctx context.Context, root string, entries []os.DirEntry) ([]authoringCatalogDocument[T], error) {
	var documents []authoringCatalogDocument[T]

	for _, entry := range entries {
		if !opscore.HasArtifactNumberPrefix(entry.Name()) {
			continue
		}

		document, err := c.read(ctx, c.repository.projectDir, filepath.ToSlash(filepath.Join(root, entry.Name(), opscore.ArtifactBodyFile())))
		if err != nil {
			return nil, err
		}

		c.size += document.size
		if c.size > 8*workflowstate.FileLimit {
			return nil, fmt.Errorf("%s inventory exceeds 8 MiB", c.kind)
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (r *fileAuthoringSessionRepository) authoringDirectoryEntries(reference string) ([]os.DirEntry, error) {
	path, err := r.authoringPath(reference)
	if err != nil {
		return nil, err
	}

	directory, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	entries, readErr := directory.ReadDir(workflowstate.RecordLimit + 1)

	closeErr := directory.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}

	if len(entries) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("authoring directory exceeds %d entries", workflowstate.RecordLimit)
	}

	return entries, nil
}
