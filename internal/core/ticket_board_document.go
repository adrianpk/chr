package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

func readBoardDocument(projectDir, path string) (IndexedTicket, error) {
	data, err := readWorkflowFile(path)
	if err != nil {
		return IndexedTicket{}, err
	}

	relative, err := filepath.Rel(projectDir, path)
	if err != nil {
		return IndexedTicket{}, err
	}

	ticket := IndexedTicket{IndexedTicketIdentity: IndexedTicketIdentity{Path: path, RelativePath: filepath.ToSlash(relative), Key: filepath.ToSlash(relative)}, IndexedTicketSummary: IndexedTicketSummary{Title: filepath.Base(path)}, IndexedTicketProjection: IndexedTicketProjection{ContentHash: workflowHash(data)}}

	front, err := parseTicketFrontmatter(string(data))
	if err != nil {
		ticket.AttentionReasons = []string{err.Error()}

		return ticket, nil //nolint:nilerr // Malformed documents remain reachable as attention cards.
	}

	ticket = boardTicketMetadata(ticket, front.Metadata)
	if ticket.Synopsis == "" {
		ticket.Synopsis = boardDescription(data)
	}

	return ticket, nil
}

func boardTicketMetadata(ticket IndexedTicket, metadata map[string]string) IndexedTicket {
	value := func(name string) string { return unquoteMetadataString(metadata[name]) }
	if value("key") != "" {
		ticket.Key = value("key")
	} else if value("id") != "" {
		ticket.Key = value("id")
	} else {
		ticket.AttentionReasons = append(ticket.AttentionReasons, "ticket key is missing")
	}

	if value("title") != "" {
		ticket.Title = value("title")
	} else {
		ticket.AttentionReasons = append(ticket.AttentionReasons, "ticket title is missing")
	}

	ticket.Status, ticket.Synopsis = value("status"), value("synopsis")
	ticket.Assignee = value("assignee")
	ticket.Priority, ticket.Severity, ticket.Type = value("priority"), value("severity"), value("type")
	ticket.UseCase, ticket.ProjectArea, ticket.Slice = value("use_case"), value("project_area"), value("slice")
	ticket.SourceKind, ticket.SourceRef, ticket.Branch = value("source_kind"), value("source_ref"), value("branch")
	ticket.ReviewState, ticket.ReviewResult = value("review_state"), value("review_result")
	ticket.Tags = parseMetadataStringList(metadata["tags"])
	ticket.BoardColumn, _ = BoardTicketStatusFor(ticket.Status)

	// Repository tickets have their own status-directory vocabulary. In that
	// contract solved means implemented, whereas legacy runtime solved awaits merge.
	if value("key") == "" && strings.HasPrefix(value("id"), "TKT-") {
		ticket.Type, ticket.ProjectArea, ticket.SourceKind = value("kind"), value("scope"), value("source")
		if ticket.Status == "solved" || ticket.Status == "closed" {
			ticket.BoardColumn = TicketStatusDone
		}

		if filepath.Base(filepath.Dir(ticket.Path)) != ticket.Status {
			ticket.AttentionReasons = append(ticket.AttentionReasons, "ticket status does not match its directory")
		}
	}

	return ticket
}

func boardDescription(data []byte) string {
	text := string(data)
	if _, body, err := splitWorkflowMarkdown(data); err == nil {
		text = string(body)
	}

	lines := []string{}

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "# Comments" || trimmed == "## Comments" || trimmed == "## Discussion" {
			break
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		lines = append(lines, trimmed)
		if len(strings.Join(lines, " ")) >= 240 {
			break
		}
	}

	return fmt.Sprintf("%.240s", strings.Join(lines, " "))
}
