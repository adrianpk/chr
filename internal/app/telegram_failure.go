package app

import (
	"errors"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramServiceError struct {
	code      string
	permanent bool
}

func (e *telegramServiceError) Error() string {
	return "telegram receiver: " + e.code + "; inspect telegram status and telegram doctor"
}

// ExitCode preserves the Telegram service's explicit correction-required exit
// status without letting a command library terminate an embedded application.
func ExitCode(err error) int {
	var failure *telegramServiceError
	if errors.As(err, &failure) && failure.permanent {
		return 78
	}

	return 1
}

func telegramCommandFailure(err error) error {
	if err == nil {
		return nil
	}

	var serviceFailure *telegramServiceError
	if errors.As(err, &serviceFailure) {
		return serviceFailure
	}

	var apiFailure *telegramcore.APIError
	if errors.As(err, &apiFailure) {
		return &telegramServiceError{code: telegramAPIFailureCode(apiFailure.Code), permanent: apiFailure.Permanent}
	}

	var shutdown *telegramShutdownError
	if errors.As(err, &shutdown) {
		return &telegramServiceError{code: "shutdown_not_quiescent"}
	}

	return &telegramServiceError{code: "local_process_failure"}
}

func telegramAPIFailureCode(code string) string {
	switch code {
	case "local_receiver_conflict", "receiver_conflict", "webhook_conflict", "bot_identity_changed", "credentials_or_access", "transport_failure", "response_limit_or_read_failure", "retry_delay_exceeds_limit", "rate_limited", "request_failed", "invalid_response":
		return code
	default:
		return "unclassified_api_failure"
	}
}
