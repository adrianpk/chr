package model

import (
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

type OutputFormatDefinition struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Strict      bool           `json:"strict"`
	Schema      map[string]any `json:"schema"`
}

func OutputFormat(request QuizModelRequest) OutputFormatDefinition {
	switch {
	case request.Action == QuizModelActionMaterializePlanTracking:
		return quizModelTrackingFormat(request)
	case request.Action == QuizModelActionGeneratePlan:
		return NewOutputFormat(
			"generated_plan",
			"Implementation plan generated from project and use-case context.",
			quizModelPlanJSONSchema(),
		)
	case request.Action == QuizModelActionGeneratePlanOutline:
		return NewOutputFormat(
			"generated_plan_outline",
			"Plan outline generated from project and use-case context.",
			quizModelPlanOutlineJSONSchema(),
		)
	case request.Action == QuizModelActionGeneratePlanSliceDetail:
		return NewOutputFormat(
			"generated_plan_slice_detail",
			"Implementation details generated for one outline slice.",
			quizModelPlanSliceDetailJSONSchema(),
		)
	case request.Action == QuizModelActionGenerateSpec:
		return NewOutputFormat(
			"generated_spec",
			"Draft specification generated from project and use-case context.",
			quizModelSpecJSONSchema(),
		)
	case request.Action == QuizModelActionGenerateAcceptanceFeature:
		return NewOutputFormat(
			"acceptance_feature",
			"Acceptance feature generated from project, use-case, quiz, dump, and spec context.",
			quizModelFeatureJSONSchema(),
		)
	case request.Action == QuizModelActionDraftSpec:
		return NewOutputFormat(
			"quiz_draft_spec",
			"Draft specification generated from project quiz context.",
			quizModelDraftSpecJSONSchema(),
		)
	case request.Mode == quizcore.ModeFast:
		return NewOutputFormat(
			"quiz_fast_questions",
			"A batch of project quiz questions.",
			quizModelFastQuestionJSONSchema(),
		)
	default:
		return NewOutputFormat(
			"quiz_next_question",
			"The next project quiz question.",
			quizModelQuestionJSONSchema(),
		)
	}
}

func NewOutputFormat(name, description string, schema map[string]any) OutputFormatDefinition {
	return OutputFormatDefinition{
		Type:        "json_schema",
		Name:        name,
		Description: description,
		Strict:      true,
		Schema:      schema,
	}
}

//nolint:funlen // JSON schema literals are clearer kept together.
func quizModelPlanTrackingJSONSchema() map[string]any {
	ticketSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"slice_id",
			"title",
			"synopsis",
			"summary",
			"body",
			"priority",
			"severity",
			"type",
			"tags",
			"dependencies",
		},
		"properties": map[string]any{
			"slice_id": map[string]any{"type": "string"},
			"title":    map[string]any{"type": "string"},
			"synopsis": map[string]any{
				"type":        "string",
				"description": "One short Board-facing sentence describing the ticket.",
			},
			"summary": map[string]any{"type": "string"},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown ticket body without front matter.",
			},
			"priority": map[string]any{
				"type": "string",
				"enum": []string{"low", "normal", "high", "critical"},
			},
			"severity": map[string]any{
				"type": "string",
				"enum": []string{"low", "normal", "high", "critical"},
			},
			"type": map[string]any{
				"type": "string",
				"enum": []string{"task", "bug", "feature", "cleanup", "research", "docs"},
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"dependencies": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body", "status", "tickets"},
		"properties": map[string]any{
			"title": map[string]any{"type": "string"},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown tracking body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"tracking"},
			},
			"tickets": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": 12,
				"items":    ticketSchema,
			},
		},
	}
}

func quizModelQuestionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             QuizModelQuestionRequiredFields,
		"properties": map[string]any{
			"question": map[string]any{
				"type":        "string",
				"description": "One concise user-facing interviewer question.",
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"follow_up_depth": map[string]any{
				"type":    "integer",
				"minimum": 0,
				"maximum": quizcore.MaxFollowUpDepth,
			},
			"completion_recommended": map[string]any{
				"type": "boolean",
			},
		},
	}
}

func quizModelFastQuestionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"questions"},
		"properties": map[string]any{
			"questions": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": QuizModelFastBatchSize,
				"items":    quizModelQuestionJSONSchema(),
			},
		},
	}
}

func quizModelDraftSpecJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body", "status"},
		"properties": map[string]any{
			"title": map[string]any{
				"type": "string",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown spec body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"draft"},
			},
		},
	}
}

func quizModelSpecJSONSchema() map[string]any {
	return quizModelDraftSpecJSONSchema()
}

