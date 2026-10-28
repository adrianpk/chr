package workflow

import (
	"context"
	"errors"
	"reflect"
	"slices"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateGuidedWorkPreview rejects omissions, additions and replaced initial contracts.
func ValidateGuidedWorkPreview(reference authoringcore.GuidedMaterializationReference, preview Preview) error {
	if preview.Selector != (workflowstate.Selector{Extent: "slice_set", Identity: reference.SetID}) || preview.Total != len(reference.Members) || len(preview.Members) != preview.Total || preview.Total == 0 || preview.Total > workflowstate.PageLimit {
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

func guidedInitialMemberMatches(member PreviewMember, expected authoringcore.GuidedWorkMember, set string) bool {
	return member.Key == expected.Key && member.Path == expected.Path && member.Revision == expected.Revision && member.SourceHash == expected.SourceHash && member.ContractHash == expected.ContractHash && member.SetID == set
}

func (s *Store) checkGuidedWorkflowAdmission(ctx context.Context, request StartRequest) error {
	if request.Admission == nil || request.Admission.Guided == nil {
		return nil
	}

	proof := request.Admission.Guided

	session, err := authoringcore.ReadOwnedSession(ctx, s.layout.ProjectDir(), proof.RunID)
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

func (s *Store) checkGuidedWorkflowTarget(ctx context.Context, request StartRequest) error {
	if request.Admission == nil || request.Admission.Guided == nil {
		return nil
	}

	return s.checkWorkflowCoordinatorTarget(ctx, request.Selection.Target)
}

func guardGuidedAdmissionProof(session authoringcore.Session, project, namespace string, request StartRequest) error {
	if err := guardGuidedAdmissionState(session, project, namespace); err != nil {
		return err
	}

	if !authoringcore.GuidedDispatchAllowed(session) {
		return authoringcore.ErrGuidedUseCaseStopped
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

func guardGuidedAdmissionCheckpoint(session authoringcore.Session, expected string) error {
	checkpoint, err := authoringcore.GuidedPreparationCheckpoint(session)
	if err != nil {
		return err
	}

	digest, err := authoringcore.GuidedCheckpointDigest(checkpoint)
	if err != nil || digest != expected {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	return nil
}

func guardGuidedAdmissionState(session authoringcore.Session, project, namespace string) error {
	state := session.Guided
	if state == nil || !authoringcore.ValidGuidedIntent(session) || session.ProjectID != project || session.Namespace != namespace || state.Stage != "preparation" || state.Work == nil || state.Work.Materialization == nil || state.Work.Admission == nil || state.PendingOperation == nil {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardGuidedAdmissionIntent(session authoringcore.Session, proof workflowstate.GuidedAdmission) error {
	state := session.Guided
	if proof.Policy != authoringcore.GuidedWorkAdmissionPolicy || proof.Actor != state.Actor || proof.RunID != session.ID || state.PendingOperation.Action != "admit" || proof.OperationDigest != state.PendingOperation.Digest || proof.MaterializationID != state.Work.Materialization.ID {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateGuidedControlProof(request workflowstate.ControlMutationRequest) error {
	proof := request.Guided
	if proof == nil {
		return nil
	}

	if !authoringcore.ValidGuidedControlAction(request.Action) || authoringstate.ValidateSessionID(proof.RunID) != nil || !validWorkflowNativeFingerprint(proof.OperationDigest) {
		return workflowstate.ErrConflict
	}

	return nil
}

type guidedAdmissionLookup struct {
	Invocation Invocation
	Found      bool
}

func (s *Store) lookupGuidedAdmission(ctx context.Context, session authoringcore.Session) (result guidedAdmissionLookup, err error) {
	err = s.withLock(ctx, func() error {
		runs, err := s.repositoryWorkflowInvocations(ctx)
		if err != nil {
			return err
		}

		request := session.Guided.Work.Admission
		for _, run := range runs {
			if run.Namespace != session.Namespace || !slices.Contains(run.RequestKeys, request.RequestKey) {
				continue
			}

			if run.ProjectID != session.ProjectID || !reflect.DeepEqual(run.Admission, request.Admission) {
				return workflowstate.ErrConflict
			}

			result = guidedAdmissionLookup{Invocation: run.Invocation, Found: true}

			return nil
		}

		return nil
	})

	return result, err
}

func findGuidedControlInvocation(ctx context.Context, root string, session authoringcore.Session) (guidedAdmissionLookup, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: root, Namespace: session.Namespace})
	if err != nil {
		return guidedAdmissionLookup{}, err
	}

	return NewWorkflowStore(layout, session.ProjectID).lookupGuidedAdmission(ctx, session)
}
