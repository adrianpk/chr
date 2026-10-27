package core

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

var ticketCommitReference = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

func validateTicketEdit(document TicketDocument, request TicketEditRequest) error {
	if err := validateTicketDescription(document, request.Body); err != nil {
		return err
	}

	for name, value := range request.Fields {
		if reason := document.ReadOnly[name]; reason != "" && value != document.Fields[name] {
			return fmt.Errorf("%s is read-only: %s", name, reason)
		}

		if err := validateTicketEditorField(document, name, value); err != nil {
			return err
		}
	}

	if err := validateTicketResolution(document, request); err != nil {
		return err
	}

	return validateTicketEditStatus(document, request)
}

func validateTicketDescription(document TicketDocument, body *string) error {
	if body == nil {
		return nil
	}

	if reason := document.ReadOnly["body"]; reason != "" && strings.TrimSpace(*body) != document.Description {
		return fmt.Errorf("description is read-only: %s", reason)
	}

	if strings.TrimSpace(*body) == "" || len(*body) > workflowstate.FileLimit/2 {
		return fmt.Errorf("ticket description is required and must fit the document bounds")
	}

	return nil
}

func validateTicketResolution(document TicketDocument, request TicketEditRequest) error {
	if request.Fields["resolution"] == "" {
		return nil
	}

	status := document.Fields["status"]
	if value, exists := request.Fields["status"]; exists {
		status = value
	}

	if !slices.Contains([]string{"solved", "closed", TicketStatusDone}, status) {
		return fmt.Errorf("resolution is available only for resolved tickets")
	}

	return nil
}

func validateTicketEditorField(document TicketDocument, name, value string) error {
	if !slices.Contains([]string{"title", "synopsis", "status", "priority", "severity", "type", "project_area", "source_kind", "resolution", "use_case", "assignee", "owner", "branch", "pr", "commits"}, name) {
		return fmt.Errorf("unsupported editable ticket field: %s", name)
	}

	if err := validateTicketScalarField(name, value); err != nil {
		return err
	}

	if name == "resolution" && value != "" && !slices.Contains([]string{"fixed", "wont_do", "duplicate", "obsolete", "moved", "answered"}, value) {
		return fmt.Errorf("invalid ticket resolution")
	}

	return validateTicketEditorClassification(document, name, value)
}

func validateTicketScalarField(name, value string) error {
	if len(value) > 8192 || name != "commits" && strings.ContainsAny(value, "\r\n\x00") {
		return fmt.Errorf("invalid or oversized ticket %s", name)
	}

	if (name == "title" || name == "synopsis") && strings.TrimSpace(value) == "" {
		return fmt.Errorf("ticket %s is required", name)
	}

	return nil
}

func validateTicketEditorClassification(document TicketDocument, name, value string) error {
	valid := true

	if repositoryTicketDocument(document) {
		values := map[string][]string{
			"priority":     {"unclassified", "low", "normal", "high", "urgent"},
			"severity":     {"unclassified", "low", "medium", "high", "critical"},
			"type":         {"unclassified", "bug", "task", "cleanup", "follow_up", "research", "decision"},
			"project_area": {"unclassified", "domain", "api", "persistence", "ui", "ops", "docs", "infra"},
			"source_kind":  {"chat", "manual_test", "review", "trace", "user_report", "implementation"},
		}
		if allowed, exists := values[name]; exists {
			valid = slices.Contains(allowed, value)
		}
	} else {
		switch name {
		case "priority":
			valid = isValidTicketPriority(value)
		case "severity":
			valid = isValidTicketSeverity(value)
		case "type":
			valid = isValidTicketType(value)
		}
	}

	if !valid {
		return fmt.Errorf("invalid ticket %s: %s", name, value)
	}

	return nil
}

func validateTicketEditStatus(document TicketDocument, request TicketEditRequest) error {
	status, changed := request.Fields["status"]
	if !changed || status == document.Fields["status"] {
		return nil
	}

	if !IsNormalBoardTicketStatus(status) {
		return fmt.Errorf("unsupported ticket destination")
	}

	if document.native != nil {
		if document.Phase != "pending" || !slices.Contains([]string{TicketStatusBacklog, TicketStatusReadyForDev}, status) {
			return fmt.Errorf("workflow phase and integration evidence own this transition")
		}

		return nil
	}

	if status == TicketStatusDone {
		return fmt.Errorf("moving to Done requires recorded integration through the workflow owner")
	}

	return nil
}

func ticketEditorStorageField(document TicketDocument, name string) string {
	if repositoryTicketDocument(document) {
		if mapped, exists := map[string]string{"type": "kind", "project_area": "scope", "source_kind": "source"}[name]; exists {
			return mapped
		}
	}

	return name
}

func repositoryTicketDocument(document TicketDocument) bool {
	return strings.HasPrefix(document.Fields["id"], "TKT-") && document.native == nil
}

func ticketEditDelivery(content string, document TicketDocument, fields map[string]string) (string, error) {
	var err error

	if value, exists := fields["commits"]; exists {
		commits := strings.Fields(value)
		for _, commit := range commits {
			if !ticketCommitReference.MatchString(commit) {
				return "", fmt.Errorf("invalid commit reference")
			}
		}

		content, err = setTicketEditorField(content, "commits", opsmetadata.QuoteList(commits))
		if err != nil {
			return "", err
		}
	}

	if value, exists := fields["pr"]; exists {
		return ticketEditPullRequest(content, document, value)
	}

	return content, nil
}

func ticketEditPullRequest(content string, document TicketDocument, value string) (string, error) {
	content, err := setTicketEditorField(content, "pr", opsmetadata.Quote(value))
	if err != nil || value == "" || repositoryTicketDocument(document) {
		return content, err
	}

	parsed, err := url.Parse(value)
	if err != nil || !validTicketPullRequestURL(parsed) {
		return "", fmt.Errorf("pull request reference requires an HTTP URL")
	}

	front, err := ticketmetadata.Parse(content)
	if err != nil {
		return "", err
	}

	for _, ref := range front.ReviewRefs {
		if ref.Type == "pull_request" && ref.URL == value {
			return content, nil
		}
	}

	return upsertTicketReviewRef(content, ticketmetadata.ReviewRef{Type: "pull_request", Provider: parsed.Hostname(), URL: value})
}

func validTicketPullRequestURL(parsed *url.URL) bool {
	return parsed != nil && parsed.Host != "" && parsed.User == nil && slices.Contains([]string{"http", "https"}, parsed.Scheme)
}
