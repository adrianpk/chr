package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowPreparationPath(id string, round int) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "preparation", fmt.Sprintf("%03d.json", round)))
}

func (s *Store) readWorkflowPreparation(run workflowInvocationFile) (workflowstate.IntegrationCandidate, error) {
	var candidate workflowstate.IntegrationCandidate
	if run.Preparation == nil || run.Preparation.ResultRef == "" {
		return candidate, workflowstate.ErrConflict
	}

	data, err := s.readWorkflowReviewImage(run.Preparation.ResultRef, run.Preparation.ResultHash)
	if err != nil {
		return candidate, err
	}

	if err := json.Unmarshal(data, &candidate); err != nil {
		return candidate, err
	}

	if err := validateWorkflowPreparationCandidate(run, candidate); err != nil {
		return candidate, err
	}

	if err := s.checkWorkflowNativeCandidate(candidate); err != nil {
		return candidate, err
	}

	if err := s.checkWorkflowRetainedPreparationInput(run, candidate); err != nil {
		return candidate, err
	}

	return candidate, nil
}

func (s *Store) checkWorkflowRetainedPreparationInput(run workflowInvocationFile, candidate workflowstate.IntegrationCandidate) error {
	result := candidate.Validation

	data, err := s.readWorkflowReviewImage(result.InputRef, result.InputHash)
	if err != nil {
		return err
	}

	var input TaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	if !workflowPreparationInputMatches(input, candidate) {
		return workflowstate.ErrConflict
	}

	expected, err := json.Marshal(candidate.Input)
	if err != nil {
		return err
	}

	actual, err := json.Marshal(input.Preparation)
	if err != nil || contenthash.Sum(actual) != contenthash.Sum(expected) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowPreparationGate(input, result)
}

func workflowPreparationInputMatches(input TaskInput, candidate workflowstate.IntegrationCandidate) bool {
	result := candidate.Validation

	return input.Preparation != nil && input.Lane == candidate.Lane && input.AttemptID == result.AttemptID && input.Generation == result.Generation && input.ContractHash == candidate.Input.Source.Slice.ContractHash
}

func validateWorkflowPreparationCandidate(run workflowInvocationFile, candidate workflowstate.IntegrationCandidate) error {
	b := run.Preparation

	if err := validateWorkflowCandidateSchema(candidate); err != nil {
		return err
	}

	if candidate.InvocationID != run.ID || candidate.Round != b.Round || candidate.Input.Source.Unit != b.Unit || !validWorkflowLane(candidate.Lane) {
		return workflowstate.ErrConflict
	}

	hash, err := workflowReviewInputHash(candidate.Input.Source)
	if err != nil || hash != b.SourceHash {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowPublicationCapability(candidate.Input.Source, candidate.Input.Capability); err != nil {
		return err
	}

	if err := validateWorkflowPreparationProofBinding(candidate); err != nil {
		return err
	}

	return validateWorkflowPreparationResult(run, candidate)
}

func validateWorkflowPreparationProofBinding(candidate workflowstate.IntegrationCandidate) error {
	proof, source := candidate.Input.Git, candidate.Input.Source

	capability := candidate.Input.Capability
	if err := checkWorkflowPreparationSigning(proof, capability.RequiredSignatures, capability.SigningProfile); err != nil {
		return err
	}

	if proof.SourceCommit != source.PullRequest.HeadCommit || proof.BaseCommit != source.TargetCommit || proof.Method != source.Policy.Method || proof.Repository != source.PullRequest.Repository || proof.TargetBranch != source.PullRequest.BaseBranch {
		return workflowstate.ErrConflict
	}

	return validateWorkflowPreparationLaneProof(candidate)
}

func validateWorkflowPreparationLaneProof(candidate workflowstate.IntegrationCandidate) error {
	proof := candidate.Input.Git
	if proof.SourceDirectory != "" || proof.ObservedTarget != proof.BaseCommit || candidate.Lane.HeadCommit != proof.ResultCommit || candidate.Lane.BaseCommit != proof.BaseCommit {
		return workflowstate.ErrConflict
	}

	return validateWorkflowIntegrationProof(proof.IntegrationQuery, proof)
}

func validateWorkflowPreparationResult(run workflowInvocationFile, candidate workflowstate.IntegrationCandidate) error {
	result := candidate.Validation
	if !result.Successful || result.Schema != "chr-integration-validation/v1" || result.Outcome != "integration_validated" || result.Commit != candidate.Input.Git.ResultCommit || result.Tree != candidate.Input.Git.PreparedTree || result.FinishedAt.IsZero() {
		return workflowstate.ErrConflict
	}

	for _, attempt := range run.Attempts {
		if workflowPreparationAttemptMatches(attempt, result, run.Preparation.ResultHash) {
			return nil
		}
	}

	return workflowstate.ErrConflict
}

func workflowPreparationAttemptMatches(attempt Attempt, result workflowstate.TaskResult, hash string) bool {
	return attempt.ID == result.AttemptID && attempt.Purpose == workflowIntegrationValidationPurpose && attempt.Outcome == "prepared" && attempt.ResultHash == hash && attempt.InputRef == result.InputRef && attempt.InputHash == result.InputHash
}

func (s *Store) persistWorkflowPreparation(run *workflowInvocationFile, record Record, candidate workflowstate.IntegrationCandidate) error {
	if err := s.checkWorkflowNativeCandidate(candidate); err != nil {
		return err
	}

	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return fmt.Errorf("prepared candidate exceeds file bounds")
	}

	data = append(data, '\n')
	ref := workflowPreparationPath(run.ID, candidate.Round)
	run.Preparation.State, run.Preparation.ResultRef, run.Preparation.ResultHash = "validated", ref, contenthash.Sum(data)
	attempt := &run.Attempts[len(run.Attempts)-1]
	attempt.Outcome, attempt.FinishedAt = "prepared", time.Now().UTC()
	attempt.ResultRef, attempt.ResultHash = workflowAttemptFile(run.ID, attempt.ID, "result.json"), contenthash.Sum(data)
	record.Task.AttemptID = ""

	run.Status, run.Outcome, run.Blocker = "waiting_review", "integration_validated", nil

	run.CandidateRepair = nil
	if run.Review != nil {
		run.Review.State = "stale"
	}

	files, err := s.workflowClaimRecordChanges(run, &record)
	if err != nil {
		return err
	}

	files = append(files, workflowstate.FileChange{Path: attempt.ResultRef, After: data})

	if candidate.Native != nil {
		file, err := s.workflowNativePreparedEntryChange(run, candidate)
		if err != nil {
			return err
		}

		files = append(files, file)
	}

	return s.commitWorkflowPullRequestFiles(run, files, workflowstate.FileChange{Path: ref, After: data})
}

func validateWorkflowPreparationGate(input TaskInput, result workflowstate.TaskResult) error {
	gate := input.Contract.Validation
	if !gate.Required {
		if result.WaiverReason != gate.WaiverReason || len(result.Validation) != 0 {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if len(result.Validation) != len(gate.Commands) {
		return workflowstate.ErrConflict
	}

	for i, command := range gate.Commands {
		observed := result.Validation[i]
		if !slices.Equal(command, observed.Command) || observed.ExitCode != 0 || observed.Truncated {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
