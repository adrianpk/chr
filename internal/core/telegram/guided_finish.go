package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedProgress contains only verified display and owner references.
// Canonical stage, control, work and result state remain in the project owners.
type GuidedProgress struct {
	SessionRef, Topic, Binding, Phase string
}

// FinishGuided atomically binds a confirmed application result, replies,
// source acknowledgement and delivery; transport failure never repeats the owner.
func (s *ReceiverStore) FinishGuided(ctx context.Context, request OperationRequest, progress GuidedProgress, facts json.RawMessage, text, configDigest string, now time.Time) error {
	if err := validateGuidedProgress(progress, facts); err != nil {
		return err
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, request.Source)
		if err != nil || !pending {
			return err
		}

		updated, err := updateGuidedBinding(ctx, tx, request, progress, now)
		if err != nil {
			return err
		}

		request.Target.ConversationRevision = updated.Revision
		if err := finishOperation(ctx, tx, request, "confirmed", progress.SessionRef, facts, text, configDigest, now); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE request_id=?`, progress.Binding, request.ID)

		return err
	})
}

func updateGuidedBinding(ctx context.Context, tx *sql.Tx, request OperationRequest, progress GuidedProgress, now time.Time) (Conversation, error) {
	p := request.Source.Peer

	c, err := scanConversation(tx.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, request.Target.ConversationID, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err != nil {
		return c, err
	}

	if err := validateBoundConversation(c, request.Target, request.Target.ProjectBindingDigest); err != nil {
		return c, err
	}

	if c.GuidedRunID == "" || c.AuthoringSessionRef != progress.SessionRef {
		return c, workflowstate.ErrConflict
	}

	c.Phase, c.QuestionBinding = progress.Phase, progress.Binding

	op := conversationTransaction{ctx: ctx, tx: tx, now: now}
	if err := op.saveCapture(&c); err != nil {
		return c, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_conversations SET topic=? WHERE id=?`, progress.Topic, c.ID)

	return c, err
}

func validateGuidedProgress(progress GuidedProgress, facts json.RawMessage) error {
	if validateAuthoringSessionRef(progress.SessionRef) != nil || len(progress.Topic) > 512 || len(progress.Binding) > 4096 || !json.Valid([]byte(progress.Binding)) {
		return workflowstate.ErrConflict
	}

	if progress.Phase != "collecting" && progress.Phase != "completed" && progress.Phase != "canceled" && progress.Phase != "failed" {
		return workflowstate.ErrConflict
	}

	if !validOperationFinish("confirmed", progress.SessionRef, facts) {
		return workflowstate.ErrConflict
	}

	return nil
}

func projectGuidedAction(ctx context.Context, tx *sql.Tx, retained ActionClaim, result ActionResult, now time.Time) (ActionClaim, error) {
	if result.Guided == nil {
		return retained, nil
	}

	if result.State != "confirmed" || retained.Request.OperationID != "guided.step" {
		return retained, workflowstate.ErrConflict
	}

	if err := validateGuidedProgress(*result.Guided, result.Facts); err != nil {
		return retained, err
	}

	updated, err := updateGuidedBinding(ctx, tx, retained.Request, *result.Guided, now)
	retained.Request.Target.ConversationRevision = updated.Revision

	return retained, err
}
