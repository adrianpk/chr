package core

import (
	"context"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) workflowAgentBudgetCreation(ctx context.Context, run *workflowInvocationFile) (workflowstate.FileChange, error) {
	if run.Policy.AgentTokenBudget == 0 {
		return workflowstate.FileChange{}, nil
	}

	limit := run.Policy.AgentTokenBudget
	limits := workflowIntegrationAdmissionLimits(run.Policy)
	budget := WorkflowIntegrationBudget{SchemaVersion: 1, InvocationID: run.ID, AdmissionTokenLimit: &limit, TokenLimit: &limit, AdmissionLimits: limits, LimitsSnapshot: limits}

	return s.workflowIntegrationBudgetChange(ctx, run, &budget)
}

func (s *WorkflowStore) reserveWorkflowAgentInput(ctx context.Context, run *workflowInvocationFile, input *WorkflowTaskInput, harness *WorkflowAgentHarness) (*WorkflowIntegrationBudget, error) {
	if harness == nil {
		return nil, s.checkWorkflowUnmeteredAgent(*run)
	}

	role := "executor"
	if input.Review != nil {
		role = "reviewer"
	}

	certificate, err := s.checkWorkflowAgentHarness(ctx, run.ID, role, *harness)
	if err != nil {
		return nil, err
	}

	budget, err := s.workflowAgentBudget(ctx, *run, input.Lane.Unit)
	if err != nil {
		return nil, err
	}

	maximum, err := workflowTokenSum(harness.InputTokens, harness.OutputTokens)
	if err != nil {
		return nil, err
	}

	allowance := WorkflowTokenAllowance{Maximum: maximum, Input: harness.InputTokens, Output: harness.OutputTokens, CapabilityRevision: certificate.CapabilityRevision, WholeCall: true, Deadline: run.Attempts[len(run.Attempts)-1].LeaseUntil}

	r := WorkflowIntegrationReservation{ID: "agent-" + input.AttemptID, Kind: "agent_call", Role: role, EntryID: workflowIntegrationEntryID(*run, input.Lane.Unit), Unit: input.Lane.Unit, ClaimGeneration: input.Generation, AttemptID: input.AttemptID, Allowance: allowance, State: "reserved"}
	if err := validateWorkflowIntegrationReservation(r); err != nil {
		return nil, err
	}

	if err := checkWorkflowIntegrationCallDeadline(run.Policy, budget, r); err != nil {
		return nil, err
	}

	if err := checkWorkflowIntegrationRoleBudget(*run, budget, r); err != nil {
		return nil, err
	}

	if len(budget.Reservations) >= workflowstate.RecordLimit {
		return nil, workflowstate.ErrConflict
	}

	budget.Reservations = append(budget.Reservations, r)
	input.AgentCall = &WorkflowAgentCall{ReservationID: r.ID, Role: role, HarnessRef: harness.Certificate, Allowance: allowance}

	return &budget, nil
}

func (s *WorkflowStore) workflowAgentBudget(ctx context.Context, run workflowInvocationFile, unit workflowstate.Ref) (WorkflowIntegrationBudget, error) {
	budget, err := s.workflowIntegrationBudgetAdmission(ctx, run, WorkflowIntegrationEntry{Unit: unit}, nil)
	if err != nil {
		return budget, err
	}

	usage, err := s.workflowIntegrationUsage(ctx, run)
	if err != nil {
		return budget, err
	}

	if err := checkWorkflowIntegrationUsage(budget.Usage, usage); err != nil {
		return budget, err
	}

	budget.Usage = usage

	return budget, nil
}

