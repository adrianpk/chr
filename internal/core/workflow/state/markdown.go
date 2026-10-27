package state

import (
	"errors"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// ErrNotWorkflow preserves the existing not-adopted-record discriminator.
var ErrNotWorkflow = errors.New("document is not an adopted workflow record")

// RequireMarkdown preserves the original native envelope presence rule.
// Callers still own schema recognition, adoption, bounds and record validation.
func RequireMarkdown(data []byte) (opsmetadata.Markdown, error) {
	parsed, err := opsmetadata.SplitMarkdown(data)
	if err != nil {
		return parsed, err
	}

	if !parsed.Present {
		return opsmetadata.Markdown{}, ErrNotWorkflow
	}

	return parsed, nil
}
