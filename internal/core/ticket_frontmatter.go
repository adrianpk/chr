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
	lineStart := strings.Index(content, key+":")
	if lineStart < 0 {
		return "", fmt.Errorf("ticket %s metadata is required", key)
	}

	lineEnd := strings.IndexByte(content[lineStart:], '\n')
	if lineEnd < 0 {
		return "", fmt.Errorf("ticket %s metadata line is incomplete", key)
	}

	lineEnd += lineStart
	line := key + ": " + value

	return content[:lineStart] + line + content[lineEnd:], nil
}

func updateTicketUpdatedAt(content string, updatedAt time.Time) (string, error) {
	return updateTicketMetadataField(content, "updated_at", yamlTimestamp(updatedAt))
}
