package telegram

import (
	"context"
	"database/sql"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedReference is a transport association, never execution authority.
type GuidedReference struct {
	Conversation Conversation
	Request      OperationRequest
}

// GuidedReferences inventories a bounded page with its latest confirmed
// transport receipt. Canonical authoring and native files retain all work state.
func (s *ReceiverStore) GuidedReferences(ctx context.Context, botID int64, after string, limit int) ([]GuidedReference, error) {
	if botID <= 0 || len(after) > 128 || limit < 1 || limit > 16 {
		return nil, workflowstate.ErrConflict
	}

	rows, err := s.db.QueryContext(ctx, `SELECT c.id,(SELECT r.id FROM telegram_requests r WHERE r.conversation_id=c.id AND r.operation_id IN ('guided.continue','guided.step') AND r.state='confirmed' ORDER BY r.rowid DESC LIMIT 1) FROM telegram_conversations c WHERE c.bot_id=? AND c.id>? AND c.guided_run_id IS NOT NULL AND EXISTS(SELECT 1 FROM telegram_requests seen WHERE seen.conversation_id=c.id AND seen.operation_id IN ('guided.continue','guided.step') AND seen.state='confirmed') ORDER BY c.id LIMIT ?`, botID, after, limit)
	if err != nil {
		return nil, err
	}

	pairs, err := readGuidedPairs(rows)
	if err != nil {
		return nil, err
	}

	var result []GuidedReference

	for _, pair := range pairs {
		ref, err := s.readGuidedReference(ctx, pair)
		if err != nil {
			return nil, err
		}

		result = append(result, ref)
	}

	return result, nil
}

func readGuidedPairs(rows *sql.Rows) (pairs [][2]string, err error) {
	defer func() { err = errors.Join(err, rows.Close()) }()

	for rows.Next() {
		var (
			id      string
			request sql.NullString
		)
		if err := rows.Scan(&id, &request); err != nil {
			return nil, err
		}

		if request.Valid {
			pairs = append(pairs, [2]string{id, request.String})
		}
	}

	return pairs, rows.Err()
}

func (s *ReceiverStore) readGuidedReference(ctx context.Context, pair [2]string) (GuidedReference, error) {
	request, err := readOperationID(ctx, s.db, pair[1])
	if err != nil {
		return GuidedReference{}, err
	}

	c, found, err := s.ConversationReference(ctx, request.Source.Peer, pair[0])
	if err != nil || !found || request.Target.ConversationID != c.ID {
		return GuidedReference{}, errors.Join(workflowstate.ErrConflict, err)
	}

	request.Target.ConversationRevision = c.Revision

	return GuidedReference{Conversation: c, Request: request}, nil
}
