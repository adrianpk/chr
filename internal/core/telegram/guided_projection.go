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
func (s *ReceiverStore) ProjectGuided(ctx context.Context, projectionInput GuidedProjectionRequest) error {
	if err := validateGuidedProgress(projectionInput.Progress, json.RawMessage(`{}`)); err != nil {
		return err
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		projection, err := guidedProjectionReady(ctx, tx, projectionInput.Reference, projectionInput.Progress)

		current, ready := projection.Conversation, projection.Ready
		if err != nil || !ready {
			return err
		}

		projectionInput.Reference.Request.Target.ConversationRevision = current.Revision

		updated, err := updateGuidedBinding(ctx, tx, projectionInput.Reference.Request, projectionInput.Progress, projectionInput.At)
		if err != nil {
			return err
		}

		native := WorkReference{Request: projectionInput.Reference.Request, ConversationRevision: updated.Revision}
		native.Request.Target.ConversationRevision = updated.Revision

		kind := "progress"
		if projectionInput.Progress.Phase != ConversationCollecting {
			kind = "result"
		} else if projectionInput.Delivery != nil {
			kind = "question"
		}

		observation := WorkObservation{Revision: uint64(updated.Revision), Phase: projectionInput.Progress.Phase, Kind: kind, Text: tgmessage.Render(projectionInput.Message).PlainFallback, Message: &projectionInput.Message}
		if err := queueWorkObservation(ctx, tx, workProjection{Reference: native, Observation: observation, ConfigDigest: projectionInput.ConfigDigest, At: projectionInput.At}); err != nil {
			return err
		}

		key := fmt.Sprintf("work:%s:%d", current.ID, updated.Revision)

		_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE bot_id=? AND event_key=?`, projectionInput.Progress.Binding, current.Peer.BotID, key)
		if err != nil {
			return err
		}

		return bindGuidedNoticeDocument(ctx, tx, guidedDocumentBinding{Request: native.Request, Key: key, Message: projectionInput.Message, Delivery: projectionInput.Delivery, ConfigDigest: projectionInput.ConfigDigest, At: projectionInput.At})
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

func bindGuidedNoticeDocument(ctx context.Context, tx *sql.Tx, binding guidedDocumentBinding) error {
	if binding.Delivery == nil {
		return nil
	}

	var id string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM telegram_outbox WHERE bot_id=? AND event_key=?`, binding.Request.Source.Peer.BotID, binding.Key).Scan(&id); err != nil {
		return err
	}

	err := tgmessage.ValidateDocument(binding.Delivery.Document)
	if err == nil && (binding.Delivery.Document.ProjectID != binding.Request.Target.ProjectID || binding.Delivery.Document.Namespace != binding.Request.Target.Namespace || binding.Delivery.Document.ProjectBindingDigest != binding.Request.Target.ProjectBindingDigest) {
		err = tgmessage.ErrDocumentUnavailable
	}

	if err == nil {
		err = bindDocumentImage(ctx, tx, documentImageBinding{DeliveryID: id, Peer: binding.Request.Source.Peer, Response: *binding.Delivery, ConfigDigest: binding.ConfigDigest, At: binding.At})
	}

	if !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		return err
	}

	binding.Message.SecondaryText = "The complete original could not be attached. Saved input is retained; restore document storage before requesting its complete image."
	presentation := tgmessage.Render(binding.Message)

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET text=?,presentation=?,presentation_hash=? WHERE id=?`, presentation.PlainFallback, string(payload), contenthash.Sum(payload), id)

	return err
}

// GuidedProjectionRequest retains one verified human view and optional original image.
type GuidedProjectionRequest struct {
	Reference    GuidedReference
	Progress     GuidedProgress
	Message      tgmessage.Blocks
	Delivery     *tgmessage.DocumentResponse
	ConfigDigest string
	At           time.Time
}

type guidedDocumentBinding struct {
	Request      OperationRequest
	Key          string
	Message      tgmessage.Blocks
	Delivery     *tgmessage.DocumentResponse
	ConfigDigest string
	At           time.Time
}
