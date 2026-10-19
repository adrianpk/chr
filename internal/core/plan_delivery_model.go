package core

import "sort"

const planDeliveryInstructions = `Complete executable ticket contracts for the supplied implementation plan and planned_workflow.
Preserve exactly every supplied slice/task ID, title, slice_order and task_order. One task becomes one ticket; do not add, remove or combine tasks. Keep the supplied human labels.
Return closing_validation and slices only. The application owns materialization identity, plan reference/hash, use case and ticket key allocation.
For each slice supply a concrete implementation contract and validation. For each task supply body, outcome, bounded scope, observable acceptance criteria, validation, priority, severity, type and tags. Ground all contracts in the plan and saved discovery/specs; do not invent product decisions.
Validation commands are argv arrays without shell interpolation. Choose proportionate development checks appropriate to the project and the slice. Do not require exhaustive, race or nightly suites for ordinary development. Do not waive validation merely to satisfy the schema. A waiver needs a specific justification in the plan.
Do not execute work, edit files, create branches/commits, open PRs or claim progress. Do not ask for credentials or secrets. Return only JSON matching the schema.`

func planDeliveryJSONSchema() map[string]any {
	return planDeliveryObject(map[string]any{
		"closing_validation": planDeliveryGateSchema(),
		"slices":             planDeliveryArray(planDeliverySliceSchema()),
	})
}

func planDeliverySliceSchema() map[string]any {
	return planDeliveryObject(map[string]any{
		"id": planDeliveryString(), "label": planDeliveryString(), "slice_order": map[string]any{"type": "integer", "minimum": 1},
		"title": planDeliveryString(), "contract": planDeliveryString(), "validation": planDeliveryGateSchema(),
		"tasks": planDeliveryArray(planDeliveryTaskSchema()),
	})
}

func planDeliveryTaskSchema() map[string]any {
	return planDeliveryObject(map[string]any{
		"id": planDeliveryString(), "label": planDeliveryString(), "task_order": map[string]any{"type": "integer", "minimum": 1},
		"title": planDeliveryString(), "body": planDeliveryString(), "outcome": planDeliveryString(), "scope": planDeliveryString(),
		"acceptance": planDeliveryArray(planDeliveryString()), "validation": planDeliveryGateSchema(),
		"priority": planDeliveryEnum("low", "normal", "high", "critical"), "severity": planDeliveryEnum("low", "normal", "high", "critical"),
		"type": planDeliveryEnum("task", "bug", "feature", "cleanup", "research", "docs"),
		"tags": map[string]any{"type": "array", "maxItems": 32, "items": planDeliveryString()},
	})
}

func planDeliveryGateSchema() map[string]any {
	return planDeliveryObject(map[string]any{
		"required": map[string]any{"type": "boolean"}, "waiver_reason": map[string]any{"type": "string"},
		"timeout_seconds": map[string]any{"type": "integer", "minimum": 0, "maximum": 86400},
		"commands":        map[string]any{"type": "array", "maxItems": 16, "items": map[string]any{"type": "array", "minItems": 1, "maxItems": 64, "items": planDeliveryString()}},
	})
}

func planDeliveryObject(properties map[string]any) map[string]any {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return map[string]any{"type": "object", "additionalProperties": false, "required": keys, "properties": properties}
}

func planDeliveryArray(items map[string]any) map[string]any {
	return map[string]any{"type": "array", "minItems": 1, "maxItems": WorkflowPageLimit, "items": items}
}

func planDeliveryString() map[string]any { return map[string]any{"type": "string", "minLength": 1} }

func planDeliveryEnum(values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values}
}

func quizModelTrackingInstructions(request QuizModelRequest) string {
	if request.Strategy == QuizModelStrategyPlanDelivery {
		return planDeliveryInstructions
	}

	return quizModelPlanTrackingInstructions()
}

func quizModelTrackingFormat(request QuizModelRequest) openAIResponsesTextFormat {
	if request.Strategy == QuizModelStrategyPlanDelivery {
		return quizModelJSONSchemaTextFormat("plan_delivery", "Ticket contracts for the retained plan and declared order.", planDeliveryJSONSchema())
	}

	return quizModelJSONSchemaTextFormat("plan_tracking_artifacts", "Plan tracking and ticket payloads derived from an existing implementation plan.", quizModelPlanTrackingJSONSchema())
}
