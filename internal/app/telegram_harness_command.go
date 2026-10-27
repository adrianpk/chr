package app

import (
	"strings"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

const telegramHarnessUsage = "/harness [worker=PROFILE|-] [generation=PROFILE|-]"

type telegramHarnessChoices struct {
	Worker     *string `json:"worker,omitempty"`
	Generation *string `json:"generation,omitempty"`
}

func parseTelegramHarness(update telegramcore.ReceivedUpdate) (telegramHarnessParsing, error) {
	var choices telegramHarnessChoices

	words := strings.Fields(update.Text)
	if update.Kind != "message" || len(words) == 0 || words[0] != "/harness" {
		return telegramHarnessParsing{Choices: choices, Recognized: false}, nil
	}

	if len(words) > 3 {
		return telegramHarnessParsing{Choices: choices, Recognized: true}, ErrRecordUnsupported
	}

	for _, word := range words[1:] {
		role, value, found := strings.Cut(word, "=")
		if !telegramHarnessAssignmentValid(found, value) {
			return telegramHarnessParsing{Choices: choices, Recognized: true}, ErrRecordUnsupported
		}

		if value == "-" {
			value = ""
		}

		if err := setTelegramHarnessChoice(&choices, role, value); err != nil {
			return telegramHarnessParsing{Choices: choices, Recognized: true}, err
		}
	}

	return telegramHarnessParsing{Choices: choices, Recognized: true}, nil
}

func telegramHarnessAssignmentValid(found bool, value string) bool {
	return found && value != "" && len(value) <= 128 && !strings.ContainsAny(value, `/\=`)
}

func setTelegramHarnessChoice(choices *telegramHarnessChoices, role, profile string) error {
	switch role {
	case "worker":
		if choices.Worker != nil {
			return ErrRecordUnsupported
		}

		choices.Worker = &profile
	case "generation":
		if choices.Generation != nil {
			return ErrRecordUnsupported
		}

		choices.Generation = &profile
	default:
		return ErrRecordUnsupported
	}

	return nil
}
