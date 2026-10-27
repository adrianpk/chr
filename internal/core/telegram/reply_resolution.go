package telegram

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ErrTelegramReplyPending retains an early reply while its exact peer's sender
// is still committing a definite prompt acknowledgement.
var ErrTelegramReplyPending = errors.New("telegram reply is waiting for its definite prompt binding")

type telegramReplyResolution int

const (
	telegramReplyUnknown telegramReplyResolution = iota
	telegramReplyKnown
	telegramReplyWaiting
)

func resolveTelegramReplyBinding(ctx context.Context, reader telegramConversationReader, update TelegramReceivedUpdate) (telegramReplyResolution, error) {
	p := update.Peer

	var known int

	err := reader.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=?`, p.BotID, p.ChatID, p.TopicID, p.UserID, update.ReplyToMessageID).Scan(&known)
	if err != nil || known > 0 {
		return telegramReplyKnown, err
	}

	var sending int

	err = reader.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_outbox WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND conversation_id IS NOT NULL AND state='sending' AND (attempts<5 OR retry_at>?)`, p.BotID, p.ChatID, p.TopicID, p.UserID, database.FormatTime(time.Now())).Scan(&sending)
	if sending > 0 {
		return telegramReplyWaiting, err
	}

	return telegramReplyUnknown, err
}

func (s *TelegramReceiverStore) telegramUnboundReply(ctx context.Context, update TelegramReceivedUpdate) (TelegramActionAnswer, bool, error) {
	state, err := resolveTelegramReplyBinding(ctx, s.db, update)
	if err != nil || state == telegramReplyKnown {
		return TelegramActionAnswer{}, false, err
	}

	if state == telegramReplyWaiting {
		return TelegramActionAnswer{}, true, ErrTelegramReplyPending
	}

	return TelegramActionAnswer{}, true, ErrTelegramActionUnavailable
}
