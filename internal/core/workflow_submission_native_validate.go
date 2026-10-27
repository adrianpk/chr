package core

import (
	"reflect"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowNativeSubmission(submission WorkflowSubmission, schema string) error {
	native := submission.Native
	if native == nil {
		return validateWorkflowLegacySubmissionSchema(submission, schema)
	}

	if schema != "chr-integration-journal/v2" || !validWorkflowNativeSubmissionOwner(submission) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordBinding(native.WorkflowNativeRecordBinding); err != nil {
		return err
	}

	if native.IntentTime != submission.SubmittedAt || !native.ObservationDeadline.After(native.IntentTime) || !validWorkflowNativeSubmissionHistory(*native) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeSubmissionRefs(*native); err != nil {
		return err
	}

	return validateWorkflowNativeApplication(submission)
}

func validWorkflowNativeSubmissionHistory(native WorkflowNativeSubmission) bool {
	return len(native.ObservationRefs) <= workflowstate.PageLimit && len(native.GitObservationRefs) <= workflowstate.PageLimit && len(native.TransportResponseHistory) <= workflowstate.PageLimit
}

func validateWorkflowLegacySubmissionSchema(submission WorkflowSubmission, schema string) error {
	if schema == "chr-integration-journal/v2" && submission.State != "unpublished" {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeSubmissionOwner(submission WorkflowSubmission) bool {
	native := submission.Native

	return native.InvocationID == submission.InvocationID && native.Unit == submission.Unit && native.ClaimGeneration > 0 && native.Mechanism == WorkflowPublicationGitAtomicDualRef && validWorkflowNativeFingerprint(native.CandidateHash) && workflowNativeRecordGitMatches(native.WorkflowNativeRecordBinding, submission.Guard.Git) && native.CanonicalPR == submission.Guard.PullRequest.URL && native.CapabilityHash == submission.Guard.Capability.Revision
}

func validateWorkflowNativeSubmissionRefs(native WorkflowNativeSubmission) error {
	refs := append([]workflowstate.WorkflowEvidenceRef{native.CandidateRef, native.ReviewGrantRef, native.FastForwardProofRef}, native.ObservationRefs...)

	refs = append(refs, native.TransportResponseHistory...)

	refs = append(refs, native.GitObservationRefs...)
	for _, ref := range []*workflowstate.WorkflowEvidenceRef{native.AdvertisementRef, native.CommandManifestRef, native.TransportResponseRef, native.ReceiptRef, native.NonApplicationRef, native.DispatchCompletionRef} {
		if ref != nil {
			refs = append(refs, *ref)
		}
	}

	for _, ref := range refs {
		if err := validateWorkflowEvidenceRef(native.InvocationID, ref); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowNativeApplication(submission WorkflowSubmission) error {
	native := submission.Native
	if !validWorkflowNativeDispatch(*native) {
		return workflowstate.ErrConflict
	}

	switch native.ApplicationState {
	case "not_started", "unknown":
		if submission.State != "awaiting_observation" || native.ReceiptRef != nil || native.NonApplicationRef != nil {
			return workflowstate.ErrConflict
		}
	case "not_applied":
		if !validWorkflowNativeUnpublished(submission) {
			return workflowstate.ErrConflict
		}
	case "git_applied":
		return validateWorkflowNativeApplied(submission)
	default:
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeDispatch(native WorkflowNativeSubmission) bool {
	if native.DispatchState == "not_started" {
		return native.DispatchIdentity == nil && native.ApplicationState == "not_started"
	}

	return slices.Contains([]string{"dispatching", "quiescent"}, native.DispatchState) && native.DispatchIdentity != nil && validWorkflowProcessIdentity(*native.DispatchIdentity)
}

func validWorkflowNativeUnpublished(submission WorkflowSubmission) bool {
	native := submission.Native

	return slices.Contains([]string{"unpublished", "awaiting_observation"}, submission.State) && native.DispatchState == "quiescent" && native.NonApplicationRef != nil && native.ReceiptRef == nil
}

func validateWorkflowNativeApplied(submission WorkflowSubmission) error {
	native := submission.Native
	if native.TransportResponseRef == nil || native.CommandManifestRef == nil || native.AdvertisementRef == nil || native.NonApplicationRef != nil || submission.State == "unpublished" {
		return workflowstate.ErrConflict
	}

	if (submission.State == "confirmed") != (native.ReceiptRef != nil) {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardWorkflowNativeSubmissionUpdate(old, next WorkflowNativeSubmission) error {
	if !sameWorkflowNativeSubmissionIntent(old, next) || !workflowNativeDispatchProgresses(old, next) || !workflowNativeApplicationProgresses(old, next) {
		return workflowstate.ErrConflict
	}

	if len(next.ObservationRefs) < len(old.ObservationRefs) || !slices.Equal(old.ObservationRefs, next.ObservationRefs[:len(old.ObservationRefs)]) {
		return workflowstate.ErrConflict
	}

	if len(next.GitObservationRefs) < len(old.GitObservationRefs) || !slices.Equal(old.GitObservationRefs, next.GitObservationRefs[:len(old.GitObservationRefs)]) {
		return workflowstate.ErrConflict
	}

	if err := guardWorkflowNativeResponseHistory(old, next); err != nil {
		return err
	}

	return guardWorkflowNativeRetainedRefs(old, next)
}

func workflowNativeDispatchProgresses(old, next WorkflowNativeSubmission) bool {
	if old.DispatchIdentity != nil && !reflect.DeepEqual(old.DispatchIdentity, next.DispatchIdentity) {
		return false
	}

	return old.DispatchState != "quiescent" || next.DispatchState == "quiescent"
}

func workflowNativeApplicationProgresses(old, next WorkflowNativeSubmission) bool {
	switch old.ApplicationState {
	case "git_applied", "not_applied":
		return next.ApplicationState == old.ApplicationState
	case "unknown":
		return next.ApplicationState != "not_started"
	default:
		return true
	}
}

func guardWorkflowNativeRetainedRefs(old, next WorkflowNativeSubmission) error {
	for _, pair := range [][2]*workflowstate.WorkflowEvidenceRef{{old.AdvertisementRef, next.AdvertisementRef}, {old.CommandManifestRef, next.CommandManifestRef}, {old.ReceiptRef, next.ReceiptRef}, {old.NonApplicationRef, next.NonApplicationRef}, {old.DispatchCompletionRef, next.DispatchCompletionRef}} {
		if pair[0] != nil && !reflect.DeepEqual(pair[0], pair[1]) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func sameWorkflowNativeSubmissionIntent(old, next WorkflowNativeSubmission) bool {
	clearWorkflowNativeSubmissionProgress(&old)
	clearWorkflowNativeSubmissionProgress(&next)

	return workflowNativeJSONEqual(old, next)
}

func clearWorkflowNativeSubmissionProgress(native *WorkflowNativeSubmission) {
	native.DispatchState, native.ApplicationState = "", ""
	native.DispatchIdentity = nil
	native.DispatchCompletionRef = nil
	native.AdvertisementRef, native.CommandManifestRef, native.TransportResponseRef = nil, nil, nil
	native.ObservationRefs = nil
	native.GitObservationRefs = nil
	native.TransportResponseHistory = nil
	native.ReceiptRef, native.NonApplicationRef = nil, nil
}

func guardWorkflowNativeResponseHistory(old, next WorkflowNativeSubmission) error {
	if len(next.TransportResponseHistory) < len(old.TransportResponseHistory) || !slices.Equal(old.TransportResponseHistory, next.TransportResponseHistory[:len(old.TransportResponseHistory)]) {
		return workflowstate.ErrConflict
	}

	if old.TransportResponseRef == nil || reflect.DeepEqual(old.TransportResponseRef, next.TransportResponseRef) {
		if len(next.TransportResponseHistory) != len(old.TransportResponseHistory) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if next.TransportResponseRef == nil || len(next.TransportResponseHistory) != len(old.TransportResponseHistory)+1 || next.TransportResponseHistory[len(old.TransportResponseHistory)] != *old.TransportResponseRef {
		return workflowstate.ErrConflict
	}

	return nil
}
