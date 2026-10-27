package core

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/database"
)

const indexedTicketColumns = `project_id, namespace, path, relative_path, key, title, synopsis, status, board_column,
    priority, severity, type, use_case, project_area, source_kind, source_ref, slice,
    branch, review_refs, review_state, review_round, review_result, review_comments_url,
    blocking_findings, merge_state, merge_result, merge_actor, merge_commit, merged_at,
    merge_blocker, created_at, updated_at, file_mtime, file_size, content_hash, indexed_at`

func (f *indexedTicketRow) values() []any {
	ticket := f.ticket

	return []any{
		ticket.ProjectID,
		indexedTicketNamespace(*ticket),
		ticket.Path,
		ticket.RelativePath,
		ticket.Key,
		ticket.Title,
		ticket.Synopsis,
		ticket.Status,
		ticket.BoardColumn,
		ticket.Priority,
		ticket.Severity,
		ticket.Type,
		ticket.UseCase,
		ticket.ProjectArea,
		ticket.SourceKind,
		ticket.SourceRef,
		ticket.Slice,
		ticket.Branch,
		mustJSON(ticket.ReviewRefs),
		ticket.ReviewState,
		ticket.ReviewRound,
		ticket.ReviewResult,
		ticket.ReviewCommentsURL,
		mustJSON(ticket.BlockingFindings),
		ticket.MergeState,
		ticket.MergeResult,
		ticket.MergeActor,
		ticket.MergeCommit,
		database.OptionalTime(ticket.MergedAt),
		ticket.MergeBlocker,
		database.OptionalTime(ticket.CreatedAt),
		database.OptionalTime(ticket.UpdatedAt),
		database.FormatTime(ticket.FileMTime),
		ticket.FileSize,
		ticket.ContentHash,
		database.FormatTime(ticket.IndexedAt),
	}
}

type indexedTicketRow struct {
	ticket *IndexedTicket

	createdAt            sql.NullString
	updatedAt            sql.NullString
	mergedAt             sql.NullString
	fileMTime            string
	indexedAt            string
	reviewRefsJSON       string
	blockingFindingsJSON string
}

func (f *indexedTicketRow) scanTargets() []any {
	ticket := f.ticket

	return []any{
		&ticket.ProjectID,
		&ticket.Namespace,
		&ticket.Path,
		&ticket.RelativePath,
		&ticket.Key,
		&ticket.Title,
		&ticket.Synopsis,
		&ticket.Status,
		&ticket.BoardColumn,
		&ticket.Priority,
		&ticket.Severity,
		&ticket.Type,
		&ticket.UseCase,
		&ticket.ProjectArea,
		&ticket.SourceKind,
		&ticket.SourceRef,
		&ticket.Slice,
		&ticket.Branch,
		&f.reviewRefsJSON,
		&ticket.ReviewState,
		&ticket.ReviewRound,
		&ticket.ReviewResult,
		&ticket.ReviewCommentsURL,
		&f.blockingFindingsJSON,
		&ticket.MergeState,
		&ticket.MergeResult,
		&ticket.MergeActor,
		&ticket.MergeCommit,
		&f.mergedAt,
		&ticket.MergeBlocker,
		&f.createdAt,
		&f.updatedAt,
		&f.fileMTime,
		&ticket.FileSize,
		&ticket.ContentHash,
		&f.indexedAt,
	}
}

func (f *indexedTicketRow) assignReviewMetadata() error {
	err := json.Unmarshal([]byte(defaultJSONList(f.reviewRefsJSON)), &f.ticket.ReviewRefs)
	if err != nil {
		return fmt.Errorf("parse indexed ticket review refs: %w", err)
	}

	err = json.Unmarshal([]byte(defaultJSONList(f.blockingFindingsJSON)), &f.ticket.BlockingFindings)
	if err != nil {
		return fmt.Errorf("parse indexed ticket blocking findings: %w", err)
	}

	return nil
}
