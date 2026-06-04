package chcapp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chc/internal/chc"
)

// ProjectMetadataSaveResult reports the saved project metadata.
type ProjectMetadataSaveResult struct {
	Message  string
	Metadata chc.ProjectMetadata
}

// ProjectMetadataPatchRequest carries editable project metadata fields.
type ProjectMetadataPatchRequest struct {
	Name        *string
	Description *string
	Context     *string
}

func (s service) SaveProjectMetadata(
	metadata chc.ProjectMetadata,
) (ProjectMetadataSaveResult, error) {
	saved, err := chc.NewProjectMetadataService(s.projectContext.ProjectDir).Save(metadata)
	if err != nil {
		return ProjectMetadataSaveResult{}, err
	}

	return ProjectMetadataSaveResult{
		Message:  "Project metadata saved.",
		Metadata: saved,
	}, nil
}

func (s service) PatchProjectMetadata(request ProjectMetadataPatchRequest) (ProjectMetadataSaveResult, error) {
	metadata, err := chc.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return ProjectMetadataSaveResult{}, err
	}

	metadata = applyProjectMetadataPatch(metadata, request)

	return s.SaveProjectMetadata(metadata)
}

// CaseMetadataSaveResult reports the saved use-case metadata.
type CaseMetadataSaveResult struct {
	Message string
	Case    chc.CaseMetadata
	Cases   []chc.CaseMetadata
}

// CaseMetadataPatchRequest carries editable use-case metadata fields.
type CaseMetadataPatchRequest struct {
	Slug        string
	Name        *string
	Description *string
	Context     *string
}

func (s service) SaveCaseMetadata(metadata chc.CaseMetadata) (CaseMetadataSaveResult, error) {
	caseService := chc.NewCaseMetadataService(s.projectContext.ProjectDir)

	saved, err := caseService.Save(metadata)
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

func (s service) loadCaseMetadata(slug string) (chc.CaseMetadata, error) {
	path := chc.CaseMetadataPath(s.projectContext.ProjectDir, slug)

	metadata, err := chc.NewCaseMetadataService(s.projectContext.ProjectDir).LoadPath(path)
	if errors.Is(err, os.ErrNotExist) {
		return chc.CaseMetadata{
			SchemaVersion: chc.CaseMetadataSchemaVersion,
			Slug:          slug,
			SourcePath:    path,
		}, nil
	}

	return metadata, err
}

func applyProjectMetadataPatch(
	metadata chc.ProjectMetadata,
	request ProjectMetadataPatchRequest,
) chc.ProjectMetadata {
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
	metadata chc.CaseMetadata,
	request CaseMetadataPatchRequest,
) chc.CaseMetadata {
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
