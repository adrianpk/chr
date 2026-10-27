package core

import (
	"encoding/json"
	"fmt"
	"strings"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// UpdateWorkflowTicketDisplay prepares native display fields without changing execution state.
func UpdateWorkflowTicketDisplay(record WorkflowRecord, title, synopsis *string) (WorkflowRecord, error) {
	if record.Task == nil {
		return record, ErrNotWorkflow
	}

	fields := map[string]string{}

	if record.Metadata == nil {
		record.Metadata = map[string]json.RawMessage{}
	}

	for _, field := range []struct {
		name  string
		value *string
	}{{"title", title}, {"synopsis", synopsis}} {
		if field.value == nil {
			continue
		}

		value := strings.TrimSpace(*field.value)
		if value == "" {
			return record, fmt.Errorf("ticket %s is required", field.name)
		}

		encoded, err := json.Marshal(value)
		if err != nil {
			return record, err
		}

		record.Metadata[field.name] = encoded
		fields[field.name] = opsmetadata.Quote(value)
	}

	if !strings.HasPrefix(string(record.Body), "---\n") {
		return record, nil
	}

	body, err := updateTicketMetadata(string(record.Body), fields)
	record.Body = []byte(body)

	return record, err
}
