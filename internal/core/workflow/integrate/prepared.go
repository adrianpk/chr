package integrate

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckPrepared checks the observed base, signing requirements and complete prepared-result proof.
func CheckPrepared(query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof, repairs RepairProof) error {
	expected := query.IntegrationQuery

	expected.ResultCommit, expected.SourceDirectory = proof.ResultCommit, ""
	if proof.BaseCommit != query.BaseCommit || proof.ObservedTarget != query.BaseCommit {
		return workflowstate.ErrConflict
	}

	if err := CheckSigning(proof, query.RequiredSignatures, query.SigningProfile); err != nil {
		return err
	}

	return CheckProof(expected, proof, repairs)
}
