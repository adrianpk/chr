package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowTrackerPreparation struct {
	Status   TrackerStatus
	Output   []byte
	Previous TrackerStatus
}

// TrackerSource binds an export to its authoritative revision and bytes.
type TrackerSource struct {
	ID       string `json:"id,omitempty"`
	Path     string `json:"path"`
	Revision uint64 `json:"revision,omitempty"`
	Hash     string `json:"hash"`
}

// TrackerStatus describes versioned order/progress views, never execution authority.
type TrackerStatus struct {
	PlanRevision *workflowstate.PlanRevisionBinding `json:"plan_revision,omitempty"`
	Schema       string                             `json:"schema"`
	ProjectID    string                             `json:"project_id"`
	Namespace    string                             `json:"namespace"`
	SetID        string                             `json:"slice_set_id"`
	Path         string                             `json:"path"`
	Sources      []TrackerSource                    `json:"sources"`
	SourceHash   string                             `json:"source_hash"`
	OutputHash   string                             `json:"output_hash"`
	PendingHash  string                             `json:"pending_output_hash,omitempty"`
	ObservedHash string                             `json:"observed_output_hash,omitempty"`
	GeneratedAt  time.Time                          `json:"generated_at"`
	Status       string                             `json:"status"`
	Reason       string                             `json:"reason,omitempty"`
	Detail       string                             `json:"detail,omitempty"`
	Checksum     string                             `json:"checksum"`
}

// TrackerOptions grants replacement only for exactly observed output bytes.
type TrackerOptions struct {
	Schema      string `json:"schema,omitempty"`
	ReplaceHash string `json:"replace_hash,omitempty"`
}

// ExportWorkflowTracker registers or repairs one derived set view under exclusion.
// It never imports tracker edits, claims work or rolls back authoritative records.
func (s *Store) ExportWorkflowTracker(ctx context.Context, setID string, opts TrackerOptions) (result TrackerStatus, err error) {
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
func (s *Store) InspectWorkflowTracker(ctx context.Context, setID string) (result TrackerStatus, err error) {
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

		trackerImage, readErr := s.versionedWorkflowTrackerImage(snapshot, setID, result.Schema)
		if readErr != nil {
			return readErr
		}

		result.ObservedHash, readErr = s.workflowTrackerOutput(setID)
		if readErr != nil || trackerImage.Status.SourceHash != result.SourceHash || result.ObservedHash != result.OutputHash {
			result.Status, result.Reason = "stale", "projection_stale"
			result.Detail = "authoritative sources or tracker output changed"
		}

		if imageErr := s.inspectRequiredOrderImages(snapshot, trackerImage.Status); imageErr != nil {
			return imageErr
		}

		return readErr
	})

	return workflowTrackerResult(result, err)
}

func workflowTrackerResult(result TrackerStatus, err error) (TrackerStatus, error) {
	if err != nil && result.SetID != "" {
		result.Status, result.Reason, result.Detail = "stale", "projection_stale", err.Error()

		return result, errors.Join(workflowstate.ErrProjectionStale, err)
	}

	return result, err
}

func (s *Store) exportWorkflowTracker(snapshot Snapshot, setID string, opts TrackerOptions) (TrackerStatus, error) {
	trackerPreparation, err := s.prepareWorkflowTrackerImage(snapshot, setID, opts)
	if err != nil {
		return trackerPreparation.Status, err
	}

	trackerPreparation.Status.ObservedHash, err = s.workflowTrackerOutput(setID)
	if err != nil {
		return s.workflowTrackerFailure(trackerPreparation.Status, err)
	}

	if !workflowTrackerMayReplace(trackerPreparation.Status, trackerPreparation.Previous, opts) {
		return s.workflowTrackerFailure(trackerPreparation.Status, fmt.Errorf("tracker edited; replacement requires its observed hash"))
	}

	if err := s.repairWorkflowTrackerCompanions(snapshot, trackerPreparation.Status, trackerPreparation.Output, opts); err != nil {
		return s.workflowTrackerFailure(trackerPreparation.Status, err)
	}

	if trackerPreparation.Previous.Status == "current" && trackerPreparation.Previous.SourceHash == trackerPreparation.Status.SourceHash && trackerPreparation.Status.ObservedHash == trackerPreparation.Status.PendingHash {
		return trackerPreparation.Previous, nil
	}

	if trackerPreparation.Status.ObservedHash != "" && trackerPreparation.Status.ObservedHash == trackerPreparation.Previous.PendingHash {
		trackerPreparation.Status.OutputHash = trackerPreparation.Status.ObservedHash
	}

	return s.writeWorkflowTracker(trackerPreparation.Status, trackerPreparation.Output)
}

