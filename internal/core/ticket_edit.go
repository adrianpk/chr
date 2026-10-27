package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TicketEditRequest patches supported fields against one observed source image.
// A comment is an independent operation and never includes form edits.
type TicketEditRequest struct {
	Key      string
	Revision string
	Fields   map[string]string
	Body     *string
	Comment  string
	Binding  workflowstate.MutationBinding
}

// UpdateTicketDocument keeps document edits behind existing persistence owners.
func UpdateTicketDocument(ctx context.Context, opts TicketBoardReadOptions, request TicketEditRequest) (TicketDocument, error) {
	if err := validateTicketEditRequest(request); err != nil {
		return TicketDocument{}, err
	}

	document, err := ReadTicketDocument(ctx, opts, request.Key)
	if err != nil {
		return document, err
	}

	if document.Revision != request.Revision {
		return document, workflowstate.ErrConflict
	}

	if document.ReadOnlyReason != "" {
		return document, fmt.Errorf("%w: %s", ErrWorkflowOwned, document.ReadOnlyReason)
	}

	content, err := prepareTicketEdit(document, request, time.Now().UTC())
	if err != nil {
		return document, err
	}

	committedRevision := contenthash.Sum([]byte(content))
	if document.native != nil {
		committedRevision, err = writeNativeTicketEdit(ctx, opts, document, request, content)
	} else {
		err = writeLegacyTicketFile(ctx, document.Ticket.Path, document.content, []byte(content))
	}

	if err != nil {
		return document, err
	}

	saved, err := ReadTicketDocument(ctx, opts, request.Key)
	// A later observation must not authorize a dirty form to overwrite an edit
	// made after this mutation. Return the owner's committed image separately.
	saved.CommittedRevision = committedRevision

	return saved, err
}

func validateTicketEditRequest(request TicketEditRequest) error {
	if request.Revision == "" {
		return fmt.Errorf("ticket revision is required; reload before saving")
	}

	if err := request.Binding.Validate(); err != nil {
		return err
	}

	if len(request.Fields) == 0 && request.Body == nil && request.Comment == "" {
		return fmt.Errorf("ticket changes or a comment are required")
	}

	return nil
}

func prepareTicketEdit(document TicketDocument, request TicketEditRequest, now time.Time) (string, error) {
	content, err := editableTicketContent(document)
	if err != nil {
		return "", err
	}

	if request.Comment != "" {
		return prepareTicketComment(content, request, now)
	}

	if err := validateTicketEdit(document, request); err != nil {
		return "", err
	}

	if request.Body != nil {
		content, err = replaceTicketDescription(content, *request.Body, document.Discussion)
		if err != nil {
			return "", err
		}
	}

	fields := ticketEditFields(document, request, now)
	for key, value := range fields {
		content, err = setTicketEditorField(content, key, value)
		if err != nil {
			return "", err
		}
	}

	content, err = ticketEditDelivery(content, document, request.Fields)
	if err != nil {
		return "", err
	}

	if err := validatePreparedTicketStatus(content, document, request); err != nil {
		return "", err
	}

	return content, nil
}

func prepareTicketComment(content string, request TicketEditRequest, now time.Time) (string, error) {
	if len(request.Fields) != 0 || request.Body != nil || len(request.Comment) > 64*1024 {
		return "", fmt.Errorf("comments require a separate bounded operation")
	}

	content, err := setTicketEditorField(content, "updated_at", opsmetadata.Timestamp(now))
	if err != nil {
		return "", err
	}

	return addTicketComment(content, ticketComment{Author: request.Binding.Operator, Body: request.Comment, CreatedAt: now})
}

func validatePreparedTicketStatus(content string, document TicketDocument, request TicketEditRequest) error {
	status := request.Fields["status"]
	if status == "" || document.native != nil || repositoryTicketDocument(document) {
		return nil
	}

	return validateTicketStatusTransitionMetadata(content, status)
}

func editableTicketContent(document TicketDocument) (string, error) {
	if strings.HasPrefix(string(document.content), "---\n") {
		return string(document.content), nil
	}

	if document.native == nil {
		return "", fmt.Errorf("repair invalid ticket frontmatter before editing")
	}

	record := document.native

	return NewTicketRenderer().RenderTicket(Ticket{
		Key: record.Task.Key, Title: document.Ticket.Title, Synopsis: boardDescription(record.Body),
		Status: record.Task.Status, UseCase: record.Task.UseCase,
		Priority: document.Ticket.Priority, Severity: document.Ticket.Severity, Type: document.Ticket.Type,
		ProjectArea: document.Ticket.ProjectArea, SourceKind: document.Ticket.SourceKind, Branch: document.Ticket.Branch,
		Body: string(record.Body), CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	})
}

