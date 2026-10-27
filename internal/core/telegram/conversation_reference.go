package telegram

import (
	"context"
	"database/sql"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ConversationReference reads one exact peer-owned conversation without
// changing focus, project selection or its revision.
func (s *ReceiverStore) ConversationReference(ctx context.Context, peer PeerIdentity, id string) (Conversation, bool, error) {
	if id == "" || len(id) > 128 {
		return Conversation{}, false, workflowstate.ErrConflict
	}

	c, err := scanConversation(s.db.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, id, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return c, false, nil
	}

	return c, err == nil, err
}
