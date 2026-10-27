package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
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
	if errors.Is(err, workflowstate.ErrConflict) {
		response.Status = webui.WriteStatusStale
		response.Summary = "The plan or its context changed. Reload this case before generating again."
		response.ReloadURL = "/cases/detail?id=" + url.QueryEscape(slug)
	}

	return response
}

func planCatalog(operations service) (webui.ArtifactCatalogView, error) {
	documents, err := operations.ListPlanDocuments()
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	cards := make([]webui.ArtifactCardView, 0, len(documents))
	for _, document := range documents {
		if document.OriginReference != "" && document.OriginReference != document.Reference {
			continue
		}

		card := webui.ArtifactCardView{Family: webArtifactFamilyPlans, ID: document.Reference, Title: document.Title, Preview: webPlanPreview(document), Revision: document.Revision}
		card.DetailURL = webArtifactDetailURL(card)
		cards = append(cards, card)
	}

	view := webArtifactCatalog(webArtifactFamilyPlans, cards)

	view.PlanChoices, err = planGenerationChoices(operations)
	if err != nil {
		return webui.ArtifactCatalogView{}, err
	}

	view.Revision = webStableRevision(view)

	return view, nil
}

func planGenerationChoices(operations service) ([]webui.PlanGenerationChoice, error) {
	cases, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadBounded(operations.ctx, core.WorkflowRecordLimit)
	if err != nil {
		return nil, err
	}

	specs, err := operations.ListSpecDocuments()
	if err != nil {
		return nil, err
	}

	titles := map[string]string{}
	for _, spec := range specs {
		titles[spec.CaseSlug] = spec.Title
	}

	var choices []webui.PlanGenerationChoice

	for _, item := range cases {
		if titles[item.Slug] == "" {
			continue
		}

		revision, err := core.PlanTargetRevision(operations.ctx, operations.projectContext.ProjectDir, operations.quizScope(core.QuizScope{CaseSlug: item.Slug}))
		if err != nil {
			return nil, err
		}

		choices = append(choices, webui.PlanGenerationChoice{CaseSlug: item.Slug, CaseTitle: item.Name, SpecTitle: titles[item.Slug], Revision: revision})
	}

	return choices, nil
}

func (p webViewProvider) writePlanDelivery(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	scope := webActionScope(request)
	if request.Family != "plans" || scope.CaseSlug == "" || request.ID != scope.CaseSlug || request.Revision == "" || request.Fields["request_id"] == "" {
		return webValidationWrite(errors.New("select a saved use case in Plans before generating")), nil
	}

	result, err := operations.GeneratePlanDelivery(ctx, PlanGenerateRequest{Scope: scope, ExpectedRevision: request.Revision, OperationID: request.Fields["request_id"]})
	if err != nil {
		response := webValidationWrite(err)
		if errors.Is(err, workflowstate.ErrConflict) {
			response.Status = webui.WriteStatusStale
		}

		return response, nil
	}

	layout, err := core.GeneratedPlanLayout(operations.projectContext.ProjectDir, operations.quizScope(scope))
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: fmt.Sprintf("Plan ready with %d tickets in %d slices. Work has not started.", len(result.Workflow.Tasks), len(result.Workflow.Set.Set.Slices)), RedirectURL: "/plans/detail?id=" + url.QueryEscape(layout.RelativePath), Operation: webOperation("plan.generate", "completed", layout.RelativePath, "tickets_created")}, nil
}

func webPlanPreview(document core.PlanDocument) string {
	body := strings.TrimSpace(document.Body)

	heading, remaining, _ := strings.Cut(body, "\n")
	if strings.TrimSpace(heading) == "# "+document.Title {
		body = strings.TrimSpace(remaining)
	}

	preview := []rune(body)
	if len(preview) > 1024 {
		preview = preview[:1024]
	}

	return string(preview)
}

func (p webViewProvider) planArtifactDetail(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	document, err := core.ReadPlanDocument(ctx, p.projectContext.ProjectDir, request.ID)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	return webui.ArtifactDetailView{Family: webArtifactFamilyPlans, ID: document.Reference, Title: document.Title, Content: document.Body, Revision: document.Revision}, nil
}
