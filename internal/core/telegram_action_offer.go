package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// OfferTelegramAction atomically checkpoints a prompt, acknowledges its source and
// queues delivery. Reoffering returns its original token without extending expiry.
func (s *TelegramReceiverStore) OfferTelegramAction(ctx context.Context, offer TelegramActionOffer, now time.Time) (TelegramAction, error) {
	var action TelegramAction
	if err := validateTelegramActionOffer(offer, now); err != nil {
		return action, err
	}

	token, err := newTelegramActionToken()
	if err != nil {
		return action, err
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		if _, err := telegramReadResponsePending(ctx, tx, offer.Input.Update); err != nil {
			return err
		}

		request, err := readTelegramOperationID(ctx, tx, offer.RequestID)
		if err != nil {
			return err
		}

		if request.Source.Digest != offer.Input.Update.Digest || request.Source.UpdateID != offer.Input.Update.UpdateID {
			return errTelegramReceiverConflict
		}

		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: offer.Input, now: now}
		if err := validateTelegramActionTarget(op, request); err != nil {
			return err
		}

		action, err = retainTelegramAction(op, request, offer, token)
		if err != nil || action.State != "pending" {
			return err
		}

		return projectTelegramAuthoringWorkPhase(ctx, tx, request, offer.ConversationPhase, database.FormatTime(now))
	})

	return action, err
}

func retainTelegramAction(op telegramConversationTransaction, request TelegramOperationRequest, offer TelegramActionOffer, token string) (TelegramAction, error) {
	var existing string

	err := op.tx.QueryRowContext(op.ctx, `SELECT token FROM telegram_actions WHERE request_id=?`, request.ID).Scan(&existing)
	if err == nil {
		return replayTelegramAction(op, existing, request, offer.Choices)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return TelegramAction{}, err
	}

	if request.State != "ready" || len(request.PreparedInput) == 0 {
		return TelegramAction{}, errTelegramReceiverConflict
	}

	choices, err := json.Marshal(offer.Choices)
	if err != nil {
		return TelegramAction{}, err
	}

	p := request.Source.Peer

	_, err = op.tx.ExecContext(op.ctx, `INSERT INTO telegram_actions(token,request_id,bot_id,chat_id,topic_id,user_id,conversation_revision,guard_digest,allowed_choices,state,expires_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,'pending',?,?,?)`, token, request.ID, p.BotID, p.ChatID, p.TopicID, p.UserID, request.Target.ConversationRevision, telegramActionGuard(request), string(choices), database.FormatTime(offer.ExpiresAt), database.FormatTime(op.now), database.FormatTime(op.now))
	if err != nil {
		return TelegramAction{}, err
	}

	if err := queueTelegramActionPrompt(op, request, offer, token); err != nil {
		return TelegramAction{}, err
	}

	if err := sealTelegramOfferedAction(op, request, offer, token); err != nil {
		return TelegramAction{}, err
	}

	return readTelegramAction(op.ctx, op.tx, token, p)
}

func queueTelegramActionPrompt(op telegramConversationTransaction, request TelegramOperationRequest, offer TelegramActionOffer, token string) error {
	result, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_requests SET state='awaiting_confirmation',updated_at=? WHERE id=? AND state='ready'`, database.FormatTime(op.now), request.ID)
	if err != nil {
		return err
	}

	if err := telegramRequireChanged(result); err != nil {
		return err
	}

	text := offer.Text + "\nRequest: " + request.ID + "\nReply with " + strings.Join(offer.Choices, ", ") + " or use /answer " + token + " <choice>."
	if err := queueTelegramActionView(op, request, offer, token, text); err != nil {
		return err
	}

	if err := bindTelegramOperationResponse(op.ctx, op.tx, request); err != nil {
		return err
	}

	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_outbox SET kind='question',action_token=?,question_binding=(SELECT question_binding FROM telegram_conversations WHERE id=?) WHERE bot_id=? AND event_key=?`, token, request.Target.ConversationID, request.Source.Peer.BotID, "update:"+strconv.FormatInt(request.Source.UpdateID, 10))

	return err
}

func replayTelegramAction(op telegramConversationTransaction, token string, request TelegramOperationRequest, choices []string) (TelegramAction, error) {
	action, err := readTelegramAction(op.ctx, op.tx, token, request.Source.Peer)
	if err == nil && (!telegramActionDigestMatches(action, request) || !slices.Equal(action.Choices, choices)) {
		return action, errTelegramReceiverConflict
	}

	return action, err
}