func quizModelFeatureJSONSchema() map[string]any {
	promptSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"id", "question", "default_answer", "reason"},
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "Stable short prompt id.",
			},
			"question": map[string]any{
				"type":        "string",
				"description": "One concise user-facing question.",
			},
			"default_answer": map[string]any{
				"type":        "string",
				"description": "Best inferred answer the user can confirm or adjust.",
			},
			"reason": map[string]any{
				"type":        "string",
				"description": "Why this answer is needed for acceptance scenarios.",
			},
		},
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"status", "body", "rejection_reason", "clarification_prompts"},
		"properties": map[string]any{
			"status": map[string]any{
				"type": "string",
				"enum": []string{
					generationcore.AcceptanceFeatureStatusGenerated,
					generationcore.AcceptanceFeatureStatusNeedsMoreInfo,
					generationcore.AcceptanceFeatureStatusRejected,
					generationcore.AcceptanceFeatureStatusNotApplicable,
				},
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Gherkin feature body when status is generated.",
			},
			"rejection_reason": map[string]any{
				"type":        "string",
				"description": "Reason generation needs more information, was rejected, or is not applicable.",
			},
			"clarification_prompts": map[string]any{
				"type":        "array",
				"description": "Default-oriented prompts required when status is needs_more_info.",
				"items":       promptSchema,
			},
		},
	}
}

func quizModelPlanJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body", "status", "order"},
		"properties": map[string]any{
			"order": planOrderJSONSchema(),
			"title": map[string]any{
				"type": "string",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Markdown implementation plan body without front matter.",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"planned"},
			},
		},
	}
}

func quizModelPlanOutlineJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"title",
			"status",
			"summary",
			"assumptions",
			"open_questions",
			"risks",
			"validation_strategy",
			"slices",
			"dependencies",
		},
		"properties": map[string]any{
			"title":               map[string]any{"type": "string"},
			"status":              map[string]any{"type": "string", "enum": []string{"planned"}},
			"summary":             map[string]any{"type": "string"},
			"assumptions":         stringArrayJSONSchema(),
			"open_questions":      stringArrayJSONSchema(),
			"risks":               stringArrayJSONSchema(),
			"validation_strategy": map[string]any{"type": "string"},
			"slices":              map[string]any{"type": "array", "items": quizModelPlanOutlineSliceJSONSchema()},
			"dependencies":        map[string]any{"type": "array", "items": quizModelPlanOutlineDependencyJSONSchema()},
		},
	}
}

func quizModelPlanOutlineSliceJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"id",
			"short_name",
			"objective",
			"rationale",
			"dependencies",
			"branch_name",
			"pr_title",
			"report_path",
			"complexity",
			"implementation_boundary",
		},
		"properties": map[string]any{
			"id":                      map[string]any{"type": "string"},
			"short_name":              map[string]any{"type": "string"},
			"objective":               map[string]any{"type": "string"},
			"rationale":               map[string]any{"type": "string"},
			"dependencies":            stringArrayJSONSchema(),
			"branch_name":             map[string]any{"type": "string"},
			"pr_title":                map[string]any{"type": "string"},
			"report_path":             map[string]any{"type": "string"},
			"complexity":              map[string]any{"type": "string"},
			"implementation_boundary": map[string]any{"type": "string"},
		},
	}
}

func quizModelPlanOutlineDependencyJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"from_slice_id", "to_slice_id", "reason"},
		"properties": map[string]any{
			"from_slice_id": map[string]any{"type": "string"},
			"to_slice_id":   map[string]any{"type": "string"},
			"reason":        map[string]any{"type": "string"},
		},
	}
}

func quizModelPlanSliceDetailJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"slice_id",
			"short_name",
			"objective",
			"implementation_notes",
			"tasks",
			"acceptance_criteria",
			"ticket_metadata",
			"branch_metadata",
			"pr_expectations",
			"validation_expectations",
			"dependencies",
			"non_goals",
		},
		"properties": map[string]any{
			"slice_id":                map[string]any{"type": "string"},
			"short_name":              map[string]any{"type": "string"},
			"objective":               map[string]any{"type": "string"},
			"implementation_notes":    stringArrayJSONSchema(),
			"tasks":                   map[string]any{"type": "array", "items": quizModelPlanSliceTaskJSONSchema()},
			"acceptance_criteria":     stringArrayJSONSchema(),
			"ticket_metadata":         quizModelPlanSliceTicketMetadataJSONSchema(),
			"branch_metadata":         quizModelPlanSliceBranchMetadataJSONSchema(),
			"pr_expectations":         stringArrayJSONSchema(),
			"validation_expectations": stringArrayJSONSchema(),
			"dependencies":            stringArrayJSONSchema(),
			"non_goals":               stringArrayJSONSchema(),
		},
	}
}

func quizModelPlanSliceTaskJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"id", "summary", "commit_hint"},
		"properties": map[string]any{
			"id":          map[string]any{"type": "string"},
			"summary":     map[string]any{"type": "string"},
			"commit_hint": map[string]any{"type": "string"},
		},
	}
}

func quizModelPlanSliceTicketMetadataJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "synopsis", "priority", "severity", "type", "tags"},
		"properties": map[string]any{
			"title":    map[string]any{"type": "string"},
			"synopsis": map[string]any{"type": "string"},
			"priority": map[string]any{"type": "string"},
			"severity": map[string]any{"type": "string"},
			"type":     map[string]any{"type": "string"},
			"tags":     stringArrayJSONSchema(),
		},
	}
}

func quizModelPlanSliceBranchMetadataJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"branch_name"},
		"properties": map[string]any{
			"branch_name": map[string]any{"type": "string"},
		},
	}
}

func stringArrayJSONSchema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": map[string]any{"type": "string"},
	}
}
