package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// StartGuidedUseCase creates or replays one explicitly authorized guided session.
func StartGuidedUseCase(ctx context.Context, root string, request GuidedUseCaseStart) (session AuthoringSession, err error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		var startErr error

		session, startErr = r.startGuidedUseCase(ctx, request)

		return startErr
	})

	return session, err
}

func (r *fileAuthoringSessionRepository) startGuidedUseCase(ctx context.Context, request GuidedUseCaseStart) (AuthoringSession, error) {
	if _, err := r.authoringSessionPaths(request.RunID); err != nil {
		return AuthoringSession{}, err
	}

	current, err := r.getGuardedSession(ctx, request.RunID)
	if err == nil {
		if err := guardGuidedScope(current, request.GuidedUseCaseScope); err != nil {
			return current, err
		}

		if current.Guided.Actor != request.Actor || current.Guided.ProgressionIntent != request.ProgressionIntent {
			return current, workflowstate.ErrConflict
		}

		return current, nil
	}

	if !errors.Is(err, ErrAuthoringSessionNotFound) {
		return current, err
	}

	now := time.Now().UTC().Truncate(time.Second)

	current = AuthoringSession{
		ID: request.RunID, ProjectID: request.ProjectID, Namespace: request.Namespace,
		Revision: 1, Topic: "New use case", Status: AuthoringSessionStatusActive,
		Mode: AuthoringSessionModeAssisted, Target: AuthoringTargetDump,
		CreatedAt: now, StartedAt: &now, UpdatedAt: now,
		Guided: &GuidedUseCaseState{
			SchemaVersion: 1, Actor: request.Actor,
			ProgressionIntent: request.ProgressionIntent, Stage: "name", Status: "waiting_input",
		},
	}
	if err := r.saveGuidedSession(current, true); err != nil {
		return current, err
	}

	return r.getGuardedSession(ctx, current.ID)
}

// ReadGuidedUseCase reads retained state without recovery, writes or dispatch.
func ReadGuidedUseCase(ctx context.Context, root string, scope GuidedUseCaseScope) (AuthoringSession, error) {
	r := &fileAuthoringSessionRepository{projectDir: root}

	session, err := r.getGuardedSession(ctx, scope.RunID)
	if err != nil {
		return session, err
	}

	return session, guardGuidedScope(session, scope)
}

func guardGuidedScope(session AuthoringSession, scope GuidedUseCaseScope) error {
	if session.Guided == nil || session.ProjectID != scope.ProjectID || session.Namespace != scope.Namespace || session.ID != scope.RunID {
		return workflowstate.ErrConflict
	}

	return nil
}

func (r *fileAuthoringSessionRepository) guidedFileChange(path string, content []byte, create bool) (workflowFileChange, error) {
	relative, err := filepath.Rel(r.projectDir, path)
	if err != nil {
		return workflowFileChange{}, err
	}

	change := workflowFileChange{Path: relative, After: content}
	if create {
		return change, nil
	}

	before, err := r.authoringRead(relative)
	if err != nil {
		return change, err
	}

	change.Before = contenthash.Sum(before)

	return change, nil
}

func (r *fileAuthoringSessionRepository) saveGuidedSession(session AuthoringSession, create bool, extra ...workflowFileChange) error {
	paths, err := r.authoringSessionPaths(session.ID)
	if err != nil {
		return err
	}

	content, err := RenderAuthoringSession(session, paths)
	if err != nil {
		return err
	}

	sessionChange, err := r.guidedFileChange(paths.SessionPath, []byte(content), create)
	if err != nil {
		return err
	}

	transcript, err := r.guidedFileChange(paths.TranscriptPath, []byte("# Transcript\n\n"+session.Guided.Capture), create)
	if err != nil {
		return err
	}
	// Replace the session last so its revision cannot advertise missing input.
	changes := append(extra, transcript, sessionChange)

	if create {
		artifact, err := r.guidedFileChange(paths.ArtifactsPath, []byte("# Artifacts\n\n"), true)
		if err != nil {
			return err
		}

		changes = append([]workflowFileChange{artifact}, changes...)
	}

	if err := r.writeAuthoringJournal(changes); err != nil {
		return err
	}

	return r.recoverAuthoringMutation()
}

// MutateGuidedUseCase serializes name and dump input with the authoring owner.
func MutateGuidedUseCase(ctx context.Context, root string, command GuidedUseCaseCommand) (session AuthoringSession, err error) {
	operation, err := guidedOperation(command)
	if err != nil {
		return session, err
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, command.RunID)
		if err != nil {
			return err
		}

		if err := guardGuidedScope(session, command.GuidedUseCaseScope); err != nil {
			return err
		}

		replay, err := guardGuidedCommand(session, command, operation)
		if err != nil || replay {
			return err
		}

		if err := r.applyGuidedInput(ctx, &session, command, operation); err != nil {
			return err
		}

		session, err = r.getGuardedSession(ctx, command.RunID)

		return err
	})

	return session, err
}

