package core

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

// ListSpecDocuments enumerates actual canonical documents across namespaces,
// with at most 4096 inventory entries and 8 MiB of retained images per read.
func ListSpecDocuments(ctx context.Context, projectDir string) ([]SpecDocument, error) {
	r := &fileAuthoringSessionRepository{projectDir: projectDir}

	namespaces, err := r.authoringDirectoryEntries(opscore.DirectoryName)
	if err != nil {
		return nil, err
	}

	var documents []SpecDocument

	count, size := len(namespaces), 0
	for _, namespace := range namespaces {
		if !namespace.IsDir() || !opsnamespace.Valid(namespace.Name()) {
			continue
		}

		root := filepath.Join(opscore.DirectoryName, namespace.Name(), "spec")

		entries, err := r.authoringDirectoryEntries(root)
		if err != nil {
			return nil, err
		}

		count += len(entries)
		if count > workflowstate.RecordLimit {
			return nil, errors.New("spec inventory exceeds 4096 entries")
		}

		batch, err := r.specDocumentsInDirectory(ctx, root, entries, &size)
		if err != nil {
			return nil, err
		}

		documents = append(documents, batch...)
	}

	sort.Slice(documents, func(i, j int) bool { return documents[i].Reference < documents[j].Reference })

	return documents, r.checkProjectDumpInspection(ctx)
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

func (r *fileAuthoringSessionRepository) specDocumentsInDirectory(ctx context.Context, root string, entries []os.DirEntry, size *int) ([]SpecDocument, error) {
	var documents []SpecDocument

	for _, entry := range entries {
		if !opscore.HasArtifactNumberPrefix(entry.Name()) {
			continue
		}

		document, err := ReadSpecDocument(ctx, r.projectDir, filepath.ToSlash(filepath.Join(root, entry.Name(), opscore.ArtifactBodyFile())))
		if err != nil {
			return nil, err
		}

		*size += len(document.images[0].After) + len(document.images[1].After)
		if *size > 8*workflowstate.FileLimit {
			return nil, errors.New("spec inventory exceeds 8 MiB")
		}

		documents = append(documents, document)
	}

	return documents, nil
}
