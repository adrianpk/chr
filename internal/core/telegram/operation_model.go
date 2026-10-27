package telegram

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// OperationModel retains a provider result independently of semantic
// interpretation and subsequent source acknowledgement or message delivery.
type OperationModel struct {
	Attempts  int
	RetryAt   time.Time
	ErrorCode string
	Response  json.RawMessage
}

// ResolveOperationModel reserves at most five attempts for an immutable
// prepared request. No database transaction spans the provider call. OS exclusion
// prevents a second owner stealing the call after a wall-clock jump.
func (s *ReceiverStore) ResolveOperationModel(ctx context.Context, env StateDirectory, request OperationRequest, now time.Time, invoke func(context.Context) (json.RawMessage, error)) (result OperationModel, err error) {
	if err := validateOperationModel(request, invoke); err != nil {
		return result, err
	}

	stripe := sha256.Sum256([]byte(request.ID))

	lock, err := acquireModelClaim(env, request.Source.Peer.BotID, fmt.Sprintf("operation-model-%d-%d.lock", request.Source.Peer.BotID, stripe[0]%16))
	if err != nil {
		return result, err
	}

	defer func() { err = errors.Join(err, lock.Close()) }()

	reservation, err := s.reserveOperationModel(ctx, request, now)

	result, ready := reservation.Model, reservation.Call
	if err != nil || !ready {
		return result, err
	}

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	response, callErr := invoke(callCtx)
	if callCtx.Err() != nil {
		callErr = callCtx.Err()
	}

	if callErr == nil && (len(response) > 65536 || !json.Valid(response)) {
		callErr = modelcore.ErrOutput
	}

	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finishCancel()

	if callErr != nil {
		result.ErrorCode = interpretationFailure(callErr)
		result.RetryAt = time.Now().UTC().Add(time.Second * time.Duration(1<<(result.Attempts-1)))
	} else {
		result.Response, result.RetryAt, result.ErrorCode = response, time.Time{}, ""
	}

	err = s.finishOperationModel(finishCtx, request, result)

	return result, err
}

func (s *ReceiverStore) reserveOperationModel(ctx context.Context, request OperationRequest, now time.Time) (operationModelReservation, error) {
	var result OperationModel

	ready := false
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, request.Source)
		if err != nil || !pending {
			return errors.Join(err, workflowstate.ErrConflict)
		}

		var retry, response string

		u := request.Source

		err = tx.QueryRowContext(ctx, `SELECT r.model_attempt,COALESCE(r.model_retry_at,''),r.model_error_code,COALESCE(r.model_result,'') FROM telegram_requests r JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id WHERE r.id=? AND r.argument_digest=? AND r.operator_id=? AND r.prepared_input=? AND r.domain_guard=? AND r.state='applying' AND u.payload_digest=? AND u.chat_id=? AND u.topic_id=? AND u.user_id=? AND u.payload=? AND u.update_kind=?`, request.ID, request.ArgumentDigest, request.Operator, string(request.PreparedInput), contenthash.Sum(request.PreparedInput), u.Digest, u.Peer.ChatID, u.Peer.TopicID, u.Peer.UserID, u.Text, u.Kind).Scan(&result.Attempts, &retry, &result.ErrorCode, &response)
		if err != nil {
			return err
		}

		result.Response = json.RawMessage(response)
		if retry != "" {
			result.RetryAt, err = time.Parse(time.RFC3339Nano, retry)
			if err != nil {
				return err
			}
		}

		if len(result.Response) > 0 || result.Attempts >= 5 || now.Before(result.RetryAt) {
			return nil
		}

		result.Attempts++
		result.RetryAt, result.ErrorCode = now.Add(61*time.Second), "interrupted"
		_, err = tx.ExecContext(ctx, `UPDATE telegram_requests SET model_attempt=?,model_retry_at=?,model_error_code='interrupted' WHERE id=? AND argument_digest=?`, result.Attempts, database.FormatTime(result.RetryAt), request.ID, request.ArgumentDigest)
		ready = err == nil

		return err
	})

	return operationModelReservation{Model: result, Call: ready}, err
}

func (s *ReceiverStore) finishOperationModel(ctx context.Context, request OperationRequest, model OperationModel) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		retry := ""
		if !model.RetryAt.IsZero() {
			retry = database.FormatTime(model.RetryAt)
		}

		changed, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET model_result=NULLIF(?,''),model_retry_at=NULLIF(?,''),model_error_code=? WHERE id=? AND argument_digest=? AND model_attempt=? AND state='applying' AND model_result IS NULL`, string(model.Response), retry, model.ErrorCode, request.ID, request.ArgumentDigest, model.Attempts)
		if err != nil {
			return err
		}

		if err := requireChanged(changed); err != nil {
			return err
		}

		return checkIntakeCapacity(ctx, tx, request.Source.Peer.BotID)
	})
}

func validateOperationModel(request OperationRequest, invoke func(context.Context) (json.RawMessage, error)) error {
	if invoke == nil || request.ID != OperationID(request.Source, request.Ordinal) || len(request.PreparedInput) == 0 {
		return workflowstate.ErrConflict
	}

	normalized, err := normalizeOperation(request)
	if err != nil || normalized.ArgumentDigest != request.ArgumentDigest {
		return errors.Join(err, workflowstate.ErrConflict)
	}

	return nil
}
