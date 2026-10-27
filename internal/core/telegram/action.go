package telegram

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

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Action binds one prompt to immutable prepared input and an exact peer.
type Action struct {
	Token                string
	RequestID            string
	Peer                 PeerIdentity
	ConversationRevision int64
	GuardDigest          string
	GuardVersion         int
	PresentationHash     string
	Choices              []string
	State                string
	ExpiresAt            time.Time
	AnswerRequestID      string
}

// ActionOffer contains current authority and one bounded, expiring prompt.
type ActionOffer struct {
	Message           *tgmessage.Blocks
	Delivery          *tgmessage.DocumentResponse
	Input             ConversationInput
	RequestID         string
	Choices           []string
	Text              string
	ExpiresAt         time.Time
	ConversationPhase string
}

// ActionClaim retains the original effect binding and the first answer request.
// Duplicate callback and text sources observe this same answer, never a new effect ID.
type ActionClaim struct {
	Action         Action
	Request        OperationRequest
	Answer         OperationRequest
	Choice         string
	ResponseSource ReceivedUpdate
	// ReceiptOnly permits recovery of an original answer, never a fresh effect.
	ReceiptOnly bool
}

const actionColumns = `token,request_id,bot_id,chat_id,topic_id,user_id,conversation_revision,guard_digest,allowed_choices,state,expires_at,COALESCE(consumed_by_request,''),guard_version,COALESCE(presentation_hash,'')`

func readAction(ctx context.Context, reader conversationReader, token string, peer PeerIdentity) (Action, error) {
	var (
		action          Action
		choices, expiry string
	)

	err := reader.QueryRowContext(ctx, `SELECT `+actionColumns+` FROM telegram_actions WHERE token=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, token, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID).Scan(&action.Token, &action.RequestID, &action.Peer.BotID, &action.Peer.ChatID, &action.Peer.TopicID, &action.Peer.UserID, &action.ConversationRevision, &action.GuardDigest, &choices, &action.State, &expiry, &action.AnswerRequestID, &action.GuardVersion, &action.PresentationHash)
	if err != nil {
		return action, err
	}

	if err := json.Unmarshal([]byte(choices), &action.Choices); err != nil {
		return action, err
	}

	action.ExpiresAt, err = database.ParseTime(expiry)

	if err == nil && action.GuardVersion != 0 && action.GuardVersion != 1 {
		return action, ErrActionUnavailable
	}

	return action, err
}

func actionGuard(request OperationRequest) string {
	return contenthash.Sum([]byte(request.ArgumentDigest + ":" + contenthash.Sum(request.PreparedInput)))
}

func validateActionOffer(offer ActionOffer, now time.Time) error {
	if !actionOfferHasView(offer) || len(offer.Text) > 16384 || !offer.ExpiresAt.After(now) || offer.ExpiresAt.After(now.Add(24*time.Hour)) {
		return errReceiverConflict
	}

	if len(offer.Choices) == 0 || len(offer.Choices) > 8 {
		return errReceiverConflict
	}

	for index, choice := range offer.Choices {
		if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,15}$`).MatchString(choice) || slices.Contains(offer.Choices[:index], choice) {
			return errReceiverConflict
		}
	}

	return nil
}

func newActionToken() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(token[:]), nil
}

func readOperationID(ctx context.Context, reader conversationReader, id string) (OperationRequest, error) {
	request, err := scanOperation(reader.QueryRowContext(ctx, `SELECT `+operationColumns+` FROM telegram_requests r WHERE r.id=?`, id))
	if err != nil {
		return request, err
	}

	err = reader.QueryRowContext(ctx, `SELECT u.bot_id,u.update_id,u.payload_digest,u.update_kind,u.chat_id,u.topic_id,u.user_id,u.message_id,u.chat_type,u.payload,u.reply_to_message_id FROM telegram_updates u JOIN telegram_requests r ON r.bot_id=u.bot_id AND r.update_id=u.update_id WHERE r.id=?`, id).Scan(&request.Source.Peer.BotID, &request.Source.UpdateID, &request.Source.Digest, &request.Source.Kind, &request.Source.Peer.ChatID, &request.Source.Peer.TopicID, &request.Source.Peer.UserID, &request.Source.MessageID, &request.Source.ChatType, &request.Source.Text, &request.Source.ReplyToMessageID)

	return request, err
}

func validateActionTarget(op conversationTransaction, request OperationRequest) error {
	user, ok := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)
	if !ok || user.OperatorID != request.Operator || request.Source.Peer != op.input.Update.Peer {
		return workflowstate.ErrConflict
	}

	if request.Target.ConversationID == "" || request.Target.ConversationRevision <= 0 {
		return workflowstate.ErrConflict
	}

	conversation, err := op.conversation(request.Target.ConversationID)
	if err != nil {
		return err
	}

	if err := validateBoundConversation(*conversation, request.Target, request.Target.ProjectBindingDigest); err != nil {
		return err
	}

	if slices.Contains([]string{"completed", "canceled", "failed"}, conversation.Phase) {
		return workflowstate.ErrConflict
	}

	return op.validateTarget(conversation)
}

func actionUnavailable(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, workflowstate.ErrConflict) || errors.Is(err, ErrActionUnavailable)
}

func actionTargetConflict(err error) bool {
	var diagnostic conversationDiagnosticError

	return errors.Is(err, workflowstate.ErrConflict) || errors.As(err, &diagnostic)
}

func actionOfferHasView(offer ActionOffer) bool {
	return offer.Text != "" || offer.Delivery != nil || offer.Message != nil
}
