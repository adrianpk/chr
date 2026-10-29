package app

import (
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiDumpActionExecutor) ExecuteDumpAction(request tui.DumpActionRequest) (tui.DumpActionResult, error) {
	result, err := e.executeDumpAction(request)
	outcome := operationResult(result.Message, err)

	outcome.OperationID = request.OperationID
	if result.Committed && result.ProjectionStale {
		outcome.Reload = true
	}

	outcome.ChangedFiles = result.Outcome.ChangedFiles
	result.Outcome = outcome

	return result, err
}

func (e tuiQuizActionExecutor) ExecuteQuizAction(request tui.QuizActionRequest) (tui.QuizActionResult, error) {
	result, err := e.executeQuizAction(request)
	outcome := operationResult(result.Message, err)
	outcome.OperationID = request.Binding.OperationID
	outcome.ChangedFiles = result.Outcome.ChangedFiles
	result.Outcome = outcome

	return result, err
}

func (e tuiSpecActionExecutor) ExecuteSpecAction(request tui.SpecActionRequest) (tui.SpecActionResult, error) {
	result, err := e.executeSpecAction(request)
	outcome := operationResult(result.Message, err)
	outcome.ChangedFiles = result.Outcome.ChangedFiles
	result.Outcome = outcome

	return result, err
}

func (e tuiPlanActionExecutor) ExecutePlanAction(request tui.PlanActionRequest) (tui.PlanActionResult, error) {
	result, err := e.executePlanAction(request)
	outcome := operationResult(result.Message, err)
	outcome.ChangedFiles = result.Outcome.ChangedFiles
	result.Outcome = outcome

	return result, err
}

func (e tuiTicketActionExecutor) ExecuteTicketAction(request tui.TicketActionRequest) (tui.TicketActionResult, error) {
	result, err := e.executeTicketAction(request)
	outcome := operationResult(result.Message, err)

	outcome.ChangedFiles = result.Outcome.ChangedFiles
	if err != nil && (request.Action == tui.TicketActionToggleChecklist || request.Action == tui.TicketActionAddChecklistItem) {
		outcome = unavailableOperation(err)
	}

	result.Outcome = outcome

	return result, err
}

func (e tuiWorkActionExecutor) ExecuteWorkAction(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	result, err := e.executeWorkAction(request)
	outcome := operationResult(result.Message, err)
	outcome.ChangedFiles = result.Outcome.ChangedFiles
	result.Outcome = outcome

	return result, err
}

func (e tuiArtifactActionExecutor) ExecuteArtifactAction(request tui.ArtifactActionID) (tui.ArtifactActionResult, error) {
	result, err := e.executeArtifactAction(request)
	outcome := operationResult(result.Message, err)

	outcome.ChangedFiles = result.Outcome.ChangedFiles
	if err != nil && (request == tui.ArtifactActionCreateContextQuiz) {
		outcome = unavailableOperation(err)
	}

	result.Outcome = outcome

	return result, err
}

func (e tuiProjectMetadataExecutor) SaveProjectMetadata(request authoringcore.ProjectMetadata) (tui.ProjectMetadataResult, error) {
	result, err := e.saveProjectMetadata(request)
	outcome := operationResult(result.Message, err)

	outcome.ChangedFiles = result.Outcome.ChangedFiles
	if err == nil {
		outcome.ChangedFiles = relativeOperationFiles(e.operations.projectContext.ProjectDir, result.Metadata.SourcePath)
	}

	if err != nil && request.SourceHash != "" {
		observed, readErr := authoringcore.NewProjectMetadataService(e.operations.projectContext.ProjectDir).Load()
		if readErr == nil && observed.SourceHash != request.SourceHash {
			outcome = staleMetadataOutcome(err)
		}
	}

	result.Outcome = outcome

	return result, err
}

func (e tuiCaseMetadataExecutor) SaveCaseMetadata(request authoringcore.CaseMetadata) (tui.CaseMetadataResult, error) {
	result, err := e.saveCaseMetadata(request)
	outcome := operationResult(result.Message, err)

	outcome.ChangedFiles = result.Outcome.ChangedFiles
	if err == nil {
		outcome.ChangedFiles = relativeOperationFiles(e.operations.projectContext.ProjectDir, result.Case.SourcePath)
	}

	if err != nil && request.SourceHash != "" {
		observed, readErr := authoringcore.NewCaseMetadataService(e.operations.projectContext.ProjectDir).LoadPath(request.SourcePath)
		if readErr == nil && observed.SourceHash != request.SourceHash {
			outcome = staleMetadataOutcome(err)
		}
	}

	result.Outcome = outcome

	return result, err
}

func relativeOperationFiles(root string, paths ...string) []string {
	var files []string

	for _, path := range paths {
		if path == "" {
			continue
		}

		relative := path
		if filepath.IsAbs(path) {
			var err error

			relative, err = filepath.Rel(root, path)
			if err != nil {
				continue
			}
		}

		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			files = append(files, filepath.ToSlash(relative))
		}
	}

	return files
}

func staleMetadataOutcome(err error) operation.Result {
	return operation.Result{Status: operation.Stale, Summary: "The metadata changed after it was loaded.", Validation: []string{err.Error()}, Reload: true}
}

func (e tuiProjectMetadataExecutor) ReadProjectMetadata() (tui.ProjectMetadataResult, error) {
	metadata, err := authoringcore.NewProjectMetadataService(e.operations.projectContext.ProjectDir).Load()

	return tui.ProjectMetadataResult{Metadata: metadata, Message: "Project metadata reloaded.", Outcome: operationResult("Project metadata reloaded.", err)}, err
}

func (e tuiCaseMetadataExecutor) ReadCaseMetadata(slug string) (tui.CaseMetadataResult, error) {
	metadata, err := authoringcore.NewCaseMetadataService(e.operations.projectContext.ProjectDir).LoadPath(authoringcore.CaseMetadataPath(e.operations.projectContext.ProjectDir, slug))

	return tui.CaseMetadataResult{Case: metadata, Message: "Case metadata reloaded.", Outcome: operationResult("Case metadata reloaded.", err)}, err
}

func unavailableOperation(err error) operation.Result {
	return operation.Result{Status: operation.Blocked, Summary: "Operation unavailable.", BlockedReason: err.Error()}
}
