package workflow

import (
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowMergerTokenSettlement(reservation *IntegrationReservation, result MergerAttempt) (workflowstate.FileChange, error) {
	usage := result.TrustedUsage
	if result.Input.TokenAllowance == nil || usage == nil {
		return workflowstate.FileChange{}, nil
	}

	if err := workflowmerge.ValidateWorkflowMergerReportedUsage(usage, result.Input.TokenAllowance); err != nil {
		return workflowstate.FileChange{}, err
	}

	proof := TokenSettlement{SchemaVersion: 1, ReservationID: reservation.ID, AttemptID: reservation.AttemptID, CapabilityRevision: reservation.Allowance.CapabilityRevision, WholeCall: true, Input: usage.Input, Output: usage.Output}

	data, err := json.Marshal(proof)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	path := "run/" + result.Input.InvocationID + "/evidence/merger-usage-" + reservation.AttemptID + ".json"
	reservation.State, reservation.ReportedInput, reservation.ReportedOutput = "settled", usage.Input, usage.Output
	reservation.SettlementProof = &workflowstate.EvidenceRef{Path: path, Hash: contenthash.Sum(data)}

	return workflowstate.FileChange{Path: path, After: data}, validateWorkflowIntegrationReservation(*reservation)
}
