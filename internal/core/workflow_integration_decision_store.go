package core

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) readWorkflowDecisionRecord(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry) (WorkflowIntegrationDecisions, error) {
	var record WorkflowIntegrationDecisions
	if entry.DecisionRef == nil {
		return record, workflowstate.ErrConflict
	}

	if err := s.readWorkflowNativeEvidence(run.ID, *entry.DecisionRef, &record); err != nil {
		return record, err
	}

	if err := validateWorkflowDecisionRecord(record); err != nil {
		return record, err
	}

	if record.Input.InvocationID != run.ID || record.Input.EntryID != entry.ID || record.Input.Unit != entry.Unit {
		return record, workflowstate.ErrConflict
	}

	if err := s.CheckWorkflowMergerEvidence(ctx, record.Input, record.Input.AdmittedAt); err != nil {
		return record, err
	}

	return record, nil
}

func (s *WorkflowStore) recordWorkflowDecisions(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationDecisionRequest, provider WorkflowPullRequestProvider) (WorkflowIntegrationDecisions, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, request.WorkflowIntegrationBudgetGuard)
	if err != nil {
		return WorkflowIntegrationDecisions{}, err
	}

	if prior, found, err := s.workflowDecisionRequestRetry(ctx, run, entry, request, provider); found || err != nil {
		return prior, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.WorkflowIntegrationBudgetGuard); err != nil {
		return WorkflowIntegrationDecisions{}, err
	}

	if err := s.checkWorkflowDecisionAdmission(ctx, run, entry, request, provider); err != nil {
		return WorkflowIntegrationDecisions{}, err
	}

	record := WorkflowIntegrationDecisions{SchemaVersion: 1, ID: request.ID, Revision: 1, CandidateRevision: entry.StateRevision + 1, State: "waiting", Input: request.Input, Requests: slices.Clone(request.Decisions), Previous: entry.DecisionRef}
	if entry.DecisionRef != nil {
		prior, err := s.readWorkflowDecisionRecord(ctx, run, entry)
		if err != nil {
			return record, err
		}

		record.Revision = prior.Revision + 1
	}

	blockWorkflowDecisionEntry(&entry, &run)

	file, err := workflowDecisionFile(&entry, record)
	if err != nil {
		return record, err
	}

	return record, s.commitWorkflowDecisionRecord(&run, entry, file)
}

func (s *WorkflowStore) checkWorkflowDecisionAdmission(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, request WorkflowIntegrationDecisionRequest, provider WorkflowPullRequestProvider) error {
	if !workflowMergerRunnable(run, entry) || !workflowstate.ValidRequestKey(request.ID) || len(request.Decisions) == 0 {
		return workflowstate.ErrConflict
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return err
	}

	if err := s.CheckWorkflowMergerEvidence(ctx, request.Input, time.Now().UTC()); err != nil {
		return err
	}

	if err := checkWorkflowMergerEntry(run, entry, request.Input); err != nil {
		return err
	}

	if err := checkWorkflowMergerRemote(ctx, provider, request.Input); err != nil {
		return err
	}

	return validateWorkflowMergerDecisions(request.Decisions, request.Input)
}

func workflowDecisionFile(entry *WorkflowIntegrationEntry, record WorkflowIntegrationDecisions) (workflowstate.FileChange, error) {
	if err := validateWorkflowDecisionRecord(record); err != nil {
		return workflowstate.FileChange{}, err
	}

	data, err := json.Marshal(record)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.FileChange{}, workflowstate.ErrConflict
	}

	hash := contenthash.Sum(data)
	path := "run/" + record.Input.InvocationID + "/extension/chr/integration/" + entry.ID + "/decisions/" + hash + ".json"
	entry.DecisionRef = &workflowstate.EvidenceRef{Path: path, Hash: hash}

	return workflowstate.FileChange{Path: path, After: data}, nil
}

func blockWorkflowDecisionEntry(entry *WorkflowIntegrationEntry, run *workflowInvocationFile) {
	entry.State, entry.UpdatedAt = "blocked", time.Now().UTC()
	entry.Revision++
	entry.StateRevision++
	entry.Blocker = &WorkflowIntegrationBlocker{Reason: "merger_decision_required", RequiredAction: "answer_decision", ResumeState: "merging"}
	run.Status, run.Outcome = "waiting_integration", "merger_waiting_decision"
}

func (s *WorkflowStore) commitWorkflowDecisionRecord(run *workflowInvocationFile, entry WorkflowIntegrationEntry, file workflowstate.FileChange, extra ...workflowstate.FileChange) error {
	entryFile, err := workflowIntegrationEntryChange(entry, run.Integration.EntryRef.Hash)
	if err != nil {
		return err
	}

	updateWorkflowIntegrationBinding(run, WorkflowIntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: entryFile.Path, Hash: contenthash.Sum(entryFile.After)}})

	return s.commitWorkflowPullRequestFiles(run, append([]workflowstate.FileChange{entryFile}, extra...), file)
}