func (s *WorkflowStore) checkWorkflowUnmeteredAgent(run workflowInvocationFile) error {
	if run.Policy.AgentTokenBudget > 0 {
		return ErrWorkflowIntegrationBudget
	}

	if run.IntegrationBudget == nil {
		return nil
	}

	budget, err := s.readWorkflowIntegrationBudget(run)
	if err != nil {
		return err
	}

	if budget.TokenLimit != nil {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func workflowAgentReservation(budget *WorkflowIntegrationBudget, input WorkflowTaskInput) (*WorkflowIntegrationReservation, error) {
	call := input.AgentCall
	if call == nil {
		return nil, workflowstate.ErrConflict
	}

	index := slices.IndexFunc(budget.Reservations, func(r WorkflowIntegrationReservation) bool {
		return r.ID == call.ReservationID && r.AttemptID == input.AttemptID && r.ClaimGeneration == input.Generation && r.Role == call.Role && r.Unit == input.Lane.Unit && r.Allowance == call.Allowance
	})
	if index < 0 {
		return nil, workflowstate.ErrConflict
	}

	return &budget.Reservations[index], nil
}

func (s *WorkflowStore) encodeWorkflowAgentInput(ctx context.Context, run *workflowInvocationFile, input WorkflowTaskInput, harness *WorkflowAgentHarness) (workflowPreparedTask, workflowstate.FileChange, workflowstate.FileChange, error) {
	budget, err := s.reserveWorkflowAgentInput(ctx, run, &input, harness)
	if err != nil {
		return workflowPreparedTask{}, workflowstate.FileChange{}, workflowstate.FileChange{}, err
	}

	prepared, file, err := encodeWorkflowTaskInput(run, input)
	if err != nil || budget == nil {
		return prepared, file, workflowstate.FileChange{}, err
	}

	reservation, err := workflowAgentReservation(budget, input)
	if err != nil {
		return prepared, file, workflowstate.FileChange{}, err
	}

	reservation.AttemptInput = &workflowstate.EvidenceRef{Path: prepared.inputRef, Hash: prepared.inputHash}
	budgetFile, err := s.workflowIntegrationBudgetChange(ctx, run, budget)

	return prepared, file, budgetFile, err
}

func appendWorkflowAgentBudgetFile(files []workflowstate.FileChange, file workflowstate.FileChange) []workflowstate.FileChange {
	if file.Path == "" {
		return files
	}

	return append(files, file)
}

func (s *WorkflowStore) workflowAgentCreationFiles(ctx context.Context, envelope WorkflowRecord, run *workflowInvocationFile) ([]workflowstate.FileChange, error) {
	file, err := s.workflowAgentBudgetCreation(ctx, run)
	if err != nil {
		return nil, err
	}

	files, err := workflowInvocationCreationFiles(envelope, *run)

	return appendWorkflowAgentBudgetFile(files, file), err
}

func (s *WorkflowStore) checkWorkflowAgentReservationEvidence(run workflowInvocationFile, r WorkflowIntegrationReservation) error {
	if !workflowAgentReservationOwned(run, r) {
		return workflowstate.ErrConflict
	}

	var input WorkflowTaskInput
	if err := s.readWorkflowNativeEvidence(run.ID, *r.AttemptInput, &input); err != nil {
		return err
	}

	budget := WorkflowIntegrationBudget{Reservations: []WorkflowIntegrationReservation{r}}
	if _, err := workflowAgentReservation(&budget, input); err != nil {
		return err
	}

	var certificate WorkflowAgentHarnessCertificate
	if err := s.readWorkflowNativeEvidence(run.ID, input.AgentCall.HarnessRef, &certificate); err != nil {
		return err
	}

	if certificate.CapabilityRevision != r.Allowance.CapabilityRevision {
		return workflowstate.ErrConflict
	}

	return checkWorkflowAgentCertificate(certificate, r.Role, input.Worker)
}

func (s *WorkflowStore) checkWorkflowAgentCeilingAdmission(run workflowInvocationFile) error {
	for _, attempt := range run.Attempts {
		if !workflowAgentAttemptPurpose(attempt.Purpose) || attempt.InputRef == "" {
			continue
		}

		var input WorkflowTaskInput
		if err := s.readWorkflowNativeEvidence(run.ID, workflowstate.EvidenceRef{Path: attempt.InputRef, Hash: attempt.InputHash}, &input); err != nil {
			return err
		}

		if len(input.Worker) > 0 {
			return ErrWorkflowIntegrationBudget
		}
	}

	return nil
}

func workflowAgentAttemptPurpose(purpose string) bool {
	return purpose == "" || purpose == workflowSliceReviewPurpose || purpose == workflowMergerPurpose
}

func (s *WorkflowStore) checkWorkflowAgentProcessBudget(run workflowInvocationFile, prepared workflowPreparedTask) error {
	attempt := run.Attempts[len(run.Attempts)-1]
	if !workflowAgentAttemptPurpose(attempt.Purpose) || len(attempt.Processes) != 1 || !slices.Equal(attempt.Processes[0].Command, prepared.input.Worker) {
		return nil
	}

	return s.checkWorkflowUnmeteredAgent(run)
}
