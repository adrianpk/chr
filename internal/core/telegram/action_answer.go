package telegram

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

// TelegramActionAnswerCommand names the deterministic action-answer command.
const TelegramActionAnswerCommand = "/answer"

// TelegramActionAnswerUsage is the canonical syntax shared with command help.
const TelegramActionAnswerUsage = TelegramActionAnswerCommand + " <action-token> <choice>"

// ErrTelegramActionUnavailable rejects an unknown, expired or mismatched prompt binding.
var ErrTelegramActionUnavailable = errors.New("telegram action is unavailable for this exact peer and prompt")

// TelegramActionAnswer contains only the opaque prompt token and a bounded choice.
type TelegramActionAnswer struct {
	Token  string
	Choice string
}

// TelegramActionCallbackData contains no actor, target, candidate or replacement guard.
func TelegramActionCallbackData(answer TelegramActionAnswer) (string, error) {
	if !validTelegramActionAnswer(answer) {
		return "", ErrTelegramActionUnavailable
	}

	return "tga:" + answer.Token + ":" + answer.Choice, nil
}

func validTelegramActionAnswer(answer TelegramActionAnswer) bool {
	if len(answer.Token) != 32 {
		return false
	}

	if _, err := hex.DecodeString(answer.Token); err != nil {
		return false
	}

	return regexp.MustCompile(`^[a-z][a-z0-9_-]{0,15}$`).MatchString(answer.Choice)
}

// ResolveTelegramActionAnswer gives replies, explicit answers and callbacks one binding.
// An invalidated action reply is still recognized so it cannot become capture text.
func (s *TelegramReceiverStore) ResolveTelegramActionAnswer(ctx context.Context, update TelegramReceivedUpdate) (TelegramActionAnswer, bool, error) {
	text := strings.TrimSpace(update.Text)
	if update.Kind == "callback_query" && strings.HasPrefix(text, "tga:") {
		parts := strings.Split(text, ":")
		if len(parts) != 3 {
			return TelegramActionAnswer{}, true, ErrTelegramActionUnavailable
		}

		return checkTelegramActionAnswer(TelegramActionAnswer{Token: parts[1], Choice: parts[2]})
	}

	if update.Kind != "message" {
		return TelegramActionAnswer{}, false, nil
	}

	words := strings.Fields(text)
	if len(words) > 0 && words[0] == TelegramActionAnswerCommand {
		if len(words) != 3 {
			return TelegramActionAnswer{}, true, ErrTelegramActionUnavailable
		}

		return checkTelegramActionAnswer(TelegramActionAnswer{Token: words[1], Choice: words[2]})
	}

	return s.telegramReplyActionAnswer(ctx, update, text)
}

func checkTelegramActionAnswer(answer TelegramActionAnswer) (TelegramActionAnswer, bool, error) {
	if !validTelegramActionAnswer(answer) {
		return answer, true, ErrTelegramActionUnavailable
	}

	return answer, true, nil
}

func (s *TelegramReceiverStore) telegramReplyActionAnswer(ctx context.Context, update TelegramReceivedUpdate, text string) (TelegramActionAnswer, bool, error) {
	var answer TelegramActionAnswer
	if update.ReplyToMessageID <= 0 {
		return answer, false, nil
	}

	var invalidated sql.NullString

	p := update.Peer

	err := s.db.QueryRowContext(ctx, `SELECT action_token,invalidated_at FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=? AND purpose='action'`, p.BotID, p.ChatID, p.TopicID, p.UserID, update.ReplyToMessageID).Scan(&answer.Token, &invalidated)
	if errors.Is(err, sql.ErrNoRows) {
		return s.telegramUnboundReply(ctx, update)
	}

	if err != nil {
		return answer, true, err
	}

	answer.Choice = text
	if invalidated.Valid {
		if err := s.telegramReplyControlReceipt(ctx, answer.Token, p); err != nil {
			return answer, true, err
		}
	}

	return checkTelegramActionAnswer(answer)
}

func (s *TelegramReceiverStore) telegramReplyControlReceipt(ctx context.Context, token string, peer TelegramPeerIdentity) error {
	action, err := readTelegramAction(ctx, s.db, token, peer)
	if err != nil {
		return err
	}

	if action.AnswerRequestID == "" {
		return ErrTelegramActionUnavailable
	}

	var operation string
	if err := s.db.QueryRowContext(ctx, `SELECT operation_id FROM telegram_requests WHERE id=?`, action.RequestID).Scan(&operation); err != nil {
		return err
	}

	if operation != "work.control" {
		return ErrTelegramActionUnavailable
	}

	return nil
}