func (s *Store) workflowTrackerPrevious(setID string) (TrackerStatus, error) {
	previous, err := s.readWorkflowTracker(setID)
	if !errors.Is(err, errWorkflowTrackerAbsent) {
		return previous, err
	}

	{
		trackerRegistration, err := s.retainedOrderRegistration(setID)

		if trackerRegistration.Found || err != nil {
			trackerRegistration.Status.Status, trackerRegistration.Status.Reason = "stale", "registration_missing"

			return trackerRegistration.Status, err
		}
	}

	ids, err := s.workflowTrackerRegistrations()
	if err != nil || len(ids) >= workflowstate.PageLimit {
		return previous, errors.Join(workflowstate.ErrProjectionStale, fmt.Errorf("cannot register tracker within 128-projection bound"), err)
	}

	return previous, nil
}

func workflowTrackerReplacement(observed string, previous TrackerStatus, opts TrackerOptions) bool {
	if opts.ReplaceHash != "" {
		return observed != "" && opts.ReplaceHash == observed
	}

	return observed == "" || observed == previous.OutputHash || observed == previous.PendingHash
}

func (s *Store) workflowTrackerFailure(result TrackerStatus, failure error) (TrackerStatus, error) {
	result.Status, result.Reason, result.Detail = "stale", "projection_stale", failure.Error()

	return result, errors.Join(workflowstate.ErrProjectionStale, failure, s.persistWorkflowTracker(result))
}

func (s *Store) prepareWorkflowTrackerImage(snapshot Snapshot, setID string, opts TrackerOptions) (workflowTrackerPreparation, error) {
	previous, err := s.workflowTrackerPrevious(setID)
	if err != nil {
		return workflowTrackerPreparation{Status: TrackerStatus{}, Output: nil, Previous: previous}, err
	}

	schema := workflowTrackerSchema(snapshot, setID, opts.Schema, previous.Schema)

	trackerImage2, err := s.versionedWorkflowTrackerImage(snapshot, setID, schema)
	if err != nil {
		return workflowTrackerPreparation{Status: trackerImage2.Status, Output: nil, Previous: previous}, err
	}

	trackerImage2.Status.OutputHash, trackerImage2.Status.PendingHash = previous.OutputHash, contenthash.Sum(trackerImage2.Output)

	trackerImage2.Status.GeneratedAt = previous.GeneratedAt
	if trackerImage2.Status.Schema == "chr-tracker/v2" && trackerImage2.Status.SourceHash != previous.SourceHash {
		trackerImage2.Status.GeneratedAt = time.Time{}
	}

	trackerImage2.Status.Status, trackerImage2.Status.Reason = "stale", "projection_stale"

	return workflowTrackerPreparation{Status: trackerImage2.Status, Output: trackerImage2.Output, Previous: previous}, nil
}

func workflowTrackerMayReplace(result, previous TrackerStatus, opts TrackerOptions) bool {
	if workflowTrackerReplacement(result.ObservedHash, previous, opts) {
		return true
	}

	return result.Schema == "chr-tracker/v2" && workflowTrackerReplacement(result.ObservedHash, previous, TrackerOptions{})
}

func (s *Store) inspectRequiredOrderImages(snapshot Snapshot, source TrackerStatus) error {
	if source.Schema != "chr-tracker/v2" {
		return nil
	}

	trackerImage3, err := s.workflowOrderTrackerImage(snapshot, source.SetID)
	if err != nil {
		return err
	}

	return s.checkRequiredOrderImages(source, trackerImage3.Output)
}

func (s *Store) repairWorkflowTrackerCompanions(snapshot Snapshot, result TrackerStatus, output []byte, opts TrackerOptions) error {
	if result.Schema != "chr-tracker/v2" {
		return nil
	}

	return s.repairRequiredOrderImages(snapshot, result, output, opts)
}
