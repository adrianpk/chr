package core

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"time"
)

// ErrTelegramNotGuided lets a transport leave existing manual conversations alone.
var ErrTelegramNotGuided = errors.New("conversation is not a guided binding")

// ReserveTelegramGuidedConversation retains only routing to a canonical session.
// It does not acknowledge the update, create authoring state or grant dispatch.
func (s *TelegramReceiverStore) ReserveTelegramGuidedConversation(ctx context.Context, input TelegramConversationInput, now time.Time) (conversation TelegramConversation, target TelegramOperationTarget, err error) {
	if _, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !allowed {
		return conversation, target, ErrWorkflowConflict
	}

	input.Intent, input.Action = "develop", TelegramConversationStart
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, input.Update)
		if err != nil || !pending {
			return errors.Join(ErrWorkflowConflict, err)
		}

		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		conversation, err = op.reserveGuidedConversation()
		if err != nil {
			return err
		}

		target, err = op.operationTarget(&conversation)

		return err
	})

	return conversation, target, err
}

func (op telegramConversationTransaction) reserveGuidedConversation() (TelegramConversation, error) {
	id := telegramConversationID(op.input.Update)
	p := op.input.Update.Peer

	c, err := scanTelegramConversation(op.tx.QueryRowContext(op.ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, id, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err == nil {
		if c.GuidedRunID == "" || c.Topic != op.input.Text {
			return c, ErrWorkflowConflict
		}

		return c, op.validateTarget(&c)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return c, err
	}

	_, created, err := op.start()
	if err != nil {
		return c, err
	}

	created.GuidedRunID = created.ID
	created.AuthoringSessionRef = filepath.ToSlash(filepath.Join(DirectoryName, "authoring", "sessions", created.ID, "session.md"))

	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_conversations SET authoring_session_ref=?,guided_run_id=? WHERE id=?`, created.AuthoringSessionRef, created.ID, created.ID)
	if err != nil {
		return c, err
	}

	op.input.Conversation, op.input.Update.ReplyToMessageID = created.ID, 0
	_, _, err = op.focus()

	return *created, err
}

// ResolveTelegramGuidedTarget uses existing exact reply/focus routing and guards.
func (s *TelegramReceiverStore) ResolveTelegramGuidedTarget(ctx context.Context, input TelegramConversationInput) (conversation TelegramConversation, target TelegramOperationTarget, err error) {
	if _, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !allowed {
		return conversation, target, ErrWorkflowConflict
	}

	input.Action = TelegramConversationCapture
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}

		selected, err := op.target()
		if err != nil {
			return err
		}

		if selected.GuidedRunID == "" {
			return ErrTelegramNotGuided
		}

		if err := op.validateTarget(selected); err != nil {
			return err
		}

		conversation = *selected
		target, err = op.operationTarget(selected)

		return err
	})

	return conversation, target, err
}
