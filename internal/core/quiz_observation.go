package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Observe reads current images without allocating artifacts or recovering a
// pending mutation. Torn snapshots fail closed and can be explicitly reloaded.
func (r *fileQuizArtifactRepository) Observe(ctx context.Context, scope QuizScope) (QuizObservation, error) {
	var observed QuizObservation
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

func (r *fileAuthoringSessionRepository) quizImage(reference string) (workflowFileChange, error) {
	image := workflowFileChange{Path: filepath.FromSlash(reference)}

	data, err := r.authoringRead(image.Path)
	if errors.Is(err, os.ErrNotExist) {
		return image, nil
	}

	image.After, image.Before = data, workflowHash(data)

	return image, err
}

func (r *fileAuthoringSessionRepository) checkQuizRead() error {
	path, err := r.authoringPath(authoringRuntimeRef("pending.json"))
	if err != nil {
		return err
	}

	_, err = os.Lstat(path)
	if err == nil {
		return ErrWorkflowConflict
	}

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func (r *fileAuthoringSessionRepository) checkQuizImages(images []workflowFileChange) error {
	for _, image := range images {
		current, err := r.quizImage(image.Path)
		if err != nil {
			return err
		}

		if current.Before != image.Before {
			return ErrWorkflowConflict
		}
	}

	return r.checkQuizRead()
}

func (r *fileAuthoringSessionRepository) quizRevision(scope QuizScope, images []workflowFileChange) (string, error) {
	root, err := filepath.Abs(r.projectDir)
	if err != nil {
		return "", err
	}

	hashes := []string{root, scope.ProjectID, scope.CaseSlug}
	for _, image := range images {
		hashes = append(hashes, image.Path, workflowHash(image.After))
	}

	return workflowHash([]byte(strings.Join(hashes, "\x00"))), nil
}

func decodeQuizObservation(observed *QuizObservation, scope QuizScope, paths QuizArtifactPaths) error {
	if observed.images[0].After == nil {
		return nil
	}

	state := &observed.State
	if err := json.Unmarshal(observed.images[0].After, state); err != nil {
		return err
	}

	if err := validateRetainedQuizState(*state); err != nil {
		return err
	}

	if !quizStateMatchesScope(*state, scope, paths) {
		return ErrWorkflowConflict
	}

	entries, err := ParseQuizRawCaptureEntries(string(observed.images[1].After))
	if err != nil {
		return err
	}

	observed.Exists, observed.Entries = true, entries

	state.Revision = observed.Revision
	if state.SchemaVersion == 0 {
		// Legacy reads retain their bytes and historical parser interpretation.
		// Only an explicit guarded mutation publishes the versioned state.
		state.SessionID = "legacy-" + observed.images[0].Before
		if state.Mode == QuizModeFast && len(state.Candidates) > 0 {
			state.FastBatchID, state.FastQueue = "legacy-buffer", state.Candidates
			state.Candidates = nil
		}
	}

	return nil
}

func quizStateMatchesScope(state QuizRuntimeState, scope QuizScope, paths QuizArtifactPaths) bool {
	return state.ProjectID == scope.ProjectID && state.CaseSlug == scope.CaseSlug && state.RawPath == paths.RawRelativePath && state.CuratedPath == paths.CuratedRelativePath
}
