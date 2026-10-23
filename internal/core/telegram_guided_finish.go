package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// TelegramGuidedProgress contains only verified display and owner references.
// Canonical stage, control, work and result state remain in the project owners.
type TelegramGuidedProgress struct {
	SessionRef, Topic, Binding, Phase string
}

// FinishTelegramGuided atomically binds a confirmed application result, replies,
// source acknowledgement and delivery; transport failure never repeats the owner.
func (s *TelegramReceiverStore) FinishTelegramGuided(ctx context.Context, request TelegramOperationRequest, progress TelegramGuidedProgress, facts json.RawMessage, text, configDigest string, now time.Time) error {
	if err := validateTelegramGuidedProgress(progress, facts); err != nil {
		return err
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, request.Source)
		if err != nil || !pending {
			return err
		}

		updated, err := updateTelegramGuidedBinding(ctx, tx, request, progress, now)
		if err != nil {
			return err
		}

		request.Target.ConversationRevision = updated.Revision
		if err := finishTelegramOperation(ctx, tx, request, "confirmed", progress.SessionRef, facts, text, configDigest, now); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE request_id=?`, progress.Binding, request.ID)

		return err
	})
}

func updateTelegramGuidedBinding(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, progress TelegramGuidedProgress, now time.Time) (TelegramConversation, error) {
	p := request.Source.Peer

	c, err := scanTelegramConversation(tx.QueryRowContext(ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, request.Target.ConversationID, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err != nil {
		return c, err
	}

	if err := validateTelegramBoundConversation(c, request.Target, request.Target.ProjectBindingDigest); err != nil {
		return c, err
	}

	if c.GuidedRunID == "" || c.AuthoringSessionRef != progress.SessionRef {
		return c, ErrWorkflowConflict
	}

	c.Phase, c.QuestionBinding = progress.Phase, progress.Binding

	op := telegramConversationTransaction{ctx: ctx, tx: tx, now: now}
	if err := op.saveCapture(&c); err != nil {
		return c, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_conversations SET topic=? WHERE id=?`, progress.Topic, c.ID)

	return c, err
}

func validateTelegramGuidedProgress(progress TelegramGuidedProgress, facts json.RawMessage) error {
	if validateTelegramAuthoringSessionRef(progress.SessionRef) != nil || len(progress.Topic) > 512 || len(progress.Binding) > 4096 || !json.Valid([]byte(progress.Binding)) {
		return ErrWorkflowConflict
	}

	if progress.Phase != "collecting" && progress.Phase != "completed" && progress.Phase != "canceled" && progress.Phase != "failed" {
		return ErrWorkflowConflict
	}

	if !validTelegramOperationFinish("confirmed", progress.SessionRef, facts) {
		return ErrWorkflowConflict
	}

	return nil
}

func projectTelegramGuidedAction(ctx context.Context, tx *sql.Tx, retained TelegramActionClaim, result TelegramActionResult, now time.Time) (TelegramActionClaim, error) {
	if result.Guided == nil {
		return retained, nil
	}

	if result.State != "confirmed" || retained.Request.OperationID != "guided.step" {
		return retained, ErrWorkflowConflict
	}

	if err := validateTelegramGuidedProgress(*result.Guided, result.Facts); err != nil {
		return retained, err
	}

	updated, err := updateTelegramGuidedBinding(ctx, tx, retained.Request, *result.Guided, now)
	retained.Request.Target.ConversationRevision = updated.Revision

	return retained, err
}
