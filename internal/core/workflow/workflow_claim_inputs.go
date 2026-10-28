package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) nextWorkflowInvocationNumber(ctx context.Context) (int, error) {
	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return 0, err
	}

	number := 1
	for _, layout := range layouts {
		number = max(number, opscore.NextArtifactNumber(layout.Root()))
	}

	return number, nil
}

func (s *Store) freezeWorkflowSets(ctx context.Context, run *workflowInvocationFile) error {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return err
	}

	ids := map[string]bool{}
	for _, member := range run.Members {
		ids[member.Admission.SetID] = true
	}

	for _, record := range selection.sets {
		if workflowFrozenSetMatches(record, run.Selection.Selector, ids) {
			run.SetImages = append(run.SetImages, InputImage{Path: record.Path, Revision: record.Revision, SourceHash: record.SourceHash, UseCase: record.Set.UseCase})
		}
	}

	slices.SortFunc(run.SetImages, func(a, b InputImage) int {
		return selection.compareSets(workflowInvocationSetID(a), workflowInvocationSetID(b), run.SetOrderRule)
	})

	if len(run.Members) == 0 && len(run.SetImages) == 0 {
		return fmt.Errorf("selected workflow has no prepared unfinished tasks or sets")
	}

	if workflowInvocationNeedsClosing(run.Invocation) {
		run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "set_closure_required"
	}

	return nil
}

func workflowFrozenSetMatches(set workflowstate.Record, selector workflowstate.Selector, memberSets map[string]bool) bool {
	if selector.Extent == "slice_set" {
		return set.ID == selector.Identity && set.Set.Status != "completed"
	}

	if selector.Extent == "use_case" {
		return set.Set.UseCase == selector.Identity && set.Set.Status != "completed"
	}

	return memberSets[set.ID]
}

func workflowInvocationNeedsClosing(run Invocation) bool {
	if run.Selection.Selector.Extent != "slice_set" && run.Selection.Selector.Extent != "use_case" {
		return false
	}

	next := workflowNextClosingSet(run)

	return next != "" && (run.Cursor == len(run.Members) || run.Members[run.Cursor].Admission.SetID != next)
}

func (s *Store) checkWorkflowInputs(ctx context.Context, run workflowInvocationFile) error {
	return s.checkWorkflowInputsPrepared(ctx, run, (*Store).recoverTransition)
}

func (s *Store) checkWorkflowInputsPrepared(ctx context.Context, run workflowInvocationFile, prepare func(*Store) error) error {
	if _, err := s.checkWorkflowCoordinator(ctx); err != nil {
		return err
	}

	if err := s.checkWorkflowFrontierOwnershipPrepared(ctx, run, prepare); err != nil {
		return err
	}

	if err := s.checkWorkflowContinuationCheckpoint(run); err != nil {
		return err
	}

	if err := s.checkWorkflowContinuationEvidence(ctx, run); err != nil {
		return err
	}

	for _, input := range run.Members {
		if err := ctx.Err(); err != nil {
			return err
		}

		record, err := s.readRecord(input.Admission.Path)
		if err != nil {
			return err
		}

		if !workflowMemberImageMatches(record, input, run.Identity) {
			return fmt.Errorf("%w: frozen task image changed", workflowstate.ErrConflict)
		}
	}

	return s.checkWorkflowSetImages(run)
}

func workflowMemberImageMatches(record workflowstate.Record, input Member, identity workflowstate.Identity) bool {
	if record.ProjectID != identity.ProjectID || record.Namespace != identity.Namespace || record.Task == nil {
		return false
	}

	return record.Task.Key == input.Admission.Key && record.Task.ContractHash == input.Admission.ContractHash && record.Task.SetID == input.Admission.SetID && record.Task.SliceID == input.Admission.SliceID && record.Task.Order == input.Admission.TaskOrder && record.Revision == input.ExpectedRevision && record.SourceHash == input.ExpectedSourceHash
}

func (s *Store) checkWorkflowSetImages(run workflowInvocationFile) error {
	for _, image := range run.SetImages {
		record, err := s.readRecord(image.Path)
		if err != nil {
			return err
		}

		if record.Set == nil || record.Revision != image.Revision || record.SourceHash != image.SourceHash || record.ProjectID != run.ProjectID || record.Namespace != run.Namespace {
			return fmt.Errorf("%w: frozen set image changed", workflowstate.ErrConflict)
		}

		if err := s.checkWorkflowSetAmendmentEvidence(record); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowClaimContract(record workflowstate.Record) error {
	contract := generationcore.GeneratedWorkflowTask{}
	if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err != nil {
		return fmt.Errorf("task requires a complete execution contract: %w", err)
	}

	if err := generationcore.ValidateWorkflowPlanTask(contract); err != nil {
		return err
	}

	encoded, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	hash := contenthash.Sum(append([]byte(record.Task.SetID+"\x00"+record.Task.SliceID+"\x00"), encoded...))
	if hash != record.Task.ContractHash {
		return fmt.Errorf("%w: execution contract image differs", workflowstate.ErrConflict)
	}

	return nil
}

func (s *Store) checkWorkflowTarget(ctx context.Context, target workflowstate.PreviewTarget) error {
	reader := workflowProofReader{store: s}

	commit, err := reader.gitOutput(ctx, "rev-parse", "--verify", target.Commit+"^{commit}")
	if err != nil || commit != target.Commit {
		return fmt.Errorf("workflow target commit is unavailable: %s", target.Commit)
	}

	return s.checkWorkflowCoordinatorTarget(ctx, target)
}

func guardWorkflowClaimImages(record workflowstate.Record, files []workflowstate.FileChange) error {
	observed := strings.Split(record.SourceHash, ":")
	if len(files) != len(observed) {
		return workflowstate.ErrConflict
	}

	for i, file := range files {
		if file.Before != observed[i] {
			return fmt.Errorf("%w: claimed task image changed before journaling", workflowstate.ErrConflict)
		}
	}

	return nil
}
