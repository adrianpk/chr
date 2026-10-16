package core

import (
	"context"
	"database/sql"
	"errors"
)

func telegramOperationCapacity(ctx context.Context, tx *sql.Tx, botID int64) (int64, error) {
	var total, pending, bytes int64

	err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(state NOT IN ('confirmed','rejected')),0),COALESCE(SUM(length(CAST(arguments AS BLOB))+COALESCE(length(CAST(prepared_input AS BLOB)),0)+COALESCE(length(CAST(result AS BLOB)),0)+COALESCE(length(CAST(model_result AS BLOB)),0)),0) FROM telegram_requests WHERE bot_id=?`, botID).Scan(&total, &pending, &bytes)
	if err != nil {
		return 0, err
	}

	if total > 8192 || pending > 1024 {
		return 0, errors.New("telegram request capacity exhausted")
	}

	var conversations, actions int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(length(CAST(capture AS BLOB))),0) FROM telegram_conversations WHERE bot_id=?`, botID).Scan(&conversations); err != nil {
		return 0, err
	}

	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(length(CAST(allowed_choices AS BLOB))),0) FROM telegram_actions WHERE bot_id=?`, botID).Scan(&actions); err != nil {
		return 0, err
	}

	return bytes + conversations + actions, nil
}
