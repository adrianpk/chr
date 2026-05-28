package chctui

import (
	"context"
	"errors"
	"strings"
)

func savedArtifactFailureStatus(actionName, artifactName string, err error) string {
	return actionName + " failed; showing last saved " + artifactName + ": " + humanStatusError(err)
}

func humanStatusError(err error) string {
	if err == nil {
		return "unknown error"
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "provider timeout"
	}

	text := strings.TrimSpace(err.Error())
	lowerText := strings.ToLower(text)

	switch {
	case strings.Contains(lowerText, "context deadline exceeded"):
		return "provider timeout"
	case strings.Contains(lowerText, "context canceled"):
		return "request cancelled"
	case text == "":
		return "unknown error"
	default:
		return text
	}
}
