package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// QueueDocumentResponse atomically freezes a primary document with its source
// acknowledgement. Replays cannot replace bytes or allocate another snapshot.
func (s *TelegramReceiverStore) QueueDocumentResponse(ctx context.Context, update TelegramReceivedUpdate, response TelegramDocumentResponse, configDigest string, now time.Time) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, update)
		if err != nil || !pending {
			return err
		}

		return queueTelegramDocument(ctx, tx, update, response, configDigest, now)
	})
}

func queueTelegramDocument(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate, response TelegramDocumentResponse, configDigest string, now time.Time) error {
	if err := validateTelegramDocument(response.Document); err != nil {
		return err
	}

	presentation := RenderTelegramDocument(response.Message)
	if presentation.PlainFallback == "" || len(configDigest) != 64 {
		return ErrTelegramDocumentUnavailable
	}

	if err := queueTelegramResponse(ctx, tx, update, presentation.PlainFallback, configDigest, now); err != nil {
		return err
	}

	key := fmt.Sprintf("update:%d", update.UpdateID)

	var deliveryID string
	if err := tx.QueryRowContext(ctx, "SELECT id FROM telegram_outbox WHERE bot_id=? AND event_key=?", update.Peer.BotID, key).Scan(&deliveryID); err != nil {
		return err
	}

	return bindTelegramDocumentImage(ctx, tx, deliveryID, update.Peer, response, configDigest, now)
}

// The caller owns the outbox row and has validated the complete source image.
func bindTelegramDocumentImage(ctx context.Context, tx *sql.Tx, deliveryID string, peer TelegramPeerIdentity, response TelegramDocumentResponse, configDigest string, now time.Time) error {
	id := "tgdoc-" + workflowHash([]byte(deliveryID))[:32]
	if err := retainTelegramDocument(ctx, tx, id, deliveryID, peer, configDigest, response.Document, now); err != nil {
		return err
	}

	presentation := RenderTelegramDocument(response.Message)
	presentation.DocumentID, presentation.DocumentHash = id, workflowHash(response.Document.Payload)

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET text=?,presentation=?,presentation_hash=?,document_id=? WHERE id=?`, presentation.PlainFallback, string(payload), workflowHash(payload), id, deliveryID)
	if err != nil {
		return err
	}

	return checkTelegramIntakeCapacity(ctx, tx, peer.BotID)
}

func retainTelegramDocument(ctx context.Context, tx *sql.Tx, id, owner string, peer TelegramPeerIdentity, configDigest string, document TelegramDocument, now time.Time) error {
	// The outbox insertion already owns SQLite's write reservation. Quota check
	// and admission therefore serialize with other admissions and cleanup.
	result, err := tx.ExecContext(ctx, `INSERT INTO telegram_document_snapshots(id,owner_kind,owner_id,bot_id,chat_id,topic_id,user_id,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at)
 SELECT ?,'receiver',?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,? WHERE (SELECT COALESCE(SUM(length(payload)),0) FROM telegram_document_snapshots)+(SELECT COALESCE(SUM(length(payload)),0) FROM telegram_notification_documents)+?<=?`, id, owner, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID, configDigest, document.ProjectID, document.Namespace, document.ProjectBindingDigest, document.SourceRef, document.SourceHash, workflowHash(document.Payload), document.Filename, document.MediaType, len(document.Payload), document.Payload, formatProjectTime(now), len(document.Payload), telegramDocumentQuota)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrTelegramDocumentUnavailable
	}

	return nil
}

// ReadDeliveryDocument resolves bytes only through the exact original outbox,
// destination, configuration and presentation hash, including during retry.
func (s *TelegramReceiverStore) ReadDeliveryDocument(ctx context.Context, delivery TelegramReceiverDelivery) (TelegramDocument, error) {
	var (
		document TelegramDocument
		hash     string
	)

	p := delivery.Peer

	err := s.db.QueryRowContext(ctx, `SELECT d.project_id,d.namespace,d.project_binding_digest,d.source_ref,d.source_hash,d.content_hash,d.filename,d.media_type,d.payload FROM telegram_document_snapshots d JOIN telegram_outbox o ON o.id=d.owner_id AND o.document_id=d.id
 WHERE o.id=? AND d.id=? AND d.bot_id=? AND d.chat_id=? AND d.topic_id=? AND d.user_id=? AND d.config_digest=? AND o.config_digest=d.config_digest AND o.bot_id=d.bot_id AND o.chat_id=d.chat_id AND o.topic_id=d.topic_id AND o.user_id=d.user_id AND d.payload IS NOT NULL AND o.state='sending'`, delivery.ID, delivery.DocumentID, p.BotID, p.ChatID, p.TopicID, p.UserID, delivery.ConfigDigest).Scan(&document.ProjectID, &document.Namespace, &document.ProjectBindingDigest, &document.SourceRef, &document.SourceHash, &hash, &document.Filename, &document.MediaType, &document.Payload)
	if err != nil {
		return document, fmt.Errorf("%w: %w", ErrTelegramDocumentUnavailable, err)
	}

	if delivery.Presentation == nil || delivery.Presentation.DocumentID != delivery.DocumentID || delivery.Presentation.DocumentHash != hash || workflowHash(document.Payload) != hash {
		return TelegramDocument{}, ErrTelegramDocumentUnavailable
	}

	return document, validateTelegramDocument(document)
}

func purgeTelegramDocuments(ctx context.Context, tx *sql.Tx, botID int64, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_document_snapshots AS d SET payload=NULL,purged_at=? WHERE d.id IN (
 SELECT s.id FROM telegram_document_snapshots s WHERE s.bot_id=? AND s.payload IS NOT NULL AND s.terminal_at IS NOT NULL AND s.terminal_at<=?
 AND NOT EXISTS(SELECT 1 FROM telegram_outbox o WHERE o.document_id=s.id AND (o.state!='delivered' OR EXISTS(SELECT 1 FROM telegram_actions a WHERE a.token=o.action_token AND a.state IN ('pending','applying'))))
 ORDER BY s.terminal_at,s.id LIMIT 20)`, formatProjectTime(now), botID, formatProjectTime(now.Add(-7*24*time.Hour)))

	return err
}