func (s *WorkflowStore) workflowMergerDecisionFiles(ctx context.Context, run *workflowInvocationFile, result WorkflowMergerAttempt) ([]workflowstate.FileChange, error) {
	if result.Proposal == nil || result.Proposal.Outcome != "needs_decision" || run.StopIntent != "" {
		return nil, nil
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return nil, err
	}

	record := WorkflowIntegrationDecisions{SchemaVersion: 1, ID: "decision-" + result.Input.InputHash[:24], Revision: 1, CandidateRevision: entry.StateRevision + 1, State: "waiting", Input: result.Input, Requests: result.Proposal.DecisionRequests, Previous: entry.DecisionRef}
	if entry.DecisionRef != nil {
		prior, err := s.readWorkflowDecisionRecord(ctx, *run, entry)
		if err != nil {
			return nil, err
		}

		record.Revision = prior.Revision + 1
	}

	blockWorkflowDecisionEntry(&entry, run)

	file, err := workflowDecisionFile(&entry, record)
	if err != nil {
		return nil, err
	}

	entryFile, err := workflowIntegrationEntryChange(entry, run.Integration.EntryRef.Hash)
	if err != nil {
		return nil, err
	}

	updateWorkflowIntegrationBinding(run, WorkflowIntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: entryFile.Path, Hash: contenthash.Sum(entryFile.After)}})

	return []workflowstate.FileChange{file, entryFile}, nil
}

func checkWorkflowMergerDecisionBinding(entry WorkflowIntegrationEntry, input WorkflowMergerInput) error {
	if input.OperatorDecisionRef == nil {
		return nil
	}

	if entry.DecisionRef == nil || *entry.DecisionRef != *input.OperatorDecisionRef {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowMergerOperatorEvidence(input WorkflowMergerInput) error {
	if input.OperatorDecisionRef == nil {
		if len(input.OperatorAnswers) != 0 {
			return workflowstate.ErrConflict
		}

		return nil
	}

	var record WorkflowIntegrationDecisions
	if err := s.readWorkflowNativeEvidence(input.InvocationID, *input.OperatorDecisionRef, &record); err != nil {
		return err
	}

	if err := validateWorkflowDecisionRecord(record); err != nil {
		return err
	}

	if record.State != "resolved" || !reflect.DeepEqual(record.Answers, input.OperatorAnswers) {
		return workflowstate.ErrConflict
	}

	expected, actual := workflowMergerFence(record.Input), workflowMergerFence(input)
	expected.CandidateRevision = record.CandidateRevision

	expected.OwnedWorkspaceID = actual.OwnedWorkspaceID
	if expected != actual || !workflowDecisionContextEqual(record.Input, input) {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowDecisionContextEqual(original, next WorkflowMergerInput) bool {
	return original.SourceManifestRef == next.SourceManifestRef && original.ValidationContractRef == next.ValidationContractRef &&
		reflect.DeepEqual(original.AllowedPaths, next.AllowedPaths) && reflect.DeepEqual(original.ProtectedPaths, next.ProtectedPaths) &&
		reflect.DeepEqual(original.ValidationCommands, next.ValidationCommands) && reflect.DeepEqual(original.Requirements, next.Requirements) &&
		reflect.DeepEqual(original.Conflicts, next.Conflicts) && reflect.DeepEqual(original.IntentRefs, next.IntentRefs) &&
		reflect.DeepEqual(original.TargetChangeRefs, next.TargetChangeRefs) && reflect.DeepEqual(original.FindingRefs, next.FindingRefs)
}

func (s *WorkflowStore) guardWorkflowDecisionTransition(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, request WorkflowIntegrationTransition) error {
	if entry.DecisionRef == nil || slices.Contains([]string{"blocked", "paused", "cancelled", "queued"}, request.State) {
		return nil
	}

	record, err := s.readWorkflowDecisionRecord(ctx, run, entry)
	if err != nil {
		return err
	}

	if record.State != "resolved" && entry.ClaimGeneration == record.Input.ClaimGeneration {
		return ErrWorkflowOwned
	}

	return nil
}

func (s *WorkflowStore) workflowDecisionRequestRetry(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, request WorkflowIntegrationDecisionRequest, provider WorkflowPullRequestProvider) (WorkflowIntegrationDecisions, bool, error) {
	if entry.DecisionRef == nil {
		return WorkflowIntegrationDecisions{}, false, nil
	}

	prior, err := s.readWorkflowDecisionRecord(ctx, run, entry)
	if err != nil {
		return prior, false, err
	}

	if prior.ID != request.ID {
		return prior, false, nil
	}

	if !reflect.DeepEqual(prior.Input, request.Input) || !reflect.DeepEqual(prior.Requests, request.Decisions) {
		return prior, false, workflowstate.ErrConflict
	}

	return prior, true, s.checkWorkflowDecisionCurrent(ctx, run, entry, prior, provider)
}
