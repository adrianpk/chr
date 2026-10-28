package metadata

import (
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// UpdateFields updates required ticket scalar fields in the original iteration order.
func UpdateFields(content string, fields map[string]string) (string, error) {
	updated := content

	for key, value := range fields {
		var err error

		updated, err = UpdateRequiredField(updated, key, value)
		if err != nil {
			return "", err
		}
	}

	return updated, nil
}

// UpdateRequiredField changes an existing scalar or returns the original missing-field error.
func UpdateRequiredField(content, key, value string) (string, error) {
	lines, err := FrontmatterLines(content)
	if err != nil {
		return "", err
	}

	for _, line := range lines {
		name, _, found := strings.Cut(line, ":")
		if found && name == key {
			return SetField(content, key, value, nil)
		}
	}

	return "", fmt.Errorf("ticket %s metadata is required", key)
}

// UpdateUpdatedAt preserves the existing UTC metadata timestamp format.
func UpdateUpdatedAt(content string, updatedAt time.Time) (string, error) {
	return UpdateRequiredField(content, "updated_at", opsmetadata.Timestamp(updatedAt))
}
