package telegram

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

// ActionAnswerCommand names the deterministic action-answer command.
const ActionAnswerCommand = "/answer"

// ActionAnswerUsage is the canonical syntax shared with command help.
const ActionAnswerUsage = ActionAnswerCommand + " <action-token> <choice>"

// ErrActionUnavailable rejects an unknown, expired or mismatched prompt binding.
var ErrActionUnavailable = errors.New("telegram action is unavailable for this exact peer and prompt")

// ActionAnswer contains only the opaque prompt token and a bounded choice.
type ActionAnswer struct {
	Token  string
	Choice string
}

// ActionCallbackData contains no actor, target, candidate or replacement guard.
func ActionCallbackData(answer ActionAnswer) (string, error) {
	if !validActionAnswer(answer) {
		return "", ErrActionUnavailable
	}

	return "tga:" + answer.Token + ":" + answer.Choice, nil
}

func validActionAnswer(answer ActionAnswer) bool {
	if len(answer.Token) != 32 {
		return false
	}

	if _, err := hex.DecodeString(answer.Token); err != nil {
		return false
	}

	return regexp.MustCompile(`^[a-z][a-z0-9_-]{0,15}$`).MatchString(answer.Choice)
}

// ResolveActionAnswer gives replies, explicit answers and callbacks one binding.
// An invalidated action reply is still recognized so it cannot become capture text.
func (s *ReceiverStore) ResolveActionAnswer(ctx context.Context, update ReceivedUpdate) (ActionAnswerResolution, error) {
	text := strings.TrimSpace(update.Text)
	if update.Kind == "callback_query" && strings.HasPrefix(text, "tga:") {
		parts := strings.Split(text, ":")
		if len(parts) != 3 {
			return ActionAnswerResolution{Answer: ActionAnswer{}, Recognized: true}, ErrActionUnavailable
		}

		return checkActionAnswer(ActionAnswer{Token: parts[1], Choice: parts[2]})
	}

	if update.Kind != "message" {
		return ActionAnswerResolution{Answer: ActionAnswer{}, Recognized: false}, nil
	}

	words := strings.Fields(text)
	if len(words) > 0 && words[0] == ActionAnswerCommand {
		if len(words) != 3 {
			return ActionAnswerResolution{Answer: ActionAnswer{}, Recognized: true}, ErrActionUnavailable
		}

		return checkActionAnswer(ActionAnswer{Token: words[1], Choice: words[2]})
	}

	return s.replyActionAnswer(ctx, update, text)
}

func checkActionAnswer(answer ActionAnswer) (ActionAnswerResolution, error) {
	if !validActionAnswer(answer) {
		return ActionAnswerResolution{Answer: answer, Recognized: true}, ErrActionUnavailable
	}

	return ActionAnswerResolution{Answer: answer, Recognized: true}, nil
}

func (s *ReceiverStore) replyActionAnswer(ctx context.Context, update ReceivedUpdate, text string) (ActionAnswerResolution, error) {
	var answer ActionAnswer
	if update.ReplyToMessageID <= 0 {
		return ActionAnswerResolution{Answer: answer, Recognized: false}, nil
	}

	var invalidated sql.NullString

	p := update.Peer

	err := s.db.QueryRowContext(ctx, `SELECT action_token,invalidated_at FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=? AND purpose='action'`, p.BotID, p.ChatID, p.TopicID, p.UserID, update.ReplyToMessageID).Scan(&answer.Token, &invalidated)
	if errors.Is(err, sql.ErrNoRows) {
		return s.unboundReply(ctx, update)
	}

	if err != nil {
		return ActionAnswerResolution{Answer: answer, Recognized: true}, err
	}

	answer.Choice = text
	if invalidated.Valid {
		if err := s.replyControlReceipt(ctx, answer.Token, p); err != nil {
			return ActionAnswerResolution{Answer: answer, Recognized: true}, err
		}
	}

	return checkActionAnswer(answer)
}

func (s *ReceiverStore) replyControlReceipt(ctx context.Context, token string, peer PeerIdentity) error {
	action, err := readAction(ctx, s.db, token, peer)
	if err != nil {
		return err
	}

	if action.AnswerRequestID == "" {
		return ErrActionUnavailable
	}

	var operation string
	if err := s.db.QueryRowContext(ctx, `SELECT operation_id FROM telegram_requests WHERE id=?`, action.RequestID).Scan(&operation); err != nil {
		return err
	}

	if operation != "work.control" {
		return ErrActionUnavailable
	}

	return nil
}
