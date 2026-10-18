package app

import (
	"context"
	"crypto/rand"
	"errors"
	"net/url"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) casePlanningView(ctx context.Context, slug, title string) *webui.CasePlanView {
	view := &webui.CasePlanView{OperationID: rand.Text(), CaseSlug: slug, CaseTitle: title, Action: webActionPlanGenerate, Label: "Generate Plan", DumpURL: "/dumps"}
	operations := p.service(ctx)
	scope := operations.quizScope(core.QuizScope{CaseSlug: slug})

	quiz, err := operations.quizRepository().Paths(scope)
	if err != nil {
		view.BlockedReason = "Planning context is unavailable. Reload this case."

		return view
	}

	view.QuizURL = quizFocusURL(quiz.CuratedRelativePath)
	if err := p.observeCasePlan(ctx, operations, scope, view); err != nil {
		view.BlockedReason = "The plan is temporarily unavailable. Reload after the pending change completes."
	}

	dumps, err := core.ObserveProjectDump(ctx, p.projectContext.ProjectDir)
	if err == nil {
		reference := dumps.Destination
		for _, dump := range dumps.Scoped {
			if dump.CaseSlug == slug {
				reference = dump.Reference

				break
			}
		}

		if reference != "" {
			view.DumpURL = "/dumps/detail?id=" + url.QueryEscape(reference)
		}
	}

	return view
}

func (p webViewProvider) observeCasePlan(ctx context.Context, operations service, scope core.QuizScope, view *webui.CasePlanView) error {
	revision, err := core.PlanTargetRevision(ctx, p.projectContext.ProjectDir, scope)
	if err != nil {
		return err
	}

	plan, exists, err := operations.existingGeneratedPlan(scope)
	if err != nil {
		return err
	}

	view.Revision = revision
	if exists {
		view.Action, view.Label = webActionPlanRegenerate, "Regenerate Plan"
		view.PlanURL = "/plans/detail?id=" + url.QueryEscape(plan.RelativePath)
	}

	return nil
}

func webCasePlanScope(ctx context.Context, operations service, request webui.WriteActionRequest) (core.QuizScope, error) {
	scope := webActionScope(request)
	if request.Surface != "case-plan" || request.Family != "cases" || scope.CaseSlug == "" || request.ID != scope.CaseSlug || request.Revision == "" {
		return scope, errors.New("open an existing named use case to generate its plan")
	}

	_, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).ReadBounded(ctx, scope.CaseSlug)

	return scope, err
}

func webPlanWriteFailure(err error, slug string) webui.WriteActionResponse {
	response := webValidationWrite(err)
	if errors.Is(err, core.ErrWorkflowConflict) {
		response.Status = webui.WriteStatusStale
		response.Summary = "The plan or its context changed. Reload this case before generating again."
		response.ReloadURL = "/cases/detail?id=" + url.QueryEscape(slug)
	}

	return response
}
