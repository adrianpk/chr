package core

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func normalizeAssistedDumpSource(opts *AssistedArtifactSynthesisOptions) error {
	expected := filepath.Join(AuthoringSessionLayout(opts.ProjectDir, opts.Session.ID).SessionDir, "curated-summary.md")

	path := filepath.FromSlash(opts.SourceRef.Path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(opts.ProjectDir, path)
	}

	if path != expected || opts.SourceRef.ID != "curated-summary-"+opts.Session.ID || opts.SourceRef.Kind != AuthoringArtifactKindDraft || opts.SourceRef.Role != AuthoringArtifactRoleOutput {
		return workflowstate.ErrConflict
	}

	relative, err := filepath.Rel(opts.ProjectDir, path)
	opts.SourceRef.Path = filepath.ToSlash(relative)

	return err
}

func (r *fileAuthoringSessionRepository) assistedDumpSourceImage(opts AssistedArtifactSynthesisOptions) (workflowFileChange, error) {
	content, err := r.readAuthoringDocument(opts.SourceRef.Path)
	if errors.Is(err, os.ErrNotExist) {
		return workflowFileChange{}, ErrAssistedCurationMissing
	}

	if err != nil {
		return workflowFileChange{}, err
	}

	fields, err := projectDumpBodyFields(content)
	if err != nil {
		return workflowFileChange{}, err
	}

	generated, err := time.Parse(time.RFC3339, fields["generated_at"])
	if err != nil {
		return workflowFileChange{}, workflowstate.ErrConflict
	}

	summary := opts.CuratedSummary

	summary.GeneratedAt = generated
	if fields["session_id"] != summary.SessionID || string(content) != RenderAssistedCuratedAnswerSummary(summary) {
		return workflowFileChange{}, workflowstate.ErrConflict
	}

	return workflowFileChange{Path: filepath.FromSlash(opts.SourceRef.Path), Before: contenthash.Sum(content), After: content}, nil
}

func (r *fileAuthoringSessionRepository) assistedDumpSessionImage(session AuthoringSession) (workflowFileChange, error) {
	paths, err := r.authoringSessionPaths(session.ID)
	if err != nil {
		return workflowFileChange{}, err
	}

	relative, err := authoringRelativePaths(r.projectDir, paths)
	if err != nil {
		return workflowFileChange{}, err
	}

	content, err := r.readAuthoringDocument(relative.SessionPath)
	if err != nil {
		return workflowFileChange{}, err
	}

	full, err := RenderAuthoringSession(session, paths)
	if err != nil {
		return workflowFileChange{}, err
	}

	local, err := RenderAuthoringSession(session, relative)
	if err != nil {
		return workflowFileChange{}, err
	}

	if !bytes.Equal(content, []byte(full)) && !bytes.Equal(content, []byte(local)) {
		return workflowFileChange{}, workflowstate.ErrConflict
	}

	return workflowFileChange{Path: relative.SessionPath, Before: contenthash.Sum(content)}, nil
}
