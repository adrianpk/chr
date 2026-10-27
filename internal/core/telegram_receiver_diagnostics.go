package core

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TelegramDeliveryDiagnostic exposes bounded metadata, never message bodies.
type TelegramDeliveryDiagnostic struct {
	ID             string
	State          string
	Attempts       int
	Cycle          int64
	TotalAttempts  int64
	RetryAt        time.Time
	ErrorCode      string
	RetryActor     string
	RetryStartedAt time.Time
}

// TelegramReceiverDiagnostics distinguishes retained transport and model waits
// from domain completion. Delivery detail is limited to twenty unresolved rows.
type TelegramReceiverDiagnostics struct {
	Receiver            TelegramReceiverState
	ActiveConversations int
	PendingUpdates      int
	PendingDeliveries   int
	ExhaustedModels     int
	Deliveries          []TelegramDeliveryDiagnostic
}

// InspectTelegramReceiverDiagnostics reads without migration or network IO.
func InspectTelegramReceiverDiagnostics(ctx context.Context, env UserEnvironment, botID int64) (result TelegramReceiverDiagnostics, err error) {
	db, err := database.OpenReadOnly(ctx, env)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	result.Receiver, err = (&TelegramReceiverStore{db: db, env: env}).Receiver(ctx, botID)
	if err != nil {
		return result, err
	}

	err = db.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM telegram_conversations WHERE bot_id=? AND phase NOT IN ('completed','canceled','failed')),
 (SELECT count(*) FROM telegram_updates WHERE bot_id=? AND state='pending'),
 (SELECT count(*) FROM telegram_outbox WHERE bot_id=? AND state!='delivered'),
 (SELECT count(*) FROM telegram_updates WHERE bot_id=? AND state='pending' AND interpretation_attempt=5 AND interpretation IS NULL)+(SELECT count(*) FROM telegram_requests WHERE bot_id=? AND state NOT IN ('confirmed','rejected') AND model_attempt=5 AND model_result IS NULL)`, botID, botID, botID, botID, botID).Scan(&result.ActiveConversations, &result.PendingUpdates, &result.PendingDeliveries, &result.ExhaustedModels)
	if err != nil {
		return result, err
	}

	result.Deliveries, err = readTelegramDeliveryDiagnostics(ctx, db, botID)

	return result, err
}

func readTelegramDeliveryDiagnostics(ctx context.Context, db *sql.DB, botID int64) (result []TelegramDeliveryDiagnostic, err error) {
	rows, err := db.QueryContext(ctx, `SELECT id,state,attempts,retry_cycle,total_attempts,COALESCE(retry_at,''),last_error_code,retry_actor,COALESCE(retry_started_at,'') FROM telegram_outbox WHERE bot_id=? AND state!='delivered' ORDER BY CASE WHEN state='blocked' OR attempts=5 THEN 0 ELSE 1 END,created_at,id LIMIT 20`, botID)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	for rows.Next() {
		var (
			item             TelegramDeliveryDiagnostic
			retryAt, started string
		)
		if err := rows.Scan(&item.ID, &item.State, &item.Attempts, &item.Cycle, &item.TotalAttempts, &retryAt, &item.ErrorCode, &item.RetryActor, &started); err != nil {
			return result, err
		}

		if retryAt != "" {
			item.RetryAt, err = database.ParseTime(retryAt)
			if err != nil {
				return result, err
			}
		}

		if started != "" {
			item.RetryStartedAt, err = database.ParseTime(started)
			if err != nil {
				return result, err
			}
		}

		result = append(result, item)
	}

	return result, rows.Err()
}
