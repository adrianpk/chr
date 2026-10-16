package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"time"
)

// ClaimTelegramAction creates the first answer and claims its prompt in one transaction.
// Applying prompts return the same retained answer for callback/text replay.
func (s *TelegramReceiverStore) ClaimTelegramAction(ctx context.Context, input TelegramConversationInput, answer TelegramActionAnswer, now time.Time) (TelegramActionClaim, error) {
	var claim TelegramActionClaim
	if !validTelegramActionAnswer(answer) {
		return claim, ErrTelegramActionUnavailable
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, input.Update)
		if err != nil {
			return err
		}

		if !pending {
			return ErrTelegramActionUnavailable
		}

		action, err := readTelegramAction(ctx, tx, answer.Token, input.Update.Peer)
		if telegramActionUnavailable(err) {
			return ErrTelegramActionUnavailable
		}

		if err != nil {
			return err
		}

		if !slices.Contains(action.Choices, answer.Choice) {
			return ErrTelegramActionUnavailable
		}

		request, err := readTelegramOperationID(ctx, tx, action.RequestID)
		if err != nil {
			return err
		}

		claim = TelegramActionClaim{Action: action, Request: request, Choice: answer.Choice, ResponseSource: input.Update}
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}

		return claimTelegramAction(op, &claim)
	})

	return claim, err
}

func claimTelegramAction(op telegramConversationTransaction, claim *TelegramActionClaim) error {
	user, ok := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)
	if !ok || user.OperatorID != claim.Request.Operator {
		return supersedeTelegramAction(op, claim, "superseded")
	}

	if claim.Action.State == "consumed" {
		return readTelegramActionClaimAnswer(op, claim)
	}

	if claim.Request.OperationID == "work.control" && claim.Action.AnswerRequestID != "" {
		return claimTelegramControlReceipt(op, claim)
	}

	if claim.Action.State == "expired" || claim.Action.State == "superseded" {
		return nil
	}

	return claimPendingTelegramAction(op, claim)
}

func claimPendingTelegramAction(op telegramConversationTransaction, claim *TelegramActionClaim) error {
	state, err := guardTelegramActionClaim(op, claim)
	if err != nil {
		return err
	}

	if state != "" {
		return supersedeTelegramAction(op, claim, state)
	}

	if claim.Action.State == "applying" {
		return readTelegramActionClaimAnswer(op, claim)
	}

	return createTelegramActionAnswer(op, claim)
}

func claimTelegramControlReceipt(op telegramConversationTransaction, claim *TelegramActionClaim) error {
	if err := readTelegramActionClaimAnswer(op, claim); err != nil {
		return err
	}

	if claim.Action.State == "expired" || claim.Action.State == "superseded" {
		claim.ReceiptOnly = true

		return nil
	}

	state, err := guardTelegramActionClaim(op, claim)
	if err != nil || state == "" {
		return err
	}

	claim.ReceiptOnly = true

	return supersedeTelegramAction(op, claim, state)
}

func readTelegramActionClaimAnswer(op telegramConversationTransaction, claim *TelegramActionClaim) error {
	answer, err := readTelegramOperationID(op.ctx, op.tx, claim.Action.AnswerRequestID)
	if err != nil {
		return err
	}

	var binding struct {
		Token     string `json:"token"`
		Choice    string `json:"choice"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(answer.Arguments, &binding); err != nil {
		return err
	}

	if binding.Token != claim.Action.Token || binding.Choice != claim.Choice || binding.RequestID != claim.Request.ID {
		return ErrTelegramActionUnavailable
	}

	claim.Answer = answer

	return nil
}

func createTelegramActionAnswer(op telegramConversationTransaction, claim *TelegramActionClaim) error {
	arguments, err := json.Marshal(struct {
		Token     string `json:"token"`
		Choice    string `json:"choice"`
		RequestID string `json:"request_id"`
	}{claim.Action.Token, claim.Choice, claim.Request.ID})
	if err != nil {
		return err
	}

	answer, err := normalizeTelegramOperation(TelegramOperationRequest{Source: op.input.Update, Operator: claim.Request.Operator, Target: claim.Request.Target, OperationID: "actions.answer", Arguments: arguments})
	if err != nil {
		return err
	}

	if err := insertTelegramOperation(op.ctx, op.tx, answer, op.now); err != nil {
		return err
	}

	retained, err := readTelegramOperationID(op.ctx, op.tx, answer.ID)
	if err != nil {
		return err
	}

	if retained.ArgumentDigest != answer.ArgumentDigest {
		return errTelegramReceiverConflict
	}

	if err := beginTelegramActionClaim(op, claim, answer); err != nil {
		return err
	}

	answer.State = "applying"
	claim.Answer, claim.Action.AnswerRequestID, claim.Action.State = answer, answer.ID, "applying"
	claim.Request.State = "applying"

	return checkTelegramIntakeCapacity(op.ctx, op.tx, answer.Source.Peer.BotID)
}

func beginTelegramActionClaim(op telegramConversationTransaction, claim *TelegramActionClaim, answer TelegramOperationRequest) error {
	result, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_actions SET state='applying',consumed_by_request=?,updated_at=? WHERE token=? AND state='pending'`, answer.ID, formatProjectTime(op.now), claim.Action.Token)
	if err != nil {
		return err
	}

	if err := telegramRequireChanged(result); err != nil {
		return err
	}

	result, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_requests SET state='applying',updated_at=? WHERE id IN (?,?) AND state IN ('prepared','awaiting_confirmation')`, formatProjectTime(op.now), answer.ID, claim.Request.ID)
	if err != nil {
		return err
	}

	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if changed != 2 {
		return errTelegramReceiverConflict
	}

	return nil
}

func supersedeTelegramAction(op telegramConversationTransaction, claim *TelegramActionClaim, state string) error {
	_, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_actions SET state=?,updated_at=? WHERE token=? AND state IN ('pending','applying')`, state, formatProjectTime(op.now), claim.Action.Token)
	if err != nil {
		return err
	}

	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_requests SET state='blocked',updated_at=? WHERE id IN (?,?) AND state NOT IN ('confirmed','rejected')`, formatProjectTime(op.now), claim.Request.ID, claim.Action.AnswerRequestID)
	claim.Action.State = state

	return err
}

func guardTelegramActionClaim(op telegramConversationTransaction, claim *TelegramActionClaim) (string, error) {
	if !claim.Action.ExpiresAt.After(op.now) {
		return "expired", nil
	}

	if claim.Action.GuardDigest != telegramActionGuard(claim.Request) || claim.Action.ConversationRevision != claim.Request.Target.ConversationRevision {
		return "superseded", nil
	}

	if err := validateTelegramActionTarget(op, claim.Request); err != nil {
		if !telegramActionTargetConflict(err) {
			return "", err
		}

		return "superseded", nil
	}

	return "", nil
}
