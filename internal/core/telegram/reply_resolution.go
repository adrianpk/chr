package telegram

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ErrReplyPending retains an early reply while its exact peer's sender
// is still committing a definite prompt acknowledgement.
var ErrReplyPending = errors.New("telegram reply is waiting for its definite prompt binding")

type replyResolution int

const (
	replyUnknown replyResolution = iota
	replyKnown
	replyWaiting
)

func resolveReplyBinding(ctx context.Context, reader conversationReader, update ReceivedUpdate) (replyResolution, error) {
	p := update.Peer

	var known int

	err := reader.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=?`, p.BotID, p.ChatID, p.TopicID, p.UserID, update.ReplyToMessageID).Scan(&known)
	if err != nil || known > 0 {
		return replyKnown, err
	}

	var sending int

	err = reader.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_outbox WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND conversation_id IS NOT NULL AND state='sending' AND (attempts<5 OR retry_at>?)`, p.BotID, p.ChatID, p.TopicID, p.UserID, database.FormatTime(time.Now())).Scan(&sending)
	if sending > 0 {
		return replyWaiting, err
	}

	return replyUnknown, err
}

func (s *ReceiverStore) unboundReply(ctx context.Context, update ReceivedUpdate) (ActionAnswerResolution, error) {
	state, err := resolveReplyBinding(ctx, s.db, update)
	if err != nil || state == replyKnown {
		return ActionAnswerResolution{Answer: ActionAnswer{}, Recognized: false}, err
	}

	if state == replyWaiting {
		return ActionAnswerResolution{Answer: ActionAnswer{}, Recognized: true}, ErrReplyPending
	}

	return ActionAnswerResolution{Answer: ActionAnswer{}, Recognized: true}, ErrActionUnavailable
}
