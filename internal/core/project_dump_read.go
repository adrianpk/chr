package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const projectDumpInventoryBytes = 8 * WorkflowFileLimit

type projectDumpInventory struct {
	repository     *fileAuthoringSessionRepository
	entries, bytes int
	artifacts      []ProjectDumpArtifact
}

func (r *fileAuthoringSessionRepository) checkProjectDumpInspection(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	pending, err := r.authoringPath(authoringRuntimeRef("pending.json"))
	if err != nil {
		return err
	}

	if _, err := os.Lstat(pending); !errors.Is(err, os.ErrNotExist) {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	return nil
}

func (r *fileAuthoringSessionRepository) projectDumpArtifacts(ctx context.Context) ([]ProjectDumpArtifact, error) {
	scan := projectDumpInventory{repository: r}

	namespaces, err := scan.readEntries(OpsDirectoryName)
	if err != nil {
		return nil, err
	}

	for _, entry := range namespaces {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := projectDumpEntryKind(entry); err != nil {
			return nil, err
		}

		if !entry.IsDir() {
			continue
		}

		if !opsnamespace.Valid(entry.Name()) {
			return nil, fmt.Errorf("%w: invalid ops namespace", ErrProjectDumpReconciliation)
		}

		if err := scan.readNamespace(ctx, entry.Name()); err != nil {
			return nil, err
		}
	}

	sort.Slice(scan.artifacts, func(i, j int) bool { return scan.artifacts[i].Reference < scan.artifacts[j].Reference })

	return scan.artifacts, nil
}

func (scan *projectDumpInventory) readNamespace(ctx context.Context, namespace string) error {
	root := filepath.Join(OpsDirectoryName, namespace, "dump")

	entries, err := scan.readEntries(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := projectDumpEntryKind(entry); err != nil {
			return err
		}

		if !entry.IsDir() {
			if err := projectDumpLooseFile(entry.Name()); err != nil {
				return err
			}

			continue
		}

		ref := filepath.ToSlash(filepath.Join(root, entry.Name(), OpsArtifactBodyFile()))

		artifact, size, err := scan.repository.readProjectDumpArtifact(ref)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrProjectDumpReconciliation, ref, err)
		}

		scan.bytes += size
		if scan.bytes > projectDumpInventoryBytes {
			return fmt.Errorf("project dump inventory exceeds 8 MiB")
		}

		scan.artifacts = append(scan.artifacts, artifact)
	}

	return nil
}

func (scan *projectDumpInventory) readEntries(relative string) ([]os.DirEntry, error) {
	path, err := scan.repository.authoringPath(relative)
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

	entries, readErr := directory.ReadDir(WorkflowRecordLimit - scan.entries + 1)

	closeErr := directory.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}

	scan.entries += len(entries)
	if scan.entries > WorkflowRecordLimit {
		return nil, fmt.Errorf("project dump inventory exceeds %d entries", WorkflowRecordLimit)
	}

	return entries, nil
}

func projectDumpEntryKind(entry os.DirEntry) error {
	if entry.Type()&os.ModeSymlink != 0 {
		return fmt.Errorf("project dump inventory cannot traverse a symlink")
	}

	return nil
}

func projectDumpLooseFile(name string) error {
	if strings.HasSuffix(strings.ToLower(name), ".md") && name != "README.md" && name != "TEMPLATE.md" {
		return fmt.Errorf("%w: loose dump document %s needs an explicit artifact mapping", ErrProjectDumpReconciliation, name)
	}

	return nil
}

func projectDumpObservation(project ProjectMetadata, artifacts []ProjectDumpArtifact) ProjectDumpObservation {
	observed := ProjectDumpObservation{Project: project}
	images := []string{project.SourcePath, projectDumpMetadataImage(project)}

	for _, artifact := range artifacts {
		images = append(images, artifact.Reference, artifact.SourceHash, artifact.MetadataHash)
		if artifact.CanonicalReference != "" {
			observed.Aliases = append(observed.Aliases, artifact)

			continue
		}

		if artifact.CaseSlug != "" {
			if !artifact.Raw {
				observed.Scoped = append(observed.Scoped, artifact)
			}

			continue
		}

		if artifact.Raw {
			observed.Components = append(observed.Components, artifact)
		} else {
			observed.Candidates = append(observed.Candidates, artifact)
		}
	}

	observed.Revision = contenthash.Sum([]byte(strings.Join(images, "\x00")))
	setProjectDumpDestination(&observed)

	return observed
}

func setProjectDumpDestination(observed *ProjectDumpObservation) {
	if observed.Project.ProjectDumpRef != "" {
		observed.Destination = observed.Project.ProjectDumpRef
		if len(observed.Candidates) != 1 || observed.Candidates[0].Reference != observed.Destination {
			observed.ReconciliationReason = "the project reference must identify the sole general dump; its target is missing, scoped, raw or conflicts with other general dumps"
		}

		return
	}

	switch len(observed.Candidates) {
	case 0:
		if len(observed.Components) > 0 {
			observed.ReconciliationReason = "raw captures exist without a general destination"
		}
	case 1:
		observed.Destination = observed.Candidates[0].Reference
	default:
		observed.ReconciliationReason = "multiple general dumps require explicit reconciliation"
	}
}

func projectDumpMetadataImage(project ProjectMetadata) string {
	if project.Exists {
		return project.SourceHash
	}

	return contenthash.Sum([]byte(RenderProjectMetadata(project)))
}
