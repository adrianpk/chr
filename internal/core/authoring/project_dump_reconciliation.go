package authoring

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ProjectDumpReconciliationPreview binds a deterministic composition to all
// observed source images. Timestamp issues remain visible in Sources.
type ProjectDumpReconciliationPreview struct {
	Revision, Destination, Content string
	Sources                        []ProjectDumpArtifact
}

// ProjectDumpReconciliationRequest confirms the destination proposed by a preview.
// OperationID must be retained across retries, including a lost acknowledgement.
type ProjectDumpReconciliationRequest struct {
	OperationID, ExpectedRevision, Destination string
}

// ProjectDumpMutationResult is the immutable committed result of one dump write.
type ProjectDumpMutationResult struct {
	OperationID, Destination, SourceHash, MetadataHash string
}

// PreviewProjectDumpReconciliation reads source images without adopting or writing.
func PreviewProjectDumpReconciliation(ctx context.Context, projectDir string) (ProjectDumpReconciliationPreview, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	observed, err := observeProjectDump(ctx, projectDir)
	if err != nil {
		return ProjectDumpReconciliationPreview{}, err
	}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}

	preview, err := r.projectDumpReconciliationPreview(observed)
	if err != nil {
		return preview, err
	}

	return preview, r.checkProjectDumpInspection(ctx)
}

func (r *fileAuthoringSessionRepository) projectDumpReconciliationPreview(observed ProjectDumpObservation) (ProjectDumpReconciliationPreview, error) {
	preview := ProjectDumpReconciliationPreview{Revision: observed.Revision, Sources: slices.Clone(observed.Candidates)}
	if len(preview.Sources) < 2 || len(preview.Sources) > projectDumpSourceLimit {
		return preview, fmt.Errorf("%w: reconciliation requires between 2 and %d general dumps", ErrProjectDumpReconciliation, projectDumpSourceLimit)
	}

	slices.SortFunc(preview.Sources, compareProjectDumpCreation)

	preview.Destination = preview.Sources[0].Reference
	content, err := r.composeProjectDumps(preview.Sources)
	preview.Content = string(content)

	return preview, err
}

func compareProjectDumpCreation(a, b ProjectDumpArtifact) int {
	if (a.TimestampIssue == "") != (b.TimestampIssue == "") {
		if a.TimestampIssue == "" {
			return -1
		}

		return 1
	}

	if a.TimestampIssue == "" && !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.Compare(b.CreatedAt)
	}

	return strings.Compare(a.Reference, b.Reference)
}

// ReconcileProjectDump retains original body and metadata images, composes the
// canonical body and publishes aliases, project reference and receipt together.
func ReconcileProjectDump(ctx context.Context, projectDir string, request ProjectDumpReconciliationRequest) (ProjectDumpMutationResult, error) {
	root, err := filepath.Abs(projectDir)
	if err != nil || strings.TrimSpace(projectDir) == "" {
		return ProjectDumpMutationResult{}, errors.New("project directory is required")
	}

	return (&fileAuthoringSessionRepository{projectDir: root}).reconcileProjectDump(ctx, request)
}

func (r *fileAuthoringSessionRepository) reconcileProjectDump(ctx context.Context, request ProjectDumpReconciliationRequest) (result ProjectDumpMutationResult, err error) {
	if err := validateProjectDumpReconciliation(request); err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	digest, err := projectDumpReconciliationDigest(r.projectDir, request)
	if err != nil {
		return result, err
	}

	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		var found bool

		receiptLookup, receiptErr := r.projectDumpReceipt(request.OperationID, digest)

		result, found, err = receiptLookup.Result, receiptLookup.Found, receiptErr
		if err != nil || found {
			return err
		}

		if err := r.projectDumpOperationAvailable(request.OperationID); err != nil {
			return err
		}

		result, err = r.applyProjectDumpReconciliation(ctx, request, digest)

		return err
	})

	return result, err
}

func validateProjectDumpReconciliation(request ProjectDumpReconciliationRequest) error {
	if err := authoringstate.ValidateSessionID(request.OperationID); err != nil {
		return fmt.Errorf("dump operation identity: %w", err)
	}

	if len(request.ExpectedRevision) != 64 || request.Destination == "" {
		return fmt.Errorf("observed dump revision and destination are required")
	}

	if _, err := hex.DecodeString(request.ExpectedRevision); err != nil {
		return fmt.Errorf("invalid observed dump revision: %w", err)
	}

	return ValidateProjectDumpReference(request.Destination)
}

func (r *fileAuthoringSessionRepository) applyProjectDumpReconciliation(ctx context.Context, request ProjectDumpReconciliationRequest, digest string) (ProjectDumpMutationResult, error) {
	observed, err := observeProjectDump(ctx, r.projectDir)
	if err != nil {
		return ProjectDumpMutationResult{}, err
	}

	if observed.Revision != request.ExpectedRevision {
		return ProjectDumpMutationResult{}, workflowstate.ErrConflict
	}

	preview, err := r.projectDumpReconciliationPreview(observed)
	if err != nil {
		return ProjectDumpMutationResult{}, err
	}

	if preview.Destination != request.Destination {
		return ProjectDumpMutationResult{}, workflowstate.ErrConflict
	}

	dumpChanges, err := r.projectDumpReconciliationChanges(observed, preview, request.OperationID)

	changes, result := dumpChanges.Changes, dumpChanges.Result
	if err != nil {
		return result, err
	}

	receipt, err := projectDumpReceiptChange(digest, result)
	if err != nil {
		return result, err
	}

	changes = append(changes, receipt)
	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, changes); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}
