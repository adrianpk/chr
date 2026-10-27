package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ListPlanDocuments enumerates concrete canonical plans across namespaces,
// including retained archives, within 4096 inventory entries and 8 MiB per read.
func ListPlanDocuments(ctx context.Context, projectDir string) ([]PlanDocument, error) {
	r := &fileAuthoringSessionRepository{projectDir: projectDir}

	namespaces, err := r.authoringDirectoryEntries(OpsDirectoryName)
	if err != nil {
		return nil, err
	}

	var documents []PlanDocument

	count, size := len(namespaces), 0
	for _, namespace := range namespaces {
		if !namespace.IsDir() || !opsnamespace.Valid(namespace.Name()) {
			continue
		}

		root := filepath.Join(OpsDirectoryName, namespace.Name(), "plan")

		entries, err := r.authoringDirectoryEntries(root)
		if err != nil {
			return nil, err
		}

		count += len(entries)
		if count > workflowstate.RecordLimit {
			return nil, errors.New("plan inventory exceeds 4096 entries")
		}

		batch, err := r.planDocumentsInDirectory(ctx, root, entries, &size)
		if err != nil {
			return nil, err
		}

		documents = append(documents, batch...)
	}

	sort.Slice(documents, func(i, j int) bool { return documents[i].Reference < documents[j].Reference })

	return documents, r.checkProjectDumpInspection(ctx)
}

func (r *fileAuthoringSessionRepository) planDocumentsInDirectory(ctx context.Context, root string, entries []os.DirEntry, size *int) ([]PlanDocument, error) {
	var documents []PlanDocument

	for _, entry := range entries {
		if !opsArtifactIDPattern.MatchString(entry.Name()) {
			continue
		}

		document, err := ReadPlanDocument(ctx, r.projectDir, filepath.ToSlash(filepath.Join(root, entry.Name(), OpsArtifactBodyFile())))
		if err != nil {
			return nil, err
		}

		*size += document.size
		if *size > 8*workflowstate.FileLimit {
			return nil, errors.New("plan inventory exceeds 8 MiB")
		}

		documents = append(documents, document)
	}

	return documents, nil
}
