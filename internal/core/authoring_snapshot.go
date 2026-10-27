package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ObserveContinuation reads a consistent bounded session after journal recovery.
func (r *fileAuthoringSessionRepository) ObserveContinuation(ctx context.Context, id string) (snapshot AuthoringSnapshot, err error) {
	err = WithAuthoringMutation(ctx, r.projectDir, func(context.Context) error {
		snapshot, err = r.readAuthoringSnapshot(id, nil)

		return err
	})

	return snapshot, err
}

func (r *fileAuthoringSessionRepository) readAuthoringSnapshot(id string, images map[string][]byte) (AuthoringSnapshot, error) {
	var snapshot AuthoringSnapshot

	paths, err := r.authoringSessionPaths(id)
	if err != nil {
		return snapshot, err
	}

	relative, err := authoringRelativePaths(r.projectDir, paths)
	if err != nil {
		return snapshot, err
	}

	snapshot.SessionRef = relative.SessionPath

	session, err := r.readAuthoringImage(relative.SessionPath, images)
	if err != nil {
		return snapshot, err
	}

	snapshot.Session, err = parseManagedAuthoringSession(session, id)
	if err != nil {
		return snapshot, err
	}

	transcript, err := r.readAuthoringImage(relative.TranscriptPath, images)
	if err != nil {
		return snapshot, err
	}

	snapshot.Events, err = ParseAssistedTranscriptEvents(string(transcript))
	if err != nil {
		return snapshot, err
	}

	if len(snapshot.Events) > 64 {
		return snapshot, errors.New("authoring transcript capacity exceeded")
	}

	snapshot.SessionHash, snapshot.TranscriptHash = contenthash.Sum(session), contenthash.Sum(transcript)
	if err := r.readAuthoringArtifacts(&snapshot, images); err != nil {
		return snapshot, err
	}

	snapshot.SourceHash, err = authoringSnapshotDigest(snapshot)
	if err != nil {
		return snapshot, err
	}

	snapshot.Question, err = authoringCurrentQuestion(snapshot)

	return snapshot, err
}

func (r *fileAuthoringSessionRepository) readAuthoringImage(relative string, images map[string][]byte) ([]byte, error) {
	if data, found := images[relative]; found {
		return data, nil
	}

	return r.authoringRead(relative)
}

func (r *fileAuthoringSessionRepository) readAuthoringArtifacts(snapshot *AuthoringSnapshot, images map[string][]byte) error {
	refs := append(append([]AuthoringArtifactRef{}, snapshot.Session.Inputs...), snapshot.Session.Outputs...)
	if len(refs) > 16 {
		return errors.New("authoring artifact context capacity exceeded")
	}

	for _, ref := range refs {
		if ref.Kind == AuthoringArtifactKindDump {
			if _, err := r.readAuthoringDumpObservations(snapshot, ref, images); err != nil {
				return err
			}

			continue
		}

		if !validAuthoringArtifactRef(snapshot.Session.Namespace, ref) {
			return workflowstate.ErrConflict
		}

		body, err := r.readAuthoringImage(filepath.FromSlash(ref.Path), images)
		if err != nil {
			return err
		}

		metadata, err := r.readAuthoringImage(filepath.Join(filepath.Dir(ref.Path), "metadata.json"), images)
		if err != nil {
			return err
		}

		snapshot.artifactBytes += len(body) + len(metadata)
		if snapshot.artifactBytes > workflowstate.FileLimit {
			return errors.New("authoring canonical context exceeds 1 MiB")
		}

		snapshot.Artifacts = append(snapshot.Artifacts, AuthoringArtifactObservation{Ref: ref, SourceHash: contenthash.Sum(body), MetadataHash: contenthash.Sum(metadata), Body: string(body)})
	}

	return validateAuthoringArtifactContext(*snapshot)
}

func validateAuthoringArtifactContext(snapshot AuthoringSnapshot) error {
	if len(snapshot.Artifacts)+len(snapshot.dumpAliases) > 32 {
		return errors.New("authoring artifact context capacity exceeded")
	}

	if snapshot.artifactBytes > workflowstate.FileLimit {
		return errors.New("authoring canonical context exceeds 1 MiB")
	}

	return nil
}

func authoringSnapshotDigest(snapshot AuthoringSnapshot) (string, error) {
	guards := []string{snapshot.SessionHash, snapshot.TranscriptHash}
	if snapshot.projectMetadataHash != "" {
		guards = append(guards, snapshot.projectMetadataHash)
	}

	for _, artifact := range append(snapshot.Artifacts[:len(snapshot.Artifacts):len(snapshot.Artifacts)], snapshot.dumpAliases...) {
		guards = append(guards, artifact.Ref.Path, artifact.SourceHash, artifact.MetadataHash)
	}

	encoded, err := json.Marshal(guards)

	return contenthash.Sum(encoded), err
}

func authoringCurrentQuestion(snapshot AuthoringSnapshot) (*AuthoringQuestionBinding, error) {
	var question *AuthoringQuestionBinding

	for _, event := range snapshot.Events {
		switch event.Type {
		case AssistedTranscriptEventQuestion:
			encoded, err := json.Marshal(event.Question)
			if err != nil {
				return nil, err
			}

			question = &AuthoringQuestionBinding{SessionID: snapshot.Session.ID, TurnID: event.TurnID, QuestionID: event.Question.ID, Revision: event.Revision, ContentDigest: contenthash.Sum(encoded)}
		case AssistedTranscriptEventAnswer:
			if question != nil && question.QuestionID == event.Answer.QuestionID {
				question = nil
			}
		}
	}

	return question, nil
}

func (r *fileAuthoringSessionRepository) guardAuthoringSnapshot(request AuthoringContinuationRequest) (AuthoringSnapshot, error) {
	snapshot, err := r.readAuthoringSnapshot(request.SessionID, nil)
	if request.Action == "create" && errors.Is(err, os.ErrNotExist) {
		return newAuthoringSnapshot(request), nil
	}

	if err != nil {
		return snapshot, err
	}

	if snapshot.Session.ProjectID != request.ProjectID || snapshot.Session.Namespace != request.Namespace {
		return snapshot, workflowstate.ErrConflict
	}

	if request.Action == "create" || snapshot.Session.Status != AuthoringSessionStatusActive {
		return snapshot, workflowstate.ErrConflict
	}

	if snapshot.Session.Revision != request.ExpectedRevision || snapshot.SourceHash != request.ExpectedSourceHash {
		return snapshot, workflowstate.ErrConflict
	}

	return snapshot, nil
}

func parseManagedAuthoringSession(content []byte, id string) (AuthoringSession, error) {
	session, err := ParseAuthoringSession(string(content))
	if err != nil {
		return session, err
	}

	if session.ID != id || session.Revision < 1 || !opsnamespace.Valid(session.Namespace) || len(session.Inputs) == 0 {
		return session, workflowstate.ErrConflict
	}

	return session, nil
}