func replaceTicketDescription(content, description, discussion string) (string, error) {
	_, end, err := ticketFrontmatterBounds(content)
	if err != nil {
		return "", err
	}

	if strings.Contains("\n"+description+"\n", "\n# Comments\n") {
		return "", fmt.Errorf("use the separate comment operation for discussion")
	}

	body := strings.TrimSpace(description) + "\n"
	if discussion != "" {
		body += "\n" + strings.TrimSpace(discussion) + "\n"
	}

	return content[:end+len("\n---")] + "\n\n" + body, nil
}

func ticketEditFields(document TicketDocument, request TicketEditRequest, now time.Time) map[string]string {
	fields := map[string]string{"updated_at": opsmetadata.Timestamp(now)}

	for key, value := range request.Fields {
		if key == "commits" || key == "pr" {
			continue
		}

		fields[ticketEditorStorageField(document, key)] = opsmetadata.Quote(strings.TrimSpace(value))
	}

	if request.Body != nil {
		fields["synopsis"] = opsmetadata.Quote(boardDescription([]byte(*request.Body)))
	}

	return fields
}

func setTicketEditorField(content, key, value string) (string, error) {
	lines, err := ticketFrontmatterLines(content)
	if err != nil {
		return "", err
	}

	for i, line := range lines {
		name, _, found := strings.Cut(line, ":")
		if !found || name != key {
			continue
		}

		end := i + 1
		for end < len(lines) && (strings.HasPrefix(lines[end], " ") || strings.HasPrefix(lines[end], "\t")) {
			end++
		}

		return replaceTicketFrontmatterLines(content, replaceStringLines(lines, i, end, []string{key + ": " + value}))
	}

	return replaceTicketFrontmatterLines(content, append(lines, key+": "+value))
}

func writeNativeTicketEdit(ctx context.Context, opts TicketBoardReadOptions, document TicketDocument, request TicketEditRequest, content string) (string, error) {
	record, err := prepareNativeTicketEdit(document, request, content)
	if err != nil {
		return "", err
	}

	result, err := NewWorkflowStore(opts.Layout, opts.ProjectID).ApplyRecordMutation(ctx, WorkflowMutationRequest{Binding: request.Binding, Record: record, ExpectedRevision: record.Revision})

	return result.Record.SourceHash, err
}

func prepareNativeTicketEdit(document TicketDocument, request TicketEditRequest, content string) (WorkflowRecord, error) {
	record := *document.native
	if err := setNativeTicketBody(&record, document, request, content); err != nil {
		return record, err
	}

	fields, err := parseTicketFrontmatter(content)
	if err != nil {
		return record, err
	}

	if err := setNativeTicketMetadata(&record, fields); err != nil {
		return record, err
	}

	if status, present := request.Fields["status"]; present {
		record.Task.Status = status
	}

	if useCase, present := request.Fields["use_case"]; present {
		record.Task.UseCase = useCase
	}

	if request.Body != nil && document.ReadOnly["body"] == "" {
		record.Task.ContractHash = contenthash.Sum([]byte(record.Task.Key + "\x00" + record.Task.UseCase + "\x00" + *request.Body))
	}

	return record, nil
}

func setNativeTicketBody(record *WorkflowRecord, document TicketDocument, request TicketEditRequest, content string) error {
	parsedMarkdown, err := workflowstate.RequireMarkdown([]byte(content))
	front, body := parsedMarkdown.Front, parsedMarkdown.Body

	if err != nil {
		return err
	}

	if strings.HasSuffix(record.Path, ".md") {
		record.Body = body

		record.Metadata["frontmatter"], err = json.Marshal(front)

		return err
	}

	if strings.HasPrefix(string(document.native.Body), "---\n") {
		record.Body = []byte(content)

		return nil
	}

	record.Body = document.native.Body
	if request.Comment != "" || request.Body != nil && document.ReadOnly["body"] == "" {
		record.Body = body
	}

	return nil
}

func setNativeTicketMetadata(record *WorkflowRecord, fields ticketFrontmatter) error {
	for _, name := range []string{"title", "synopsis", "priority", "severity", "type", "project_area", "assignee", "owner", "branch", "pr", "source_kind", "resolution"} {
		if _, exists := fields.Metadata[name]; !exists {
			continue
		}

		encoded, err := json.Marshal(opsmetadata.Unquote(fields.Metadata[name]))
		if err != nil {
			return err
		}

		record.Metadata[name] = encoded
	}

	if value, exists := fields.Metadata["commits"]; exists {
		encoded, err := json.Marshal(opsmetadata.ParseList(value))
		if err != nil {
			return err
		}

		record.Metadata["commits"] = encoded
	}

	return nil
}
