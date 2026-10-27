package app

import (
	"errors"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
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

func (e tuiWorkspaceRefreshExecutor) RefreshWorkspace() (tui.WorkspaceRefreshResult, error) {
	result, err := e.operations.RefreshWorkspace()
	if err != nil {
		return tui.WorkspaceRefreshResult{}, err
	}

	return tui.WorkspaceRefreshResult{
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
	metadata authoringcore.ProjectMetadata,
) (tui.ProjectMetadataResult, error) {
	result, err := e.operations.SaveProjectMetadata(metadata)
	if err != nil {
		return tui.ProjectMetadataResult{}, err
	}

	return tui.ProjectMetadataResult{
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
	metadata authoringcore.CaseMetadata,
) (tui.CaseMetadataResult, error) {
	result, err := e.operations.SaveCaseMetadata(metadata)
	if err != nil {
		return tui.CaseMetadataResult{}, err
	}

	return tui.CaseMetadataResult{
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
	action tui.ArtifactActionID,
) (tui.ArtifactActionResult, error) {
	switch action {
	case tui.ArtifactActionCreateTicket:
		result, err := e.operations.CreateDefaultTicket()
		if err != nil {
			return tui.ArtifactActionResult{}, err
		}

		return e.artifactActionResult(result)
	case tui.ArtifactActionCreateAuthoringSession:
		result, err := e.operations.CreateTUIAuthoringSession()
		if err != nil {
			return tui.ArtifactActionResult{}, err
		}

		return e.artifactActionResult(result)
	case tui.ArtifactActionCreateContextQuiz:
		return tui.ArtifactActionResult{}, errors.New(
			"context-aware quiz creation requires provider-backed question generation and is not available from this TUI action yet",
		)
	default:
		return tui.ArtifactActionResult{}, errors.New("unknown artifact action")
	}
}

func (e tuiArtifactActionExecutor) artifactActionResult(
	result ArtifactOperationResult,
) (tui.ArtifactActionResult, error) {
	documents, err := tui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return tui.ArtifactActionResult{}, err
	}

	return tui.ArtifactActionResult{
		Message:           result.Message,
		Tickets:           result.Tickets,
		Documents:         documents,
		AuthoringSessions: result.AuthoringSessions,
	}, nil
}
