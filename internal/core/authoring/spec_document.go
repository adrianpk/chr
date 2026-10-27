package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SpecDocument is one concrete spec, with its editable body separated from
// managed frontmatter and retained generation provenance.
type SpecDocument struct {
	Reference, Revision, Title, Body, Status, ProjectID, CaseSlug string
	front, provenance                                             string
	heading                                                       bool
	images                                                        []workflowstate.FileChange
	metadata                                                      map[string]any
}

// NormalizeSpecStatus interprets native approval aliases without changing files.
func NormalizeSpecStatus(status string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "draft":
		return "draft", nil
	case "approved", "accepted", "final":
		return "approved", nil
	default:
		return "", errors.New("unsupported spec lifecycle")
	}
}

func validateSpecReference(reference string) error {
	if filepath.ToSlash(filepath.Clean(reference)) != reference || !filepath.IsLocal(reference) {
		return errors.New("invalid spec reference")
	}

	parts := strings.Split(reference, "/")
	if len(parts) != 5 || parts[0] != opscore.DirectoryName || !opsnamespace.Valid(parts[1]) || parts[2] != "spec" || parts[4] != opscore.ArtifactBodyFile() {
		return errors.New("spec reference must identify a canonical spec artifact")
	}

	return nil
}

// ReadSpecDocument observes body and metadata together without recovery or writes.
func ReadSpecDocument(ctx context.Context, projectDir, reference string) (SpecDocument, error) {
	if err := validateSpecReference(reference); err != nil {
		return SpecDocument{}, err
	}

	r := &fileSessionRepository{projectDir: projectDir}
	if err := r.checkProjectDumpInspection(ctx); err != nil {
		return SpecDocument{}, err
	}

	document := SpecDocument{Reference: reference}
	for _, path := range []string{reference, filepath.Join(filepath.Dir(reference), "metadata.json")} {
		data, err := r.readAuthoringDocument(path)
		if err != nil {
			return document, err
		}

		document.images = append(document.images, workflowstate.FileChange{Path: path, Before: contenthash.Sum(data), After: data})
	}

	if err := decodeSpecDocument(&document); err != nil {
		return document, err
	}

	root, err := filepath.Abs(projectDir)
	if err != nil {
		return document, err
	}

	document.Revision = contenthash.Sum([]byte(strings.Join([]string{root, reference, document.images[0].Before, document.images[1].Before}, "\x00")))

	return document, r.checkQuizImages(document.images)
}

func decodeSpecDocument(document *SpecDocument) error {
	if err := decodeSpecMetadata(document); err != nil {
		return err
	}

	parsedMarkdown, err := opsmetadata.SplitMarkdown(document.images[0].After)
	front, body := parsedMarkdown.Front, parsedMarkdown.Body

	if err != nil {
		return err
	}

	fields, err := ProjectDocumentFields(front)
	if err != nil {
		return err
	}

	document.front, document.Title = front, fields["title"]
	if document.Title == "" {
		document.Title, err = specMetadataString(document.metadata, "title")
		if err != nil {
			return err
		}
	}

	document.ProjectID = fields["project_id"]

	document.CaseSlug, err = projectDumpScope(fields["case_slug"], fields["case"], fields["use_case"])
	if err != nil {
		return err
	}

	document.Status, err = NormalizeSpecStatus(fields["status"])
	if err != nil {
		return err
	}

	document.Body = string(body)
	if fields["source_request_id"] != "" || fields["source_quiz_path"] != "" {
		separateSpecProvenance(document)
	}

	return nil
}

func separateSpecProvenance(document *SpecDocument) {
	if at := strings.LastIndex(document.Body, "\n## Provenance\n"); at >= 0 {
		document.Body, document.provenance = document.Body[:at], document.Body[at:]
	}

	heading := "# " + document.Title
	if strings.HasPrefix(strings.TrimLeft(document.Body, "\n"), heading+"\n") {
		document.heading = true
		document.Body = strings.TrimPrefix(strings.TrimLeft(document.Body, "\n"), heading+"\n")
		document.Body = strings.TrimPrefix(document.Body, "\n")
	}
}

func decodeSpecMetadata(document *SpecDocument) error {
	if err := json.Unmarshal(document.images[1].After, &document.metadata); err != nil {
		return err
	}

	parts := strings.Split(document.Reference, "/")
	for key, want := range map[string]string{"schema_version": "ops-layout/v1", "namespace": parts[1], "type": "spec", "id": parts[3]} {
		if document.metadata[key] != want {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func specMetadataString(metadata map[string]any, key string) (string, error) {
	value, ok := metadata[key].(string)
	if !ok {
		return "", errors.New("invalid spec metadata field: " + key)
	}

	return value, nil
}
