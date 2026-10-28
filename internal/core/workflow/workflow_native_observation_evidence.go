package workflow

import (
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowNativeCanonicalObservationValid(submission workflowstate.Submission, observation workflowstate.NativePublicationObservation) bool {
	if !workflowNativeObservationTargetValid(observation) {
		return false
	}

	binding, valid := s.workflowNativeCanonicalEvidenceBinding(submission)
	if !valid {
		return false
	}

	condition := workflowNativeObservationCondition(binding, observation.PullRequest)
	if observation.State == "reconciliation_required" {
		return workflowNativeReconciliationObservationValid(submission, binding, observation, condition)
	}

	if condition != "" || observation.Condition != "" || observation.Git == nil {
		return false
	}

	return workflowNativeCanonicalResultValid(binding, observation)
}

func workflowNativeCanonicalResultValid(binding workflowstate.NativePublicationBinding, observation workflowstate.NativePublicationObservation) bool {
	contractMatches := observation.ObservationContract == binding.Capability.ObservationContract

	sourceResult := contractMatches && workflowNativeSourceResultObserved(observation)
	if observation.State == "merged" {
		return observation.PullRequest.Merged && (observation.PullRequest.ResultCommit == binding.Git.ResultCommit || sourceResult)
	}

	pendingSource := contractMatches && workflowNativeSourceResultCompatible(binding.Capability.ObservationContract, binding.Git, observation.PullRequest) && !sourceResult

	return observation.State == "observing" && (!observation.PullRequest.Merged || pendingSource)
}

func workflowNativeObservationTargetValid(observation workflowstate.NativePublicationObservation) bool {
	if observation.ObservedTargetCommit == "" {
		return true
	}

	if !workflowstate.ValidCommitID(observation.ObservedTargetCommit) {
		return false
	}

	if observation.Git != nil {
		return observation.ObservedTargetCommit == observation.Git.ObservedTarget
	}

	return observation.State == "reconciliation_required" && observation.Condition == "canonical_target_history_diverged"
}

func (s *Store) workflowNativeCanonicalEvidenceBinding(submission workflowstate.Submission) (workflowstate.NativePublicationBinding, bool) {
	native := submission.Native

	binding := workflowstate.NativePublicationBinding{PullRequest: submission.Guard.PullRequest, Git: submission.Guard.Git}
	if s.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &binding.Capability) != nil {
		return binding, false
	}

	if native.TransportResponseRef == nil || s.readWorkflowNativeEvidence(native.InvocationID, *native.TransportResponseRef, &binding.Publication) != nil {
		return binding, false
	}

	for _, ref := range native.GitObservationRefs {
		var prior workflowstate.NativeGitObservation
		if s.readWorkflowNativeEvidence(native.InvocationID, ref, &prior) != nil {
			return binding, false
		}

		if prior.Git != nil && prior.SourceCommit == native.PreparedResult {
			binding.GitApplication = &prior

			break
		}
	}

	return binding, true
}

func workflowNativeReconciliationObservationValid(submission workflowstate.Submission, binding workflowstate.NativePublicationBinding, observation workflowstate.NativePublicationObservation, condition string) bool {
	native := submission.Native

	switch observation.Condition {
	case "canonical_pr_observation_deadline":
		return workflowNativeDeadlineEvidenceValid(*native, observation)
	case "canonical_target_unverified", "canonical_observation_unavailable":
		return observation.CanonicalPR == native.CanonicalPR
	case "canonical_target_history_diverged", "canonical_receipt_pr_unmerged":
		return workflowNativeReceiptReconciliationValid(*native, observation, condition)
	case "published_source_unattributed":
		return binding.Publication.State == "unknown" && !workflowNativeApplicationAttributed(binding) && observation.Git != nil && observation.ObservedSourceCommit != native.PreparedResult
	default:
		return condition != "" && condition == observation.Condition
	}
}

func workflowNativeDeadlineEvidenceValid(native workflowstate.NativeSubmission, observation workflowstate.NativePublicationObservation) bool {
	return observation.CanonicalPR == native.CanonicalPR && !time.Now().Before(native.ObservationDeadline)
}

func workflowNativeReceiptReconciliationValid(native workflowstate.NativeSubmission, observation workflowstate.NativePublicationObservation, condition string) bool {
	if observation.Condition == "canonical_receipt_pr_unmerged" {
		return native.ReceiptRef != nil && condition == "" && !observation.PullRequest.Merged
	}

	return observation.CanonicalPR == native.CanonicalPR && workflowstate.ValidCommitID(observation.ObservedTargetCommit) && observation.ObservedTargetCommit != native.PreparedResult && observation.Git == nil
}
