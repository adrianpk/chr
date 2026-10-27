package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func readBoardDocument(projectDir, path string) (ticketcore.Indexed, error) {
	data, err := opsfile.Read(path)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	relative, err := filepath.Rel(projectDir, path)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket := ticketcore.Indexed{IndexedIdentity: ticketcore.IndexedIdentity{Path: path, RelativePath: filepath.ToSlash(relative), Key: filepath.ToSlash(relative)}, IndexedSummary: ticketcore.IndexedSummary{Title: filepath.Base(path)}, IndexedProjection: ticketcore.IndexedProjection{ContentHash: contenthash.Sum(data)}}

	front, err := ticketmetadata.Parse(string(data))
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

func boardTicketMetadata(ticket ticketcore.Indexed, metadata map[string]string) ticketcore.Indexed {
	value := func(name string) string { return opsmetadata.Unquote(metadata[name]) }
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

	ticket.Status, ticket.Synopsis = ticketcore.Status(value("status")), value("synopsis")
	ticket.Assignee = value("assignee")
	ticket.Priority, ticket.Severity, ticket.Type = value("priority"), value("severity"), value("type")
	ticket.UseCase, ticket.ProjectArea, ticket.Slice = value("use_case"), value("project_area"), value("slice")
	ticket.SourceKind, ticket.SourceRef, ticket.Branch = value("source_kind"), value("source_ref"), value("branch")
	ticket.ReviewState, ticket.ReviewResult = ticketcore.ReviewState(value("review_state")), ticketcore.ReviewResult(value("review_result"))
	ticket.Tags = opsmetadata.ParseList(metadata["tags"])
	ticket.BoardColumn, _ = ticketcore.BoardStatus(ticket.Status)

	// Repository tickets have their own status-directory vocabulary. In that
	// contract solved means implemented, whereas legacy runtime solved awaits merge.
	if value("key") == "" && strings.HasPrefix(value("id"), "TKT-") {
		ticket.Type, ticket.ProjectArea, ticket.SourceKind = value("kind"), value("scope"), value("source")
		if ticket.Status == "solved" || ticket.Status == "closed" {
			ticket.BoardColumn = ticketcore.StatusDone
		}

		if filepath.Base(filepath.Dir(ticket.Path)) != string(ticket.Status) {
			ticket.AttentionReasons = append(ticket.AttentionReasons, "ticket status does not match its directory")
		}
	}

	return ticket
}

func boardDescription(data []byte) string {
	text := string(data)
	if parsed, err := workflowstate.RequireMarkdown(data); err == nil {
		text = string(parsed.Body)
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
