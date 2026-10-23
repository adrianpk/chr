package core

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

// FinishGuidedUseCaseDump seals meaningful capture without dispatching later stages.
func FinishGuidedUseCaseDump(ctx context.Context, root string, command GuidedUseCaseCommand) (session AuthoringSession, err error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.finishGuidedUseCaseDump(ctx, command)

		return err
	})

	return session, err
}

func (r *fileAuthoringSessionRepository) finishGuidedUseCaseDump(ctx context.Context, command GuidedUseCaseCommand) (AuthoringSession, error) {
	session, err := r.getGuardedSession(ctx, command.RunID)
	if err != nil {
		return session, err
	}

	if err := guardGuidedScope(session, command.GuidedUseCaseScope); err != nil {
		return session, err
	}

	operation, err := guidedOperation(command)
	if err != nil {
		return session, err
	}

	replay, err := guardGuidedCommand(session, command, operation)
	if err != nil || replay {
		return session, err
	}

	return r.completeGuidedCapture(ctx, session, command, operation)
}

func (r *fileAuthoringSessionRepository) completeGuidedCapture(ctx context.Context, session AuthoringSession, command GuidedUseCaseCommand, operation GuidedUseCaseOperation) (AuthoringSession, error) {
	if command.Action != "finish" || session.Guided.Stage != "dump" {
		return session, ErrWorkflowConflict
	}

	if session.Guided.PendingOperation == nil && command.Content != "" {
		session.Guided.Capture = command.Content
	}

	if strings.TrimSpace(session.Guided.Capture) == "" {
		return session, errors.New("provide meaningful capture before finishing the dump")
	}

	if session.Guided.PendingOperation == nil {
		session.Guided.PendingOperation = &operation

		session.Revision++
		if err := r.saveGuidedSession(session, false); err != nil {
			return session, err
		}
	}

	checkpoint, err := r.sealGuidedDump(ctx, &session, operation)
	if err != nil {
		return session, err
	}

	session.Guided.Checkpoints = append(session.Guided.Checkpoints, checkpoint)
	session.Guided.PendingOperation, session.Guided.PendingCheckpoint = nil, nil
	session.Guided.CompletedOperations = append(session.Guided.CompletedOperations, GuidedUseCaseOperation{ID: operation.ID, Action: operation.Action, Digest: operation.Digest})
	session.Guided.Stage, session.Guided.Status = "quiz", "blocked"
	session.Revision++

	session.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	if err := r.saveGuidedSession(session, false); err != nil {
		return session, err
	}

	return r.getGuardedSession(ctx, session.ID)
}

func (r *fileAuthoringSessionRepository) sealGuidedDump(ctx context.Context, session *AuthoringSession, operation GuidedUseCaseOperation) (GuidedStageCheckpoint, error) {
	if session.Guided.PendingCheckpoint == nil {
		checkpoint, err := r.prepareGuidedDump(ctx, *session, operation)
		if err != nil {
			return checkpoint, err
		}

		session.Guided.PendingCheckpoint = &checkpoint

		session.Revision++
		if err := r.saveGuidedSession(*session, false); err != nil {
			return checkpoint, err
		}
	}

	pending := session.Guided.PendingCheckpoint
	if pending.OperationID != operation.ID || pending.Ref != guidedCheckpointRef(session.ID, operation.Digest) {
		return *pending, ErrWorkflowConflict
	}

	checkpoint, err := r.commitGuidedCheckpoint(ctx, *pending, operation.Digest)
	if err != nil {
		return checkpoint, err
	}

	if r.afterGuidedCheckpoint != nil {
		return checkpoint, r.afterGuidedCheckpoint()
	}

	return checkpoint, nil
}

func (r *fileAuthoringSessionRepository) prepareGuidedDump(ctx context.Context, session AuthoringSession, operation GuidedUseCaseOperation) (GuidedStageCheckpoint, error) {
	request := DumpMutationRequest{OperationID: "guided-" + workflowHash([]byte(session.ID+"-dump")), Action: "capture", Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug, Title: session.Guided.CaseName, Content: session.Guided.Capture, ExpectedRevision: session.Guided.CaseHash}

	result, err := MutateDump(ctx, r.projectDir, request, nil)
	if err != nil {
		return GuidedStageCheckpoint{}, err
	}

	artifact, err := ReadDumpDocument(ctx, r.projectDir, result.Destination)
	if err != nil {
		return GuidedStageCheckpoint{}, err
	}

	if artifact.Artifact.SourceHash != result.SourceHash || artifact.Artifact.MetadataHash != result.MetadataHash {
		return GuidedStageCheckpoint{}, ErrWorkflowConflict
	}

	fields, err := projectDumpBodyFields([]byte(artifact.Content))
	if err != nil {
		return GuidedStageCheckpoint{}, err
	}

	paths := []string{result.Destination, filepath.ToSlash(filepath.Join(filepath.Dir(result.Destination), "metadata.json")), fields["source"], filepath.ToSlash(filepath.Join(filepath.Dir(fields["source"]), "metadata.json"))}
	casePath := CaseMetadataPath(r.projectDir, session.Guided.CaseSlug)

	relative, err := filepath.Rel(r.projectDir, casePath)
	if err != nil {
		return GuidedStageCheckpoint{}, err
	}

	paths = append(paths, filepath.ToSlash(relative))

	checkpoint, err := r.observeGuidedCheckpoint(ctx, session.ID, operation, "dump", paths)
	if err != nil {
		return checkpoint, err
	}

	return checkpoint, guardGuidedCaseImage(checkpoint, filepath.ToSlash(relative), session.Guided.CaseHash)
}

func guardGuidedCaseImage(checkpoint GuidedStageCheckpoint, path, hash string) error {
	for _, image := range checkpoint.Artifacts {
		if image.Path == path && image.Hash != hash {
			return ErrWorkflowConflict
		}
	}

	return nil
}
