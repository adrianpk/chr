package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowReviewEvidence struct {
	Schema       string                       `json:"schema"`
	Origin       string                       `json:"origin"`
	Decision     workflowstate.ReviewDecision `json:"decision"`
	PreviousRef  string                       `json:"previous_ref"`
	PreviousHash string                       `json:"previous_hash"`
	RecordedAt   time.Time                    `json:"recorded_at"`
}

func workflowReviewPath(id string, round int, name string) string {
	return filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "review", fmt.Sprintf("%03d", round), name))
}

func (s *Store) persistWorkflowReviewInput(run *workflowInvocationFile, observation workflowSliceObservation, input workflowstate.ReviewInput) error {
	if run.ReviewCount >= workflowstate.PageLimit {
		return fmt.Errorf("workflow review history budget exhausted")
	}

	data, err := workflowreview.InputBytes(input)
	if err != nil {
		return err
	}

	run.ReviewCount++
	ref := workflowReviewPath(run.ID, run.ReviewCount, "input.json")
	run.Review = &workflowstate.ReviewBinding{Round: run.ReviewCount, Unit: input.Unit, State: "pending", InputRef: ref, InputHash: contenthash.Sum(data)}
	run.Status, run.Outcome, run.Blocker = workflowstate.InvocationWaitingReview, "slice_review_pending", nil

	files, err := s.workflowReviewDomainChanges(run, observation, false)
	if err != nil {
		return err
	}

	return s.commitWorkflowPullRequestFiles(run, files, workflowstate.FileChange{Path: ref, After: data})
}

func (s *Store) persistWorkflowReviewDecision(run *workflowInvocationFile, observation workflowSliceObservation, decision workflowstate.ReviewDecision, state string, extra ...workflowstate.FileChange) error {
	if run.Review.DecisionCount >= 2 {
		return fmt.Errorf("workflow review decision history budget exhausted")
	}

	origin, err := s.workflowReviewOrigin(*run, state)
	if err != nil {
		return err
	}

	evidence := workflowReviewEvidence{Schema: "chr-workflow-review/v1", Origin: origin, Decision: decision, PreviousRef: run.Review.ResultRef, PreviousHash: run.Review.ResultHash, RecordedAt: time.Now().UTC()}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return fmt.Errorf("review decision exceeds file bounds")
	}

	data = append(data, '\n')
	run.Review.DecisionCount++
	ref := workflowReviewPath(run.ID, run.Review.Round, fmt.Sprintf("decision-%d.json", run.Review.DecisionCount))
	run.Review.ResultRef, run.Review.ResultHash, run.Review.State = ref, contenthash.Sum(data), state
	run.Review.RepairTaskKey = decision.RepairTaskKey

	run.Status, run.Outcome, run.Blocker = workflowstate.InvocationWaitingReview, "slice_review_"+state, nil
	if state == "authorized" {
		run.Status = workflowstate.InvocationWaitingIntegration
	}

	if state == "insufficient_evidence" || state == "human_required" {
		run.Blocker = &Blocker{Dependency: run.Review.Unit, Reason: "review_" + state, NextActionOwner: "workflow_operator"}
	}

	files, err := s.workflowReviewDomainChanges(run, observation, state == "authorized")
	if err != nil {
		return err
	}

	files = append(files, workflowReviewAttemptResult(run, data)...)
	files = append(files, extra...)

	return s.commitWorkflowPullRequestFiles(run, files, workflowstate.FileChange{Path: ref, After: data})
}

func workflowReviewAttemptResult(run *workflowInvocationFile, data []byte) []workflowstate.FileChange {
	if len(run.Attempts) == 0 {
		return nil
	}

	attempt := &run.Attempts[len(run.Attempts)-1]
	if attempt.Purpose != workflowSliceReviewPurpose || attempt.Outcome != "reviewed" || attempt.ResultRef != "" {
		return nil
	}

	attempt.ResultRef = workflowAttemptFile(run.ID, attempt.ID, "result.json")
	attempt.ResultHash = contenthash.Sum(data)

	return []workflowstate.FileChange{{Path: attempt.ResultRef, After: data}}
}

