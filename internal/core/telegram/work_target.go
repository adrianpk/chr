package telegram

import (
	"context"
	"database/sql"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ResolveTelegramWorkTarget creates a source-stable conversation when /work has
// no explicit session. No canonical work is admitted by this local allocation.
func (s *TelegramReceiverStore) ResolveTelegramWorkTarget(ctx context.Context, input TelegramConversationInput, now time.Time) (TelegramOperationTarget, error) {
	return s.resolveTelegramActionTarget(ctx, input, "develop", "Native work selection", now)
}

// ResolveTelegramControlTarget gives a standalone control its own retry-stable
// confirmation conversation. Allocation grants no native execution authority.
func (s *TelegramReceiverStore) ResolveTelegramControlTarget(ctx context.Context, input TelegramConversationInput, now time.Time) (TelegramOperationTarget, error) {
	return s.resolveTelegramActionTarget(ctx, input, "prepare", "Native workflow control", now)
}

func (s *TelegramReceiverStore) resolveTelegramActionTarget(ctx context.Context, input TelegramConversationInput, intent, topic string, now time.Time) (TelegramOperationTarget, error) {
	var target TelegramOperationTarget
	if _, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !allowed {
		return target, workflowstate.ErrConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		c, err := op.sourceActionConversation(intent, topic)
		if err != nil {
			return err
		}

		target, err = op.operationTarget(c)

		return err
	})

	return target, err
}

func (op telegramConversationTransaction) sourceActionConversation(intent, topic string) (*TelegramConversation, error) {
	c, err := op.explicitTarget()
	if err != nil || c != nil {
		return c, err
	}

	id := telegramConversationID(op.input.Update)

	var exists string

	err = op.tx.QueryRowContext(op.ctx, `SELECT id FROM telegram_conversations WHERE id=?`, id).Scan(&exists)
	if err == nil {
		return op.conversation(exists)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	op.input.Intent, op.input.Text = intent, topic

	_, c, err = op.start()
	if err != nil {
		return nil, err
	}

	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_conversations SET phase='prepared' WHERE id=?`, c.ID)
	c.Phase = "prepared"

	return c, err
}
