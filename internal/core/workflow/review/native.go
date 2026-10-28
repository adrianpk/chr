package review

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativeProofs checks retained evidence through the host's existing validators.
// The review stage receives no Store, claim, configuration or write capability.
type NativeProofs interface {
	CheckNativeRecordBinding(workflowstate.NativeRecordBinding) error
	CheckNativePreparation(workflowstate.ReviewInput) error
}

// CheckNativeInput checks exact native candidate, owner, Git and capability binding.
func CheckNativeInput(input workflowstate.ReviewInput, proofs NativeProofs) error {
	if input.Native == nil {
		return nil
	}

	native := input.Native

	binding := native.Candidate.NativeRecordBinding
	if native.ClaimGeneration == 0 || !workflowstate.ValidFingerprint(native.Candidate.CandidateHash) || input.PreparedGit == nil {
		return workflowstate.ErrConflict
	}

	if err := proofs.CheckNativeRecordBinding(binding); err != nil {
		return err
	}

	if !nativeOwnerMatches(input) {
		return workflowstate.ErrConflict
	}

	if binding.CapabilityHash != native.Capability.Revision || binding.PolicyHash != native.Capability.PolicyFingerprint {
		return workflowstate.ErrConflict
	}

	return proofs.CheckNativePreparation(input)
}

func nativeOwnerMatches(input workflowstate.ReviewInput) bool {
	binding := input.Native.Candidate.NativeRecordBinding

	return binding.InvocationID == input.InvocationID && binding.Unit == input.Unit && binding.CanonicalPR == input.PullRequest.URL && binding.ContractHash == input.Slice.ContractHash && workflowstate.NativeRecordGitMatches(binding, *input.PreparedGit)
}
