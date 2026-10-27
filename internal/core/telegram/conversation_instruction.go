package telegram

import (
	"context"
	"database/sql"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ConversationInstruction reads the original retained source. A nonempty
// DevelopmentRef proves the explicit control still maps to the current operator;
// the final prepared scope must independently bind that source decision.
type ConversationInstruction struct {
	Text           string
	DevelopmentRef string
}

// ReadConversationInstruction recovers original user intent without using
// a model summary or treating an earlier operator mapping as a permanent grant.
func (s *ReceiverStore) ReadConversationInstruction(ctx context.Context, input ConversationInput) (ConversationInstruction, error) {
	var result ConversationInstruction
	if _, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !allowed || input.Conversation == "" {
		return result, workflowstate.ErrConflict
	}

	peer := input.Update.Peer

	c, err := scanConversation(s.db.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, input.Conversation, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID))
	if err != nil {
		return result, err
	}

	op := conversationTransaction{ctx: ctx, input: input, now: time.Now()}
	if err := op.validateTarget(&c); err != nil {
		return result, err
	}

	// Revision one is the original creation response, retained before delivery.
	// Reply invalidation and focus changes cannot change this source association.
	err = s.db.QueryRowContext(ctx, `SELECT u.payload FROM telegram_outbox o JOIN telegram_updates u ON u.bot_id=o.bot_id AND o.event_key='update:'||u.update_id WHERE o.conversation_id=? AND o.conversation_revision=1 AND o.bot_id=? AND o.chat_id=? AND o.topic_id=? AND o.user_id=? ORDER BY o.part_index LIMIT 1`, c.ID, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID).Scan(&result.Text)
	if errors.Is(err, sql.ErrNoRows) {
		return result, conversationDiagnosticError("The original instruction is unavailable. Provide a new explicit instruction; retained authority cannot be guessed.")
	}

	if err != nil {
		return result, err
	}

	op.input.Update.Text = result.Text
	if expected := op.developmentInstructionRef(c); expected != "" && expected == c.AuthorizationRef {
		result.DevelopmentRef = expected
	}

	return result, nil
}
