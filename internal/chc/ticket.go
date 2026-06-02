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
	Key               string
	Title             string
	Synopsis          string
	Status            string
	Priority          string
	Severity          string
	Type              string
	Tags              []string
	UseCase           string
	ProjectArea       string
	SourceKind        string
	SourceRef         string
	Slice             string
	Branch            string
	ReviewRefs        []TicketReviewRef
	ReviewState       string
	ReviewRound       int
	ReviewResult      string
	ReviewCommentsURL string
	BlockingFindings  []string
	MergeState        string
	MergeResult       string
	MergeActor        string
	MergeCommit       string
	MergedAt          time.Time
	MergeBlocker      string
	Body              string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TicketReviewRef struct {
	Type     string
	Provider string
	URL      string
	ID       string
	Status   string
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

	if err := renderTicketCoreFrontmatter(output, ticket); err != nil {
		return err
	}

	err = writeTicketListField(output, "tags", ticket.Tags)
	if err != nil {
		return err
	}

	if err := renderTicketSourceFrontmatter(output, ticket); err != nil {
		return err
	}

	if err := renderTicketReviewFrontmatter(output, ticket); err != nil {
		return err
	}

	if err := renderTicketMergeFrontmatter(output, ticket); err != nil {
		return err
	}

	if err := renderTicketTimeFrontmatter(output, ticket); err != nil {
		return err
	}

	_, err = output.WriteString("---\n\n")
	if err != nil {
		return fmt.Errorf("write ticket frontmatter: %w", err)
	}

	return nil
}

func renderTicketCoreFrontmatter(output *bytes.Buffer, ticket Ticket) error {
	return writeTicketStringFields(output, []ticketStringField{
		{key: "key", value: ticket.Key},
		{key: "title", value: ticket.Title},
		{key: "synopsis", value: ticket.Synopsis},
		{key: "status", value: ticket.Status},
		{key: "priority", value: ticket.Priority},
		{key: "severity", value: ticket.Severity},
		{key: "type", value: ticket.Type},
	})
}

func renderTicketSourceFrontmatter(output *bytes.Buffer, ticket Ticket) error {
	return writeTicketStringFields(output, []ticketStringField{
		{key: "use_case", value: ticket.UseCase},
		{key: "project_area", value: ticket.ProjectArea},
		{key: "source_kind", value: ticket.SourceKind},
		{key: "source_ref", value: ticket.SourceRef},
		{key: "slice", value: ticket.Slice},
		{key: "branch", value: ticket.Branch},
	})
}

func renderTicketReviewFrontmatter(output *bytes.Buffer, ticket Ticket) error {
	if err := writeTicketReviewRefsField(output, ticket.ReviewRefs); err != nil {
		return err
	}

	if err := writeTicketStringFields(output, []ticketStringField{
		{key: "review_state", value: ticket.ReviewState},
		{key: "review_result", value: ticket.ReviewResult},
		{key: "review_comments_url", value: ticket.ReviewCommentsURL},
	}); err != nil {
		return err
	}

	if err := writeTicketReviewRoundField(output, ticket.ReviewRound); err != nil {
		return err
	}

	return writeTicketListField(output, "blocking_findings", ticket.BlockingFindings)
}

func renderTicketMergeFrontmatter(output *bytes.Buffer, ticket Ticket) error {
	if err := writeTicketStringFields(output, []ticketStringField{
		{key: "merge_state", value: ticket.MergeState},
		{key: "merge_result", value: ticket.MergeResult},
		{key: "merge_actor", value: ticket.MergeActor},
		{key: "merge_commit", value: ticket.MergeCommit},
		{key: "merge_blocker", value: ticket.MergeBlocker},
	}); err != nil {
		return err
	}

	return writeTicketEncodedField(output, "merged_at", yamlTimestamp(ticket.MergedAt))
}

func renderTicketTimeFrontmatter(output *bytes.Buffer, ticket Ticket) error {
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

	return nil
}

type ticketStringField struct {
	key   string
	value string
}

func writeTicketStringFields(output *bytes.Buffer, fields []ticketStringField) error {
	for _, field := range fields {
		if err := writeTicketStringField(output, field.key, field.value); err != nil {
			return err
		}
	}

	return nil
}

func writeTicketStringField(output *bytes.Buffer, key, value string) error {
	return writeTicketEncodedField(output, key, yamlString(value))
}

func writeTicketListField(output *bytes.Buffer, key string, values []string) error {
	return writeTicketEncodedField(output, key, yamlStringList(values))
}

func writeTicketReviewRoundField(output *bytes.Buffer, value int) error {
	if value == 0 {
		return writeTicketEncodedField(output, "review_round", yamlString(""))
	}

	return writeTicketEncodedField(output, "review_round", strconv.Itoa(value))
}

func writeTicketReviewRefsField(output *bytes.Buffer, refs []TicketReviewRef) error {
	if len(refs) == 0 {
		return writeTicketEncodedField(output, "review_refs", "[]")
	}

	_, err := output.WriteString("review_refs:\n")
	if err != nil {
		return fmt.Errorf("write ticket frontmatter: %w", err)
	}

	for _, ref := range refs {
		for index, field := range []struct {
			key   string
			value string
		}{
			{key: "type", value: ref.Type},
			{key: "provider", value: ref.Provider},
			{key: "url", value: ref.URL},
			{key: "id", value: ref.ID},
			{key: "status", value: ref.Status},
		} {
			prefix := "    "
			if index == 0 {
				prefix = "  - "
			}

			_, err = fmt.Fprintf(output, "%s%s: %s\n", prefix, field.key, yamlString(field.value))
			if err != nil {
				return fmt.Errorf("write ticket frontmatter: %w", err)
			}
		}
	}

	return nil
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
