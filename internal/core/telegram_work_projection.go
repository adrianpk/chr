package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// TelegramWorkObservation is a read projection of a canonical native revision.
// It cannot grant authority or change native execution state.
type TelegramWorkObservation struct {
	Revision uint64
	Phase    string
	Kind     string
	Text     string
	Message  *TelegramMessageBlocks
}

// ProjectTelegramWork reserves an observed notice and local phase atomically.
// Original admission results and all native state remain unchanged.
func (s *TelegramReceiverStore) ProjectTelegramWork(ctx context.Context, ref TelegramWorkReference, observation TelegramWorkObservation, configDigest string, now time.Time) error {
	if observation.Revision == 0 || observation.Revision > math.MaxInt64 || observation.Text == "" || len(observation.Text) > 16384 || len(configDigest) != 64 {
		return ErrWorkflowConflict
	}

	if !validTelegramWorkObservation(observation) {
		return ErrWorkflowConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		target := ref.Request.Target
		key := fmt.Sprintf("work:%s:%d", target.ConversationID, observation.Revision)

		changed, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET last_observed_revision=?,phase=?,updated_at=? WHERE id=? AND revision=? AND invocation_id=? AND authorization_ref=? AND phase IN ('executing','waiting') AND COALESCE(last_observed_revision,0)<=? AND NOT EXISTS(SELECT 1 FROM telegram_outbox WHERE conversation_id=? AND event_key=?)`, observation.Revision, observation.Phase, formatProjectTime(now), target.ConversationID, ref.ConversationRevision, ref.Invocation.ID, ref.Request.ID, observation.Revision, target.ConversationID, key)
		if err != nil {
			return err
		}

		count, err := changed.RowsAffected()
		if err != nil || count == 0 {
			return err
		}

		return queueTelegramWorkObservation(ctx, tx, ref, observation, configDigest, now)
	})
}

func validTelegramWorkObservation(observation TelegramWorkObservation) bool {
	switch observation.Phase {
	case "completed", "canceled", "failed":
		return observation.Kind == "result"
	case "executing", "waiting":
		return observation.Kind == "progress" || observation.Kind == "diagnostic"
	default:
		return false
	}
}

func queueTelegramWorkObservation(ctx context.Context, tx *sql.Tx, ref TelegramWorkReference, observation TelegramWorkObservation, configDigest string, now time.Time) error {
	// Only never-attempted routine notices can coalesce. Questions, terminal
	// results, sending rows and uncertain retries keep their exact identity.
	if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_outbox WHERE conversation_id=? AND kind='progress' AND state='pending' AND attempts=0 AND total_attempts=0`, ref.Request.Target.ConversationID); err != nil {
		return err
	}

	var pending int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_outbox WHERE bot_id=? AND state!='delivered'`, ref.Request.Source.Peer.BotID).Scan(&pending); err != nil {
		return err
	}

	key := fmt.Sprintf("work:%s:%d", ref.Request.Target.ConversationID, observation.Revision)

	parts := TelegramResponseParts(observation.Text, "workflow:"+ref.Invocation.ID)

	var presentation []byte

	if observation.Message != nil {
		view := RenderTelegramMessage(*observation.Message)
		parts = []string{view.PlainFallback}

		var err error

		presentation, err = json.Marshal(view)
		if err != nil {
			return err
		}
	}

	if pending+len(parts) > 1024 {
		return fmt.Errorf("telegram outbox capacity exhausted")
	}

	for index, part := range parts {
		if err := insertTelegramWorkNotice(ctx, tx, ref, observation.Kind, key, part, configDigest, index+1, len(parts), now); err != nil {
			return err
		}

		if len(presentation) != 0 {
			if _, err := tx.ExecContext(ctx, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE bot_id=? AND event_key=?", string(presentation), workflowHash(presentation), ref.Request.Source.Peer.BotID, key); err != nil {
				return err
			}
		}
	}

	return checkTelegramIntakeCapacity(ctx, tx, ref.Request.Source.Peer.BotID)
}

func insertTelegramWorkNotice(ctx context.Context, tx *sql.Tx, ref TelegramWorkReference, kind, key, text, configDigest string, part, parts int, now time.Time) error {
	peer, target := ref.Request.Source.Peer, ref.Request.Target
	id := "tgd-" + workflowHash(fmt.Appendf(nil, "%d:%s:%d", peer.BotID, key, part))[:32]
	_, err := tx.ExecContext(ctx, `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,conversation_id,conversation_revision,request_id,event_key,config_digest,kind,text,part_index,part_count,state,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,'pending',?,?)`, id, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID, target.ConversationID, ref.ConversationRevision, ref.Request.ID, key, configDigest, kind, text, part, parts, formatProjectTime(now), formatProjectTime(now))

	return err
}
