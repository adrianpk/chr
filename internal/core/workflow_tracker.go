package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowProjectionStale preserves domain progress when a derived export fails.
var ErrWorkflowProjectionStale = errors.New("projection_stale")

// WorkflowTrackerSource binds an export to its authoritative revision and bytes.
type WorkflowTrackerSource struct {
	ID       string `json:"id,omitempty"`
	Path     string `json:"path"`
	Revision uint64 `json:"revision,omitempty"`
	Hash     string `json:"hash"`
}

// WorkflowTrackerStatus describes versioned order/progress views, never execution authority.
type WorkflowTrackerStatus struct {
	PlanRevision *generationcore.PlanRevisionBinding `json:"plan_revision,omitempty"`
	Schema       string                              `json:"schema"`
	ProjectID    string                              `json:"project_id"`
	Namespace    string                              `json:"namespace"`
	SetID        string                              `json:"slice_set_id"`
	Path         string                              `json:"path"`
	Sources      []WorkflowTrackerSource             `json:"sources"`
	SourceHash   string                              `json:"source_hash"`
	OutputHash   string                              `json:"output_hash"`
	PendingHash  string                              `json:"pending_output_hash,omitempty"`
	ObservedHash string                              `json:"observed_output_hash,omitempty"`
	GeneratedAt  time.Time                           `json:"generated_at"`
	Status       string                              `json:"status"`
	Reason       string                              `json:"reason,omitempty"`
	Detail       string                              `json:"detail,omitempty"`
	Checksum     string                              `json:"checksum"`
}

// WorkflowTrackerOptions grants replacement only for exactly observed output bytes.
type WorkflowTrackerOptions struct {
	Schema      string `json:"schema,omitempty"`
	ReplaceHash string `json:"replace_hash,omitempty"`
}

// ExportWorkflowTracker registers or repairs one derived set view under exclusion.
// It never imports tracker edits, claims work or rolls back authoritative records.
func (s *WorkflowStore) ExportWorkflowTracker(ctx context.Context, setID string, opts WorkflowTrackerOptions) (result WorkflowTrackerStatus, err error) {
	err = s.withWorkflowPlanMutation(ctx, func(ctx context.Context) error {
		snapshot, readErr := s.workflowTrackerSnapshot(ctx, setID, opts.Schema)
		if readErr != nil {
			return readErr
		}

		result, readErr = s.exportWorkflowTracker(snapshot, setID, opts)

		return readErr
	})

	return workflowTrackerResult(result, err)
}

// InspectWorkflowTracker observes source and output hashes without repairing files.
func (s *WorkflowStore) InspectWorkflowTracker(ctx context.Context, setID string) (result WorkflowTrackerStatus, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		var readErr error

		result, readErr = s.readWorkflowTracker(setID)
		if readErr != nil {
			return readErr
		}

		snapshot, readErr := s.workflowTrackerSnapshot(ctx, setID, result.Schema)
		if readErr != nil {
			return readErr
		}

		source, _, readErr := s.versionedWorkflowTrackerImage(snapshot, setID, result.Schema)
		if readErr != nil {
			return readErr
		}

		result.ObservedHash, readErr = s.workflowTrackerOutput(setID)
		if readErr != nil || source.SourceHash != result.SourceHash || result.ObservedHash != result.OutputHash {
			result.Status, result.Reason = "stale", "projection_stale"
			result.Detail = "authoritative sources or tracker output changed"
		}

		if imageErr := s.inspectRequiredOrderImages(snapshot, source); imageErr != nil {
			return imageErr
		}

		return readErr
	})

	return workflowTrackerResult(result, err)
}

func workflowTrackerResult(result WorkflowTrackerStatus, err error) (WorkflowTrackerStatus, error) {
	if err != nil && result.SetID != "" {
		result.Status, result.Reason, result.Detail = "stale", "projection_stale", err.Error()

		return result, errors.Join(ErrWorkflowProjectionStale, err)
	}

	return result, err
}

