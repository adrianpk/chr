package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	_ "modernc.org/sqlite"

	"forge.adrianpk.com/chr/internal/core/database"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// ErrTicketNotFound identifies a package boundary error.
var ErrTicketNotFound = errors.New("ticket not found")

// TicketIndex defines a package boundary interface.
type TicketIndex interface {
	UpsertTicket(ctx context.Context, ticket ticketcore.Indexed) error
	ListProjectTickets(ctx context.Context, projectID string) ([]ticketcore.Indexed, error)
	GetProjectTicketByKey(ctx context.Context, projectID, key string) (ticketcore.Indexed, error)
	DeleteProjectTickets(ctx context.Context, projectID string) error
	Close() error
}

type sqliteTicketIndex struct {
	db *sql.DB
}

// NewTicketIndexFromDB constructs the corresponding package service or value.
func NewTicketIndexFromDB(ctx context.Context, db *sql.DB, assetsFS fs.FS) (TicketIndex, error) {
	migrator := database.NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate ticket index database: %w", err)
	}

	return &sqliteTicketIndex{db: db}, nil
}

// UpsertTicket executes the corresponding package operation.
func (i *sqliteTicketIndex) UpsertTicket(ctx context.Context, ticket ticketcore.Indexed) error {
	err := validateIndexedTicket(ticket)
	if err != nil {
		return err
	}

	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ticket index transaction: %w", err)
	}

	err = upsertIndexedTicket(ctx, tx, ticket)
	if err != nil {
		return database.RollbackOnError(tx, err)
	}

	err = replaceIndexedTicketTags(ctx, tx, ticket)
	if err != nil {
		return database.RollbackOnError(tx, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM workflow_snapshots WHERE project_id = ?", ticket.ProjectID); err != nil {
		return database.RollbackOnError(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit ticket index transaction: %w", err)
	}

	return nil
}

// ListProjectTickets lists the requested data.
func (i *sqliteTicketIndex) ListProjectTickets(ctx context.Context, projectID string) ([]ticketcore.Indexed, error) {
	rows, err := i.db.QueryContext(
		ctx,
		`SELECT `+indexedTicketColumns+` FROM indexed_tickets
WHERE project_id = ?
ORDER BY key`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list indexed tickets: %w", err)
	}

	tickets := []ticketcore.Indexed{}

	for rows.Next() {
		ticket, err := scanIndexedTicket(rows)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list indexed ticket rows: %w", err)
	}

	err = rows.Close()
	if err != nil {
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

// GetProjectTicketByKey executes the corresponding package operation.
func (i *sqliteTicketIndex) GetProjectTicketByKey(ctx context.Context, projectID, key string) (ticketcore.Indexed, error) {
	row := i.db.QueryRowContext(
		ctx,
		`SELECT `+indexedTicketColumns+` FROM indexed_tickets
WHERE project_id = ? AND key = ?`,
		projectID,
		key,
	)

	ticket, err := scanIndexedTicket(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ticketcore.Indexed{}, ErrTicketNotFound
	}

	if err != nil {
		return ticketcore.Indexed{}, err
	}

	tags, err := i.ticketTags(ctx, ticket.ProjectID, ticket.Path)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	ticket.Tags = tags

	return ticket, nil
}

// DeleteProjectTickets executes the corresponding package operation.
func (i *sqliteTicketIndex) DeleteProjectTickets(ctx context.Context, projectID string) error {
	if _, err := i.db.ExecContext(ctx, "DELETE FROM workflow_snapshots WHERE project_id = ?", projectID); err != nil {
		return err
	}

	_, err := i.db.ExecContext(ctx, "DELETE FROM indexed_tickets WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("delete indexed project tickets: %w", err)
	}

	return nil
}

// Close executes the corresponding package operation.
func (i *sqliteTicketIndex) Close() error {
	err := i.db.Close()
	if err != nil {
		return fmt.Errorf("close ticket index database: %w", err)
	}

	return nil
}

const upsertIndexedTicketSQL = `INSERT INTO indexed_tickets (` + indexedTicketColumns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, path) DO UPDATE SET
    namespace = excluded.namespace,
    relative_path = excluded.relative_path,
    key = excluded.key,
    title = excluded.title,
    synopsis = excluded.synopsis,
    status = excluded.status,
    board_column = excluded.board_column,
    priority = excluded.priority,
    severity = excluded.severity,
    type = excluded.type,
    use_case = excluded.use_case,
    project_area = excluded.project_area,
    source_kind = excluded.source_kind,
    source_ref = excluded.source_ref,
    slice = excluded.slice,
    branch = excluded.branch,
    review_refs = excluded.review_refs,
    review_state = excluded.review_state,
    review_round = excluded.review_round,
    review_result = excluded.review_result,
    review_comments_url = excluded.review_comments_url,
    blocking_findings = excluded.blocking_findings,
    merge_state = excluded.merge_state,
    merge_result = excluded.merge_result,
    merge_actor = excluded.merge_actor,
    merge_commit = excluded.merge_commit,
    merged_at = excluded.merged_at,
    merge_blocker = excluded.merge_blocker,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at,
    file_mtime = excluded.file_mtime,
    file_size = excluded.file_size,
    content_hash = excluded.content_hash,
    indexed_at = excluded.indexed_at`

func upsertIndexedTicket(ctx context.Context, tx *sql.Tx, ticket ticketcore.Indexed) error {
	_, err := tx.ExecContext(ctx, upsertIndexedTicketSQL, (&indexedTicketRow{ticket: &ticket}).values()...)
	if err != nil {
		return fmt.Errorf("upsert indexed ticket: %w", err)
	}

	return nil
}

func replaceIndexedTicketTags(ctx context.Context, tx *sql.Tx, ticket ticketcore.Indexed) error {
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
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO indexed_ticket_tags (project_id, ticket_path, tag) VALUES (?, ?, ?)",
			ticket.ProjectID,
			ticket.Path,
			tag,
		)
		if err != nil {
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

		err = rows.Scan(&tag)
		if err != nil {
			return nil, fmt.Errorf("scan indexed ticket tag: %w", err)
		}

		tags = append(tags, tag)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("list indexed ticket tag rows: %w", err)
	}

	return tags, nil
}

func scanIndexedTicket(scanner indexedTicketScanner) (ticketcore.Indexed, error) {
	var ticket ticketcore.Indexed

	fields := indexedTicketRow{ticket: &ticket}

	err := scanner.Scan(fields.scanTargets()...)
	if err != nil {
		return ticketcore.Indexed{}, fmt.Errorf("scan indexed ticket: %w", err)
	}

	err = fields.assignReviewMetadata()
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	err = fields.assignTimes()
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	return ticket, nil
}

func mustJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func defaultJSONList(value string) string {
	if strings.TrimSpace(value) == "" {
		return "[]"
	}

	return value
}

func validateIndexedTicket(ticket ticketcore.Indexed) error {
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

func indexedTicketNamespace(ticket ticketcore.Indexed) string {
	if ticket.Namespace == "" {
		return opsnamespace.Default
	}

	return ticket.Namespace
}

// The indexed-ticket row reader consumes only the SQL scan operation.
type indexedTicketScanner interface{ Scan(dest ...any) error }
