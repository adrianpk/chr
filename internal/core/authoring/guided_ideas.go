package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedLaterIdea is an attributable request, excluded from all sealed inputs.
type GuidedLaterIdea struct {
	GuidedIdeaReference
	Content string
	State   string
}

// RetainGuidedIdea writes one bounded idea and its reference through the authoring
// journal, without replacing a pending stage, control, checkpoint or native scope.
func RetainGuidedIdea(ctx context.Context, root string, command GuidedUseCaseCommand) (session Session, err error) {
	operation, err := guidedOperation(command)
	if err != nil || command.Action != "idea" || strings.TrimSpace(command.Content) == "" || len(command.Content) > 8192 {
		return session, errors.Join(workflowstate.ErrConflict, err)
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, command.RunID)
		if err != nil {
			return err
		}

		if err := ValidateGuidedScope(session, command.GuidedUseCaseScope); err != nil {
			return err
		}

		replay, err := guardGuidedIdea(session, operation)
		if err != nil || replay {
			return err
		}

		return r.retainGuidedIdea(&session, operation)
	})

	return session, err
}

func guardGuidedIdea(session Session, operation GuidedUseCaseOperation) (bool, error) {
	for _, idea := range session.Guided.LaterIdeas {
		if idea.ID == operation.ID {
			if idea.Digest != operation.Digest {
				return false, workflowstate.ErrConflict
			}

			return true, nil
		}
	}

	if session.Revision != operation.ExpectedRevision || len(session.Guided.LaterIdeas) >= 64 || len(session.Guided.Checkpoints) == 0 {
		return false, workflowstate.ErrConflict
	}

	if guidedControlOwnsID(session.Guided.Control, operation.ID) {
		return false, workflowstate.ErrConflict
	}

	return false, guardGuidedControlIdentity(session.Guided, operation.ID)
}

func guidedIdeaOwnsID(state *GuidedUseCaseState, id string) bool {
	for _, idea := range state.LaterIdeas {
		if idea.ID == id {
			return true
		}
	}

	return false
}

func (r *fileAuthoringSessionRepository) retainGuidedIdea(session *Session, operation GuidedUseCaseOperation) error {
	idea := GuidedLaterIdea{GuidedIdeaReference: GuidedIdeaReference{ID: operation.ID, Digest: operation.Digest, Actor: session.Guided.Actor, CreatedAt: time.Now().UTC()}, Content: operation.Content, State: "pending_review"}
	idea.Path = authoringRuntimeRef(filepath.Join("sessions", session.ID, "later-ideas", operation.ID+".json"))

	encoded, err := json.Marshal(idea)
	if err != nil {
		return err
	}

	idea.Hash = contenthash.Sum(encoded)

	path, err := r.authoringPath(idea.Path)
	if err != nil {
		return err
	}

	change, err := r.guidedFileChange(path, encoded, true)
	if err != nil {
		return err
	}

	session.Guided.LaterIdeas = append(session.Guided.LaterIdeas, idea.GuidedIdeaReference)
	session.Revision++
	session.UpdatedAt = idea.CreatedAt

	return r.saveGuidedSession(*session, false, change)
}

// ReadGuidedIdeas verifies original separately retained entries without recovery.
func ReadGuidedIdeas(ctx context.Context, root string, session Session) ([]GuidedLaterIdea, error) {
	if session.Guided == nil {
		return nil, workflowstate.ErrConflict
	}

	r := &fileAuthoringSessionRepository{projectDir: root}

	result := make([]GuidedLaterIdea, 0, len(session.Guided.LaterIdeas))
	for _, ref := range session.Guided.LaterIdeas {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		idea, err := r.readGuidedIdea(session.ID, ref)
		if err != nil {
			return nil, err
		}

		result = append(result, idea)
	}

	return result, nil
}

func (r *fileAuthoringSessionRepository) readGuidedIdea(id string, ref GuidedIdeaReference) (GuidedLaterIdea, error) {
	if ref.Path != authoringRuntimeRef(filepath.Join("sessions", id, "later-ideas", ref.ID+".json")) {
		return GuidedLaterIdea{}, workflowstate.ErrConflict
	}

	data, err := r.authoringRead(ref.Path)
	if err != nil || contenthash.Sum(data) != ref.Hash {
		return GuidedLaterIdea{}, errors.Join(workflowstate.ErrConflict, err)
	}

	var idea GuidedLaterIdea
	if err := authoringstate.Decode(data, &idea); err != nil {
		return idea, err
	}

	expected := ref

	expected.Hash = ""
	if idea.GuidedIdeaReference != expected || idea.State != "pending_review" || strings.TrimSpace(idea.Content) == "" {
		return idea, workflowstate.ErrConflict
	}

	idea.Hash = ref.Hash

	return idea, nil
}
