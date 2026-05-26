package chcapp

import "forge.adrianpk.com/chc/internal/chctui"

func loadTUIDumps(operations service) ([]chctui.DumpView, error) {
	result, err := operations.ListDumps()
	if err != nil {
		return nil, err
	}

	dumps := make([]chctui.DumpView, 0, len(result.Dumps))
	for index, dump := range result.Dumps {
		dumps = append(dumps, chctui.DumpView{
			Index:        index,
			NodeID:       dump.NodeID,
			UseCaseTitle: dump.UseCaseTitle,
			Title:        dump.Title,
			Status:       dump.Status,
			Scope:        chctui.DumpScope{CaseSlug: dump.Scope.CaseSlug},
			RelativePath: dump.RelativePath,
			UpdatedAt:    dump.UpdatedAt,
			Preview:      dump.Content,
		})
	}

	return dumps, nil
}

func loadTUIQuizzes(operations service) ([]chctui.QuizView, error) {
	result, err := operations.ListQuizzes()
	if err != nil {
		return nil, err
	}

	quizzes := make([]chctui.QuizView, 0, len(result.Views))
	for index, quiz := range result.Views {
		quizzes = append(quizzes, chctui.QuizView{
			Index:                 index,
			Title:                 quiz.Title,
			Scope:                 chctui.QuizScope{CaseSlug: quiz.Scope.CaseSlug},
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

func loadTUISpecs(operations service) ([]chctui.SpecView, error) {
	result, err := operations.ListSpecs()
	if err != nil {
		return nil, err
	}

	specs := make([]chctui.SpecView, 0, len(result.Specs))
	for index, spec := range result.Specs {
		specs = append(specs, chctui.SpecView{
			Index:        index,
			UseCaseTitle: spec.UseCaseTitle,
			Title:        spec.Title,
			Status:       spec.Status,
			Scope:        chctui.QuizScope{CaseSlug: spec.Scope.CaseSlug},
			RelativePath: spec.RelativePath,
			UpdatedAt:    spec.UpdatedAt,
			Content:      spec.Content,
		})
	}

	return specs, nil
}

func loadTUIPlans(operations service) ([]chctui.PlanView, error) {
	result, err := operations.ListPlans()
	if err != nil {
		return nil, err
	}

	plans := make([]chctui.PlanView, 0, len(result.Plans))
	for index, plan := range result.Plans {
		plans = append(plans, chctui.PlanView{
			Index:        index,
			UseCaseTitle: plan.UseCaseTitle,
			Title:        plan.Title,
			Status:       plan.Status,
			Scope:        chctui.QuizScope{CaseSlug: plan.Scope.CaseSlug},
			RelativePath: plan.RelativePath,
			UpdatedAt:    plan.UpdatedAt,
			Content:      plan.Content,
		})
	}

	return plans, nil
}
