package core

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedWorkAdmissionPolicy binds existing progression intent to native admission.
const GuidedWorkAdmissionPolicy = "guided-work/v1"

// GuidedWorkflowAdmission records authority provenance without extending native policy.
type GuidedWorkflowAdmission struct {
	RunID, Actor, Policy, OperationDigest, CheckpointDigest, MaterializationID string
}

// GuidedInvocationReference addresses the native owner; its mutable state stays there.
type GuidedInvocationReference struct {
	ID, RequestKey string
	Revision       uint64
}

// GuidedPrerequisiteError carries a fixed, user-facing corrective action.
// Underlying provider diagnostics must not be interpolated into Action.
type GuidedPrerequisiteError struct {
	Action string
	Cause  error
}

func (e *GuidedPrerequisiteError) Error() string { return e.Action }
func (e *GuidedPrerequisiteError) Unwrap() error { return e.Cause }

// RetainGuidedHostBlocker reports a stopped activation without changing native facts.
func RetainGuidedHostBlocker(ctx context.Context, root string, scope GuidedUseCaseScope, invocation string) error {
	return WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		r := &fileAuthoringSessionRepository{projectDir: root}

		session, err := r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := guardGuidedScope(session, scope); err != nil {
			return err
		}

		state := session.Guided
		if state.Work == nil || state.Work.Invocation == nil || state.Work.Invocation.ID != invocation {
			return workflowstate.ErrConflict
		}

		if state.Status == "paused" || state.Status == "cancelled" || state.Status == "completed" {
			return nil
		}

		state.Status = "blocked"
		state.Blocker = "Managed activation stopped. Inspect the retained native work and correct its prerequisite before continuing."

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})
}

// GuidedPreparationCheckpoint returns the sealed canonical preparation input.
func GuidedPreparationCheckpoint(session AuthoringSession) (GuidedStageCheckpoint, error) {
	if session.Guided == nil {
		return GuidedStageCheckpoint{}, workflowstate.ErrConflict
	}

	return guidedStageCheckpoint(session, "preparation")
}

// ValidateGuidedWorkPreview rejects omissions, additions and replaced initial contracts.
func ValidateGuidedWorkPreview(reference GuidedMaterializationReference, preview WorkflowPreview) error {
	if preview.Selector != (WorkflowSelector{Extent: "slice_set", Identity: reference.SetID}) || preview.Total != len(reference.Members) || len(preview.Members) != preview.Total || preview.Total == 0 || preview.Total > workflowstate.WorkflowPageLimit {
		return workflowstate.ErrConflict
	}

	for i, expected := range reference.Members {
		member := preview.Members[i]
		if !guidedInitialMemberMatches(member, expected, reference.SetID) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func guidedInitialMemberMatches(member WorkflowPreviewMember, expected GuidedWorkMember, set string) bool {
	return member.Key == expected.Key && member.Path == expected.Path && member.Revision == expected.Revision && member.SourceHash == expected.SourceHash && member.ContractHash == expected.ContractHash && member.SetID == set
}

func (s *WorkflowStore) checkGuidedWorkflowAdmission(ctx context.Context, request WorkflowStartRequest) error {
	if request.Admission == nil || request.Admission.Guided == nil {
		return nil
	}

	absolute, err := filepath.Abs(s.layout.ProjectDir())
	if err != nil || !holdsAuthoringOwner(ctx, absolute) {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	proof := request.Admission.Guided
	r := &fileAuthoringSessionRepository{projectDir: s.layout.ProjectDir()}

	session, err := r.getGuardedSession(ctx, proof.RunID)
	if err != nil {
		return err
	}

	if err := guardGuidedAdmissionProof(session, s.projectID, s.layout.Namespace(), request); err != nil {
		return err
	}

	preview, err := s.freezeWorkflowSelection(ctx, request.Selection)
	if err != nil {
		return err
	}

	return ValidateGuidedWorkPreview(*session.Guided.Work.Materialization, preview)
}

func (s *WorkflowStore) checkGuidedWorkflowTarget(ctx context.Context, request WorkflowStartRequest) error {
	if request.Admission == nil || request.Admission.Guided == nil {
		return nil
	}

	return s.checkWorkflowCoordinatorTarget(ctx, request.Selection.Target)
}

func guardGuidedAdmissionProof(session AuthoringSession, project, namespace string, request WorkflowStartRequest) error {
	if err := guardGuidedAdmissionState(session, project, namespace); err != nil {
		return err
	}

	if !GuidedDispatchAllowed(session) {
		return ErrGuidedUseCaseStopped
	}

	state := session.Guided

	proof := request.Admission.Guided
	if err := guardGuidedAdmissionIntent(session, *proof); err != nil {
		return err
	}

	if err := guardGuidedAdmissionCheckpoint(session, proof.CheckpointDigest); err != nil {
		return err
	}

	retained := state.Work.Admission

	selection, err := normalizeWorkflowPreviewRequest(retained.Selection)
	if err != nil {
		return err
	}

	if retained.RequestKey != request.RequestKey || !reflect.DeepEqual(selection, request.Selection) || !reflect.DeepEqual(retained.Admission, request.Admission) || request.Policy.IncludeLaterTickets {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardGuidedAdmissionCheckpoint(session AuthoringSession, expected string) error {
	checkpoint, err := GuidedPreparationCheckpoint(session)
	if err != nil {
		return err
	}

	digest, err := GuidedCheckpointDigest(checkpoint)
	if err != nil || digest != expected {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	return nil
}

func guardGuidedAdmissionState(session AuthoringSession, project, namespace string) error {
	state := session.Guided
	if state == nil || !validGuidedIntent(session) || session.ProjectID != project || session.Namespace != namespace || state.Stage != "preparation" || state.Work == nil || state.Work.Materialization == nil || state.Work.Admission == nil || state.PendingOperation == nil {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardGuidedAdmissionIntent(session AuthoringSession, proof GuidedWorkflowAdmission) error {
	state := session.Guided
	if proof.Policy != GuidedWorkAdmissionPolicy || proof.Actor != state.Actor || proof.RunID != session.ID || state.PendingOperation.Action != "admit" || proof.OperationDigest != state.PendingOperation.Digest || proof.MaterializationID != state.Work.Materialization.ID {
		return workflowstate.ErrConflict
	}

	return nil
}
