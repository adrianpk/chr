package core

import "encoding/json"

func workflowPlainTicketDisplay(record WorkflowRecord) (IndexedTicket, error) {
	data, err := json.Marshal(record.Metadata)
	if err != nil {
		return IndexedTicket{}, err
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
		return IndexedTicket{}, err
	}

	return IndexedTicket{
		Title: display.Title, Synopsis: display.Synopsis, Priority: display.Priority,
		Assignee: display.Assignee,
		Severity: display.Severity, Type: display.Type, Tags: display.Tags,
		Branch: display.Branch, ProjectArea: display.ProjectArea, SourceKind: display.SourceKind, SourceRef: display.SourceRef,
	}, nil
}
