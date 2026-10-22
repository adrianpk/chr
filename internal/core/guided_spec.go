package core

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
)

func (r *fileAuthoringSessionRepository) guardGuidedSpecApproval(ctx context.Context, request SpecMutationRequest, document SpecDocument) error {
	approval := request.GuidedApproval
	if approval == nil {
		return nil
	}

	if err := guardGuidedApprovalDocument(request, document); err != nil {
		return err
	}

	session, err := r.getGuardedSession(ctx, approval.RunID)
	if err != nil {
		return err
	}

	if err := guardGuidedApprovalRun(session, request); err != nil {
		return err
	}

	return guardGuidedApprovalSource(session, request, document)
}

func guardGuidedApprovalDocument(request SpecMutationRequest, document SpecDocument) error {
	approval := request.GuidedApproval
	if request.Status != "approved" || request.Body != document.Body || request.Title != document.Title {
		return ErrWorkflowConflict
	}

	if approval.GeneratedRevision != document.Revision || approval.ContentHash != workflowHash([]byte(document.Body)) {
		return ErrWorkflowConflict
	}

	return nil
}

func guardGuidedApprovalSource(session AuthoringSession, request SpecMutationRequest, document SpecDocument) error {
	approval := request.GuidedApproval

	draft := session.Guided.SpecDraft
	if draft == nil || draft.Reference != document.Reference || draft.Revision != document.Revision || draft.ContentHash != approval.ContentHash {
		return ErrWorkflowConflict
	}

	checkpoint, err := guidedStageCheckpoint(session, "quiz")
	if err != nil {
		return err
	}

	digest, err := GuidedCheckpointDigest(checkpoint)
	if err != nil {
		return err
	}

	if digest != approval.InputDigest {
		return ErrWorkflowConflict
	}

	return nil
}

func guardGuidedApprovalRun(session AuthoringSession, request SpecMutationRequest) error {
	approval := request.GuidedApproval
	if approval == nil || validateSpecReference(request.Reference) != nil {
		return ErrWorkflowConflict
	}

	scope := GuidedUseCaseScope{RunID: approval.RunID, ProjectID: request.ProjectID, Namespace: strings.Split(request.Reference, "/")[1]}
	if err := guardGuidedScope(session, scope); err != nil {
		return err
	}

	state := session.Guided
	if state.Stage != "specification" || !state.ProgressionIntent || state.Actor != approval.Actor || approval.Policy != GuidedSpecApprovalPolicy {
		return ErrWorkflowConflict
	}

	return guardGuidedPreparedSpec(state.PendingOperation, request)
}

func guardGuidedPreparedSpec(operation *GuidedUseCaseOperation, request SpecMutationRequest) error {
	approval := request.GuidedApproval

	if operation == nil || operation.Action != "spec" || operation.PreparedSpec == nil {
		return ErrWorkflowConflict
	}

	prepared := operation.PreparedSpec
	if prepared.Title != request.Title || prepared.Body != request.Body || prepared.InputDigest != approval.InputDigest {
		return ErrWorkflowConflict
	}

	return nil
}

func guidedStageCheckpoint(session AuthoringSession, stage string) (GuidedStageCheckpoint, error) {
	for _, checkpoint := range session.Guided.Checkpoints {
		if checkpoint.Stage == stage && len(checkpoint.Artifacts) > 0 {
			return checkpoint, nil
		}
	}

	return GuidedStageCheckpoint{}, errors.New("required guided checkpoint is unavailable")
}

// SealGuidedSpec retains the exact approved revision and committed dependency chain.
func SealGuidedSpec(ctx context.Context, root string, session *AuthoringSession) error {
	approval := session.Guided.SpecApproval
	if approval == nil || session.Guided.PendingOperation == nil {
		return ErrWorkflowConflict
	}

	source, err := guidedStageCheckpoint(*session, "quiz")
	if err != nil {
		return err
	}

	if session.Guided.PendingCheckpoint == nil {
		document, err := ReadSpecDocument(ctx, root, approval.Reference)
		if err != nil {
			return err
		}

		if document.Revision != approval.Revision || document.Status != "approved" || workflowHash([]byte(document.Body)) != approval.ContentHash {
			return ErrWorkflowConflict
		}
	}

	paths := []string{approval.Reference, filepath.ToSlash(filepath.Join(filepath.Dir(approval.Reference), "metadata.json"))}
	if err := sealGuidedStage(ctx, root, session, "specification", paths, source); err != nil {
		return err
	}

	session.Guided.Stage, session.Guided.Status = "preparation", "blocked"
	session.Guided.Blocker = "Specification is approved. Continue the retained run to prepare its native work."

	return nil
}
