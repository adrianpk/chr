package app

import (
	"context"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func (s service) ListDumps() (DumpListResult, error) {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	s.ctx = ctx

	observed, err := s.ReadProjectDump()
	if err != nil {
		return DumpListResult{}, err
	}

	artifacts := append(observed.Candidates, observed.Scoped...)
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Reference < artifacts[j].Reference })

	result := DumpListResult{Message: "Dumps loaded.", Dumps: make([]DumpView, 0, len(artifacts))}
	for _, artifact := range artifacts {
		document, err := authoringcore.ReadDumpDocument(s.ctx, s.projectContext.ProjectDir, artifact.Reference)
		if err != nil {
			return result, err
		}

		result.Dumps = append(result.Dumps, s.dumpView(document))
	}

	return result, nil
}

func (s service) dumpView(document authoringcore.DumpDocumentObservation) DumpView {
	artifact := document.Artifact

	return DumpView{
		NodeID: string(core.WorkspaceNodeDump) + ":" + artifact.Reference, Revision: document.Revision,
		UseCaseTitle: s.dumpUseCaseTitle(artifact.CaseSlug), Title: artifact.Title, Status: artifact.Status, DocumentType: artifact.DocumentType,
		SourcePath: artifact.RawReference, Scope: DumpScope{CaseSlug: artifact.CaseSlug}, RelativePath: artifact.Reference, UpdatedAt: artifact.UpdatedAt, Content: document.Content,
	}
}

func (s service) dumpUseCaseTitle(slug string) string {
	if strings.TrimSpace(slug) == "" {
		return "Project-wide"
	}

	metadata, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).ReadBounded(s.ctx, slug)
	if err != nil || strings.TrimSpace(metadata.Name) == "" {
		return slug
	}

	return metadata.Name
}
