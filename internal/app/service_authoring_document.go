package app

import (
	"context"
	"errors"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

var errAuthoringDocumentUnavailable = errors.New("requested original authoring document is unavailable")

func validAuthoringDocumentKind(kind authoringcore.ArtifactKind) bool {
	switch kind {
	case authoringcore.ArtifactKindDump, authoringcore.ArtifactKindSpec, authoringcore.ArtifactKindPlan:
		return true
	default:
		return false
	}
}

func (s service) authoringDocument(ctx context.Context, sessionID string, kind authoringcore.ArtifactKind) (authoringcore.ArtifactObservation, error) {
	snapshot, err := s.ObserveAuthoring(ctx, sessionID)
	if err != nil {
		return authoringcore.ArtifactObservation{}, err
	}

	if snapshot.Session.ProjectID != s.projectContext.ProjectID {
		return authoringcore.ArtifactObservation{}, workflowstate.ErrConflict
	}

	for index := len(snapshot.Artifacts) - 1; index >= 0; index-- {
		artifact := snapshot.Artifacts[index]
		if artifact.Ref.Kind == kind && artifact.Body != "" {
			return artifact, nil
		}
	}

	return authoringcore.ArtifactObservation{}, errAuthoringDocumentUnavailable
}

// ReadPreparedAuthoringDocument reads the exact selected canonical artifact.
// A changed source is a conflict; an excerpt or generated answer is never a substitute.
func (s service) ReadPreparedAuthoringDocument(ctx context.Context, prepared PreparedAuthoringExplanation) (authoringcore.ArtifactObservation, error) {
	if err := validateAuthoringExplanation(prepared.Request, s.projectContext); err != nil {
		return authoringcore.ArtifactObservation{}, err
	}

	if prepared.ProjectID != s.projectContext.ProjectID || !validAuthoringDocumentKind(prepared.Request.DocumentKind) || len(prepared.Sources) != 1 {
		return authoringcore.ArtifactObservation{}, workflowstate.ErrConflict
	}

	artifact, err := s.authoringDocument(ctx, prepared.Request.SessionID, prepared.Request.DocumentKind)
	if err != nil {
		return artifact, err
	}

	source := prepared.Sources[0]
	if artifact.Ref.Path != source.Reference || artifact.SourceHash != source.SourceHash {
		return authoringcore.ArtifactObservation{}, workflowstate.ErrConflict
	}

	return artifact, nil
}
