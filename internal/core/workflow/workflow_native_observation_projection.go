package workflow

import (
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) projectWorkflowNativeCanonicalObservation(entry *IntegrationEntry, submission workflowstate.Submission) error {
	if entry.State == "blocked" && !workflowNativeCanonicalBlocker(entry.Blocker) {
		return nil
	}

	refs := submission.Native.ObservationRefs

	var observation workflowstate.NativePublicationObservation
	if err := s.readWorkflowNativeEvidence(submission.InvocationID, refs[len(refs)-1], &observation); err != nil {
		return err
	}

	if observation.State == "reconciliation_required" {
		entry.State = "blocked"
		entry.Blocker = &IntegrationBlocker{Reason: observation.Condition, RequiredAction: "reconcile the original canonical PR and exact target history", ResumeState: "observing"}
	} else {
		entry.State, entry.Blocker = "observing", nil
	}

	return nil
}

func workflowNativeCanonicalBlocker(blocker *IntegrationBlocker) bool {
	return blocker != nil && (strings.HasPrefix(blocker.Reason, "canonical_") || blocker.Reason == "published_source_unattributed")
}