func guardGuidedCommand(session AuthoringSession, command GuidedUseCaseCommand, operation GuidedUseCaseOperation) (bool, error) {
	for _, completed := range session.Guided.CompletedOperations {
		if completed.ID == operation.ID {
			return completed.Digest == operation.Digest, guidedDigestConflict(completed, operation)
		}
	}

	if !GuidedDispatchAllowed(session) {
		return false, ErrGuidedUseCaseStopped
	}

	if guidedControlOwnsID(session.Guided.Control, operation.ID) {
		return false, workflowstate.ErrConflict
	}

	if guidedIdeaOwnsID(session.Guided, operation.ID) {
		return false, workflowstate.ErrConflict
	}

	if pending := session.Guided.PendingOperation; pending != nil {
		return false, guidedDigestConflict(*pending, operation)
	}

	if session.Revision != command.ExpectedRevision || len(session.Guided.CompletedOperations) >= 256 {
		return false, workflowstate.ErrConflict
	}

	return false, nil
}

func guidedDigestConflict(retained, received GuidedUseCaseOperation) error {
	if retained.ID != received.ID || retained.Digest != received.Digest {
		return workflowstate.ErrConflict
	}

	return nil
}

func (r *fileAuthoringSessionRepository) applyGuidedInput(ctx context.Context, session *AuthoringSession, command GuidedUseCaseCommand, operation GuidedUseCaseOperation) error {
	switch command.Action {
	case "name":
		if session.Guided.Stage != "name" {
			return workflowstate.ErrConflict
		}

		if err := r.nameGuidedUseCase(ctx, session, operation); err != nil {
			return err
		}
	case "append":
		if session.Guided.Stage != "dump" || session.Guided.PendingOperation != nil {
			return workflowstate.ErrConflict
		}

		if command.Content == "" {
			return errors.New("capture content is required")
		}

		session.Guided.Capture += command.Content + "\n\n"
	default:
		return errors.New("unsupported guided input action")
	}

	session.Guided.CompletedOperations = append(session.Guided.CompletedOperations, GuidedUseCaseOperation{ID: operation.ID, Action: operation.Action, Digest: operation.Digest})
	session.Guided.PendingOperation = nil
	session.Revision++
	session.UpdatedAt = time.Now().UTC().Truncate(time.Second)

	return r.saveGuidedSession(*session, false)
}

func (r *fileAuthoringSessionRepository) nameGuidedUseCase(ctx context.Context, session *AuthoringSession, operation GuidedUseCaseOperation) error {
	metadata := CaseMetadata{SchemaVersion: CaseMetadataSchemaVersion, Name: operation.Content}
	if len(operation.Content) > 512 {
		return errors.New("use-case name exceeds 512 bytes")
	}

	if err := ValidateCaseMetadata(metadata); err != nil {
		return err
	}

	if session.Guided.PendingOperation == nil {
		session.Guided.PendingOperation = &operation

		session.Revision++
		if err := r.saveGuidedSession(*session, false); err != nil {
			return err
		}
	}

	owner := NewCaseMetadataService(r.projectDir)

	result, err := owner.SaveBound(ctx, CaseMutationRequest{
		Binding:   workflowstate.MutationBinding{RequestID: "guided-" + contenthash.Sum([]byte(session.ID+"-name")), Operator: session.Guided.Actor},
		ProjectID: session.ProjectID, Namespace: session.Namespace, Metadata: metadata,
	})
	if err != nil {
		return err
	}

	session.Guided.CaseSlug, session.Guided.CaseName = result.Metadata.Slug, result.Metadata.Name
	session.Guided.CaseHash = result.Metadata.SourceHash
	session.Guided.Stage, session.Topic = "dump", result.Metadata.Name

	return nil
}

// ListGuidedUseCases inspects scoped sessions without recovering effects.
func ListGuidedUseCases(ctx context.Context, root, project, namespace string) ([]AuthoringSession, error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	if _, err := r.authoringPath(authoringRuntimeRef("sessions")); err != nil {
		return nil, err
	}

	sessions, err := r.listSessions(ctx)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	result := []AuthoringSession{}

	for _, session := range sessions {
		if session.Guided != nil && session.ProjectID == project && session.Namespace == namespace {
			result = append(result, session)
		}
	}

	return result, nil
}
