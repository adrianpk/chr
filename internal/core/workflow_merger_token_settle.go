package core

import (
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func workflowMergerTokenSettlement(reservation *WorkflowIntegrationReservation, result WorkflowMergerAttempt) (workflowFileChange, error) {
	usage := result.TrustedUsage
	if result.Input.TokenAllowance == nil || usage == nil {
		return workflowFileChange{}, nil
	}

	if err := validateWorkflowMergerReportedUsage(usage, result.Input.TokenAllowance); err != nil {
		return workflowFileChange{}, err
	}

	proof := WorkflowTokenSettlement{SchemaVersion: 1, ReservationID: reservation.ID, AttemptID: reservation.AttemptID, CapabilityRevision: reservation.Allowance.CapabilityRevision, WholeCall: true, Input: usage.Input, Output: usage.Output}

	data, err := json.Marshal(proof)
	if err != nil {
		return workflowFileChange{}, err
	}

	path := "run/" + result.Input.InvocationID + "/evidence/merger-usage-" + reservation.AttemptID + ".json"
	reservation.State, reservation.ReportedInput, reservation.ReportedOutput = "settled", usage.Input, usage.Output
	reservation.SettlementProof = &WorkflowEvidenceRef{Path: path, Hash: contenthash.Sum(data)}

	return workflowFileChange{Path: path, After: data}, validateWorkflowIntegrationReservation(*reservation)
}
