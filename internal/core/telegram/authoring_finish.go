package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// AuthoringProgress projects confirmed owner references into a session.
// Question revisions belong to the authoring owner, not the conversation row.
type AuthoringProgress struct {
	SessionRef string                `json:"session_ref"`
	Question   *ConversationQuestion `json:"question,omitempty"`
}

// FinishAuthoring atomically links the confirmed owner result, current
// question, source acknowledgement and reply association. A failed CAS changes
// none of them; the caller can reconcile the immutable owner receipt separately.
func (s *ReceiverStore) FinishAuthoring(ctx context.Context, request OperationRequest, progress AuthoringProgress, receipt string, facts json.RawMessage, text, configDigest string, now time.Time) error {
	if err := validateAuthoringProgress(progress); err != nil {
		return err
	}

	if !validOperationFinish("confirmed", receipt, facts) {
		return workflowstate.ErrConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, request.Source)
		if err != nil || !pending {
			return err
		}

		updated, err := updateAuthoringConversation(ctx, tx, request, progress, now)
		if err != nil {
			return err
		}

		request.Target.ConversationRevision = updated.Revision
		if err := finishOperation(ctx, tx, request, "confirmed", receipt, facts, text, configDigest, now); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE request_id=?`, updated.QuestionBinding, request.ID)

		return err
	})
}

func updateAuthoringConversation(ctx context.Context, tx *sql.Tx, request OperationRequest, progress AuthoringProgress, now time.Time) (Conversation, error) {
	p := request.Source.Peer

	c, err := scanConversation(tx.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, request.Target.ConversationID, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err != nil {
		return c, err
	}

	if err := validateBoundConversation(c, request.Target, request.Target.ProjectBindingDigest); err != nil {
		return c, err
	}

	if !authoringEligible(c) {
		return c, workflowstate.ErrConflict
	}

	c.Phase, c.QuestionBinding = "prepared", ""

	if progress.Question != nil {
		encoded, err := json.Marshal(progress.Question)
		if err != nil {
			return c, err
		}

		c.Phase, c.QuestionBinding = "clarifying", string(encoded)
	}

	op := conversationTransaction{ctx: ctx, tx: tx, now: now}
	if err := op.saveCapture(&c); err != nil {
		return c, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_conversations SET authoring_session_ref=? WHERE id=?`, progress.SessionRef, c.ID)
	c.AuthoringSessionRef = progress.SessionRef

	return c, err
}

func validateAuthoringProgress(progress AuthoringProgress) error {
	if err := validateAuthoringSessionRef(progress.SessionRef); err != nil {
		return err
	}

	if progress.Question != nil {
		if err := progress.Question.validate(); err != nil {
			return err
		}

		if progress.Question.SessionRef != progress.SessionRef {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validateAuthoringSessionRef(ref string) error {
	parts := strings.Split(ref, "/")
	if len(parts) != 5 || !filepath.IsLocal(ref) || filepath.Clean(ref) != ref {
		return workflowstate.ErrConflict
	}

	if parts[0] != workspacecore.DirectoryName || parts[1] != "authoring" || parts[2] != "sessions" || parts[4] != "session.md" {
		return workflowstate.ErrConflict
	}

	if err := authoringstate.ValidateSessionID(parts[3]); err != nil {
		return err
	}

	return nil
}
