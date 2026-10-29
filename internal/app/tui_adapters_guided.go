package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

type tuiGuidedExecutor struct{ operations service }

func (e tuiGuidedExecutor) ExecuteGuided(request tui.GuidedRequest) (tui.GuidedView, error) {
	ctx, cancel := context.WithTimeout(e.operations.ctx, 10*time.Second)
	defer cancel()

	if request.Command.ProjectID != e.operations.projectContext.ProjectID {
		return tui.GuidedView{}, workflowstate.ErrConflict
	}

	if request.Action == "open" {
		return e.openCase(ctx, request)
	}

	if err := e.apply(ctx, request); err != nil {
		view, readErr := e.read(ctx, tui.GuidedRequest{Action: "inspect", Command: request.Command})

		return view, errors.Join(err, readErr)
	}

	return e.read(ctx, request)
}

func (e tuiGuidedExecutor) read(ctx context.Context, request tui.GuidedRequest) (tui.GuidedView, error) {
	observed, err := e.operations.InspectGuidedUseCase(ctx, request.Command.RunID, request.Command.Namespace)
	if err != nil {
		return tui.GuidedView{}, err
	}

	view := tuiGuidedView(observed)
	if err := e.projectPreparedOriginals(ctx, observed, &view); err != nil {
		return view, err
	}

	cases, err := authoringcore.NewCaseMetadataService(e.operations.projectContext.ProjectDir).LoadBounded(ctx, workflowstate.RecordLimit)
	if err != nil {
		return view, err
	}

	view.Cases = append([]authoringcore.CaseMetadata{}, cases...)

	if request.Action == "document" && request.Document == "results" {
		if observed.Session == nil {
			return view, workflowstate.ErrConflict
		}

		payload, err := json.MarshalIndent(guidedResultSnapshot(observed), "", "  ")
		view.Document = &tui.GuidedDocument{Kind: "results export", Content: string(payload)}

		return view, err
	}

	if request.Action == "document" {
		original, err := e.operations.ReadGuidedUseCaseOriginal(ctx, request.Command.RunID, request.Command.Namespace, request.Document)
		if err != nil {
			return view, err
		}

		view.Document = &tui.GuidedDocument{Kind: request.Document, Reference: original.Reference, Commit: original.Commit, Content: original.Content}
	}

	return view, nil
}

func (e tuiGuidedExecutor) apply(ctx context.Context, request tui.GuidedRequest) error {
	var err error

	switch request.Action {
	case "inspect", "document":
		return nil
	case "start":
		_, err = e.operations.StartNamedGuidedUseCase(ctx, authoringcore.GuidedUseCaseStart{GuidedUseCaseScope: request.Command.GuidedUseCaseScope, Actor: "local"}, request.Command.Content)
	case "input":
		_, err = e.operations.ChangeGuidedUseCase(ctx, request.Command)
	default:
		return workflowstate.ErrConflict
	}

	var prerequisite *authoringcore.GuidedPrerequisiteError
	if errors.As(err, &prerequisite) {
		return nil
	}

	return err
}

func tuiGuidedView(observed GuidedUseCaseInspection) tui.GuidedView {
	view := tui.GuidedView{Project: observed.ProjectName, Namespace: observed.Namespace, Session: observed.Session, Runs: observed.Runs, Policy: observed.Policy, Problems: observed.Problems}
	if observed.Session == nil {
		return view
	}

	view.Originals = tuiWizardOriginals(observed)
	view.Quiz = tuiGuidedQuiz(observed)
	state := observed.Session.Guided
	projectTUIGuidedStatus(&view, observed)

	if state.Blocker != "" {
		view.Problems = append(view.Problems, state.Blocker)
	}

	view.Prompt = tuiGuidedStagePrompt(observed)
	if !authoringcore.GuidedDispatchAllowed(*observed.Session) || state.PendingOperation != nil {
		return view
	}

	switch state.Stage {
	case "name":
		view.InputAction = "name"
	case "dump":
		view.InputAction = "append"
	case "quiz":
		if observed.Question != "" {
			view.InputAction = "answer"
		}
	}

	return view
}
