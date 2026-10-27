package model

import (
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// SchemaDefinition retains the named schema used by generation backends.
type SchemaDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema"`
}

// OutputSchema selects the shared strict output contract for an action.
func OutputSchema(request Request) SchemaDefinition {
	switch {
	case request.Action == ActionMaterializePlanTracking:
		return trackingFormat(request)
	case request.Action == ActionGeneratePlan:
		return newSchemaDefinition(
			"generated_plan",
			"Implementation plan generated from project and use-case context.",
			planJSONSchema(),
		)
	case request.Action == ActionGeneratePlanOutline:
		return newSchemaDefinition(
			"generated_plan_outline",
			"Plan outline generated from project and use-case context.",
			planOutlineJSONSchema(),
		)
	case request.Action == ActionGeneratePlanSliceDetail:
		return newSchemaDefinition(
			"generated_plan_slice_detail",
			"Implementation details generated for one outline slice.",
			planSliceDetailJSONSchema(),
		)
	case request.Action == ActionGenerateSpec:
		return newSchemaDefinition(
			"generated_spec",
			"Draft specification generated from project and use-case context.",
			specJSONSchema(),
		)
	case request.Action == ActionGenerateAcceptanceFeature:
		return newSchemaDefinition(
			"acceptance_feature",
			"Acceptance feature generated from project, use-case, quiz, dump, and spec context.",
			featureJSONSchema(),
		)
	case request.Action == ActionDraftSpec:
		return newSchemaDefinition(
			"quiz_draft_spec",
			"Draft specification generated from project quiz context.",
			draftSpecJSONSchema(),
		)
	case request.Mode == quizcore.ModeFast:
		return newSchemaDefinition(
			"quiz_fast_questions",
			"A batch of project quiz questions.",
			fastQuestionJSONSchema(),
		)
	default:
		return newSchemaDefinition(
			"quiz_next_question",
			"The next project quiz question.",
			questionJSONSchema(),
		)
	}
}

// newSchemaDefinition binds one named JSON schema without changing its fields.
func newSchemaDefinition(name, description string, schema map[string]any) SchemaDefinition {
	return SchemaDefinition{
		Name:        name,
		Description: description,
		Schema:      schema,
	}
}

//nolint:funlen // JSON schema literals are clearer kept together.
func planTrackingJSONSchema() map[string]any {
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

func questionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             QuestionRequiredFields,
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

func fastQuestionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"questions"},
		"properties": map[string]any{
			"questions": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": FastBatchSize,
				"items":    questionJSONSchema(),
			},
		},
	}
}

func draftSpecJSONSchema() map[string]any {
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

func specJSONSchema() map[string]any {
	return draftSpecJSONSchema()
}

func featureJSONSchema() map[string]any {
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

func planJSONSchema() map[string]any {
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

func planOutlineJSONSchema() map[string]any {
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
			"slices":              map[string]any{"type": "array", "items": planOutlineSliceJSONSchema()},
			"dependencies":        map[string]any{"type": "array", "items": planOutlineDependencyJSONSchema()},
		},
	}
}

func planOutlineSliceJSONSchema() map[string]any {
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

func planOutlineDependencyJSONSchema() map[string]any {
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

func planSliceDetailJSONSchema() map[string]any {
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
			"tasks":                   map[string]any{"type": "array", "items": planSliceTaskJSONSchema()},
			"acceptance_criteria":     stringArrayJSONSchema(),
			"ticket_metadata":         planSliceTicketMetadataJSONSchema(),
			"branch_metadata":         planSliceBranchMetadataJSONSchema(),
			"pr_expectations":         stringArrayJSONSchema(),
			"validation_expectations": stringArrayJSONSchema(),
			"dependencies":            stringArrayJSONSchema(),
			"non_goals":               stringArrayJSONSchema(),
		},
	}
}

func planSliceTaskJSONSchema() map[string]any {
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

func planSliceTicketMetadataJSONSchema() map[string]any {
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

func planSliceBranchMetadataJSONSchema() map[string]any {
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
