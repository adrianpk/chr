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
		Message:      result.Message,
		Graph:        result.Graph,
		Index:        result.Index,
		IndexHealth:  result.IndexHealth,
		IndexChanges: result.IndexChanges,
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
	case chctui.ArtifactActionCreateDraft:
		return chctui.ArtifactActionResult{}, errors.New(
			"draft file creation requires a source quiz promotion; create an authoring session first",
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

type tuiTicketActionExecutor struct {
	operations service
}

func newTUITicketActionExecutor(
	operations service,
) tuiTicketActionExecutor {
	return tuiTicketActionExecutor{
		operations: operations,
	}
}

func (e tuiTicketActionExecutor) ExecuteTicketAction(
	request chctui.TicketActionRequest,
) (chctui.TicketActionResult, error) {
	switch request.Action {
	case chctui.TicketActionAddComment:
		result, err := e.operations.AddSelectedTicketComment(request.Key)
		if err != nil {
			return chctui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case chctui.TicketActionMoveStatus:
		result, err := e.operations.MoveSelectedTicketStatus(request.Key, request.Status)
		if err != nil {
			return chctui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case chctui.TicketActionToggleChecklist:
		return chctui.TicketActionResult{}, errors.New(
			"checklist toggle requires a safe checklist item writer and is not available from this TUI action yet",
		)
	case chctui.TicketActionAddChecklistItem:
		return chctui.TicketActionResult{}, errors.New(
			"checklist item creation requires a safe checklist item writer and is not available from this TUI action yet",
		)
	default:
		return chctui.TicketActionResult{}, errors.New("unknown ticket action")
	}
}

func ticketActionResult(result TicketOperationResult) chctui.TicketActionResult {
	return chctui.TicketActionResult{
		Message: result.Message,
		Tickets: result.Tickets,
	}
}

type tuiWorkActionExecutor struct {
	operations service
}

func newTUIWorkActionExecutor(
	operations service,
) tuiWorkActionExecutor {
	return tuiWorkActionExecutor{
		operations: operations,
	}
}

func (e tuiWorkActionExecutor) ExecuteWorkAction(
	request chctui.WorkActionRequest,
) (chctui.WorkActionResult, error) {
	switch request.Action {
	case chctui.WorkActionStartTicket:
		result, err := e.operations.StartSelectedTicket(request.TicketKey)
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case chctui.WorkActionStartReadyLoop:
		result, err := e.operations.StartReadyTicketLoop()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case chctui.WorkActionRequestStop:
		result, err := e.operations.RequestDaemonStop()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	default:
		return chctui.WorkActionResult{}, errors.New("unknown work action")
	}
}

func (e tuiWorkActionExecutor) workActionResult(
	result WorkOperationResult,
) (chctui.WorkActionResult, error) {
	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	return chctui.WorkActionResult{
		Message:       result.Message,
		Tickets:       result.Tickets,
		Observability: observability,
	}, nil
}
