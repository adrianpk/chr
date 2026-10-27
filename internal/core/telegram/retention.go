package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// PreparePoll performs bounded retention and starts one cursor epoch
// after a remote retention gap. Local pending input and retry budgets survive.
func (s *ReceiverStore) PreparePoll(ctx context.Context, botID int64, now time.Time) (ReceiverState, error) {
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE telegram_receivers SET updated_at=updated_at WHERE bot_id=? AND schema_version=1`, botID); err != nil {
			return err
		}

		if err := purgeDocuments(ctx, tx, botID, now); err != nil {
			return err
		}

		if err := compactConversations(ctx, tx, botID, now); err != nil {
			return err
		}

		if err := compactStandaloneRequests(ctx, tx, botID, now); err != nil {
			return err
		}

		if err := retireSources(ctx, tx, botID, now); err != nil {
			return err
		}

		return prepareEpoch(ctx, tx, botID, now)
	})
	if err != nil {
		return ReceiverState{}, err
	}

	return s.Receiver(ctx, botID)
}

func prepareEpoch(ctx context.Context, tx *sql.Tx, botID int64, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_receivers SET cursor_epoch=cursor_epoch+1,next_update_id=NULL,retention_gap_at=?,retention_gap_from_poll_at=last_poll_at,updated_at=? WHERE bot_id=? AND schema_version=1 AND last_poll_at<? AND (retention_gap_from_poll_at IS NULL OR retention_gap_from_poll_at!=last_poll_at)`, database.FormatTime(now), database.FormatTime(now), botID, database.FormatTime(now.Add(-24*time.Hour)))

	return err
}

// A terminal source remains protected by any unresolved request, current
// conversation, active prompt/capture, or unacknowledged outbound result.
const retirableSource = `u.bot_id=? AND u.state IN ('processed','ignored','rejected') AND u.updated_at<=?
AND NOT EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id)
AND NOT EXISTS(SELECT 1 FROM telegram_capture_parts p WHERE p.bot_id=u.bot_id AND p.update_id=u.update_id)
AND NOT EXISTS(SELECT 1 FROM telegram_outbox o WHERE o.bot_id=u.bot_id AND o.event_key='update:'||u.update_id AND (o.state!='delivered' OR EXISTS(SELECT 1 FROM telegram_conversations c WHERE c.id=o.conversation_id AND (c.phase NOT IN ('completed','canceled','failed') OR EXISTS(SELECT 1 FROM telegram_peers p WHERE p.focused_conversation_id=c.id)))))`

