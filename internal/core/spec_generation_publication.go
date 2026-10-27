package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SpecTargetRevision binds generation to either the current document or its
// observed absence. It does not allocate an envelope or recover a journal.
func SpecTargetRevision(ctx context.Context, projectDir, reference string) (string, error) {
	document, err := observeSpecTarget(ctx, projectDir, reference)

	return document.Revision, err
}

func observeSpecTarget(ctx context.Context, projectDir, reference string) (SpecDocument, error) {
	document, err := ReadSpecDocument(ctx, projectDir, reference)
	if !errors.Is(err, os.ErrNotExist) {
		return document, err
	}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}

	document = SpecDocument{Reference: reference}
	for _, path := range []string{reference, filepath.Join(filepath.Dir(reference), "metadata.json")} {
		if _, err := r.authoringRead(path); !errors.Is(err, os.ErrNotExist) {
			return document, errors.Join(workflowstate.ErrConflict, err)
		}

		document.images = append(document.images, workflowstate.FileChange{Path: path})
	}

	root, err := filepath.Abs(projectDir)
	if err != nil {
		return document, err
	}

	document.Revision = contenthash.Sum([]byte(strings.Join([]string{root, reference, "", ""}, "\x00")))

	return document, r.checkProjectDumpInspection(ctx)
}

func publishGeneratedSpec(ctx context.Context, projectDir string, paths GeneratedSpecPaths, spec GeneratedSpec) error {
	return publishSpecContent(ctx, projectDir, paths, spec, RenderGeneratedSpec(spec))
}

func publishSpecContent(ctx context.Context, projectDir string, paths GeneratedSpecPaths, spec GeneratedSpec, content string) error {
	document, err := observeSpecTarget(ctx, projectDir, paths.RelativePath)
	if err != nil {
		return err
	}

	if err := validateSpecGenerationTarget(document, spec); err != nil {
		return err
	}

	document, action, err := prepareGeneratedSpecDocument(projectDir, paths, document, spec, content)
	if err != nil {
		return err
	}

	changes, err := specDocumentChanges(document, SpecMutationRequest{Action: action, ProjectID: spec.Scope.ProjectID, CaseSlug: spec.Scope.CaseSlug, Title: spec.Title, Body: document.Body, Status: "draft"})
	if err != nil {
		return err
	}

	r := &fileAuthoringSessionRepository{projectDir: projectDir}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, changes); err != nil {
		return err
	}

	return r.recoverAuthoringMutation()
}

func mergeGeneratedSpecFields(front string, fields map[string]string) string {
	lines := strings.Split(front, "\n")

	keys := make([]string, 0, len(fields))
	for key := range fields {
		if key != "created_at" {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	for _, key := range keys {
		lines = replaceProjectMetadataField(lines, key, yamlString(fields[key]))
	}

	// Creation time remains owned by the existing envelope on regeneration.
	return strings.Join(replaceProjectMetadataField(lines, "updated_at", time.Now().UTC().Format(time.RFC3339Nano)), "\n")
}

func prepareGeneratedSpecDocument(projectDir string, paths GeneratedSpecPaths, document SpecDocument, spec GeneratedSpec, content string) (SpecDocument, string, error) {
	action := "save"
	if document.metadata == nil {
		action = "create"
		parts := strings.Split(paths.RelativePath, "/")

		layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: projectDir, Namespace: parts[1]})
		if err != nil {
			return document, action, err
		}

		document.metadata = opscore.NewArtifactMetadata(opscore.ArtifactOptions{Layout: layout, Type: "spec", Title: spec.Title, Producer: "chr"}, parts[3], spec.CreatedAt)
	}

	front, body, err := splitWorkflowMarkdown([]byte(content))
	if err != nil {
		return document, action, err
	}

	fields, err := projectDocumentFields(front)
	if err != nil {
		return document, action, err
	}

	document.front = mergeGeneratedSpecFields(document.front, fields)
	document.Title, document.Body = spec.Title, string(body)
	document.provenance, document.heading = "", false
	separateSpecProvenance(&document)

	return document, action, nil
}

func validateSpecGenerationTarget(document SpecDocument, spec GeneratedSpec) error {
	if spec.ExpectedRevision != "" && document.Revision != spec.ExpectedRevision {
		return workflowstate.ErrConflict
	}

	if document.metadata == nil {
		return nil
	}

	if document.CaseSlug != spec.Scope.CaseSlug {
		return workflowstate.ErrConflict
	}

	if document.ProjectID != "" && document.ProjectID != spec.Scope.ProjectID {
		return workflowstate.ErrConflict
	}

	return nil
}
