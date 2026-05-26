package chc

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DefaultTicketStatus = TicketStatusOpenLegacy

type Ticket struct {
	Key         string
	Title       string
	Synopsis    string
	Status      string
	Priority    string
	Severity    string
	Type        string
	Tags        []string
	UseCase     string
	ProjectArea string
	SourceKind  string
	SourceRef   string
	Slice       string
	Body        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TicketRenderer interface {
	RenderTicket(ticket Ticket) (string, error)
}

type ticketRenderer struct{}

func NewTicketRenderer() TicketRenderer {
	return &ticketRenderer{}
}

func (r *ticketRenderer) RenderTicket(ticket Ticket) (string, error) {
	if ticket.Status == "" {
		ticket.Status = DefaultTicketStatus
	}

	var output bytes.Buffer

	err := renderTicketFrontmatter(&output, ticket)
	if err != nil {
		return "", err
	}

	renderTicketBody(&output, ticket.Body)

	return output.String(), nil
}

func renderTicketFrontmatter(output *bytes.Buffer, ticket Ticket) error {
	_, err := output.WriteString("---\n")
	if err != nil {
		return fmt.Errorf("write ticket frontmatter: %w", err)
	}

	for _, field := range []struct {
		key   string
		value string
	}{
		{key: "key", value: ticket.Key},
		{key: "title", value: ticket.Title},
		{key: "synopsis", value: ticket.Synopsis},
		{key: "status", value: ticket.Status},
		{key: "priority", value: ticket.Priority},
		{key: "severity", value: ticket.Severity},
		{key: "type", value: ticket.Type},
	} {
		if err := writeTicketStringField(output, field.key, field.value); err != nil {
			return err
		}
	}

	err = writeTicketListField(output, "tags", ticket.Tags)
	if err != nil {
		return err
	}

	for _, field := range []struct {
		key   string
		value string
	}{
		{key: "use_case", value: ticket.UseCase},
		{key: "project_area", value: ticket.ProjectArea},
		{key: "source_kind", value: ticket.SourceKind},
		{key: "source_ref", value: ticket.SourceRef},
		{key: "slice", value: ticket.Slice},
	} {
		if err := writeTicketStringField(output, field.key, field.value); err != nil {
			return err
		}
	}

	for _, field := range []struct {
		key   string
		value time.Time
	}{
		{key: "created_at", value: ticket.CreatedAt},
		{key: "updated_at", value: ticket.UpdatedAt},
	} {
		if err := writeTicketEncodedField(output, field.key, yamlTimestamp(field.value)); err != nil {
			return err
		}
	}

	_, err = output.WriteString("---\n\n")
	if err != nil {
		return fmt.Errorf("write ticket frontmatter: %w", err)
	}

	return nil
}

func writeTicketStringField(output *bytes.Buffer, key, value string) error {
	return writeTicketEncodedField(output, key, yamlString(value))
}

func writeTicketListField(output *bytes.Buffer, key string, values []string) error {
	return writeTicketEncodedField(output, key, yamlStringList(values))
}

func writeTicketEncodedField(output *bytes.Buffer, key, value string) error {
	_, err := fmt.Fprintf(output, "%s: %s\n", key, value)
	if err != nil {
		return fmt.Errorf("write ticket frontmatter: %w", err)
	}

	return nil
}

func renderTicketBody(output *bytes.Buffer, body string) {
	if strings.TrimSpace(body) != "" {
		output.WriteString(strings.TrimSpace(body))
		output.WriteString("\n")

		return
	}

	output.WriteString("# Description\n\n")
	output.WriteString("# Checklist\n\n")
	output.WriteString("# Acceptance\n\n")
	output.WriteString("# Comments\n\n")
	output.WriteString("# Blocker\n\n")
	output.WriteString("# Validation\n")
}

func yamlString(value string) string {
	return strconv.Quote(value)
}

func yamlStringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}

	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, yamlString(value))
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

func yamlTimestamp(value time.Time) string {
	if value.IsZero() {
		return yamlString("")
	}

	return yamlString(value.UTC().Format(time.RFC3339))
}
