package core

import (
	"fmt"
	"strings"
	"time"
)

func updateTicketMetadata(content string, fields map[string]string) (string, error) {
	updated := content

	for key, value := range fields {
		var err error

		updated, err = updateTicketMetadataField(updated, key, value)
		if err != nil {
			return "", err
		}
	}

	return updated, nil
}

func updateTicketMetadataField(content, key, value string) (string, error) {
	lines, err := ticketFrontmatterLines(content)
	if err != nil {
		return "", err
	}

	for _, line := range lines {
		name, _, found := strings.Cut(line, ":")
		if found && name == key {
			return setTicketFrontmatterScalarField(content, key, value, nil)
		}
	}

	return "", fmt.Errorf("ticket %s metadata is required", key)
}

func updateTicketUpdatedAt(content string, updatedAt time.Time) (string, error) {
	return updateTicketMetadataField(content, "updated_at", yamlTimestamp(updatedAt))
}
