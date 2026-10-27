package telegram

import (
	"context"
	"database/sql"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ResolveAuthoringTarget observes an eligible session and the exact
// frozen target together, without acknowledging its source or invoking an owner.
func (s *ReceiverStore) ResolveAuthoringTarget(ctx context.Context, input ConversationInput) (ConversationBinding, error) {
	var (
		conversation Conversation
		target       OperationTarget
		err          error
	)

	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return ConversationBinding{Conversation: conversation, Target: target}, workflowstate.ErrConflict
	}

	input.Action = ConversationAuthor
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		selected, err := op.target()
		if err != nil {
			return err
		}

		if !authoringEligible(*selected) {
			return workflowstate.ErrConflict
		}

		target, err = op.operationTarget(selected)
		conversation = *selected

		return err
	})

	return ConversationBinding{Conversation: conversation, Target: target}, err
}

// ReadConversationBinding observes a retained request's exact peer/root
// binding even after the conversation advances. It grants no new write revision.
func (s *ReceiverStore) ReadConversationBinding(ctx context.Context, input ConversationInput, target OperationTarget) (Conversation, error) {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return Conversation{}, workflowstate.ErrConflict
	}

	p := input.Update.Peer

	c, err := scanConversation(s.db.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, target.ConversationID, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if err != nil {
		return c, err
	}

	op := conversationTransaction{ctx: ctx, input: input}
	if err := op.validateTarget(&c); err != nil {
		return c, err
	}

	target.ConversationRevision = c.Revision

	return c, validateBoundConversation(c, target, target.ProjectBindingDigest)
}
