package authoring

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrProjectDumpReconciliation leaves existing material unchanged until an
// explicit, observed reconciliation can resolve the conflicting destination.
var ErrProjectDumpReconciliation = errors.New("project dump reconciliation required")

// ProjectDumpArtifact is one observed ops dump with its component and scope identity.
type ProjectDumpArtifact struct {
	Reference, Title, SourceHash, MetadataHash string
	CaseSlug, RawReference, CuratedReference   string
	CanonicalReference                         string
	Status, DocumentType                       string
	Raw                                        bool
	CreatedAt                                  time.Time
	UpdatedAt                                  time.Time
	TimestampIssue                             string
}

// ProjectDumpObservation is a read-only project-wide inventory, independent of
// the active namespace and rebuildable indexes. Revision binds the project root and every source image.
type ProjectDumpObservation struct {
	Project              ProjectMetadata
	Candidates           []ProjectDumpArtifact
	Components           []ProjectDumpArtifact
	Aliases              []ProjectDumpArtifact
	Scoped               []ProjectDumpArtifact
	Destination          string
	Revision             string
	ReconciliationReason string
}

// ValidateProjectDumpReference accepts only a canonical project-relative dump body.
// Existence and document ownership are checked by ObserveProjectDump.
func ValidateProjectDumpReference(reference string) error {
	if reference == "" {
		return nil
	}

	parts := strings.Split(reference, "/")
	if len(reference) > 4096 || path.Clean(reference) != reference || strings.ContainsAny(reference, "\\\x00\r\n") || len(parts) != 5 {
		return fmt.Errorf("project_dump_ref must be a canonical project-relative dump body")
	}

	if !validProjectDumpEnvelope(parts) {
		return fmt.Errorf("project_dump_ref must identify an ops dump body in a valid namespace")
	}

	return nil
}

func validProjectDumpEnvelope(parts []string) bool {
	return parts[0] == opscore.DirectoryName && parts[2] == "dump" && parts[4] == opscore.ArtifactBodyFile() && opsnamespace.Valid(parts[1]) && strings.TrimSpace(parts[1]) == parts[1]
}

// ObserveProjectDump inspects canonical documents without creating metadata,
// adopting a destination, rebuilding an index or recovering a pending mutation.
func ObserveProjectDump(ctx context.Context, projectDir string) (ProjectDumpObservation, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return observeProjectDump(ctx, projectDir)
}

func observeProjectDump(ctx context.Context, projectDir string) (ProjectDumpObservation, error) {
	if strings.TrimSpace(projectDir) == "" {
		return ProjectDumpObservation{}, fmt.Errorf("project directory is required")
	}

	if err := ctx.Err(); err != nil {
		return ProjectDumpObservation{}, err
	}

	root, err := filepath.Abs(projectDir)
	if err != nil {
		return ProjectDumpObservation{}, err
	}

	projectDir = root

	project, err := NewProjectMetadataService(projectDir).Load()
	if err != nil {
		return ProjectDumpObservation{}, err
	}

	repository := &fileAuthoringSessionRepository{projectDir: projectDir}

	return repository.observeProjectDump(ctx, project)
}

func (r *fileAuthoringSessionRepository) observeProjectDump(ctx context.Context, project ProjectMetadata) (ProjectDumpObservation, error) {
	if err := r.checkProjectDumpInspection(ctx); err != nil {
		return ProjectDumpObservation{}, err
	}

	artifacts, err := r.projectDumpArtifacts(ctx)
	if err != nil {
		return ProjectDumpObservation{}, err
	}

	if err := validateProjectDumpAliases(artifacts); err != nil {
		return ProjectDumpObservation{}, err
	}

	observed := projectDumpObservation(project, artifacts)
	if err := r.checkProjectDumpInspection(ctx); err != nil {
		return observed, err
	}

	current, err := NewProjectMetadataService(r.projectDir).Load()
	if err != nil {
		return observed, err
	}

	if projectDumpMetadataImage(current) != projectDumpMetadataImage(project) {
		return observed, workflowstate.ErrConflict
	}

	return observed, nil
}

// AdoptProjectDump binds the sole existing general dump under the shared
// authoring owner. It does not allocate, move or rewrite artifact components.
func (s ProjectMetadataService) AdoptProjectDump(ctx context.Context, expectedRevision string) (saved ProjectMetadata, err error) {
	if expectedRevision == "" {
		return saved, fmt.Errorf("observed project dump revision is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = WithAuthoringMutation(ctx, s.projectDir, func(ctx context.Context) error {
		observed, readErr := observeProjectDump(ctx, s.projectDir)
		if readErr != nil {
			return readErr
		}

		if observed.Revision != expectedRevision {
			return workflowstate.ErrConflict
		}

		if observed.ReconciliationReason != "" || observed.Destination == "" {
			return fmt.Errorf("%w: %s", ErrProjectDumpReconciliation, projectDumpAdoptionReason(observed))
		}

		if observed.Project.ProjectDumpRef != "" {
			saved = observed.Project

			return nil
		}

		observed.Project.ProjectDumpRef = observed.Destination
		saved, readErr = s.save(observed.Project, true)

		return readErr
	})

	return saved, err
}

func projectDumpAdoptionReason(observed ProjectDumpObservation) string {
	if observed.ReconciliationReason != "" {
		return observed.ReconciliationReason
	}

	return "there is no existing general dump to adopt"
}
