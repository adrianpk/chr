package chcapp

import "forge.adrianpk.com/chc/internal/chc"

type ProjectMetadataSaveResult struct {
	Message  string
	Metadata chc.ProjectMetadata
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

type CaseMetadataSaveResult struct {
	Message string
	Case    chc.CaseMetadata
	Cases   []chc.CaseMetadata
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