func retireSources(ctx context.Context, tx *sql.Tx, botID int64, now time.Time) error {
	intervals, err := readRetired(ctx, tx, botID)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `SELECT u.update_id FROM telegram_updates u WHERE `+retirableSource+` ORDER BY u.update_id LIMIT 100`, botID, database.FormatTime(now.Add(-7*24*time.Hour)))
	if err != nil {
		return err
	}

	ids, err := readRetirementIDs(rows)
	if err != nil {
		return err
	}

	for _, id := range ids {
		candidate := addRetired(intervals, id)
		if len(candidate) > 1024 {
			continue
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_outbox WHERE bot_id=? AND event_key='update:'||? AND state='delivered' AND request_id IS NULL AND action_token IS NULL`, botID, id); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_updates WHERE bot_id=? AND update_id=?`, botID, id); err != nil {
			return err
		}

		intervals = candidate
	}

	encoded, err := json.Marshal(intervals)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_receivers SET retired_update_ranges=? WHERE bot_id=?`, string(encoded), botID)

	return err
}

func readRetirementIDs(rows *sql.Rows) (ids []int64, err error) {
	defer func() { err = errors.Join(err, rows.Close()) }()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		if id < 0 || id == math.MaxInt64 {
			return nil, errors.New("invalid telegram retirement identity")
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

const compactableConversation = `c.bot_id=? AND c.phase IN ('completed','canceled','failed') AND c.updated_at<=? AND c.compacted_at IS NULL
AND NOT EXISTS(SELECT 1 FROM telegram_peers p WHERE p.focused_conversation_id=c.id)
AND NOT EXISTS(SELECT 1 FROM telegram_requests r WHERE r.conversation_id=c.id AND r.state NOT IN ('confirmed','rejected'))
AND NOT EXISTS(SELECT 1 FROM telegram_actions a JOIN telegram_requests r ON r.id=a.request_id WHERE r.conversation_id=c.id AND a.state IN ('pending','applying'))
AND NOT EXISTS(SELECT 1 FROM telegram_outbox o WHERE o.conversation_id=c.id AND o.state!='delivered')`

func compactConversations(ctx context.Context, tx *sql.Tx, botID int64, now time.Time) error {
	rows, err := tx.QueryContext(ctx, `SELECT c.id FROM telegram_conversations c WHERE `+compactableConversation+` ORDER BY c.updated_at,c.id LIMIT 20`, botID, database.FormatTime(now.Add(-7*24*time.Hour)))
	if err != nil {
		return err
	}

	ids, err := readCompactionIDs(rows)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if err := compactConversation(ctx, tx, id, now); err != nil {
			return err
		}
	}

	return nil
}

func readCompactionIDs(rows *sql.Rows) (ids []string, err error) {
	defer func() { err = errors.Join(err, rows.Close()) }()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func compactConversation(ctx context.Context, tx *sql.Tx, id string, now time.Time) error {
	queries := []string{
		`DELETE FROM telegram_outbox WHERE conversation_id=? AND state='delivered'`,
		`DELETE FROM telegram_replies WHERE conversation_id=?`,
		`DELETE FROM telegram_actions WHERE request_id IN (SELECT id FROM telegram_requests WHERE conversation_id=?)`,
		`DELETE FROM telegram_capture_parts WHERE conversation_id=?`,
		`UPDATE telegram_requests SET arguments='{}',prepared_input=NULL,domain_guard='',result=NULL,model_result=NULL,compacted_at=updated_at WHERE conversation_id=? AND state IN ('confirmed','rejected')`,
		`DELETE FROM telegram_conversations WHERE id=? AND NOT EXISTS(SELECT 1 FROM telegram_requests WHERE conversation_id=telegram_conversations.id)`,
	}
	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query, id); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE telegram_updates AS u SET payload='',interpretation=NULL WHERE u.state IN ('processed','ignored','rejected') AND EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id AND r.conversation_id=?) AND NOT EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id AND (r.conversation_id IS NULL OR r.conversation_id!=?)) AND NOT EXISTS(SELECT 1 FROM telegram_capture_parts p WHERE p.bot_id=u.bot_id AND p.update_id=u.update_id) AND NOT EXISTS(SELECT 1 FROM telegram_outbox o WHERE o.bot_id=u.bot_id AND o.event_key='update:'||u.update_id AND o.state!='delivered')`, id, id); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET capture='',question_binding='',compacted_at=? WHERE id=?`, database.FormatTime(now), id)

	return err
}

func compactStandaloneRequests(ctx context.Context, tx *sql.Tx, botID int64, now time.Time) error {
	rows, err := tx.QueryContext(ctx, `SELECT r.id FROM telegram_requests r JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id WHERE r.bot_id=? AND r.conversation_id IS NULL AND r.compacted_at IS NULL AND r.state IN ('confirmed','rejected') AND r.updated_at<=? AND u.state IN ('processed','ignored','rejected')
 AND NOT EXISTS(SELECT 1 FROM telegram_conversations c WHERE c.authorization_ref=r.id)
 AND NOT EXISTS(SELECT 1 FROM telegram_actions a WHERE a.request_id=r.id OR a.consumed_by_request=r.id)
 AND NOT EXISTS(SELECT 1 FROM telegram_outbox o WHERE o.request_id=r.id AND (o.state!='delivered' OR o.conversation_id IS NOT NULL)) ORDER BY r.updated_at,r.id LIMIT 100`, botID, database.FormatTime(now.Add(-7*24*time.Hour)))
	if err != nil {
		return err
	}

	ids, err := readCompactionIDs(rows)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_outbox WHERE request_id=? AND state='delivered'`, id); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET arguments='{}',prepared_input=NULL,domain_guard='',result=NULL,model_result=NULL,compacted_at=? WHERE id=?`, database.FormatTime(now), id); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE telegram_updates AS u SET payload='',interpretation=NULL WHERE EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id AND r.id=?) AND NOT EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id AND r.compacted_at IS NULL) AND NOT EXISTS(SELECT 1 FROM telegram_capture_parts p WHERE p.bot_id=u.bot_id AND p.update_id=u.update_id)`, id); err != nil {
			return err
		}
	}

	return nil
}
