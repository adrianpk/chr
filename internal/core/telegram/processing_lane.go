package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ProcessingLane bounds adapter calls independently of native workers.
type ProcessingLane string

// Processing lanes reserve four of the receiver's sixteen application calls.
const (
	ProcessingOrdinary ProcessingLane = "ordinary"
	ProcessingDecision ProcessingLane = "decision"
	ProcessingUrgent   ProcessingLane = "urgent"
)

func controlLane(action string) ProcessingLane {
	switch action {
	case "status", "pause", "cancel", "reconcile":
		return ProcessingUrgent
	case "resume", "approve", "grant", "answer-decision", "answer_decision":
		return ProcessingDecision
	default:
		// A confirmed continuation can execute a worker, so it cannot consume
		// the slots reserved for status, stop and decisions.
		return ProcessingOrdinary
	}
}

func inputLane(ctx context.Context, tx *sql.Tx, update ReceivedUpdate) (ProcessingLane, bool, error) {
	if update.Kind != "message" && update.Kind != "callback_query" {
		return ProcessingOrdinary, false, nil
	}

	if lane, control := ExplicitControlLane(update); control {
		return lane, true, nil
	}

	token, answer, err := inputActionToken(ctx, tx, update)
	if err != nil {
		return ProcessingOrdinary, false, err
	}

	if !answer {
		return unboundReplyLane(ctx, tx, update)
	}

	return actionInputLane(ctx, tx, update, token)
}

// ExplicitControlLane classifies deterministic control syntax. This
// scheduling hint grants no target, actor, execution or mutation authority.
func ExplicitControlLane(update ReceivedUpdate) (ProcessingLane, bool) {
	if update.Kind == "message" {
		switch strings.ToLower(strings.TrimSpace(update.Text)) {
		case "pause use case", "cancel use case", "show current step", "show use case":
			return ProcessingUrgent, true
		case "resume use case", "show controls":
			return ProcessingDecision, true
		}
	}

	words := strings.Fields(update.Text)
	if update.Kind != "message" || len(words) == 0 {
		return ProcessingOrdinary, false
	}

	switch words[0] {
	case "/status", "/pause", "/cancel", "/reconcile":
		return ProcessingUrgent, true
	case "/resume", "/approve", "/grant", "/answer-decision":
		return ProcessingDecision, true
	case "/continue":
		return ProcessingOrdinary, true
	default:
		return ProcessingOrdinary, false
	}
}

func actionInputLane(ctx context.Context, tx *sql.Tx, update ReceivedUpdate, token string) (ProcessingLane, bool, error) {
	var operation, arguments string

	p := update.Peer

	err := tx.QueryRowContext(ctx, `SELECT r.operation_id,r.arguments FROM telegram_actions a JOIN telegram_requests r ON r.id=a.request_id WHERE a.token=? AND a.bot_id=? AND a.chat_id=? AND a.topic_id=? AND a.user_id=?`, token, p.BotID, p.ChatID, p.TopicID, p.UserID).Scan(&operation, &arguments)
	if errors.Is(err, sql.ErrNoRows) {
		return ProcessingDecision, true, nil
	}

	if err != nil || operation != "work.control" && operation != "guided.step" {
		return ProcessingOrdinary, false, err
	}

	lane, err := ActionChoiceProcessingLane(OperationRequest{OperationID: operation, Arguments: json.RawMessage(arguments)}, inputActionChoice(update))

	return lane, true, err
}

// ActionProcessingLane derives execution class from the immutable
// retained operation, including a reply whose binding appeared after intake.
func ActionProcessingLane(request OperationRequest) (ProcessingLane, error) {
	if request.OperationID != "work.control" {
		return ProcessingOrdinary, nil
	}

	var args struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(request.Arguments, &args); err != nil {
		return "", err
	}

	return controlLane(args.Action), nil
}

