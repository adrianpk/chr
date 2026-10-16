package app

import (
	"slices"
	"strings"
)

type telegramControlParameters struct {
	Kind         string `json:"kind,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	QuestionID   string `json:"question_id,omitempty"`
	AnswerAction string `json:"answer_action,omitempty"`
	Option       int    `json:"option,omitempty"`
	Reason       string `json:"reason"`
}

func validateTelegramControlParameters(action string, p *telegramControlParameters) error {
	if action != "grant" && action != "answer_decision" {
		if p != nil {
			return ErrRecordUnsupported
		}

		return nil
	}

	if p == nil || strings.TrimSpace(p.Reason) == "" || len(p.Reason) > 4096 {
		return ErrRecordUnsupported
	}

	if action == "grant" {
		return validateTelegramGrantParameters(*p)
	}

	return validateTelegramAnswerParameters(*p)
}

func validateTelegramAnswerParameters(p telegramControlParameters) error {
	if p.Kind != "" || p.Amount != 0 || !validWorkflowControlIdentity(p.QuestionID) || p.Option < 1 || p.Option > 128 || !slices.Contains([]string{"resolve", "defer"}, p.AnswerAction) {
		return ErrRecordUnsupported
	}

	return nil
}

func validateTelegramGrantParameters(p telegramControlParameters) error {
	if p.QuestionID != "" || p.AnswerAction != "" || p.Option != 0 || p.Amount <= 0 {
		return ErrRecordUnsupported
	}

	if !slices.Contains([]string{"worker", "review", "closing", "merger", "drift", "submission", "tokens"}, p.Kind) {
		return ErrRecordUnsupported
	}

	return nil
}
