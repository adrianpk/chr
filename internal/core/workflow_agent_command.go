package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowRunAgentCommand(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, command []string) (WorkflowCommandResult, error) {
	if err := harnesshost.CheckHarnessExecutable(prepared.input.HarnessBinding, command); err != nil {
		return WorkflowCommandResult{}, err
	}

	if prepared.input.AgentCall == nil {
		return workflowRunCommand(ctx, claim, prepared, command)
	}

	if err := claim.store.checkWorkflowAgentInput(ctx, prepared.input); err != nil {
		return WorkflowCommandResult{}, err
	}

	result, executionErr := workflowRunCommand(ctx, claim, prepared, command)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	response, usageErr := parseWorkflowAgentResponse(result, prepared.input)

	err := claim.store.settleWorkflowAgentCall(cleanup, claim, prepared, response, errors.Join(executionErr, usageErr))
	if executionErr == nil && usageErr == nil && err == nil {
		result.Stdout = response.Result
	}

	return result, errors.Join(executionErr, usageErr, err)
}

func (s *WorkflowStore) checkWorkflowAgentInput(ctx context.Context, input WorkflowTaskInput) error {
	call := input.AgentCall
	if call == nil {
		return workflowstate.ErrConflict
	}

	certificate, err := s.checkWorkflowAgentHarness(ctx, input.InvocationID, call.Role, WorkflowAgentHarness{Command: input.Worker, Certificate: call.HarnessRef})
	if err == nil && certificate.CapabilityRevision != call.Allowance.CapabilityRevision {
		err = ErrWorkflowIntegrationBudget
	}

	return err
}

func parseWorkflowAgentResponse(result WorkflowCommandResult, input WorkflowTaskInput) (WorkflowAgentResponse, error) {
	var response WorkflowAgentResponse
	if result.ExitCode != 0 || result.Truncated {
		return response, ErrWorkflowIntegrationBudget
	}

	decoder := json.NewDecoder(strings.NewReader(result.Stdout))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&response); err != nil {
		return response, errors.Join(ErrWorkflowIntegrationBudget, err)
	}

	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return response, ErrWorkflowIntegrationBudget
	}

	return response, checkWorkflowAgentResponse(response, input)
}

func checkWorkflowAgentResponse(response WorkflowAgentResponse, input WorkflowTaskInput) error {
	call := input.AgentCall
	if call == nil || response.Protocol != workflowAgentCallProtocol || response.ReservationID != call.ReservationID || response.AttemptID != input.AttemptID || response.Input == nil || response.Output == nil {
		return ErrWorkflowIntegrationBudget
	}

	usage := WorkflowMergerUsage{CapabilityRevision: response.CapabilityRevision, WholeCall: response.WholeCall, Input: *response.Input, Output: *response.Output}
	if err := validateWorkflowMergerReportedUsage(&usage, &call.Allowance); err != nil {
		return errors.Join(ErrWorkflowIntegrationBudget, err)
	}

	return nil
}

func (s *WorkflowStore) settleWorkflowAgentCall(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, response WorkflowAgentResponse, executionErr error) error {
	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
			return err
		}

		budget, err := claim.store.readWorkflowIntegrationBudget(run)
		if err != nil {
			return err
		}

		reservation, err := workflowAgentReservation(&budget, prepared.input)
		if err != nil || reservation.State != "launched" {
			return errors.Join(workflowstate.ErrConflict, err)
		}

		proofFile, err := claim.store.workflowAgentSettlement(ctx, reservation, prepared.input, response, executionErr)
		if err != nil {
			return err
		}

		budgetFile, err := claim.store.workflowIntegrationBudgetChange(ctx, &run, &budget)
		if err != nil {
			return err
		}

		return claim.store.commitWorkflowPullRequestFiles(&run, []workflowstate.FileChange{budgetFile}, proofFile)
	})
}

func (s *WorkflowStore) workflowAgentSettlement(ctx context.Context, reservation *WorkflowIntegrationReservation, input WorkflowTaskInput, response WorkflowAgentResponse, executionErr error) (workflowstate.FileChange, error) {
	reservation.State = "unknown"

	if !workflowAgentUsageKnown(executionErr) {
		return workflowstate.FileChange{}, nil
	}

	if err := s.checkWorkflowAgentInput(ctx, input); err != nil {
		return workflowstate.FileChange{}, err
	}

	proof := WorkflowTokenSettlement{SchemaVersion: 1, ReservationID: reservation.ID, AttemptID: reservation.AttemptID, CapabilityRevision: reservation.Allowance.CapabilityRevision, WholeCall: true, Input: *response.Input, Output: *response.Output}

	data, err := json.Marshal(proof)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	path := "run/" + input.InvocationID + "/evidence/agent-usage-" + input.AttemptID + ".json"
	reservation.State, reservation.ReportedInput, reservation.ReportedOutput = "settled", proof.Input, proof.Output
	reservation.SettlementProof = &workflowstate.EvidenceRef{Path: path, Hash: contenthash.Sum(data)}

	return workflowstate.FileChange{Path: path, After: data}, validateWorkflowIntegrationReservation(*reservation)
}

func workflowAgentUsageKnown(err error) bool { return err == nil }
