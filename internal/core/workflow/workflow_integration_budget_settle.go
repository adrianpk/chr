package workflow

import (
	"context"
	"reflect"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) settleWorkflowIntegration(ctx context.Context, run workflowInvocationFile, request IntegrationSettlement) (IntegrationBudget, error) {
	budgetObservation, err := s.currentWorkflowIntegrationBudget(ctx, run, request.IntegrationBudgetGuard)

	entry, budget := budgetObservation.Entry, budgetObservation.Budget

	if err != nil {
		return budget, err
	}

	index := slices.IndexFunc(budget.Reservations, func(r IntegrationReservation) bool { return r.ID == request.ReservationID })
	if index < 0 {
		return budget, workflowstate.ErrConflict
	}

	if request.State == "launched" && workflowIntegrationReservationStopped(run, entry) {
		return budget, ErrWorkflowStop
	}

	previous := budget.Reservations[index]
	next := previous

	next.State, next.ReportedInput, next.ReportedOutput, next.SettlementProof = request.State, request.Input, request.Output, request.Proof
	if reflect.DeepEqual(previous, next) {
		return budget, nil
	}

	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.IntegrationBudgetGuard); err != nil {
		return budget, err
	}

	if err := s.checkWorkflowIntegrationSettlement(run.ID, previous, next); err != nil {
		return budget, err
	}

	budget.Reservations[index] = next

	return budget, s.commitWorkflowIntegrationBudget(ctx, &run, &budget)
}

func (s *Store) checkWorkflowIntegrationSettlement(id string, old, next IntegrationReservation) error {
	if old.Kind != "agent_call" || !workflowIntegrationSettlementAllowed(old.State, next.State) {
		return workflowstate.ErrConflict
	}

	if next.State == "settled" {
		if err := s.checkWorkflowTokenSettlement(id, next); err != nil {
			return err
		}
	}

	return validateWorkflowIntegrationReservation(next)
}

func workflowIntegrationSettlementAllowed(old, next string) bool {
	switch next {
	case "launched", "released_before_launch":
		return old == "reserved"
	case "unknown":
		return old == "reserved" || old == "launched"
	case "settled":
		return old == "launched" || old == "unknown"
	default:
		return false
	}
}

func (s *Store) checkWorkflowTokenSettlement(id string, reservation IntegrationReservation) error {
	if reservation.SettlementProof == nil {
		return workflowstate.ErrConflict
	}

	var proof TokenSettlement
	if err := s.readWorkflowNativeEvidence(id, *reservation.SettlementProof, &proof); err != nil {
		return err
	}

	expected := TokenSettlement{SchemaVersion: 1, ReservationID: reservation.ID, AttemptID: reservation.AttemptID, CapabilityRevision: reservation.Allowance.CapabilityRevision, WholeCall: true, Input: reservation.ReportedInput, Output: reservation.ReportedOutput}
	if proof != expected {
		return workflowstate.ErrConflict
	}

	total, err := workflowstate.TokenSum(proof.Input, proof.Output)
	if err != nil || total > reservation.Allowance.Maximum || proof.Input > reservation.Allowance.Input || proof.Output > reservation.Allowance.Output {
		return workflowstate.ErrConflict
	}

	return nil
}
