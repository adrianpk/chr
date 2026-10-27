package model

func ProviderPayload(request QuizModelRequest) map[string]any {
	payload := map[string]any{
		"project":          request.Payload.Project,
		"quiz":             request.Payload.Quiz,
		"recent_entries":   request.Payload.RecentEntries,
		"current_question": request.Payload.CurrentQuestion,
		"output_contract":  request.Payload.OutputContract,
	}
	if request.Payload.AcceptanceReadiness != nil {
		payload["acceptance_readiness"] = request.Payload.AcceptanceReadiness
	}

	hasContextAction := request.Action == QuizModelActionNextQuestion ||
		request.Action == QuizModelActionDraftSpec ||
		request.Action == QuizModelActionGenerateSpec ||
		request.Action == QuizModelActionGeneratePlan ||
		request.Action == QuizModelActionGeneratePlanOutline ||
		request.Action == QuizModelActionGeneratePlanSliceDetail ||
		request.Action == QuizModelActionMaterializePlanTracking
	if hasContextAction && len(request.ContextSections) > 0 {
		payload["context_sections"] = quizModelProviderContextSections(request.ContextSections)
	}

	return payload
}

func quizModelProviderContextSections(sections []QuizModelContextSection) []map[string]any {
	payloads := make([]map[string]any, 0, len(sections))

	for _, section := range sections {
		payloads = append(payloads, map[string]any{
			"name":      section.Name,
			"layer":     section.Layer,
			"authority": section.Authority,
			"content":   section.Content,
			"metadata":  section.Metadata,
		})
	}

	return payloads
}
