package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// ErrTelegramRequestCapacity rejects a new reservation before any owner effect.
// The host can acknowledge its retained input with a durable capacity diagnostic.
var ErrTelegramRequestCapacity = errors.New("telegram request capacity exhausted")

// Each of the 32 reserved control inputs can retain a prompt and its answer.
// Their 64 request records remain within the existing 1,024/8,192 ceilings.
func checkTelegramRequestAllowance(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest) error {
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_requests WHERE id=?`, request.ID).Scan(&existing); err != nil || existing != 0 {
		return err
	}

	var total, pending int

	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(state NOT IN ('confirmed','rejected')),0) FROM telegram_requests WHERE bot_id=?`, request.Source.Peer.BotID).Scan(&total, &pending); err != nil {
		return err
	}

	if total >= 8192 || pending >= 1024 {
		return ErrTelegramRequestCapacity
	}

	if total < 8128 && pending < 960 {
		return nil
	}

	control, err := telegramRequestControlAllowance(ctx, tx, request)
	if err != nil || control {
		return err
	}

	return ErrTelegramRequestCapacity
}

func telegramRequestControlAllowance(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest) (bool, error) {
	if request.OperationID == "work.control" {
		return true, nil
	}

	if request.OperationID != "actions.answer" {
		return false, nil
	}

	var answer struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(request.Arguments, &answer); err != nil {
		return false, err
	}

	p := request.Source.Peer

	var count int

	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_requests r JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id WHERE r.id=? AND r.operation_id='work.control' AND r.operator_id=? AND r.project_id=? AND r.namespace=? AND u.bot_id=? AND u.chat_id=? AND u.topic_id=? AND u.user_id=?`, answer.RequestID, request.Operator, request.Target.ProjectID, request.Target.Namespace, p.BotID, p.ChatID, p.TopicID, p.UserID).Scan(&count)

	return count == 1, err
}
