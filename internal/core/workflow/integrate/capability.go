package integrate

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckCapability checks exact source identity and the observed atomic-result and signing guarantees.
func CheckCapability(source workflowstate.ReviewInput, capability workflowstate.PublicationCapability) error {
	if capability.Repository != source.PullRequest.Repository || capability.PullRequestURL != source.PullRequest.URL || capability.Method != source.Policy.Method || capability.Mechanism == "" || len(capability.Mechanism) > 128 || len(capability.Revision) != 64 {
		return workflowstate.ErrConflict
	}

	if !capability.AtomicSourceBase || !capability.ExactResult {
		return workflowstate.ErrAtomicGuard
	}

	return checkSignatureCapability(capability)
}

func checkSignatureCapability(capability workflowstate.PublicationCapability) error {
	if capability.RequiredSignatures != (capability.SigningProfile != "") || (capability.RequiredSignatures && !workflowstate.ValidFingerprint(capability.SigningProfile)) {
		return workflowstate.ErrAtomicGuard
	}

	return nil
}
