package model

import (
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// NewOutputContract constructs the corresponding package service or value.
func NewOutputContract(action Action, mode quizcore.Mode) OutputContractPayload {
	if action == ActionGeneratePlanSliceDetail {
		return OutputContractPayload{
			RequiredFields:    planSliceDetailRequiredFields(),
			RenderedTextField: "objective",
		}
	}

	if action == ActionGeneratePlanOutline {
		return OutputContractPayload{
			RequiredFields:    planOutlineRequiredFields(),
			RenderedTextField: "summary",
		}
	}

	if action == ActionGeneratePlan {
		return OutputContractPayload{RequiredFields: []string{"title", "body", "status", "order"}, RenderedTextField: "body"}
	}

	if action == ActionDraftSpec || action == ActionGenerateSpec {
		return OutputContractPayload{
			RequiredFields:    []string{"title", "body", "status"},
			RenderedTextField: "body",
		}
	}

	if action == ActionGenerateAcceptanceFeature {
		return OutputContractPayload{
			RequiredFields:    []string{"status"},
			RenderedTextField: "body",
		}
	}

	if action == ActionMaterializePlanTracking {
		return OutputContractPayload{
			RequiredFields:    []string{"title", "body", "status", "tickets"},
			RenderedTextField: "body",
		}
	}

	if mode == quizcore.ModeFast {
		return OutputContractPayload{
			RequiredFields:    []string{"questions"},
			RenderedTextField: "questions",
		}
	}

	return OutputContractPayload{
		RequiredFields:    append([]string(nil), QuestionRequiredFields...),
		RenderedTextField: "question",
	}
}

func planOutlineRequiredFields() []string {
	return []string{
		"title",
		"status",
		"summary",
		"assumptions",
		"open_questions",
		"risks",
		"validation_strategy",
		"slices",
		"dependencies",
	}
}

func planSliceDetailRequiredFields() []string {
	return []string{
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
	}
}
