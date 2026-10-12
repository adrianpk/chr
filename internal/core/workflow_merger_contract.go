package core

import (
	"context"
	"encoding/json"
	"reflect"
)

func (s *WorkflowStore) checkWorkflowMergerCommands(ctx context.Context, run workflowInvocationFile, input WorkflowMergerInput) error {
	observation, err := s.observeWorkflowSlice(ctx, run)
	if err != nil {
		return err
	}

	gate, err := s.workflowMergerValidationContract(run, observation)
	if err != nil {
		return err
	}

	var retained WorkflowValidation
	if err := s.readWorkflowNativeEvidence(run.ID, input.ValidationContractRef, &retained); err != nil {
		return err
	}

	if !reflect.DeepEqual(gate.Commands, input.ValidationCommands) || !reflect.DeepEqual(gate, retained) {
		return ErrWorkflowConflict
	}

	return nil
}

func checkWorkflowMergerBudget(budget WorkflowIntegrationBudget, input WorkflowMergerInput) error {
	if err := checkWorkflowIntegrationCounterCharge(budget, WorkflowIntegrationReservation{Unit: input.Unit, Kind: "agent_call", Role: "merger"}, "merger"); err != nil {
		return err
	}

	if input.TokenAllowance == nil {
		if budget.TokenLimit != nil {
			return ErrWorkflowIntegrationBudget
		}

		return nil
	}

	total, err := workflowTokenSum(budget.TokensReserved, budget.TokensSettled)
	if err == nil {
		total, err = workflowTokenSum(total, input.TokenAllowance.Maximum)
	}

	if err != nil || budget.TokenLimit != nil && total > *budget.TokenLimit {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}

func (s *WorkflowStore) workflowMergerValidationContract(run workflowInvocationFile, observation workflowSliceObservation) (WorkflowValidation, error) {
	if run.Lane.Unit.Kind == "slice" {
		contract, err := s.readWorkflowSliceContract(observation.descriptor)

		return contract.Validation, err
	}

	var contract GeneratedWorkflowTask

	err := json.Unmarshal(observation.tasks[0].Domain["execution_contract"], &contract)

	return contract.Validation, err
}
