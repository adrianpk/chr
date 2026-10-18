package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TicketDocument is an observed editable document with its canonical write guards.
type TicketDocument struct {
	CommittedRevision string
	Ticket            IndexedTicket
	Description       string
	Discussion        string
	Fields            map[string]string
	ReadOnly          map[string]string
	ReadOnlyReason    string
	Revision          string
	Phase             string
	content           []byte
	native            *WorkflowRecord
}

// ReadTicketDocument resolves the same identity and authority as the Board.
func ReadTicketDocument(ctx context.Context, opts TicketBoardReadOptions, key string) (document TicketDocument, err error) {
	if key == "" || len(key) > 512 || opts.ProjectID == "" || opts.Layout.ProjectDir() == "" {
		return document, fmt.Errorf("ticket identity and scope are required")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	store := NewWorkflowStore(opts.Layout, opts.ProjectID)

	read := func() error {
		var readErr error

		document, readErr = store.readTicketDocument(ctx, opts.NativeWorkflow, key)

		return readErr
	}
	if opts.NativeWorkflow {
		err = store.withWorkflowInspection(ctx, read)
	} else {
		err = read()
	}

	return document, err
}

func (s *WorkflowStore) readTicketDocument(ctx context.Context, native bool, key string) (TicketDocument, error) {
	tickets, err := s.readBoardTickets(ctx, native)
	if err != nil {
		return TicketDocument{}, err
	}

	for _, ticket := range tickets {
		if ticket.Key != key {
			continue
		}

		document, err := loadTicketDocument(ticket)
		if err != nil || !ticket.NativeWorkflow {
			return document, err
		}

		return s.nativeTicketDocument(ctx, document)
	}

	return TicketDocument{}, fmt.Errorf("ticket not found: %s", key)
}

func loadTicketDocument(ticket IndexedTicket) (TicketDocument, error) {
	content, err := readWorkflowFile(ticket.Path)
	if err != nil {
		return TicketDocument{}, err
	}

	document := TicketDocument{Ticket: ticket, content: content, Revision: workflowHash(content), Fields: map[string]string{}, ReadOnly: map[string]string{}}

	_, body, splitErr := splitWorkflowMarkdown(content)
	if splitErr != nil {
		if !ticket.NativeWorkflow {
			document.ReadOnlyReason = splitErr.Error()
		}

		return completeTicketDocument(document, content), nil
	}

	if err := loadTicketDocumentFields(&document); err != nil {
		return document, err
	}

	observed := document.Fields["key"]
	if observed == "" {
		observed = document.Fields["id"]
	}

	if !ticket.NativeWorkflow && observed != ticket.Key {
		return document, ErrWorkflowConflict
	}

	return completeTicketDocument(document, body), nil
}

func completeTicketDocument(document TicketDocument, body []byte) TicketDocument {
	document.Description, document.Discussion = splitTicketDiscussion(string(body))
	if document.Fields["title"] == "" {
		document.Fields["title"] = document.Ticket.Title
	}

	return document
}

func loadTicketDocumentFields(document *TicketDocument) error {
	fields, err := parseTicketFrontmatter(string(document.content))
	if err != nil {
		return err
	}

	documentFields(document, fields)

	return nil
}

func documentFields(document *TicketDocument, parsed ticketFrontmatter) {
	for name, value := range parsed.Metadata {
		document.Fields[name] = unquoteMetadataString(value)
	}

	fields := document.Fields
	if fields["key"] == "" && strings.HasPrefix(fields["id"], "TKT-") {
		fields["type"], fields["project_area"], fields["source_kind"] = fields["kind"], fields["scope"], fields["source"]
		document.ReadOnly["status"] = "Repository ticket transitions use the repository ticket workflow."
	}

	if fields["pr"] == "" {
		for _, ref := range parsed.ReviewRefs {
			if ref.Type == "pull_request" {
				fields["pr"] = ref.URL

				break
			}
		}
	}

	if value, present := fields["commits"]; present {
		fields["commits"] = strings.Join(parseMetadataStringList(value), "\n")
	}
}

func (s *WorkflowStore) nativeTicketDocument(ctx context.Context, document TicketDocument) (TicketDocument, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return document, err
	}

	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.Key != document.Ticket.Key {
			continue
		}

		if !nativeTicketDocumentImage(document, record) {
			return document, ErrWorkflowConflict
		}

		document.native, document.Revision, document.Phase = &record, record.SourceHash, record.Task.Phase
		applyNativeTicketFields(&document, record)

		document.Fields["key"], document.Fields["status"], document.Fields["use_case"] = record.Task.Key, record.Task.Status, record.Task.UseCase
		nativeTicketReadOnly(&document, record)
		nativeTicketOrderReadOnly(&document, record, snapshot)

		return document, nil
	}

	return document, ErrWorkflowConflict
}

func nativeTicketDocumentImage(document TicketDocument, record WorkflowRecord) bool {
	if strings.HasSuffix(record.Path, ".md") {
		return workflowHash(document.content) == record.SourceHash
	}

	return bytes.Equal(document.content, record.Body)
}

func applyNativeTicketFields(document *TicketDocument, record WorkflowRecord) {
	for _, name := range []string{"title", "synopsis", "priority", "severity", "type", "project_area", "assignee", "owner", "branch", "pr", "resolution", "source_kind"} {
		if _, present := document.Fields[name]; present {
			continue
		}

		var value string
		if json.Unmarshal(record.Metadata[name], &value) == nil && value != "" {
			document.Fields[name] = value
		}
	}

	var commits []string
	if _, present := document.Fields["commits"]; !present && json.Unmarshal(record.Metadata["commits"], &commits) == nil && len(commits) > 0 {
		document.Fields["commits"] = strings.Join(commits, "\n")
	}
}

func nativeTicketReadOnly(document *TicketDocument, record WorkflowRecord) {
	if record.Task.AttemptID != "" {
		document.ReadOnlyReason = "An active workflow claim owns this ticket."
	}

	if record.Task.SetID != "" {
		document.ReadOnly["use_case"] = "The approved slice set owns this association."
	}

	if record.Task.SetID != "" || len(record.Domain["execution_contract"]) > 0 {
		document.ReadOnly["body"] = "The approved execution contract owns this description; use its amendment operation."
	}

	if record.Task.Phase != "pending" {
		document.ReadOnly["status"] = "Workflow execution and verified integration own this status."
	}

	if record.Task.Delivery != nil {
		document.Fields["branch"], document.Fields["pr"], document.Fields["commits"] = record.Task.Delivery.Branch, record.Task.Delivery.PullRequest, record.Task.Delivery.ResultCommit

		for _, name := range []string{"branch", "pr", "commits"} {
			document.ReadOnly[name] = "The verified delivery receipt owns this value."
		}
	}
}

func nativeTicketOrderReadOnly(document *TicketDocument, record WorkflowRecord, snapshot WorkflowSnapshot) {
	for _, owner := range snapshot.Records {
		if owner.Set != nil && owner.ID == record.Task.SetID && owner.Set.PlanRevision != nil {
			document.ReadOnly["title"] = "The approved plan order owns this title; use its amendment operation."

			return
		}
	}
}

func splitTicketDiscussion(body string) (string, string) {
	start := strings.Index("\n"+body, "\n# Comments\n")
	if start < 0 {
		return strings.TrimSpace(body), ""
	}

	end := len(body)
	if next := strings.Index(body[start+len("# Comments\n"):], "\n# "); next >= 0 {
		end = start + len("# Comments\n") + next + 1
	}

	return strings.TrimSpace(body[:start] + body[end:]), body[start:end]
}
