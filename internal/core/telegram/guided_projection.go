package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// ProjectGuided retains a changed human view and its next reply binding.
// It does not acknowledge input, recover owners, or advance canonical work.
func (s *ReceiverStore) ProjectGuided(ctx context.Context, ref GuidedReference, progress GuidedProgress, message tgmessage.Blocks, delivery *tgmessage.DocumentResponse, configDigest string, now time.Time) error {
	if err := validateGuidedProgress(progress, json.RawMessage(`{}`)); err != nil {
		return err
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		projection, err := guidedProjectionReady(ctx, tx, ref, progress)

		current, ready := projection.Conversation, projection.Ready
		if err != nil || !ready {
			return err
		}

		ref.Request.Target.ConversationRevision = current.Revision

		updated, err := updateGuidedBinding(ctx, tx, ref.Request, progress, now)
		if err != nil {
			return err
		}

		native := WorkReference{Request: ref.Request, ConversationRevision: updated.Revision}
		native.Request.Target.ConversationRevision = updated.Revision

		kind := "progress"
		if progress.Phase != "collecting" {
			kind = "result"
		} else if delivery != nil {
			kind = "question"
		}

		observation := WorkObservation{Revision: uint64(updated.Revision), Phase: progress.Phase, Kind: kind, Text: tgmessage.Render(message).PlainFallback, Message: &message}
		if err := queueWorkObservation(ctx, tx, native, observation, configDigest, now); err != nil {
			return err
		}

		key := fmt.Sprintf("work:%s:%d", current.ID, updated.Revision)

		_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE bot_id=? AND event_key=?`, progress.Binding, current.Peer.BotID, key)
		if err != nil {
			return err
		}

		return bindGuidedNoticeDocument(ctx, tx, native.Request, key, message, delivery, configDigest, now)
	})
}

func guidedProjectionReady(ctx context.Context, tx *sql.Tx, ref GuidedReference, progress GuidedProgress) (guidedProjection, error) {
	p := ref.Conversation.Peer

	c, err := scanConversation(tx.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, ref.Conversation.ID, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err != nil {
		return guidedProjection{Conversation: c, Ready: false}, err
	}

	if c.Revision != ref.Conversation.Revision || c.GuidedRunID != ref.Conversation.GuidedRunID {
		return guidedProjection{Conversation: c, Ready: false}, nil
	}

	if sameGuidedView(c.QuestionBinding, progress.Binding) {
		return guidedProjection{Conversation: c, Ready: false}, nil
	}

	var active int

	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_requests WHERE conversation_id=? AND operation_id IN ('guided.continue','guided.step') AND state IN ('ready','applying')`, c.ID).Scan(&active)

	return guidedProjection{Conversation: c, Ready: active == 0}, err
}

func sameGuidedView(before, after string) bool {
	var prior, current struct {
		ViewDigest string `json:"view_digest"`
	}
	if json.Unmarshal([]byte(before), &prior) != nil || json.Unmarshal([]byte(after), &current) != nil {
		return false
	}

	return prior.ViewDigest != "" && prior.ViewDigest == current.ViewDigest
}

func bindGuidedNoticeDocument(ctx context.Context, tx *sql.Tx, request OperationRequest, key string, message tgmessage.Blocks, delivery *tgmessage.DocumentResponse, configDigest string, now time.Time) error {
	if delivery == nil {
		return nil
	}

	var id string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM telegram_outbox WHERE bot_id=? AND event_key=?`, request.Source.Peer.BotID, key).Scan(&id); err != nil {
		return err
	}

	err := tgmessage.ValidateDocument(delivery.Document)
	if err == nil && (delivery.Document.ProjectID != request.Target.ProjectID || delivery.Document.Namespace != request.Target.Namespace || delivery.Document.ProjectBindingDigest != request.Target.ProjectBindingDigest) {
		err = tgmessage.ErrDocumentUnavailable
	}

	if err == nil {
		err = bindDocumentImage(ctx, tx, id, request.Source.Peer, *delivery, configDigest, now)
	}

	if !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		return err
	}

	message.SecondaryText = "The complete original could not be attached. Saved input is retained; restore document storage before requesting its complete image."
	presentation := tgmessage.Render(message)

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET text=?,presentation=?,presentation_hash=? WHERE id=?`, presentation.PlainFallback, string(payload), contenthash.Sum(payload), id)

	return err
}
