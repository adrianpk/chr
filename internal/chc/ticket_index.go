package chc

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type TicketIndex interface {
	UpsertTicket(ctx context.Context, ticket IndexedTicket) error
	ListProjectTickets(ctx context.Context, projectID string) ([]IndexedTicket, error)
	DeleteProjectTickets(ctx context.Context, projectID string) error
	Close() error
}

type sqliteTicketIndex struct {
	db *sql.DB
}

func NewTicketIndex(ctx context.Context, env UserEnvironment, assetsFS fs.FS) (TicketIndex, error) {
	if err := os.MkdirAll(env.DataDir(), 0o755); err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open ticket index database: %w", err)
	}

	db.SetMaxOpenConns(1)

	index, err := NewTicketIndexFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return index, nil
}

func NewTicketIndexFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (TicketIndex, error) {
	migrator := NewMigrator(assetsFS)
	if err := migrator.Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migrate ticket index database: %w", err)
	}

	return &sqliteTicketIndex{db: db}, nil
}

func (i *sqliteTicketIndex) UpsertTicket(ctx context.Context, ticket IndexedTicket) error {
	if err := validateIndexedTicket(ticket); err != nil {
		return err
	}

	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ticket index transaction: %w", err)
	}

	if err := upsertIndexedTicket(ctx, tx, ticket); err != nil {
		return rollbackWithError(tx, err)
	}

	if err := replaceIndexedTicketTags(ctx, tx, ticket); err != nil {
		return rollbackWithError(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit ticket index transaction: %w", err)
	}

	return nil
}

func (i *sqliteTicketIndex) ListProjectTickets(ctx context.Context, projectID string) ([]IndexedTicket, error) {
	rows, err := i.db.QueryContext(
		ctx,
		`SELECT project_id, path, relative_path, key, title, status, priority, severity, type,
    use_case, project_area, created_at, updated_at, file_mtime, file_size, content_hash, indexed_at
FROM indexed_tickets
WHERE project_id = ?
ORDER BY key`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list indexed tickets: %w", err)
	}

	tickets := []IndexedTicket{}

	for rows.Next() {
		ticket, err := scanIndexedTicket(rows)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list indexed ticket rows: %w", err)
	}

	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close indexed ticket rows: %w", err)
	}

	for ticketIndex := range tickets {
		tags, err := i.ticketTags(ctx, tickets[ticketIndex].ProjectID, tickets[ticketIndex].Path)
		if err != nil {
			return nil, err
		}

		tickets[ticketIndex].Tags = tags
	}

	return tickets, nil
}

func (i *sqliteTicketIndex) DeleteProjectTickets(ctx context.Context, projectID string) error {
	_, err := i.db.ExecContext(ctx, "DELETE FROM indexed_tickets WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("delete indexed project tickets: %w", err)
	}

	return nil
}

func (i *sqliteTicketIndex) Close() error {
	if err := i.db.Close(); err != nil {
		return fmt.Errorf("close ticket index database: %w", err)
	}

	return nil
}

func upsertIndexedTicket(ctx context.Context, tx *sql.Tx, ticket IndexedTicket) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO indexed_tickets (
    project_id, path, relative_path, key, title, status, priority, severity, type,
    use_case, project_area, created_at, updated_at, file_mtime, file_size, content_hash, indexed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, path) DO UPDATE SET
    relative_path = excluded.relative_path,
    key = excluded.key,
    title = excluded.title,
    status = excluded.status,
    priority = excluded.priority,
    severity = excluded.severity,
    type = excluded.type,
    use_case = excluded.use_case,
    project_area = excluded.project_area,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at,
    file_mtime = excluded.file_mtime,
    file_size = excluded.file_size,
    content_hash = excluded.content_hash,
    indexed_at = excluded.indexed_at`,
		ticket.ProjectID,
		ticket.Path,
		ticket.RelativePath,
		ticket.Key,
		ticket.Title,
		ticket.Status,
		ticket.Priority,
		ticket.Severity,
		ticket.Type,
		ticket.UseCase,
		ticket.ProjectArea,
		formatOptionalProjectTime(ticket.CreatedAt),
		formatOptionalProjectTime(ticket.UpdatedAt),
		formatProjectTime(ticket.FileMTime),
		ticket.FileSize,
		ticket.ContentHash,
		formatProjectTime(ticket.IndexedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert indexed ticket: %w", err)
	}

	return nil
}

func replaceIndexedTicketTags(ctx context.Context, tx *sql.Tx, ticket IndexedTicket) error {
	_, err := tx.ExecContext(
		ctx,
		"DELETE FROM indexed_ticket_tags WHERE project_id = ? AND ticket_path = ?",
		ticket.ProjectID,
		ticket.Path,
	)
	if err != nil {
		return fmt.Errorf("delete indexed ticket tags: %w", err)
	}

	for _, tag := range ticket.Tags {
		if _, err := tx.ExecContext(
			ctx,
			"INSERT INTO indexed_ticket_tags (project_id, ticket_path, tag) VALUES (?, ?, ?)",
			ticket.ProjectID,
			ticket.Path,
			tag,
		); err != nil {
			return fmt.Errorf("insert indexed ticket tag: %w", err)
		}
	}

	return nil
}

func (i *sqliteTicketIndex) ticketTags(ctx context.Context, projectID, ticketPath string) ([]string, error) {
	rows, err := i.db.QueryContext(
		ctx,
		`SELECT tag
