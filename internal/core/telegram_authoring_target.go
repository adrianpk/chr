package core

import (
	"context"
	"database/sql"
	"time"
)

// ResolveTelegramAuthoringTarget observes an eligible session and the exact
// frozen target together, without acknowledging its source or invoking an owner.
func (s *TelegramReceiverStore) ResolveTelegramAuthoringTarget(ctx context.Context, input TelegramConversationInput) (conversation TelegramConversation, target TelegramOperationTarget, err error) {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return conversation, target, ErrWorkflowConflict
	}

	input.Action = TelegramConversationAuthor
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		selected, err := op.target()
		if err != nil {
			return err
		}

		if !telegramAuthoringEligible(*selected) {
			return ErrWorkflowConflict
		}

		target, err = op.operationTarget(selected)
		conversation = *selected

		return err
	})

	return conversation, target, err
}

// ReadTelegramConversationBinding observes a retained request's exact peer/root
// binding even after the conversation advances. It grants no new write revision.
func (s *TelegramReceiverStore) ReadTelegramConversationBinding(ctx context.Context, input TelegramConversationInput, target TelegramOperationTarget) (TelegramConversation, error) {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return TelegramConversation{}, ErrWorkflowConflict
	}

	p := input.Update.Peer

	c, err := scanTelegramConversation(s.db.QueryRowContext(ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, target.ConversationID, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err != nil {
		return c, err
	}

	op := telegramConversationTransaction{ctx: ctx, input: input}
	if err := op.validateTarget(&c); err != nil {
		return c, err
	}

	target.ConversationRevision = c.Revision

	return c, validateTelegramBoundConversation(c, target, target.ProjectBindingDigest)
}
