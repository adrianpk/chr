package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (r *fileAuthoringSessionRepository) readProjectDumpArtifact(reference string) (ProjectDumpArtifact, int, error) {
	if err := ValidateProjectDumpReference(reference); err != nil {
		return ProjectDumpArtifact{}, 0, err
	}

	body, err := r.readAuthoringDocument(reference)
	if err != nil {
		return ProjectDumpArtifact{}, 0, err
	}

	encoded, err := r.readAuthoringDocument(filepath.Join(filepath.Dir(reference), "metadata.json"))
	if err != nil {
		return ProjectDumpArtifact{}, 0, err
	}

	metadata, err := decodeProjectDumpMetadata(reference, encoded)
	if err != nil {
		return ProjectDumpArtifact{}, 0, err
	}

	fields, err := projectDumpBodyFields(body)
	if err != nil {
		return ProjectDumpArtifact{}, 0, err
	}

	artifact, err := projectDumpArtifact(reference, body, encoded, fields, metadata)

	return artifact, len(body) + len(encoded), err
}

func (r *fileAuthoringSessionRepository) readAuthoringDocument(relative string) ([]byte, error) {
	path, err := r.authoringPath(filepath.FromSlash(relative))
	if err != nil {
		return nil, err
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("authoring document is not a regular file")
	}

	return readWorkflowFile(path)
}

func decodeProjectDumpMetadata(reference string, data []byte) (map[string]string, error) {
	var metadata map[string]json.RawMessage
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	fields := projectMetadataStringFields(metadata)

	parts := strings.Split(reference, "/")
	if fields["schema_version"] != "ops-layout/v1" || fields["namespace"] != parts[1] || fields["type"] != "dump" || fields["id"] != parts[3] {
		return nil, fmt.Errorf("dump metadata identity does not match its canonical body")
	}

	return fields, nil
}

func projectDumpBodyFields(body []byte) (map[string]string, error) {
	if !strings.HasPrefix(string(body), "---\n") {
		return map[string]string{}, nil
	}

	front, _, err := splitWorkflowMarkdown(body)
	if err != nil {
		return nil, err
	}

	return projectDocumentFields(front)
}

func projectDumpArtifact(reference string, body, encoded []byte, fields, metadata map[string]string) (ProjectDumpArtifact, error) {
	scope, err := projectDumpScope(fields["case"], fields["use_case"], metadata["case"], metadata["use_case"])
	if err != nil {
		return ProjectDumpArtifact{}, err
	}

	artifact := ProjectDumpArtifact{
		Reference: reference, Title: metadata["title"], SourceHash: workflowHash(body), MetadataHash: workflowHash(encoded), CaseSlug: scope,
		Raw: fields["type"] == "dump_raw" || fields["status"] == "raw" && fields["curated"] != "", CuratedReference: fields["curated"],
		CanonicalReference: metadata["canonical_ref"],
	}
	if err := validateProjectDumpAlias(artifact); err != nil {
		return ProjectDumpArtifact{}, err
	}

	if fields["title"] != "" {
		artifact.Title = fields["title"]
	}

	if source := fields["source"]; source != "" && ValidateProjectDumpReference(source) == nil {
		artifact.RawReference = source
	}

	artifact.CreatedAt, err = time.Parse(time.RFC3339Nano, metadata["created_at"])
	if err != nil {
		artifact.TimestampIssue = "creation timestamp is missing or invalid"
	}

	return artifact, nil
}

func validateProjectDumpAlias(artifact ProjectDumpArtifact) error {
	if artifact.CanonicalReference == "" {
		return nil
	}

	if err := ValidateProjectDumpReference(artifact.CanonicalReference); err != nil || artifact.Raw || artifact.CaseSlug != "" {
		return fmt.Errorf("invalid canonical dump alias")
	}

	return nil
}

func projectDumpScope(values ...string) (string, error) {
	selected := ""

	for _, value := range values {
		if value == "" {
			continue
		}

		if selected != "" && selected != value {
			return "", fmt.Errorf("dump scope declarations conflict")
		}

		selected = value
	}

	return selected, nil
}
