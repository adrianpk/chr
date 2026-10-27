package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Observe reads current images without allocating artifacts or recovering a
// pending mutation. Torn snapshots fail closed and can be explicitly reloaded.
type observedQuizImages struct {
	quizcore.Observation
	images []workflowstate.FileChange
}

func (r *fileQuizArtifactRepository) Observe(ctx context.Context, scope quizcore.Scope) (quizcore.Observation, error) {
	observed, err := r.observeImages(ctx, scope)

	return observed.Observation, err
}

func (r *fileQuizArtifactRepository) observeImages(ctx context.Context, scope quizcore.Scope) (observedQuizImages, error) {
	var observed observedQuizImages
	if err := ctx.Err(); err != nil {
		return observed, err
	}

	paths, err := r.Paths(scope)
	if err != nil {
		return observed, err
	}

	owner := &fileAuthoringSessionRepository{projectDir: r.projectDir}
	if err := owner.checkQuizRead(); err != nil {
		return observed, err
	}

	for _, reference := range quizImageReferences(paths) {
		image, err := owner.quizImage(reference)
		if err != nil {
			return observed, err
		}

		observed.images = append(observed.images, image)
	}

	if err := owner.checkQuizImages(observed.images); err != nil {
		return observed, err
	}

	observed.Revision, err = owner.quizRevision(scope, observed.images)
	if err != nil {
		return observed, err
	}

	if err := decodeQuizObservation(&observed, scope, paths); err != nil {
		return observed, err
	}

	return observed, nil
}

func quizImageReferences(paths QuizArtifactPaths) []string {
	return []string{
		paths.RuntimeRelativePath, paths.RawRelativePath, paths.CuratedRelativePath,
		filepath.Join(filepath.Dir(paths.RawRelativePath), "metadata.json"),
		filepath.Join(filepath.Dir(paths.CuratedRelativePath), "metadata.json"),
	}
}

func (r *fileAuthoringSessionRepository) quizImage(reference string) (workflowstate.FileChange, error) {
	image := workflowstate.FileChange{Path: filepath.FromSlash(reference)}

	data, err := r.authoringRead(image.Path)
	if errors.Is(err, os.ErrNotExist) {
		return image, nil
	}

	image.After, image.Before = data, contenthash.Sum(data)

	return image, err
}

func (r *fileAuthoringSessionRepository) checkQuizRead() error {
	path, err := r.authoringPath(authoringRuntimeRef("pending.json"))
	if err != nil {
		return err
	}

	_, err = os.Lstat(path)
	if err == nil {
		return workflowstate.ErrConflict
	}

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func (r *fileAuthoringSessionRepository) checkQuizImages(images []workflowstate.FileChange) error {
	for _, image := range images {
		current, err := r.quizImage(image.Path)
		if err != nil {
			return err
		}

		if current.Before != image.Before {
			return workflowstate.ErrConflict
		}
	}

	return r.checkQuizRead()
}

func (r *fileAuthoringSessionRepository) quizRevision(scope quizcore.Scope, images []workflowstate.FileChange) (string, error) {
	root, err := filepath.Abs(r.projectDir)
	if err != nil {
		return "", err
	}

	hashes := []string{root, scope.ProjectID, scope.CaseSlug}
	for _, image := range images {
		hashes = append(hashes, image.Path, contenthash.Sum(image.After))
	}

	return contenthash.Sum([]byte(strings.Join(hashes, "\x00"))), nil
}

func decodeQuizObservation(observed *observedQuizImages, scope quizcore.Scope, paths QuizArtifactPaths) error {
	entries, err := quizcore.ParseRawCaptureEntries(string(observed.images[1].After))
	if err != nil {
		return err
	}

	observed.Entries = entries
	if observed.images[0].After == nil {
		return nil
	}

	state := &observed.State
	if err := json.Unmarshal(observed.images[0].After, state); err != nil {
		return err
	}

	if err := quizcore.ValidateRetainedState(*state); err != nil {
		return err
	}

	if !quizStateMatchesScope(*state, scope, paths) {
		return workflowstate.ErrConflict
	}

	observed.Exists = true

	state.Revision = observed.Revision
	if state.SchemaVersion == 0 {
		// Legacy reads retain their bytes and historical parser interpretation.
		// Only an explicit guarded mutation publishes the versioned state.
		state.SessionID = "legacy-" + observed.images[0].Before
		if state.Mode == quizcore.ModeFast && len(state.Candidates) > 0 {
			state.FastBatchID, state.FastQueue = "legacy-buffer", state.Candidates
			state.Candidates = nil
		}
	}

	return nil
}

func quizStateMatchesScope(state quizcore.RuntimeState, scope quizcore.Scope, paths QuizArtifactPaths) bool {
	return state.ProjectID == scope.ProjectID && state.CaseSlug == scope.CaseSlug && state.RawPath == paths.RawRelativePath && state.CuratedPath == paths.CuratedRelativePath
}
