package core

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
	"time"
)

// TelegramAction binds one prompt to immutable prepared input and an exact peer.
type TelegramAction struct {
	Token                string
	RequestID            string
	Peer                 TelegramPeerIdentity
	ConversationRevision int64
	GuardDigest          string
	Choices              []string
	State                string
	ExpiresAt            time.Time
	AnswerRequestID      string
}

// TelegramActionOffer contains current authority and one bounded, expiring prompt.
type TelegramActionOffer struct {
	Input             TelegramConversationInput
	RequestID         string
	Choices           []string
	Text              string
	ExpiresAt         time.Time
	ConversationPhase string
}

// TelegramActionClaim retains the original effect binding and the first answer request.
// Duplicate callback and text sources observe this same answer, never a new effect ID.
type TelegramActionClaim struct {
	Action         TelegramAction
	Request        TelegramOperationRequest
	Answer         TelegramOperationRequest
	Choice         string
	ResponseSource TelegramReceivedUpdate
	// ReceiptOnly permits recovery of an original answer, never a fresh effect.
	ReceiptOnly bool
}

const telegramActionColumns = `token,request_id,bot_id,chat_id,topic_id,user_id,conversation_revision,guard_digest,allowed_choices,state,expires_at,COALESCE(consumed_by_request,'')`

func readTelegramAction(ctx context.Context, reader telegramConversationReader, token string, peer TelegramPeerIdentity) (TelegramAction, error) {
	var (
		action          TelegramAction
		choices, expiry string
	)

	err := reader.QueryRowContext(ctx, `SELECT `+telegramActionColumns+` FROM telegram_actions WHERE token=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, token, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID).Scan(&action.Token, &action.RequestID, &action.Peer.BotID, &action.Peer.ChatID, &action.Peer.TopicID, &action.Peer.UserID, &action.ConversationRevision, &action.GuardDigest, &choices, &action.State, &expiry, &action.AnswerRequestID)
	if err != nil {
		return action, err
	}

	if err := json.Unmarshal([]byte(choices), &action.Choices); err != nil {
		return action, err
	}

	action.ExpiresAt, err = parseProjectTime(expiry)

	return action, err
}

func telegramActionGuard(request TelegramOperationRequest) string {
	return workflowHash([]byte(request.ArgumentDigest + ":" + workflowHash(request.PreparedInput)))
}

func validateTelegramActionOffer(offer TelegramActionOffer, now time.Time) error {
	if len(offer.Text) == 0 || len(offer.Text) > 16384 || !offer.ExpiresAt.After(now) || offer.ExpiresAt.After(now.Add(24*time.Hour)) {
		return errTelegramReceiverConflict
	}

	if len(offer.Choices) == 0 || len(offer.Choices) > 8 {
		return errTelegramReceiverConflict
	}

	for index, choice := range offer.Choices {
		if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,15}$`).MatchString(choice) || slices.Contains(offer.Choices[:index], choice) {
			return errTelegramReceiverConflict
		}
	}

	return nil
}

func newTelegramActionToken() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(token[:]), nil
}

func readTelegramOperationID(ctx context.Context, reader telegramConversationReader, id string) (TelegramOperationRequest, error) {
	request, err := scanTelegramOperation(reader.QueryRowContext(ctx, `SELECT `+telegramOperationColumns+` FROM telegram_requests r WHERE r.id=?`, id))
	if err != nil {
		return request, err
	}

	err = reader.QueryRowContext(ctx, `SELECT u.bot_id,u.update_id,u.payload_digest,u.update_kind,u.chat_id,u.topic_id,u.user_id,u.message_id,u.chat_type,u.payload,u.reply_to_message_id FROM telegram_updates u JOIN telegram_requests r ON r.bot_id=u.bot_id AND r.update_id=u.update_id WHERE r.id=?`, id).Scan(&request.Source.Peer.BotID, &request.Source.UpdateID, &request.Source.Digest, &request.Source.Kind, &request.Source.Peer.ChatID, &request.Source.Peer.TopicID, &request.Source.Peer.UserID, &request.Source.MessageID, &request.Source.ChatType, &request.Source.Text, &request.Source.ReplyToMessageID)

	return request, err
}

func validateTelegramActionTarget(op telegramConversationTransaction, request TelegramOperationRequest) error {
	user, ok := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)
	if !ok || user.OperatorID != request.Operator || request.Source.Peer != op.input.Update.Peer {
		return ErrWorkflowConflict
	}

	if request.Target.ConversationID == "" || request.Target.ConversationRevision <= 0 {
		return ErrWorkflowConflict
	}

	conversation, err := op.conversation(request.Target.ConversationID)
	if err != nil {
		return err
	}

	if err := validateTelegramBoundConversation(*conversation, request.Target, request.Target.ProjectBindingDigest); err != nil {
		return err
	}

	if slices.Contains([]string{"completed", "canceled", "failed"}, conversation.Phase) {
		return ErrWorkflowConflict
	}

	return op.validateTarget(conversation)
}

func telegramActionUnavailable(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrWorkflowConflict)
}

func telegramActionTargetConflict(err error) bool {
	var diagnostic telegramConversationDiagnostic

	return errors.Is(err, ErrWorkflowConflict) || errors.As(err, &diagnostic)
}
