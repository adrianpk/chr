package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

// ProjectMetadataSaveResult reports the saved project metadata.
type ProjectMetadataSaveResult struct {
	Message  string
	Metadata core.ProjectMetadata
}

// ProjectMetadataPatchRequest carries editable project metadata fields.
type ProjectMetadataPatchRequest struct {
	Name        *string
	Description *string
	Context     *string
}

func (s service) SaveProjectMetadata(
	metadata core.ProjectMetadata,
) (ProjectMetadataSaveResult, error) {
	saved, err := core.NewProjectMetadataService(s.projectContext.ProjectDir).SaveContext(s.ctx, metadata)
	if err != nil {
		return ProjectMetadataSaveResult{}, err
	}

	return ProjectMetadataSaveResult{
		Message:  "Project metadata saved.",
		Metadata: saved,
	}, nil
}

func (s service) PatchProjectMetadata(request ProjectMetadataPatchRequest) (ProjectMetadataSaveResult, error) {
	metadata, err := core.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return ProjectMetadataSaveResult{}, err
	}

	metadata = applyProjectMetadataPatch(metadata, request)

	return s.SaveProjectMetadata(metadata)
}

// CaseMetadataSaveResult reports the saved use-case metadata.
type CaseMetadataSaveResult struct {
	Message string
	Case    core.CaseMetadata
	Cases   []core.CaseMetadata
}

// CaseMetadataPatchRequest carries editable use-case metadata fields.
type CaseMetadataPatchRequest struct {
	Slug        string
	Name        *string
	Description *string
	Context     *string
}

func (s service) SaveCaseMetadata(metadata core.CaseMetadata) (CaseMetadataSaveResult, error) {
	caseService := core.NewCaseMetadataService(s.projectContext.ProjectDir)

	saved, err := caseService.SaveContext(s.ctx, metadata)
	if err != nil {
		return CaseMetadataSaveResult{}, err
	}

	cases, err := caseService.LoadAll()
	if err != nil {
		return CaseMetadataSaveResult{}, err
	}

	return CaseMetadataSaveResult{
		Message: "Case metadata saved.",
		Case:    saved,
		Cases:   cases,
	}, nil
}

func (s service) PatchCaseMetadata(request CaseMetadataPatchRequest) (CaseMetadataSaveResult, error) {
	metadata, err := s.loadCaseMetadata(request.Slug)
	if err != nil {
		return CaseMetadataSaveResult{}, err
	}

	metadata = applyCaseMetadataPatch(metadata, request)
	if strings.TrimSpace(metadata.Name) == "" {
		metadata.Name = caseNameFromSlug(request.Slug)
	}

	return s.SaveCaseMetadata(metadata)
}

func (s service) loadCaseMetadata(slug string) (core.CaseMetadata, error) {
	path := core.CaseMetadataPath(s.projectContext.ProjectDir, slug)

	metadata, err := core.NewCaseMetadataService(s.projectContext.ProjectDir).LoadPath(path)
	if errors.Is(err, os.ErrNotExist) {
		return core.CaseMetadata{
			SchemaVersion: core.CaseMetadataSchemaVersion,
			Slug:          slug,
			SourcePath:    path,
		}, nil
	}

	return metadata, err
}

func applyProjectMetadataPatch(
	metadata core.ProjectMetadata,
	request ProjectMetadataPatchRequest,
) core.ProjectMetadata {
	if request.Name != nil {
		metadata.Name = *request.Name
	}

	if request.Description != nil {
		metadata.Description = *request.Description
	}

	if request.Context != nil {
		metadata.Context = *request.Context
	}

	return metadata
}

func applyCaseMetadataPatch(
	metadata core.CaseMetadata,
	request CaseMetadataPatchRequest,
) core.CaseMetadata {
	if request.Name != nil {
		metadata.Name = *request.Name
	}

	if request.Description != nil {
		metadata.Description = *request.Description
	}

	if request.Context != nil {
		metadata.Context = *request.Context
	}

	return metadata
}

func validateCaseSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("metadata case slug is required")
	}

	for _, current := range slug {
		if current >= 'a' && current <= 'z' {
			continue
		}

		if current >= '0' && current <= '9' {
			continue
		}

		if current == '-' {
			continue
		}

		return fmt.Errorf("metadata case slug must contain only lowercase letters, numbers, and dashes")
	}

	return nil
}

func caseNameFromSlug(slug string) string {
	parts := strings.Split(filepath.Base(slug), "-")
	for index, part := range parts {
		if part == "" {
			continue
		}

		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}
