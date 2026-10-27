package core

import (
	"encoding/json"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func workflowPlainTicketDisplay(record WorkflowRecord) (ticketcore.IndexedTicket, error) {
	data, err := json.Marshal(record.Metadata)
	if err != nil {
		return ticketcore.IndexedTicket{}, err
	}

	var display struct {
		Title       string
		Assignee    string
		Synopsis    string
		Priority    string
		Severity    string
		Type        string
		Tags        []string
		Branch      string
		ProjectArea string `json:"project_area"`
		SourceKind  string `json:"source_kind"`
		SourceRef   string `json:"source_ref"`
	}
	if err := json.Unmarshal(data, &display); err != nil {
		return ticketcore.IndexedTicket{}, err
	}

	return ticketcore.IndexedTicket{
		IndexedTicketSummary: ticketcore.IndexedTicketSummary{
			Title: display.Title, Synopsis: display.Synopsis,
			Assignee: display.Assignee,
		}, IndexedTicketClassification: ticketcore.IndexedTicketClassification{
			Priority: display.Priority,

			Severity: display.Severity, Type: display.Type, Tags: display.Tags,
			ProjectArea: display.ProjectArea,
		}, IndexedTicketSource: ticketcore.IndexedTicketSource{Branch: display.Branch, SourceKind: display.SourceKind, SourceRef: display.SourceRef},
	}, nil
}
