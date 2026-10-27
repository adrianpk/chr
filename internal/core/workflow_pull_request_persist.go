package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowPullRequestObservation struct {
	Schema      string                        `json:"schema"`
	BindingID   string                        `json:"binding_id"`
	CandidateID string                        `json:"candidate_id"`
	Reason      string                        `json:"reason"`
	Snapshots   []WorkflowPullRequestSnapshot `json:"snapshots"`
	ObservedAt  time.Time                     `json:"observed_at"`
}

func (s *WorkflowStore) workflowPullRequestObservation(run *workflowInvocationFile, snapshots []WorkflowPullRequestSnapshot) (workflowstate.FileChange, error) {
	binding := run.PullRequest

	if len(snapshots) > workflowstate.RecordLimit {
		return workflowstate.FileChange{}, fmt.Errorf("workflow PR observation bound exceeded")
	}

	observation := workflowPullRequestObservation{Schema: "chr-slice-pr-observation/v1", BindingID: binding.ID, CandidateID: binding.CandidateID, Reason: binding.Reason, Snapshots: snapshots, ObservedAt: time.Now().UTC()}
	if binding.ObservationRef != "" {
		previous, err := s.readWorkflowPullRequestObservation(*binding)
		if err != nil {
			return workflowstate.FileChange{}, err
		}

		if previous.CandidateID == observation.CandidateID && previous.Reason == observation.Reason && slices.Equal(previous.Snapshots, observation.Snapshots) {
			return workflowstate.FileChange{}, nil
		}
	}

	if binding.ObservationCount >= workflowstate.PageLimit {
		return workflowstate.FileChange{}, fmt.Errorf("workflow PR observation history bound exceeded")
	}

	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.FileChange{}, fmt.Errorf("workflow PR observation exceeds file bounds")
	}

	data = append(data, '\n')
	binding.ObservationCount++
	binding.ObservationRef = workflowUnitPullRequestObservationPath(run.ID, binding.ID, binding.ObservationCount)
	binding.ObservationHash = contenthash.Sum(data)

	return workflowstate.FileChange{Path: binding.ObservationRef, After: data}, nil
}

func workflowPullRequestObservationPath(id string, count int) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "pull-request", "observations", fmt.Sprintf("%03d.json", count)))
}

func workflowUnitPullRequestObservationPath(id, bindingID string, count int) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "pull-request", "observations", bindingID, fmt.Sprintf("%03d.json", count)))
}

func (s *WorkflowStore) readWorkflowPullRequestObservation(binding WorkflowPullRequestBinding) (workflowPullRequestObservation, error) {
	path, err := s.safePath(binding.ObservationRef)
	if err != nil {
		return workflowPullRequestObservation{}, err
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != binding.ObservationHash {
		return workflowPullRequestObservation{}, workflowstate.ErrConflict
	}

	var observation workflowPullRequestObservation
	if err := json.Unmarshal(data, &observation); err != nil {
		return observation, err
	}

	if observation.Schema != "chr-slice-pr-observation/v1" || observation.BindingID != binding.ID || observation.ObservedAt.IsZero() {
		return observation, workflowstate.ErrConflict
	}

	return observation, nil
}

func (s *WorkflowStore) workflowPullRequestDomainChanges(run *workflowInvocationFile, observation workflowSliceObservation) ([]workflowstate.FileChange, error) {
	files, err := s.workflowPullRequestTaskChanges(run, observation.tasks)
	if err != nil {
		return nil, err
	}

	changes, err := s.workflowPullRequestSetChanges(run, observation.set)

	return append(files, changes...), err
}

func (s *WorkflowStore) workflowPullRequestTaskChanges(run *workflowInvocationFile, tasks []WorkflowRecord) ([]workflowstate.FileChange, error) {
	files := []workflowstate.FileChange{}

	for _, record := range tasks {
		if record.Task.Phase == "waiting_slice" && slices.Contains([]string{ticketcore.TicketStatusReadyForReview, ticketcore.TicketStatusReadyForMerge}, record.Task.Status) {
			continue
		}

		record.Task.Phase, record.Task.Status = "waiting_slice", ticketcore.TicketStatusReadyForReview

		changes, err := s.workflowClaimRecordChanges(run, &record)
		if err != nil {
			return nil, err
		}

		files = append(files, changes...)
	}

	return files, nil
}

func (s *WorkflowStore) workflowPullRequestSetChanges(run *workflowInvocationFile, set WorkflowRecord) ([]workflowstate.FileChange, error) {
	if set.Set == nil {
		return nil, nil
	}

	for i := range set.Set.Slices {
		descriptor := &set.Set.Slices[i]
		if descriptor.ID != run.Lane.Unit.ID {
			continue
		}

		if descriptor.Branch == run.Lane.Branch && descriptor.CandidateRef == run.CandidateRef && descriptor.PullRequest == run.PullRequest.URL {
			return nil, nil
		}

		descriptor.Phase, descriptor.Branch = "reviewing", run.Lane.Branch
		descriptor.CandidateRef, descriptor.PullRequest = run.CandidateRef, run.PullRequest.URL
	}

	return s.workflowSetRecordChanges(run, &set)
}

func (s *WorkflowStore) workflowSetRecordChanges(run *workflowInvocationFile, record *WorkflowRecord) ([]workflowstate.FileChange, error) {
	record.Revision++
	record.UpdatedAt = time.Now().UTC()

	files, err := s.encodeChanges(*record)
	if err != nil {
		return nil, err
	}

	if err := guardWorkflowClaimImages(*record, files); err != nil {
		return nil, err
	}

	record.SourceHash = workflowClaimRecordHash(*record, files)
	for i := range run.SetImages {
		if run.SetImages[i].Path == record.Path {
			run.SetImages[i].Revision, run.SetImages[i].SourceHash = record.Revision, record.SourceHash
		}
	}

	return files, record.Validate()
}

func (s *WorkflowStore) commitWorkflowPullRequestFiles(run *workflowInvocationFile, files []workflowstate.FileChange, observation workflowstate.FileChange) error {
	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	if observation.Path != "" {
		files = append(files, observation)
	}

	return s.commitFiles(append(files, state))
}
