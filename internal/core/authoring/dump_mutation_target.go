package authoring

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type dumpMutationTarget struct {
	Project                   ProjectMetadata
	Artifact                  ProjectDumpArtifact
	Destination, RawReference string
}

func (r *fileSessionRepository) dumpMutationTarget(ctx context.Context, request DumpMutationRequest) (dumpMutationTarget, error) {
	if request.Action != "capture" {
		return r.dumpEditTarget(ctx, request)
	}

	if request.CaseSlug != "" {
		if err := r.validateDumpCase(ctx, request.CaseSlug, request.ExpectedRevision); err != nil {
			return dumpMutationTarget{}, err
		}

		return r.newDumpMutationTarget(request.Namespace, request.Title)
	}

	observed, err := observeProjectDump(ctx, r.projectDir)
	if err != nil {
		return dumpMutationTarget{}, err
	}

	if observed.ReconciliationReason != "" {
		return dumpMutationTarget{}, fmt.Errorf("%w: %s", ErrProjectDumpReconciliation, observed.ReconciliationReason)
	}

	if request.ExpectedRevision != "" && request.ExpectedRevision != observed.Revision {
		return dumpMutationTarget{}, workflowstate.ErrConflict
	}

	if observed.Destination == "" {
		target, err := r.newDumpMutationTarget(opsnamespace.Default, request.Title)
		target.Project = observed.Project

		return target, err
	}

	target := dumpMutationTarget{Project: observed.Project, Artifact: observed.Candidates[0], Destination: observed.Destination}
	parts := strings.Split(target.Destination, "/")

	target.RawReference = strings.Join([]string{opscore.DirectoryName, parts[1], "dump", "00000-" + contenthash.Sum([]byte(request.OperationID)) + "-raw", opscore.ArtifactBodyFile()}, "/")
	if err := r.requireNewDumpEnvelope(target.RawReference); err != nil {
		return target, err
	}

	return target, nil
}

func (r *fileSessionRepository) dumpEditTarget(ctx context.Context, request DumpMutationRequest) (dumpMutationTarget, error) {
	observed, err := ReadDumpDocument(ctx, r.projectDir, request.Reference)
	if err != nil {
		return dumpMutationTarget{}, err
	}

	if observed.Revision != request.ExpectedRevision || observed.Artifact.Raw {
		return dumpMutationTarget{}, workflowstate.ErrConflict
	}

	target := dumpMutationTarget{Artifact: observed.Artifact, Destination: observed.Artifact.Reference}
	if observed.Artifact.CaseSlug != "" {
		return target, r.validateDumpCase(ctx, observed.Artifact.CaseSlug, "")
	}

	project, err := observeProjectDump(ctx, r.projectDir)
	if err != nil {
		return target, err
	}

	if project.ReconciliationReason != "" || project.Destination != target.Destination {
		return target, ErrProjectDumpReconciliation
	}

	target.Project = project.Project

	return target, nil
}

func (r *fileSessionRepository) validateDumpCase(ctx context.Context, slug, expectedHash string) error {
	metadata, err := NewCaseMetadataService(r.projectDir).ReadBounded(ctx, slug)
	if err != nil {
		return err
	}

	if metadata.Slug != slug || expectedHash != "" && metadata.SourceHash != expectedHash {
		return workflowstate.ErrConflict
	}

	return nil
}

func (r *fileSessionRepository) newDumpMutationTarget(namespace, title string) (dumpMutationTarget, error) {
	slug := opscore.ArtifactSlugOrDefault(title)
	for attempt := 1; attempt <= 100; attempt++ {
		id := "00000-" + slug
		if attempt > 1 {
			id += fmt.Sprintf("-%d", attempt)
		}

		root := strings.Join([]string{opscore.DirectoryName, namespace, "dump"}, "/")

		target := dumpMutationTarget{Destination: root + "/" + id + "/artifact.md", RawReference: root + "/" + id + "-raw/artifact.md"}
		if err := r.requireNewDumpEnvelope(target.Destination); errors.Is(err, os.ErrExist) {
			continue
		} else if err != nil {
			return target, err
		}

		if err := r.requireNewDumpEnvelope(target.RawReference); errors.Is(err, os.ErrExist) {
			continue
		} else if err != nil {
			return target, err
		}

		return target, nil
	}

	return dumpMutationTarget{}, errors.New("dump artifact allocation exceeds 100 candidates")
}

func (r *fileSessionRepository) requireNewDumpEnvelope(reference string) error {
	if err := ValidateProjectDumpReference(reference); err != nil {
		return err
	}

	path, err := r.authoringPath(filepath.Dir(filepath.FromSlash(reference)))
	if err != nil {
		return err
	}

	_, err = os.Lstat(path)
	if err == nil {
		return os.ErrExist
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
