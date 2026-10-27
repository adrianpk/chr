package model

// ProviderPayload projects the shared request context for a generation backend.
func ProviderPayload(request Request) map[string]any {
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

	hasContextAction := request.Action == ActionNextQuestion ||
		request.Action == ActionDraftSpec ||
		request.Action == ActionGenerateSpec ||
		request.Action == ActionGeneratePlan ||
		request.Action == ActionGeneratePlanOutline ||
		request.Action == ActionGeneratePlanSliceDetail ||
		request.Action == ActionMaterializePlanTracking
	if hasContextAction && len(request.ContextSections) > 0 {
		payload["context_sections"] = providerContextSections(request.ContextSections)
	}

	return payload
}

func providerContextSections(sections []ContextSection) []map[string]any {
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
