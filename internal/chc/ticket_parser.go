package chc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type TicketParser interface {
	ParseTicket(ctx context.Context, opts ParseTicketOptions) (IndexedTicket, error)
}

type ParseTicketOptions struct {
	ProjectID    string
	ProjectDir   string
	TicketsDir   string
	Path         string
	IndexedAt    time.Time
	RelativeBase string
}

type ticketParser struct{}

func NewTicketParser() TicketParser {
	return &ticketParser{}
}

func (p *ticketParser) ParseTicket(ctx context.Context, opts ParseTicketOptions) (IndexedTicket, error) {
	err := ctx.Err()
	if err != nil {
		return IndexedTicket{}, err
	}

	content, err := os.ReadFile(opts.Path)
	if err != nil {
		return IndexedTicket{}, fmt.Errorf("read ticket file: %w", err)
	}

	info, err := os.Stat(opts.Path)
	if err != nil {
		return IndexedTicket{}, fmt.Errorf("stat ticket file: %w", err)
	}

	metadata, err := parseTicketFrontmatter(string(content))
	if err != nil {
		return IndexedTicket{}, err
	}

	relativePath, err := relativeTicketPath(opts)
	if err != nil {
		return IndexedTicket{}, err
	}

	return indexedTicketFromMetadata(opts, relativePath, info, content, metadata)
}

func parseTicketFrontmatter(content string) (map[string]string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("ticket frontmatter opening delimiter is required")
	}

	remaining := strings.TrimPrefix(content, "---\n")

	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return nil, fmt.Errorf("ticket frontmatter closing delimiter is required")
	}

	frontmatter := remaining[:end]
	metadata := map[string]string{}

	for _, line := range strings.Split(frontmatter, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		metadata[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return metadata, nil
}

func indexedTicketFromMetadata(
	opts ParseTicketOptions,
	relativePath string,
	info os.FileInfo,
	content []byte,
	metadata map[string]string,
) (IndexedTicket, error) {
	createdAt, err := optionalTicketTime(metadata["created_at"])
	if err != nil {
		return IndexedTicket{}, err
	}

	updatedAt, err := optionalTicketTime(metadata["updated_at"])
	if err != nil {
		return IndexedTicket{}, err
	}

	ticket := IndexedTicket{
		ProjectID:    opts.ProjectID,
		Path:         opts.Path,
		RelativePath: relativePath,
		Key:          unquoteMetadataString(metadata["key"]),
		Title:        unquoteMetadataString(metadata["title"]),
		Status:       unquoteMetadataString(metadata["status"]),
		Priority:     unquoteMetadataString(metadata["priority"]),
		Severity:     unquoteMetadataString(metadata["severity"]),
		Type:         unquoteMetadataString(metadata["type"]),
		Tags:         parseMetadataStringList(metadata["tags"]),
		UseCase:      unquoteMetadataString(metadata["use_case"]),
		ProjectArea:  unquoteMetadataString(metadata["project_area"]),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		FileMTime:    info.ModTime().UTC(),
		FileSize:     info.Size(),
		ContentHash:  contentHash(content),
		IndexedAt:    opts.IndexedAt.UTC(),
	}

	err = ValidateIndexedTicket(ticket)
	if err != nil {
		return IndexedTicket{}, err
	}

	return ticket, nil
}

func relativeTicketPath(opts ParseTicketOptions) (string, error) {
	base := opts.RelativeBase
	if base == "" {
		base = opts.TicketsDir
	}

	if base == "" {
		base = filepath.Join(opts.ProjectDir, DirectoryName, "tickets")
	}

	relativePath, err := filepath.Rel(base, opts.Path)
	if err != nil {
		return "", fmt.Errorf("resolve relative ticket path: %w", err)
	}

	return filepath.ToSlash(relativePath), nil
}

func unquoteMetadataString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}

	return value
}

func parseMetadataStringList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return nil
	}

	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	parts := strings.Split(value, ",")

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		parsed := unquoteMetadataString(part)
		if parsed != "" {
			values = append(values, parsed)
		}
	}

	return values
}

func optionalTicketTime(value string) (*time.Time, error) {
	value = unquoteMetadataString(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("parse ticket timestamp: %w", err)
	}

	parsed = parsed.UTC()

	return &parsed, nil
}

func contentHash(content []byte) string {
	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:])
}
