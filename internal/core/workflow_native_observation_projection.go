package core

import "strings"

func (s *WorkflowStore) projectWorkflowNativeCanonicalObservation(entry *WorkflowIntegrationEntry, submission WorkflowSubmission) error {
	if entry.State == "blocked" && !workflowNativeCanonicalBlocker(entry.Blocker) {
		return nil
	}

	refs := submission.Native.ObservationRefs

	var observation WorkflowNativePublicationObservation
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, refs[len(refs)-1], &observation); err != nil {
		return err
	}

	if observation.State == "reconciliation_required" {
		entry.State = "blocked"
		entry.Blocker = &WorkflowIntegrationBlocker{Reason: observation.Condition, RequiredAction: "reconcile the original canonical PR and exact target history", ResumeState: "observing"}
	} else {
		entry.State, entry.Blocker = "observing", nil
	}

	return nil
}

func workflowNativeCanonicalBlocker(blocker *WorkflowIntegrationBlocker) bool {
	return blocker != nil && (strings.HasPrefix(blocker.Reason, "canonical_") || blocker.Reason == "published_source_unattributed")
}
