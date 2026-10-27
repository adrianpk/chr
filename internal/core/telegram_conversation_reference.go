package core

import (
	"context"
	"database/sql"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TelegramConversationReference reads one exact peer-owned conversation without
// changing focus, project selection or its revision.
func (s *TelegramReceiverStore) TelegramConversationReference(ctx context.Context, peer TelegramPeerIdentity, id string) (TelegramConversation, bool, error) {
	if id == "" || len(id) > 128 {
		return TelegramConversation{}, false, workflowstate.ErrConflict
	}

	c, err := scanTelegramConversation(s.db.QueryRowContext(ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, id, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return c, false, nil
	}

	return c, err == nil, err
}
