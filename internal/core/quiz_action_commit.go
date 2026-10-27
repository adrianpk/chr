package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ApplyAction compares the original binding and publishes all quiz images and
// its acknowledgement through the existing recoverable authoring journal.
func (r *fileQuizArtifactRepository) ApplyAction(ctx context.Context, request QuizActionRequest, transition QuizTransitionFunc) (QuizActionResult, error) {
	var result QuizActionResult
	if err := validateQuizAction(request); err != nil {
		return result, err
	}

	err := WithAuthoringMutation(ctx, r.projectDir, func(owned context.Context) error {
		owner := &fileAuthoringSessionRepository{projectDir: r.projectDir, afterMutationReplace: r.afterMutationReplace}

		digest, err := owner.quizActionDigest(request)
		if err != nil {
			return err
		}

		retained, exists, err := owner.quizReceipt(request.Binding.OperationID, digest)
		if err != nil || exists {
			result = retained

			return err
		}

		if err := owner.projectDumpOperationAvailable(request.Binding.OperationID); err != nil {
			return err
		}

		before, err := r.Observe(owned, request.Scope)
		if err != nil {
			return err
		}

		if err := compareQuizBinding(before, request); err != nil {
			return err
		}

		next, err := prepareBoundedQuizTransition(owned, before, request, transition)
		if err != nil {
			return err
		}

		result, err = owner.commitQuizTransition(owned, before, next, request, digest)

		return err
	})

	return result, err
}

func compareQuizBinding(before QuizObservation, request QuizActionRequest) error {
	binding := request.Binding
	if binding.ExpectedRevision != before.Revision || binding.SessionID != before.State.SessionID {
		return workflowstate.ErrConflict
	}

	questionID := ""
	if before.State.ActiveQuestion != nil {
		questionID = before.State.ActiveQuestion.ID
	}

	if binding.QuestionID != questionID || (!before.Exists && request.Action != "start") {
		return workflowstate.ErrConflict
	}

	return nil
}

func prepareBoundedQuizTransition(ctx context.Context, before QuizObservation, request QuizActionRequest, transition QuizTransitionFunc) (QuizTransition, error) {
	bounded, cancel := context.WithTimeout(ctx, quizProviderTimeout)
	defer cancel()

	next, err := transition(bounded, before)
	if err != nil {
		return next, err
	}

	return prepareQuizTransition(before, next, request)
}

func (r *fileAuthoringSessionRepository) commitQuizTransition(ctx context.Context, before QuizObservation, next QuizTransition, request QuizActionRequest, digest string) (QuizActionResult, error) {
	result := QuizActionResult{OperationID: request.Binding.OperationID, State: next.State, Message: next.Message}

	changes, err := r.quizTransitionChanges(before, next, request.Scope)
	if err != nil {
		return result, err
	}

	result.State.Revision, err = r.quizRevision(request.Scope, changes)
	if err != nil {
		return result, err
	}

	receipt, err := quizReceiptChange(digest, result)
	if err != nil {
		return result, err
	}

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, append(changes, receipt)); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}

func (r *fileAuthoringSessionRepository) quizTransitionChanges(before QuizObservation, next QuizTransition, scope quizcore.QuizScope) ([]workflowFileChange, error) {
	changes := append([]workflowFileChange(nil), before.images...)

	encoded, err := json.MarshalIndent(next.State, "", "  ")
	if err != nil {
		return nil, err
	}

	changes[0].After = append(encoded, '\n')

	entries := append([]quizcore.QuizEntry(nil), before.Entries...)
	if next.Entry != nil {
		entries = append(entries, *next.Entry)
		changes[1].After = append(append([]byte(nil), changes[1].After...), []byte(RenderQuizRawCapture(next.State, *next.Entry))...)
	}

	paths, err := QuizArtifactLayout(r.projectDir, scope)
	if err != nil {
		return nil, err
	}

	changes[2].After = []byte(RenderQuizCuratedNotes(paths, quizcore.CurateQuizEntries(scope, entries, next.State.UpdatedAt)))
	for index := 3; index < len(changes); index++ {
		changes[index].After, err = r.quizMetadata(changes[index], next.State.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	// Empty raw captures are retained on session creation, keeping every journal
	// image present and its revision independent of filesystem absence.
	if changes[1].After == nil {
		changes[1].After = []byte{}
	}

	return changes, nil
}

func (r *fileAuthoringSessionRepository) quizMetadata(image workflowFileChange, now time.Time) ([]byte, error) {
	metadata, err := r.quizMetadataBefore(image, now)
	if err != nil {
		return nil, err
	}

	metadata["updated_at"] = now.UTC().Format(time.RFC3339Nano)

	return json.MarshalIndent(metadata, "", "  ")
}

func (r *fileAuthoringSessionRepository) quizMetadataBefore(image workflowFileChange, now time.Time) (map[string]any, error) {
	if image.After == nil {
		return r.newQuizMetadata(image.Path, now)
	}

	var metadata map[string]any
	if err := json.Unmarshal(image.After, &metadata); err != nil {
		return nil, err
	}

	if metadata == nil {
		return nil, errors.New("quiz metadata is empty")
	}

	if err := advanceAuthoringMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (r *fileAuthoringSessionRepository) newQuizMetadata(reference string, now time.Time) (map[string]any, error) {
	parts := strings.Split(filepath.ToSlash(reference), "/")

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: r.projectDir, Namespace: parts[1]})
	if err != nil {
		return nil, err
	}

	return newOpsArtifactMetadata(CreateOpsArtifactOptions{Layout: layout, Type: "quiz", Title: strings.TrimPrefix(parts[3], "00000-"), Producer: "chr"}, parts[3], now), nil
}