func unboundReplyLane(ctx context.Context, tx *sql.Tx, update ReceivedUpdate) (ProcessingLane, bool, error) {
	if update.ReplyToMessageID <= 0 {
		return ProcessingOrdinary, false, nil
	}

	state, err := resolveReplyBinding(ctx, tx, update)
	if state == replyKnown {
		return ProcessingOrdinary, false, err
	}

	// Unknown replies only wait or produce diagnostics. They cannot execute a
	// model and receive no reserved intake allowance before their binding exists.
	return ProcessingDecision, false, err
}

func inputActionToken(ctx context.Context, tx *sql.Tx, update ReceivedUpdate) (string, bool, error) {
	text := strings.TrimSpace(update.Text)
	if update.Kind == "callback_query" && strings.HasPrefix(text, "tga:") {
		parts := strings.Split(text, ":")
		if len(parts) == 3 {
			return parts[1], true, nil
		}

		return "", true, nil
	}

	words := strings.Fields(text)
	if len(words) > 0 && words[0] == ActionAnswerCommand {
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
func (s *ReceiverStore) PendingApplicationInputs(ctx context.Context, botID int64, lane ProcessingLane, now time.Time) ([]ReceivedUpdate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT u.update_id,u.payload_digest,u.update_kind,u.chat_id,u.topic_id,u.user_id,u.message_id,u.chat_type,u.payload,u.reply_to_message_id
FROM telegram_updates u WHERE u.bot_id=? AND u.state='pending' AND u.processing_lane=?
AND (u.interpretation IS NOT NULL OR u.interpretation_attempt>=5 OR u.interpretation_retry_at IS NULL OR u.interpretation_retry_at<=?)
AND NOT EXISTS(SELECT 1 FROM telegram_requests r WHERE r.bot_id=u.bot_id AND r.update_id=u.update_id AND r.state='applying' AND r.model_result IS NULL AND r.model_attempt<5 AND r.model_retry_at>?)
AND (u.processing_lane!='ordinary' OR NOT EXISTS(SELECT 1 FROM telegram_updates earlier WHERE earlier.bot_id=u.bot_id AND earlier.chat_id=u.chat_id AND earlier.topic_id=u.topic_id AND earlier.user_id=u.user_id AND earlier.state='pending' AND earlier.processing_lane='ordinary' AND earlier.update_id<u.update_id))
ORDER BY CASE WHEN u.processing_lane='decision' THEN u.control_input ELSE 0 END DESC,u.update_id LIMIT 50`, botID, lane, database.FormatTime(now), database.FormatTime(now))
	if err != nil {
		return nil, err
	}

	return readPendingRows(rows, botID)
}

func readPendingRows(rows *sql.Rows, botID int64) ([]ReceivedUpdate, error) {
	defer func() {
		//nolint:errcheck // rows.Err is checked after the bounded iteration.
		_ = rows.Close()
	}()

	var result []ReceivedUpdate

	for rows.Next() {
		update := ReceivedUpdate{Peer: PeerIdentity{BotID: botID}}
		if err := rows.Scan(&update.UpdateID, &update.Digest, &update.Kind, &update.Peer.ChatID, &update.Peer.TopicID, &update.Peer.UserID, &update.MessageID, &update.ChatType, &update.Text, &update.ReplyToMessageID); err != nil {
			return nil, err
		}

		result = append(result, update)
	}

	return result, rows.Err()
}

// ActionChoiceProcessingLane is a scheduling hint for an already sealed
// choice. The action owner still validates every peer, choice and mutation guard.
func ActionChoiceProcessingLane(request OperationRequest, choice string) (ProcessingLane, error) {
	if request.OperationID == "guided.step" {
		return controlLane(choice), nil
	}

	return ActionProcessingLane(request)
}

func inputActionChoice(update ReceivedUpdate) string {
	if update.Kind == "callback_query" {
		parts := strings.Split(update.Text, ":")
		if len(parts) == 3 {
			return parts[2]
		}
	}

	words := strings.Fields(update.Text)
	if len(words) == 3 && words[0] == ActionAnswerCommand {
		return words[2]
	}

	return strings.TrimSpace(update.Text)
}
