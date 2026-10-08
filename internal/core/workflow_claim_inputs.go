package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *WorkflowStore) nextWorkflowInvocationNumber(ctx context.Context) (int, error) {
	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return 0, err
	}

	number := 1
	for _, layout := range layouts {
		number = max(number, nextOpsArtifactNumber(layout.Root()))
	}

	return number, nil
}

func (s *WorkflowStore) freezeWorkflowSets(ctx context.Context, run *workflowInvocationFile) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	ids := map[string]bool{}
	for _, member := range run.Members {
		ids[member.Admission.SetID] = true
	}

	for _, record := range snapshot.Records {
		if record.Set != nil && ids[record.ID] {
			run.SetImages = append(run.SetImages, WorkflowInputImage{Path: record.Path, Revision: record.Revision, SourceHash: record.SourceHash, UseCase: record.Set.UseCase})
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowInputs(ctx context.Context, run workflowInvocationFile) error {
	for _, input := range run.Members {
		if err := ctx.Err(); err != nil {
			return err
		}

		record, err := s.readRecord(input.Admission.Path)
		if err != nil {
			return err
		}

		if !workflowMemberImageMatches(record, input, run.WorkflowIdentity) {
			return fmt.Errorf("%w: frozen task image changed", ErrWorkflowConflict)
		}
	}

	return s.checkWorkflowSetImages(run)
}

func workflowMemberImageMatches(record WorkflowRecord, input WorkflowMember, identity WorkflowIdentity) bool {
	if record.ProjectID != identity.ProjectID || record.Namespace != identity.Namespace || record.Task == nil {
		return false
	}

	return record.Task.Key == input.Admission.Key && record.Task.ContractHash == input.Admission.ContractHash && record.Revision == input.ExpectedRevision && record.SourceHash == input.ExpectedSourceHash
}

func (s *WorkflowStore) checkWorkflowSetImages(run workflowInvocationFile) error {
	for _, image := range run.SetImages {
		record, err := s.readRecord(image.Path)
		if err != nil {
			return err
		}

		if record.Set == nil || record.Revision != image.Revision || record.SourceHash != image.SourceHash || record.ProjectID != run.ProjectID || record.Namespace != run.Namespace {
			return fmt.Errorf("%w: frozen set image changed", ErrWorkflowConflict)
		}
	}

	return nil
}

func validateWorkflowClaimContract(record WorkflowRecord) error {
	contract := GeneratedWorkflowTask{}
	if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err != nil {
		return fmt.Errorf("task requires a complete execution contract: %w", err)
	}

	if err := validateWorkflowPlanTask(contract); err != nil {
		return err
	}

	encoded, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	hash := workflowHash(append([]byte(record.Task.SetID+"\x00"+record.Task.SliceID+"\x00"), encoded...))
	if hash != record.Task.ContractHash {
		return fmt.Errorf("%w: execution contract image differs", ErrWorkflowConflict)
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowTarget(ctx context.Context, target WorkflowPreviewTarget) error {
	reader := workflowProofReader{store: s}

	commit, err := reader.gitOutput(ctx, "rev-parse", "--verify", target.Commit+"^{commit}")
	if err != nil || commit != target.Commit {
		return fmt.Errorf("workflow target commit is unavailable: %s", target.Commit)
	}

	return nil
}

func guardWorkflowClaimImages(record WorkflowRecord, files []workflowFileChange) error {
	observed := strings.Split(record.SourceHash, ":")
	if len(files) != len(observed) {
		return ErrWorkflowConflict
	}

	for i, file := range files {
		if file.Before != observed[i] {
			return fmt.Errorf("%w: claimed task image changed before journaling", ErrWorkflowConflict)
		}
	}

	return nil
}
