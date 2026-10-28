package integrate

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckWorkflowPreparedProof checks the original immutable candidate observations and retained proof.
func CheckWorkflowPreparedProof(query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof, repairs RepairProof) error {
	expected := query.IntegrationQuery

	expected.ResultCommit, expected.SourceDirectory = proof.ResultCommit, ""
	if proof.BaseCommit != query.BaseCommit || proof.ObservedTarget != query.BaseCommit {
		return workflowstate.ErrConflict
	}

	if err := CheckSigning(proof, query.RequiredSignatures, query.SigningProfile); err != nil {
		return err
	}

	return ValidateWorkflowIntegrationProof(expected, proof, repairs)
}
