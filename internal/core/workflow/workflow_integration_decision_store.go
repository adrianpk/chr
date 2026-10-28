package workflow

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) readWorkflowDecisionRecord(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry) (IntegrationDecisions, error) {
	var record IntegrationDecisions
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

func (s *Store) recordWorkflowDecisions(ctx context.Context, run workflowInvocationFile, request IntegrationDecisionRequest, provider PullRequestProvider) (IntegrationDecisions, error) {
	budgetObservation, err := s.currentWorkflowIntegrationBudget(ctx, run, request.IntegrationBudgetGuard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	if err != nil {
		return IntegrationDecisions{}, err
	}

	if decisionRetry, err := s.workflowDecisionRequestRetry(ctx, run, entry, request, provider); decisionRetry.Found || err != nil {
		return decisionRetry.Record, err
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.IntegrationBudgetGuard); err != nil {
		return IntegrationDecisions{}, err
	}

	if err := s.checkWorkflowDecisionAdmission(ctx, run, entry, request, provider); err != nil {
		return IntegrationDecisions{}, err
	}

	record := IntegrationDecisions{SchemaVersion: 1, ID: request.ID, Revision: 1, CandidateRevision: entry.StateRevision + 1, State: "waiting", Input: request.Input, Requests: slices.Clone(request.Decisions), Previous: entry.DecisionRef}
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

func (s *Store) checkWorkflowDecisionAdmission(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, request IntegrationDecisionRequest, provider PullRequestProvider) error {
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

	return workflowmerge.CheckDecisions(request.Decisions, request.Input)
}

func workflowDecisionFile(entry *IntegrationEntry, record IntegrationDecisions) (workflowstate.FileChange, error) {
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

func blockWorkflowDecisionEntry(entry *IntegrationEntry, run *workflowInvocationFile) {
	entry.State, entry.UpdatedAt = "blocked", time.Now().UTC()
	entry.Revision++
	entry.StateRevision++
	entry.Blocker = &IntegrationBlocker{Reason: "merger_decision_required", RequiredAction: "answer_decision", ResumeState: "merging"}
	run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "merger_waiting_decision"
}

func (s *Store) commitWorkflowDecisionRecord(run *workflowInvocationFile, entry IntegrationEntry, file workflowstate.FileChange, extra ...workflowstate.FileChange) error {
	entryFile, err := workflowIntegrationEntryChange(entry, run.Integration.EntryRef.Hash)
	if err != nil {
		return err
	}

	updateWorkflowIntegrationBinding(run, workflowstate.IntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: entryFile.Path, Hash: contenthash.Sum(entryFile.After)}})

	return s.commitWorkflowPullRequestFiles(run, append([]workflowstate.FileChange{entryFile}, extra...), file)
}

func (s *Store) workflowMergerDecisionFiles(ctx context.Context, run *workflowInvocationFile, result MergerAttempt) ([]workflowstate.FileChange, error) {
	if result.Proposal == nil || result.Proposal.Outcome != "needs_decision" || run.StopIntent != "" {
		return nil, nil
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return nil, err
	}

	record := IntegrationDecisions{SchemaVersion: 1, ID: "decision-" + result.Input.InputHash[:24], Revision: 1, CandidateRevision: entry.StateRevision + 1, State: "waiting", Input: result.Input, Requests: result.Proposal.DecisionRequests, Previous: entry.DecisionRef}
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

	updateWorkflowIntegrationBinding(run, workflowstate.IntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: workflowstate.EvidenceRef{Path: entryFile.Path, Hash: contenthash.Sum(entryFile.After)}})

	return []workflowstate.FileChange{file, entryFile}, nil
}

func checkWorkflowMergerDecisionBinding(entry IntegrationEntry, input workflowstate.MergerInput) error {
	if input.OperatorDecisionRef == nil {
		return nil
	}

	if entry.DecisionRef == nil || *entry.DecisionRef != *input.OperatorDecisionRef {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowMergerOperatorEvidence(input workflowstate.MergerInput) error {
	if input.OperatorDecisionRef == nil {
		if len(input.OperatorAnswers) != 0 {
			return workflowstate.ErrConflict
		}

		return nil
	}

	var record IntegrationDecisions
	if err := s.readWorkflowNativeEvidence(input.InvocationID, *input.OperatorDecisionRef, &record); err != nil {
		return err
	}

	if err := validateWorkflowDecisionRecord(record); err != nil {
		return err
	}

	if record.State != "resolved" || !reflect.DeepEqual(record.Answers, input.OperatorAnswers) {
		return workflowstate.ErrConflict
	}

	expected, actual := workflowmerge.InputFence(record.Input), workflowmerge.InputFence(input)
	expected.CandidateRevision = record.CandidateRevision

	expected.OwnedWorkspaceID = actual.OwnedWorkspaceID
	if expected != actual || !workflowDecisionContextEqual(record.Input, input) {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowDecisionContextEqual(original, next workflowstate.MergerInput) bool {
	return original.SourceManifestRef == next.SourceManifestRef && original.ValidationContractRef == next.ValidationContractRef &&
		reflect.DeepEqual(original.AllowedPaths, next.AllowedPaths) && reflect.DeepEqual(original.ProtectedPaths, next.ProtectedPaths) &&
		reflect.DeepEqual(original.ValidationCommands, next.ValidationCommands) && reflect.DeepEqual(original.Requirements, next.Requirements) &&
		reflect.DeepEqual(original.Conflicts, next.Conflicts) && reflect.DeepEqual(original.IntentRefs, next.IntentRefs) &&
		reflect.DeepEqual(original.TargetChangeRefs, next.TargetChangeRefs) && reflect.DeepEqual(original.FindingRefs, next.FindingRefs)
}

func (s *Store) guardWorkflowDecisionTransition(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, request IntegrationTransition) error {
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

func (s *Store) workflowDecisionRequestRetry(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, request IntegrationDecisionRequest, provider PullRequestProvider) (workflowDecisionRetry, error) {
	if entry.DecisionRef == nil {
		return workflowDecisionRetry{Record: IntegrationDecisions{}, Found: false}, nil
	}

	prior, err := s.readWorkflowDecisionRecord(ctx, run, entry)
	if err != nil {
		return workflowDecisionRetry{Record: prior, Found: false}, err
	}

	if prior.ID != request.ID {
		return workflowDecisionRetry{Record: prior, Found: false}, nil
	}

	if !reflect.DeepEqual(prior.Input, request.Input) || !reflect.DeepEqual(prior.Requests, request.Decisions) {
		return workflowDecisionRetry{Record: prior, Found: false}, workflowstate.ErrConflict
	}

	return workflowDecisionRetry{Record: prior, Found: true}, s.checkWorkflowDecisionCurrent(ctx, run, entry, prior, provider)
}
