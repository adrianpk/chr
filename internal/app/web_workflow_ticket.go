package app

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) workflowTicketRecords(ctx context.Context) ([]core.WorkflowRecord, error) {
	store, err := p.service(ctx).workflowStore()
	if err != nil {
		return nil, err
	}

	snapshot, err := store.InspectWorkflowRecords(ctx)
	records := []core.WorkflowRecord{}

	for _, record := range snapshot.Records {
		if record.Task != nil {
			records = append(records, record)
		}
	}

	slices.SortFunc(records, func(a, b core.WorkflowRecord) int { return strings.Compare(a.Task.Key, b.Task.Key) })

	return records, err
}

func (p webViewProvider) workflowTicketDetail(ctx context.Context, key string) (webui.TicketDetailView, bool, error) {
	records, err := p.workflowTicketRecords(ctx)
	if err != nil {
		return webui.TicketDetailView{}, false, err
	}

	for _, record := range records {
		if record.Task.Key != key {
			continue
		}

		view := webui.TicketDetailView{NativeWorkflow: true, WorkflowEnabled: true, Key: record.Task.Key, Title: webWorkflowTicketTitle(record), Synopsis: string(record.Body), Status: record.Task.Status, UseCase: record.Task.UseCase, Slice: record.Task.SliceID, RelativePath: record.Path}
		view.Revision = webStableRevision(view)

		return view, true, nil
	}

	return webui.TicketDetailView{}, false, nil
}

func webWorkflowTicketTitle(record core.WorkflowRecord) string {
	var title string
	if err := json.Unmarshal(record.Metadata["title"], &title); err == nil && title != "" {
		return title
	}

	return record.Task.Key
}