FROM indexed_ticket_tags
WHERE project_id = ? AND ticket_path = ?
ORDER BY tag`,
		projectID,
		ticketPath,
	)
	if err != nil {
		return nil, fmt.Errorf("list indexed ticket tags: %w", err)
	}

	defer func() {
		//nolint:errcheck // Row iteration errors are checked through rows.Err below.
		_ = rows.Close()
	}()

	tags := []string{}

	for rows.Next() {
		var tag string

		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan indexed ticket tag: %w", err)
		}

		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list indexed ticket tag rows: %w", err)
	}

	return tags, nil
}

func scanIndexedTicket(scanner projectScanner) (IndexedTicket, error) {
	var ticket IndexedTicket

	var createdAt sql.NullString

	var updatedAt sql.NullString

	var fileMTime string

	var indexedAt string

	err := scanner.Scan(
		&ticket.ProjectID,
		&ticket.Path,
		&ticket.RelativePath,
		&ticket.Key,
		&ticket.Title,
		&ticket.Status,
		&ticket.Priority,
		&ticket.Severity,
		&ticket.Type,
		&ticket.UseCase,
		&ticket.ProjectArea,
		&createdAt,
		&updatedAt,
		&fileMTime,
		&ticket.FileSize,
		&ticket.ContentHash,
		&indexedAt,
	)
	if err != nil {
		return IndexedTicket{}, fmt.Errorf("scan indexed ticket: %w", err)
	}

	if err := assignIndexedTicketTimes(&ticket, createdAt, updatedAt, fileMTime, indexedAt); err != nil {
		return IndexedTicket{}, err
	}

	return ticket, nil
}

func assignIndexedTicketTimes(
	ticket *IndexedTicket,
	createdAt sql.NullString,
	updatedAt sql.NullString,
	fileMTime string,
	indexedAt string,
) error {
	if createdAt.Valid {
		parsedCreatedAt, err := parseProjectTime(createdAt.String)
		if err != nil {
			return err
		}

		ticket.CreatedAt = &parsedCreatedAt
	}

	if updatedAt.Valid {
		parsedUpdatedAt, err := parseProjectTime(updatedAt.String)
		if err != nil {
			return err
		}

		ticket.UpdatedAt = &parsedUpdatedAt
	}

	parsedFileMTime, err := parseProjectTime(fileMTime)
	if err != nil {
		return err
	}

	parsedIndexedAt, err := parseProjectTime(indexedAt)
	if err != nil {
		return err
	}

	ticket.FileMTime = parsedFileMTime
	ticket.IndexedAt = parsedIndexedAt

	return nil
}

func validateIndexedTicket(ticket IndexedTicket) error {
	if ticket.ProjectID == "" {
		return fmt.Errorf("indexed ticket project id is required")
	}

	if ticket.Path == "" {
		return fmt.Errorf("indexed ticket path is required")
	}

	if ticket.RelativePath == "" {
		return fmt.Errorf("indexed ticket relative path is required")
	}

	if ticket.Key == "" {
		return fmt.Errorf("indexed ticket key is required")
	}

	if ticket.IndexedAt.IsZero() {
		return fmt.Errorf("indexed ticket indexed timestamp is required")
	}

	return nil
}

func timePtr(value time.Time) *time.Time {
	return &value
}
