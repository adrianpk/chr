package authoring

import (
	"context"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// AppendRawCapture retains the legacy explicit repository operation under the
// shared writer exclusion. Interactive actions use ApplyAction for atomic state.
func (r *fileQuizArtifactRepository) AppendRawCapture(ctx context.Context, state quizcore.RuntimeState, entry quizcore.Entry) (QuizArtifactPaths, error) {
	var paths QuizArtifactPaths

	err := WithMutation(ctx, r.projectDir, func(owned context.Context) error {
		if err := r.checkQuizWritePaths(quizScopeFromRuntime(state)); err != nil {
			return err
		}

		var err error

		paths, err = r.appendRawCapture(owned, state, entry)

		return err
	})

	return paths, err
}

// WriteCuratedNotes shares exclusion with interactive quiz and authoring writes.
func (r *fileQuizArtifactRepository) WriteCuratedNotes(ctx context.Context, notes quizcore.CuratedNotes) (QuizArtifactPaths, error) {
	var paths QuizArtifactPaths

	err := WithMutation(ctx, r.projectDir, func(owned context.Context) error {
		if err := r.checkQuizWritePaths(notes.Scope); err != nil {
			return err
		}

		var err error

		paths, err = r.writeCuratedNotes(owned, notes)

		return err
	})

	return paths, err
}

// SaveRuntimeState retains the legacy import boundary under shared exclusion.
// Session actions must use ApplyAction with their observed binding.
func (r *fileQuizArtifactRepository) SaveRuntimeState(ctx context.Context, state quizcore.RuntimeState) (QuizArtifactPaths, error) {
	var paths QuizArtifactPaths

	err := WithMutation(ctx, r.projectDir, func(owned context.Context) error {
		if err := r.checkQuizWritePaths(quizScopeFromRuntime(state)); err != nil {
			return err
		}

		var err error

		paths, err = r.saveRuntimeState(owned, state)

		return err
	})

	return paths, err
}

func (r *fileQuizArtifactRepository) checkQuizWritePaths(scope quizcore.Scope) error {
	paths, err := r.Paths(scope)
	if err != nil {
		return err
	}

	owner := &fileSessionRepository{projectDir: r.projectDir}
	for _, reference := range quizImageReferences(paths) {
		if _, err := owner.authoringPath(reference); err != nil {
			return err
		}
	}

	return nil
}
