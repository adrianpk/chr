package workflow

import (
	"context"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowEncodedAgentInput struct {
	Prepared workflowPreparedTask
	Input    workflowstate.FileChange
	Budget   workflowstate.FileChange
}

func (s *Store) workflowAgentBudgetCreation(ctx context.Context, run *workflowInvocationFile) (workflowstate.FileChange, error) {
	if run.Policy.AgentTokenBudget == 0 {
		return workflowstate.FileChange{}, nil
	}

	limit := run.Policy.AgentTokenBudget
	limits := workflowIntegrationAdmissionLimits(run.Policy)
	budget := IntegrationBudget{SchemaVersion: 1, InvocationID: run.ID, AdmissionTokenLimit: &limit, TokenLimit: &limit, AdmissionLimits: limits, LimitsSnapshot: limits}

	return s.workflowIntegrationBudgetChange(ctx, run, &budget)
}

func (s *Store) reserveWorkflowAgentInput(ctx context.Context, run *workflowInvocationFile, input *TaskInput, harness *AgentHarness) (*IntegrationBudget, error) {
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

	maximum, err := workflowstate.TokenSum(harness.InputTokens, harness.OutputTokens)
	if err != nil {
		return nil, err
	}

	allowance := workflowstate.TokenAllowance{Maximum: maximum, Input: harness.InputTokens, Output: harness.OutputTokens, CapabilityRevision: certificate.CapabilityRevision, WholeCall: true, Deadline: run.Attempts[len(run.Attempts)-1].LeaseUntil}

	r := IntegrationReservation{ID: "agent-" + input.AttemptID, Kind: "agent_call", Role: role, EntryID: workflowIntegrationEntryID(*run, input.Lane.Unit), Unit: input.Lane.Unit, ClaimGeneration: input.Generation, AttemptID: input.AttemptID, Allowance: allowance, State: "reserved"}
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
	input.AgentCall = &AgentCall{ReservationID: r.ID, Role: role, HarnessRef: harness.Certificate, Allowance: allowance}

	return &budget, nil
}

func (s *Store) workflowAgentBudget(ctx context.Context, run workflowInvocationFile, unit workflowstate.Ref) (IntegrationBudget, error) {
	budget, err := s.workflowIntegrationBudgetAdmission(ctx, run, IntegrationEntry{Unit: unit}, nil)
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

func (s *Store) checkWorkflowUnmeteredAgent(run workflowInvocationFile) error {
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

func workflowAgentReservation(budget *IntegrationBudget, input TaskInput) (*IntegrationReservation, error) {
	call := input.AgentCall
	if call == nil {
		return nil, workflowstate.ErrConflict
	}

	index := slices.IndexFunc(budget.Reservations, func(r IntegrationReservation) bool {
		return r.ID == call.ReservationID && r.AttemptID == input.AttemptID && r.ClaimGeneration == input.Generation && r.Role == call.Role && r.Unit == input.Lane.Unit && r.Allowance == call.Allowance
	})
	if index < 0 {
		return nil, workflowstate.ErrConflict
	}

	return &budget.Reservations[index], nil
}

func (s *Store) encodeWorkflowAgentInput(ctx context.Context, run *workflowInvocationFile, input TaskInput, harness *AgentHarness) (workflowEncodedAgentInput, error) {
	budget, err := s.reserveWorkflowAgentInput(ctx, run, &input, harness)
	if err != nil {
		return workflowEncodedAgentInput{Prepared: workflowPreparedTask{}, Input: workflowstate.FileChange{}, Budget: workflowstate.FileChange{}}, err
	}

	encodedTaskInput, err := encodeWorkflowTaskInput(run, input)

	if err != nil || budget == nil {
		return workflowEncodedAgentInput{Prepared: encodedTaskInput.Prepared, Input: encodedTaskInput.Change, Budget: workflowstate.FileChange{}}, err
	}

	reservation, err := workflowAgentReservation(budget, input)
	if err != nil {
		return workflowEncodedAgentInput{Prepared: encodedTaskInput.Prepared, Input: encodedTaskInput.Change, Budget: workflowstate.FileChange{}}, err
	}

	reservation.AttemptInput = &workflowstate.EvidenceRef{Path: encodedTaskInput.Prepared.inputRef, Hash: encodedTaskInput.Prepared.inputHash}
	budgetFile, err := s.workflowIntegrationBudgetChange(ctx, run, budget)

	return workflowEncodedAgentInput{Prepared: encodedTaskInput.Prepared, Input: encodedTaskInput.Change, Budget: budgetFile}, err
}

func appendWorkflowAgentBudgetFile(files []workflowstate.FileChange, file workflowstate.FileChange) []workflowstate.FileChange {
	if file.Path == "" {
		return files
	}

	return append(files, file)
}

func (s *Store) workflowAgentCreationFiles(ctx context.Context, envelope workflowstate.Record, run *workflowInvocationFile) ([]workflowstate.FileChange, error) {
	file, err := s.workflowAgentBudgetCreation(ctx, run)
	if err != nil {
		return nil, err
	}

	files, err := workflowInvocationCreationFiles(envelope, *run)

	return appendWorkflowAgentBudgetFile(files, file), err
}

func (s *Store) checkWorkflowAgentReservationEvidence(run workflowInvocationFile, r IntegrationReservation) error {
	if !workflowAgentReservationOwned(run, r) {
		return workflowstate.ErrConflict
	}

	var input TaskInput
	if err := s.readWorkflowNativeEvidence(run.ID, *r.AttemptInput, &input); err != nil {
		return err
	}

	budget := IntegrationBudget{Reservations: []IntegrationReservation{r}}
	if _, err := workflowAgentReservation(&budget, input); err != nil {
		return err
	}

	var certificate AgentHarnessCertificate
	if err := s.readWorkflowNativeEvidence(run.ID, input.AgentCall.HarnessRef, &certificate); err != nil {
		return err
	}

	if certificate.CapabilityRevision != r.Allowance.CapabilityRevision {
		return workflowstate.ErrConflict
	}

	return checkWorkflowAgentCertificate(certificate, r.Role, input.Worker)
}

func (s *Store) checkWorkflowAgentCeilingAdmission(run workflowInvocationFile) error {
	for _, attempt := range run.Attempts {
		if !workflowAgentAttemptPurpose(attempt.Purpose) || attempt.InputRef == "" {
			continue
		}

		var input TaskInput
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

func (s *Store) checkWorkflowAgentProcessBudget(run workflowInvocationFile, prepared workflowPreparedTask) error {
	attempt := run.Attempts[len(run.Attempts)-1]
	if !workflowAgentAttemptPurpose(attempt.Purpose) || len(attempt.Processes) != 1 || !slices.Equal(attempt.Processes[0].Command, prepared.input.Worker) {
		return nil
	}

	return s.checkWorkflowUnmeteredAgent(run)
}
