package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// TelegramProcessingLane bounds adapter calls independently of native workers.
type TelegramProcessingLane string

// Processing lanes reserve four of the receiver's sixteen application calls.
const (
	TelegramProcessingOrdinary TelegramProcessingLane = "ordinary"
	TelegramProcessingDecision TelegramProcessingLane = "decision"
	TelegramProcessingUrgent   TelegramProcessingLane = "urgent"
)

func telegramControlLane(action string) TelegramProcessingLane {
	switch action {
	case "status", "pause", "cancel", "reconcile":
		return TelegramProcessingUrgent
	case "resume", "approve", "grant", "answer-decision", "answer_decision":
		return TelegramProcessingDecision
	default:
		// A confirmed continuation can execute a worker, so it cannot consume
		// the slots reserved for status, stop and decisions.
		return TelegramProcessingOrdinary
	}
}

func telegramInputLane(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate) (TelegramProcessingLane, bool, error) {
	if update.Kind != "message" && update.Kind != "callback_query" {
		return TelegramProcessingOrdinary, false, nil
	}

	if lane, control := TelegramExplicitControlLane(update); control {
		return lane, true, nil
	}

	token, answer, err := telegramInputActionToken(ctx, tx, update)
	if err != nil {
		return TelegramProcessingOrdinary, false, err
	}

	if !answer {
		return telegramUnboundReplyLane(ctx, tx, update)
	}

	return telegramActionInputLane(ctx, tx, update.Peer, token)
}

// TelegramExplicitControlLane classifies deterministic control syntax. This
// scheduling hint grants no target, actor, execution or mutation authority.
func TelegramExplicitControlLane(update TelegramReceivedUpdate) (TelegramProcessingLane, bool) {
	words := strings.Fields(update.Text)
	if update.Kind != "message" || len(words) == 0 {
		return TelegramProcessingOrdinary, false
	}

	switch words[0] {
	case "/status", "/pause", "/cancel", "/reconcile":
		return TelegramProcessingUrgent, true
	case "/resume", "/approve", "/grant", "/answer-decision":
		return TelegramProcessingDecision, true
	case "/continue":
		return TelegramProcessingOrdinary, true
	default:
		return TelegramProcessingOrdinary, false
	}
}

func telegramActionInputLane(ctx context.Context, tx *sql.Tx, p TelegramPeerIdentity, token string) (TelegramProcessingLane, bool, error) {
	var operation, arguments string

	err := tx.QueryRowContext(ctx, `SELECT r.operation_id,r.arguments FROM telegram_actions a JOIN telegram_requests r ON r.id=a.request_id WHERE a.token=? AND a.bot_id=? AND a.chat_id=? AND a.topic_id=? AND a.user_id=?`, token, p.BotID, p.ChatID, p.TopicID, p.UserID).Scan(&operation, &arguments)
	if errors.Is(err, sql.ErrNoRows) {
		return TelegramProcessingDecision, true, nil
	}

	if err != nil || operation != "work.control" {
		return TelegramProcessingOrdinary, false, err
	}

	lane, err := TelegramActionProcessingLane(TelegramOperationRequest{OperationID: operation, Arguments: json.RawMessage(arguments)})

	return lane, true, err
}

// TelegramActionProcessingLane derives execution class from the immutable
// retained operation, including a reply whose binding appeared after intake.
func TelegramActionProcessingLane(request TelegramOperationRequest) (TelegramProcessingLane, error) {
	if request.OperationID != "work.control" {
		return TelegramProcessingOrdinary, nil
	}

	var args struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(request.Arguments, &args); err != nil {
		return "", err
	}

	return telegramControlLane(args.Action), nil
}

func telegramUnboundReplyLane(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate) (TelegramProcessingLane, bool, error) {
	if update.ReplyToMessageID <= 0 {
		return TelegramProcessingOrdinary, false, nil
	}

	state, err := resolveTelegramReplyBinding(ctx, tx, update)
	if state == telegramReplyKnown {
		return TelegramProcessingOrdinary, false, err
	}

	// Unknown replies only wait or produce diagnostics. They cannot execute a
	// model and receive no reserved intake allowance before their binding exists.
	return TelegramProcessingDecision, false, err
}

func telegramInputActionToken(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate) (string, bool, error) {
	text := strings.TrimSpace(update.Text)
	if update.Kind == "callback_query" && strings.HasPrefix(text, "tga:") {
		parts := strings.Split(text, ":")
		if len(parts) == 3 {
			return parts[1], true, nil
		}

		return "", true, nil
	}

	words := strings.Fields(text)
	if len(words) > 0 && words[0] == "/answer" {
		if len(words) == 3 {
			return words[1], true, nil
		}

		return "", true, nil
	}

	if update.ReplyToMessageID <= 0 {
		return "", false, nil
	}

	var token string

	p := update.Peer

	err := tx.QueryRowContext(ctx, `SELECT action_token FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=? AND purpose='action'`, p.BotID, p.ChatID, p.TopicID, p.UserID, update.ReplyToMessageID).Scan(&token)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}

	return token, err == nil, err
}

// PendingApplicationInputs returns a bounded lane without overtaking an earlier
// ordinary input for the same peer. Controls may bypass a blocked model call.
func (s *TelegramReceiverStore) PendingApplicationInputs(ctx context.Context, botID int64, lane TelegramProcessingLane, now time.Time) ([]TelegramReceivedUpdate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT u.update_id,u.payload_digest,u.update_kind,u.chat_id,u.topic_id,u.user_id,u.message_id,u.chat_type,u.payload,u.reply_to_message_id
FROM telegram_updates u WHERE u.bot_id=? AND u.state='pending' AND u.processing_lane=?
AND (u.interpretation IS NOT NULL OR u.interpretation_attempt>=5 OR u.interpretation_retry_at IS NULL OR u.interpretation_retry_at<=?)
AND NOT EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id AND r.state='applying' AND r.model_result IS NULL AND r.model_attempt<5 AND r.model_retry_at>?)
AND (u.processing_lane!='ordinary' OR NOT EXISTS(SELECT 1 FROM telegram_updates earlier WHERE earlier.bot_id=u.bot_id AND earlier.chat_id=u.chat_id AND earlier.topic_id=u.topic_id AND earlier.user_id=u.user_id AND earlier.state='pending' AND earlier.processing_lane='ordinary' AND earlier.update_id<u.update_id))
ORDER BY CASE WHEN u.processing_lane='decision' THEN u.control_input ELSE 0 END DESC,u.update_id LIMIT 50`, botID, lane, formatProjectTime(now), formatProjectTime(now))
	if err != nil {
		return nil, err
	}

	return readTelegramPendingRows(rows, botID)
}

func readTelegramPendingRows(rows *sql.Rows, botID int64) ([]TelegramReceivedUpdate, error) {
	defer func() {
		//nolint:errcheck // rows.Err is checked after the bounded iteration.
		_ = rows.Close()
	}()

	var result []TelegramReceivedUpdate

	for rows.Next() {
		update := TelegramReceivedUpdate{Peer: TelegramPeerIdentity{BotID: botID}}
		if err := rows.Scan(&update.UpdateID, &update.Digest, &update.Kind, &update.Peer.ChatID, &update.Peer.TopicID, &update.Peer.UserID, &update.MessageID, &update.ChatType, &update.Text, &update.ReplyToMessageID); err != nil {
			return nil, err
		}

		result = append(result, update)
	}

	return result, rows.Err()
}