func (s *WorkflowStore) exportWorkflowTracker(snapshot WorkflowSnapshot, setID string, opts WorkflowTrackerOptions) (WorkflowTrackerStatus, error) {
	result, output, previous, err := s.prepareWorkflowTrackerImage(snapshot, setID, opts)
	if err != nil {
		return result, err
	}

	result.ObservedHash, err = s.workflowTrackerOutput(setID)
	if err != nil {
		return s.workflowTrackerFailure(result, err)
	}

	if !workflowTrackerMayReplace(result, previous, opts) {
		return s.workflowTrackerFailure(result, fmt.Errorf("tracker edited; replacement requires its observed hash"))
	}

	if err := s.repairWorkflowTrackerCompanions(snapshot, result, output, opts); err != nil {
		return s.workflowTrackerFailure(result, err)
	}

	if previous.Status == "current" && previous.SourceHash == result.SourceHash && result.ObservedHash == result.PendingHash {
		return previous, nil
	}

	if result.ObservedHash != "" && result.ObservedHash == previous.PendingHash {
		result.OutputHash = result.ObservedHash
	}

	return s.writeWorkflowTracker(result, output)
}

func (s *WorkflowStore) workflowTrackerPrevious(setID string) (WorkflowTrackerStatus, error) {
	previous, err := s.readWorkflowTracker(setID)
	if !errors.Is(err, errWorkflowTrackerAbsent) {
		return previous, err
	}

	if retained, found, err := s.retainedOrderRegistration(setID); found || err != nil {
		retained.Status, retained.Reason = "stale", "registration_missing"

		return retained, err
	}

	ids, err := s.workflowTrackerRegistrations()
	if err != nil || len(ids) >= workflowstate.PageLimit {
		return previous, errors.Join(ErrWorkflowProjectionStale, fmt.Errorf("cannot register tracker within 128-projection bound"), err)
	}

	return previous, nil
}

func workflowTrackerReplacement(observed string, previous WorkflowTrackerStatus, opts WorkflowTrackerOptions) bool {
	if opts.ReplaceHash != "" {
		return observed != "" && opts.ReplaceHash == observed
	}

	return observed == "" || observed == previous.OutputHash || observed == previous.PendingHash
}

func (s *WorkflowStore) workflowTrackerFailure(result WorkflowTrackerStatus, failure error) (WorkflowTrackerStatus, error) {
	result.Status, result.Reason, result.Detail = "stale", "projection_stale", failure.Error()

	return result, errors.Join(ErrWorkflowProjectionStale, failure, s.persistWorkflowTracker(result))
}

func (s *WorkflowStore) prepareWorkflowTrackerImage(snapshot WorkflowSnapshot, setID string, opts WorkflowTrackerOptions) (WorkflowTrackerStatus, []byte, WorkflowTrackerStatus, error) {
	previous, err := s.workflowTrackerPrevious(setID)
	if err != nil {
		return WorkflowTrackerStatus{}, nil, previous, err
	}

	schema := workflowTrackerSchema(snapshot, setID, opts.Schema, previous.Schema)

	result, output, err := s.versionedWorkflowTrackerImage(snapshot, setID, schema)
	if err != nil {
		return result, nil, previous, err
	}

	result.OutputHash, result.PendingHash = previous.OutputHash, contenthash.Sum(output)

	result.GeneratedAt = previous.GeneratedAt
	if result.Schema == "chr-tracker/v2" && result.SourceHash != previous.SourceHash {
		result.GeneratedAt = time.Time{}
	}

	result.Status, result.Reason = "stale", "projection_stale"

	return result, output, previous, nil
}

func workflowTrackerMayReplace(result, previous WorkflowTrackerStatus, opts WorkflowTrackerOptions) bool {
	if workflowTrackerReplacement(result.ObservedHash, previous, opts) {
		return true
	}

	return result.Schema == "chr-tracker/v2" && workflowTrackerReplacement(result.ObservedHash, previous, WorkflowTrackerOptions{})
}

func (s *WorkflowStore) inspectRequiredOrderImages(snapshot WorkflowSnapshot, source WorkflowTrackerStatus) error {
	if source.Schema != "chr-tracker/v2" {
		return nil
	}

	_, output, err := s.workflowOrderTrackerImage(snapshot, source.SetID)
	if err != nil {
		return err
	}

	return s.checkRequiredOrderImages(source, output)
}

func (s *WorkflowStore) repairWorkflowTrackerCompanions(snapshot WorkflowSnapshot, result WorkflowTrackerStatus, output []byte, opts WorkflowTrackerOptions) error {
	if result.Schema != "chr-tracker/v2" {
		return nil
	}

	return s.repairRequiredOrderImages(snapshot, result, output, opts)
}
