package workflow

import (
	"reflect"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func validateWorkflowNativeSubmission(submission workflowstate.Submission, schema string) error {
	native := submission.Native
	if native == nil {
		return validateWorkflowLegacySubmissionSchema(submission, schema)
	}

	if schema != "chr-integration-journal/v2" || !validWorkflowNativeSubmissionOwner(submission) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordBinding(native.NativeRecordBinding); err != nil {
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

func validWorkflowNativeSubmissionHistory(native workflowstate.NativeSubmission) bool {
	return len(native.ObservationRefs) <= workflowstate.PageLimit && len(native.GitObservationRefs) <= workflowstate.PageLimit && len(native.TransportResponseHistory) <= workflowstate.PageLimit
}

func validateWorkflowLegacySubmissionSchema(submission workflowstate.Submission, schema string) error {
	if schema == "chr-integration-journal/v2" && submission.State != workflowstate.SubmissionUnpublished {
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeSubmissionOwner(submission workflowstate.Submission) bool {
	native := submission.Native

	return native.InvocationID == submission.InvocationID && native.Unit == submission.Unit && native.ClaimGeneration > 0 && native.Mechanism == workspaceconfig.WorkflowPublicationGitAtomicDualRef && workflowstate.ValidFingerprint(native.CandidateHash) && workflowstate.NativeRecordGitMatches(native.NativeRecordBinding, submission.Guard.Git) && native.CanonicalPR == submission.Guard.PullRequest.URL && native.CapabilityHash == submission.Guard.Capability.Revision
}

func validateWorkflowNativeSubmissionRefs(native workflowstate.NativeSubmission) error {
	refs := append([]workflowstate.EvidenceRef{native.CandidateRef, native.ReviewGrantRef, native.FastForwardProofRef}, native.ObservationRefs...)

	refs = append(refs, native.TransportResponseHistory...)

	refs = append(refs, native.GitObservationRefs...)
	for _, ref := range []*workflowstate.EvidenceRef{native.AdvertisementRef, native.CommandManifestRef, native.TransportResponseRef, native.ReceiptRef, native.NonApplicationRef, native.DispatchCompletionRef} {
		if ref != nil {
			refs = append(refs, *ref)
		}
	}

	for _, ref := range refs {
		if err := workflowstate.CheckEvidenceRef(native.InvocationID, ref); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowNativeApplication(submission workflowstate.Submission) error {
	native := submission.Native
	if !validWorkflowNativeDispatch(*native) {
		return workflowstate.ErrConflict
	}

	switch native.ApplicationState {
	case workflowstate.NativeApplicationNotStarted, workflowstate.NativeApplicationUnknown:
		if submission.State != workflowstate.SubmissionAwaitingObservation || native.ReceiptRef != nil || native.NonApplicationRef != nil {
			return workflowstate.ErrConflict
		}
	case workflowstate.NativeApplicationNotApplied:
		if !validWorkflowNativeUnpublished(submission) {
			return workflowstate.ErrConflict
		}
	case workflowstate.NativeApplicationGitApplied:
		return validateWorkflowNativeApplied(submission)
	default:
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeDispatch(native workflowstate.NativeSubmission) bool {
	if native.DispatchState == workflowstate.NativeDispatchNotStarted {
		return native.DispatchIdentity == nil && native.ApplicationState == workflowstate.NativeApplicationNotStarted
	}

	return slices.Contains([]workflowstate.NativeDispatchState{workflowstate.NativeDispatching, workflowstate.NativeDispatchQuiescent}, native.DispatchState) && native.DispatchIdentity != nil && validWorkflowProcessIdentity(*native.DispatchIdentity)
}

func validWorkflowNativeUnpublished(submission workflowstate.Submission) bool {
	native := submission.Native

	return slices.Contains([]workflowstate.SubmissionState{workflowstate.SubmissionUnpublished, workflowstate.SubmissionAwaitingObservation}, submission.State) && native.DispatchState == workflowstate.NativeDispatchQuiescent && native.NonApplicationRef != nil && native.ReceiptRef == nil
}

func validateWorkflowNativeApplied(submission workflowstate.Submission) error {
	native := submission.Native
	if native.TransportResponseRef == nil || native.CommandManifestRef == nil || native.AdvertisementRef == nil || native.NonApplicationRef != nil || submission.State == workflowstate.SubmissionUnpublished {
		return workflowstate.ErrConflict
	}

	if (submission.State == workflowstate.SubmissionConfirmed) != (native.ReceiptRef != nil) {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardWorkflowNativeSubmissionUpdate(old, next workflowstate.NativeSubmission) error {
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

func workflowNativeDispatchProgresses(old, next workflowstate.NativeSubmission) bool {
	if old.DispatchIdentity != nil && !reflect.DeepEqual(old.DispatchIdentity, next.DispatchIdentity) {
		return false
	}

	return old.DispatchState != workflowstate.NativeDispatchQuiescent || next.DispatchState == workflowstate.NativeDispatchQuiescent
}

func workflowNativeApplicationProgresses(old, next workflowstate.NativeSubmission) bool {
	switch old.ApplicationState {
	case workflowstate.NativeApplicationGitApplied, workflowstate.NativeApplicationNotApplied:
		return next.ApplicationState == old.ApplicationState
	case workflowstate.NativeApplicationUnknown:
		return next.ApplicationState != workflowstate.NativeApplicationNotStarted
	default:
		return true
	}
}

func guardWorkflowNativeRetainedRefs(old, next workflowstate.NativeSubmission) error {
	for _, pair := range [][2]*workflowstate.EvidenceRef{{old.AdvertisementRef, next.AdvertisementRef}, {old.CommandManifestRef, next.CommandManifestRef}, {old.ReceiptRef, next.ReceiptRef}, {old.NonApplicationRef, next.NonApplicationRef}, {old.DispatchCompletionRef, next.DispatchCompletionRef}} {
		if pair[0] != nil && !reflect.DeepEqual(pair[0], pair[1]) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func sameWorkflowNativeSubmissionIntent(old, next workflowstate.NativeSubmission) bool {
	clearWorkflowNativeSubmissionProgress(&old)
	clearWorkflowNativeSubmissionProgress(&next)

	return workflowNativeJSONEqual(old, next)
}

func clearWorkflowNativeSubmissionProgress(native *workflowstate.NativeSubmission) {
	native.DispatchState, native.ApplicationState = "", ""
	native.DispatchIdentity = nil
	native.DispatchCompletionRef = nil
	native.AdvertisementRef, native.CommandManifestRef, native.TransportResponseRef = nil, nil, nil
	native.ObservationRefs = nil
	native.GitObservationRefs = nil
	native.TransportResponseHistory = nil
	native.ReceiptRef, native.NonApplicationRef = nil, nil
}

func guardWorkflowNativeResponseHistory(old, next workflowstate.NativeSubmission) error {
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
