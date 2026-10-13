package core

import "time"

func (s *WorkflowStore) workflowNativeCanonicalObservationValid(submission WorkflowSubmission, observation WorkflowNativePublicationObservation) bool {
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

	if observation.State == "merged" {
		return observation.PullRequest.Merged
	}

	return observation.State == "observing" && !observation.PullRequest.Merged
}

func (s *WorkflowStore) workflowNativeCanonicalEvidenceBinding(submission WorkflowSubmission) (WorkflowNativePublicationBinding, bool) {
	native := submission.Native

	binding := WorkflowNativePublicationBinding{PullRequest: submission.Guard.PullRequest, Git: submission.Guard.Git}
	if native.TransportResponseRef == nil || s.readWorkflowNativeEvidence(native.InvocationID, *native.TransportResponseRef, &binding.Publication) != nil {
		return binding, false
	}

	for _, ref := range native.GitObservationRefs {
		var prior WorkflowNativeGitObservation
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

func workflowNativeReconciliationObservationValid(submission WorkflowSubmission, binding WorkflowNativePublicationBinding, observation WorkflowNativePublicationObservation, condition string) bool {
	native := submission.Native

	switch observation.Condition {
	case "canonical_pr_observation_deadline":
		return observation.CanonicalPR == native.CanonicalPR && !time.Now().Before(native.ObservationDeadline)
	case "canonical_target_unverified", "canonical_observation_unavailable":
		return observation.CanonicalPR == native.CanonicalPR
	case "published_source_unattributed":
		return binding.Publication.State == "unknown" && !workflowNativeApplicationAttributed(binding) && observation.Git != nil && observation.ObservedSourceCommit != native.PreparedResult
	default:
		return condition != "" && condition == observation.Condition
	}
}
