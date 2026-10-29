package app

import (
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

func loadTUIDumps(operations service) ([]tui.DumpView, error) {
	result, err := operations.ListDumps()
	if err != nil {
		return nil, err
	}

	dumps := make([]tui.DumpView, 0, len(result.Dumps))
	for index, dump := range result.Dumps {
		dumps = append(dumps, tui.DumpView{
			Index: index, Revision: dump.Revision,
			NodeID:       dump.NodeID,
			UseCaseTitle: dump.UseCaseTitle,
			Title:        dump.Title,
			Status:       dump.Status,
			SourcePath:   dump.SourcePath,
			Scope:        tui.DumpScope{CaseSlug: dump.Scope.CaseSlug},
			RelativePath: dump.RelativePath,
			UpdatedAt:    dump.UpdatedAt,
			Preview:      dump.Content,
		})
	}

	return dumps, nil
}

func loadTUIQuizzes(operations service) ([]tui.QuizView, error) {
	result, err := operations.ListQuizzes()
	if err != nil {
		return nil, err
	}

	quizzes := make([]tui.QuizView, 0, len(result.Views))
	for index, quiz := range result.Views {
		quizzes = append(quizzes, tui.QuizView{
			History:               tuiQuizHistory(quiz.History),
			Binding:               tui.QuizActionBinding{SessionID: quiz.Binding.SessionID, QuestionID: quiz.Binding.QuestionID, ExpectedRevision: quiz.Binding.ExpectedRevision},
			Index:                 index,
			Title:                 quiz.Title,
			Scope:                 tui.QuizScope{CaseSlug: quiz.Scope.CaseSlug},
			Status:                quiz.Status,
			Mode:                  string(quiz.Mode),
			RawPath:               quiz.RawPath,
			CuratedPath:           quiz.CuratedPath,
			RuntimePath:           quiz.RuntimePath,
			QuestionBudget:        quiz.QuestionBudget,
			QuestionsAsked:        quiz.QuestionsAsked,
			AnswersRecorded:       quiz.AnswersRecorded,
			SkipsRecorded:         quiz.SkipsRecorded,
			CandidateCount:        quiz.CandidateCount,
			BranchReturnReason:    quiz.BranchReturnReason,
			UpdatedAt:             quiz.UpdatedAt,
			ActiveQuestionSummary: quiz.ActiveQuestionSummary,
			TrailSummary:          quiz.TrailSummary,
			CuratedPreview:        quiz.CuratedPreview,
		})
	}

	return quizzes, nil
}

func loadTUISpecs(operations service) ([]tui.SpecView, error) {
	documents, err := operations.ListSpecDocuments()
	if err != nil {
		return nil, err
	}

	cases, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadBounded(operations.ctx, workflowstate.RecordLimit)
	if err != nil {
		return nil, err
	}

	names := map[string]string{"": "Project-wide"}
	for _, metadata := range cases {
		names[metadata.Slug] = metadata.Name
	}

	specs := make([]tui.SpecView, 0, len(documents))
	for index, spec := range documents {
		specs = append(specs, tui.SpecView{
			Index:        index,
			UseCaseTitle: names[spec.CaseSlug],
			Title:        spec.Title,
			Status:       spec.Status,
			Scope:        tui.QuizScope{CaseSlug: spec.CaseSlug},
			RelativePath: spec.Reference,
			Revision:     spec.Revision,
			Content:      spec.Body,
		})
	}

	return specs, nil
}

func loadTUIPlans(operations service) ([]tui.PlanView, error) {
	result, err := operations.ListPlans()
	if err != nil {
		return nil, err
	}

	plans := make([]tui.PlanView, 0, len(result.Plans))
	for index, plan := range result.Plans {
		plans = append(plans, tui.PlanView{
			Index: index, Revision: plan.Revision,
			UseCaseTitle: plan.UseCaseTitle,
			Title:        plan.Title,
			Status:       plan.Status,
			Scope:        tui.QuizScope{CaseSlug: plan.Scope.CaseSlug},
			RelativePath: plan.RelativePath,
			UpdatedAt:    plan.UpdatedAt,
			Content:      plan.Content,
		})
	}

	return plans, nil
}
