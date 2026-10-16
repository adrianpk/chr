package core

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

// TelegramConversationQuestion identifies one canonical authoring question or turn.
// The authoring owner must supply the verified canonical binding, not quoted text.
type TelegramConversationQuestion struct {
	SessionRef    string `json:"session_ref"`
	QuestionID    string `json:"question_id"`
	TurnID        string `json:"turn_id,omitempty"`
	Revision      int64  `json:"revision"`
	ContentDigest string `json:"content_digest"`
}

func (q TelegramConversationQuestion) validate() error {
	if len(q.TurnID) > 128 {
		return errors.New("invalid canonical turn binding")
	}

	digest, err := hex.DecodeString(q.ContentDigest)
	if err != nil || len(digest) != 32 || q.Revision < 1 || q.SessionRef == "" || len(q.SessionRef) > 1024 || q.QuestionID == "" || len(q.QuestionID) > 128 {
		return errors.New("invalid canonical question binding")
	}

	return nil
}

// BindTelegramConversationQuestion replaces only this session's prompt using its
// observed revision. It neither calls an authoring provider nor answers a question.
func (s *TelegramReceiverStore) BindTelegramConversationQuestion(ctx context.Context, input TelegramConversationInput, observed int64, question TelegramConversationQuestion, now time.Time) error {
	if err := question.validate(); err != nil {
		return err
	}

	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return errors.New("telegram peer is no longer authorized")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		c, err := op.conversation(input.Conversation)
		if err != nil {
			return err
		}

		if err := op.validateTarget(c); err != nil {
			return err
		}

		if c.Revision != observed || !telegramCaptureEligible(*c) {
			return errors.New("stale or finished conversation question")
		}

		encoded, err := json.Marshal(question)
		if err != nil {
			return err
		}

		c.QuestionBinding, c.Phase = string(encoded), "clarifying"
		if err := op.saveCapture(c); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE telegram_conversations SET authoring_session_ref=? WHERE id=?`, question.SessionRef, c.ID)

		return err
	})
}
