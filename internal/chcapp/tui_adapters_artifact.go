package chcapp

import (
	"errors"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
)

type tuiWorkspaceRefreshExecutor struct {
	operations service
}

func newTUIWorkspaceRefreshExecutor(
	operations service,
) tuiWorkspaceRefreshExecutor {
	return tuiWorkspaceRefreshExecutor{
		operations: operations,
	}
}

func (e tuiWorkspaceRefreshExecutor) RefreshWorkspace() (chctui.WorkspaceRefreshResult, error) {
	result, err := e.operations.RefreshWorkspace()
	if err != nil {
		return chctui.WorkspaceRefreshResult{}, err
	}

	return chctui.WorkspaceRefreshResult{
		Message:              result.Message,
		Graph:                result.Graph,
		Index:                result.Index,
		IndexHealth:          result.IndexHealth,
		IndexChanges:         result.IndexChanges,
		IndexCache:           result.IndexCache,
		SemanticDigestStates: result.SemanticDigestStates,
	}, nil
}

type tuiProjectMetadataExecutor struct {
	operations service
}

func newTUIProjectMetadataExecutor(
	operations service,
) tuiProjectMetadataExecutor {
	return tuiProjectMetadataExecutor{
		operations: operations,
	}
}

func (e tuiProjectMetadataExecutor) SaveProjectMetadata(
	metadata chc.ProjectMetadata,
) (chctui.ProjectMetadataResult, error) {
	result, err := e.operations.SaveProjectMetadata(metadata)
	if err != nil {
		return chctui.ProjectMetadataResult{}, err
	}

	return chctui.ProjectMetadataResult{
		Message:  result.Message,
		Metadata: result.Metadata,
	}, nil
}

type tuiCaseMetadataExecutor struct {
	operations service
}

func newTUICaseMetadataExecutor(
	operations service,
) tuiCaseMetadataExecutor {
	return tuiCaseMetadataExecutor{
		operations: operations,
	}
}

func (e tuiCaseMetadataExecutor) SaveCaseMetadata(
	metadata chc.CaseMetadata,
) (chctui.CaseMetadataResult, error) {
	result, err := e.operations.SaveCaseMetadata(metadata)
	if err != nil {
		return chctui.CaseMetadataResult{}, err
	}

	return chctui.CaseMetadataResult{
		Message: result.Message,
		Case:    result.Case,
		Cases:   result.Cases,
	}, nil
}

type tuiArtifactActionExecutor struct {
	operations service
}

func newTUIArtifactActionExecutor(
	operations service,
) tuiArtifactActionExecutor {
	return tuiArtifactActionExecutor{
		operations: operations,
	}
}

func (e tuiArtifactActionExecutor) ExecuteArtifactAction(
	action chctui.ArtifactActionID,
) (chctui.ArtifactActionResult, error) {
	switch action {
	case chctui.ArtifactActionCreateTicket:
		result, err := e.operations.CreateDefaultTicket()
		if err != nil {
			return chctui.ArtifactActionResult{}, err
		}

		return e.artifactActionResult(result)
	case chctui.ArtifactActionCreateAuthoringSession:
		result, err := e.operations.CreateTUIAuthoringSession()
		if err != nil {
			return chctui.ArtifactActionResult{}, err
		}

		return e.artifactActionResult(result)
	case chctui.ArtifactActionCreateContextQuiz:
		return chctui.ArtifactActionResult{}, errors.New(
			"context-aware quiz creation requires provider-backed question generation and is not available from this TUI action yet",
		)
	default:
		return chctui.ArtifactActionResult{}, errors.New("unknown artifact action")
	}
}

func (e tuiArtifactActionExecutor) artifactActionResult(
	result ArtifactOperationResult,
) (chctui.ArtifactActionResult, error) {
	documents, err := chctui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	return chctui.ArtifactActionResult{
		Message:           result.Message,
		Tickets:           result.Tickets,
		Documents:         documents,
		AuthoringSessions: result.AuthoringSessions,
	}, nil
}