func (s *Store) workflowReviewDomainChanges(run *workflowInvocationFile, observation workflowSliceObservation, authorized bool) ([]workflowstate.FileChange, error) {
	status, phase := string(ticketcore.StatusReadyForReview), "reviewing"
	if authorized {
		status, phase = string(ticketcore.StatusReadyForMerge), "ready_for_merge"
	}

	files := []workflowstate.FileChange{}

	for _, record := range observation.tasks {
		if record.Task.Status == status && record.Task.AttemptID == "" {
			continue
		}

		record.Task.Status, record.Task.Phase, record.Task.AttemptID = status, "waiting_slice", ""

		changes, err := s.workflowClaimRecordChanges(run, &record)
		if err != nil {
			return nil, err
		}

		files = append(files, changes...)
	}

	set := observation.set
	if set.Set == nil {
		return files, nil
	}

	for i := range set.Set.Slices {
		if set.Set.Slices[i].ID != run.Lane.Unit.ID || set.Set.Slices[i].Phase == phase {
			continue
		}

		set.Set.Slices[i].Phase = phase
		changes, err := s.workflowSetRecordChanges(run, &set)

		return append(files, changes...), err
	}

	return files, nil
}

func (s *Store) readWorkflowReview(run workflowInvocationFile) (workflowReviewRecord, error) {
	var input workflowstate.ReviewInput
	if run.Review == nil {
		return workflowReviewRecord{Input: input, Decision: workflowstate.ReviewDecision{}}, workflowstate.ErrConflict
	}

	data, err := s.readWorkflowReviewImage(run.Review.InputRef, run.Review.InputHash)
	if err != nil {
		return workflowReviewRecord{Input: input, Decision: workflowstate.ReviewDecision{}}, err
	}

	if err := json.Unmarshal(data, &input); err != nil {
		return workflowReviewRecord{Input: input, Decision: workflowstate.ReviewDecision{}}, err
	}

	if input.Schema != "chr-workflow-review-input/v1" || input.InvocationID != run.ID || input.Unit != run.Review.Unit || !workflowstate.ValidCommitID(input.TargetCommit) {
		return workflowReviewRecord{Input: input, Decision: workflowstate.ReviewDecision{}}, workflowstate.ErrConflict
	}

	if err := workflowreview.CheckPolicy(input.Policy); err != nil {
		return workflowReviewRecord{Input: input, Decision: workflowstate.ReviewDecision{}}, err
	}

	if err := validateWorkflowNativeReviewInput(input); err != nil {
		return workflowReviewRecord{Input: input, Decision: workflowstate.ReviewDecision{}}, err
	}

	decision, err := s.readWorkflowReviewDecisions(run, input)

	return workflowReviewRecord{Input: input, Decision: decision}, err
}

func (s *Store) readWorkflowReviewImage(ref, hash string) ([]byte, error) {
	path, err := s.safePath(ref)
	if err != nil {
		return nil, err
	}

	data, err := opsfile.Read(path)
	if err != nil || contenthash.Sum(data) != hash {
		return nil, workflowstate.ErrConflict
	}

	return data, nil
}

func (s *Store) workflowReviewOrigin(run workflowInvocationFile, state string) (string, error) {
	if state != "authorized" {
		return "agentic", nil
	}

	reviewRecord, err := s.readWorkflowReview(run)

	input := reviewRecord.Input

	if err != nil {
		return "", err
	}

	if input.Policy.Mode == workflowstate.ReviewTrustedAutoMerge {
		return "trusted", nil
	}

	if run.Review.State == "waiting_approval" || run.Review.State == "human_required" {
		return "human", nil
	}

	return "agentic", nil
}
