package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ClaimAction creates the first answer and claims its prompt in one transaction.
// Applying prompts return the same retained answer for callback/text replay.
func (s *ReceiverStore) ClaimAction(ctx context.Context, input ConversationInput, answer ActionAnswer, now time.Time) (ActionClaim, error) {
	var claim ActionClaim
	if !validActionAnswer(answer) {
		return claim, ErrActionUnavailable
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, input.Update)
		if err != nil {
			return err
		}

		if !pending {
			return ErrActionUnavailable
		}

		action, err := readAction(ctx, tx, answer.Token, input.Update.Peer)
		if actionUnavailable(err) {
			return ErrActionUnavailable
		}

		if err != nil {
			return err
		}

		if !slices.Contains(action.Choices, answer.Choice) {
			return ErrActionUnavailable
		}

		request, err := readOperationID(ctx, tx, action.RequestID)
		if err != nil {
			return err
		}

		claim = ActionClaim{Action: action, Request: request, Choice: answer.Choice, ResponseSource: input.Update}
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}

		return claimAction(op, &claim)
	})

	return claim, err
}

func claimAction(op conversationTransaction, claim *ActionClaim) error {
	user, ok := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)
	if !ok || user.OperatorID != claim.Request.Operator {
		return supersedeAction(op, claim, ActionSuperseded)
	}

	if claim.Action.State == ActionConsumed {
		return readActionClaimAnswer(op, claim)
	}

	if claim.Request.OperationID == "work.control" && claim.Action.AnswerRequestID != "" {
		return claimControlReceipt(op, claim)
	}

	if guidedActionRecovering(claim) {
		return claimGuidedRecovery(op, claim)
	}

	if claim.Action.State == ActionExpired || claim.Action.State == ActionSuperseded {
		return nil
	}

	return claimPendingAction(op, claim)
}

func claimPendingAction(op conversationTransaction, claim *ActionClaim) error {
	state, err := guardActionClaim(op, claim)
	if err != nil {
		return err
	}

	if state != "" {
		return supersedeAction(op, claim, state)
	}

	if claim.Action.State == ActionApplying {
		return readActionClaimAnswer(op, claim)
	}

	return createActionAnswer(op, claim)
}

func claimControlReceipt(op conversationTransaction, claim *ActionClaim) error {
	if err := readActionClaimAnswer(op, claim); err != nil {
		return err
	}

	if claim.Action.State == ActionExpired || claim.Action.State == ActionSuperseded {
		claim.ReceiptOnly = true

		return nil
	}

	state, err := guardActionClaim(op, claim)
	if err != nil || state == "" {
		return err
	}

	claim.ReceiptOnly = true

	return supersedeAction(op, claim, state)
}

func readActionClaimAnswer(op conversationTransaction, claim *ActionClaim) error {
	answer, err := readOperationID(op.ctx, op.tx, claim.Action.AnswerRequestID)
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
		return ErrActionUnavailable
	}

	claim.Answer = answer

	return nil
}

func createActionAnswer(op conversationTransaction, claim *ActionClaim) error {
	arguments, err := json.Marshal(struct {
		Token     string `json:"token"`
		Choice    string `json:"choice"`
		RequestID string `json:"request_id"`
	}{claim.Action.Token, claim.Choice, claim.Request.ID})
	if err != nil {
		return err
	}

	answer, err := normalizeOperation(OperationRequest{Source: op.input.Update, Operator: claim.Request.Operator, Target: claim.Request.Target, OperationID: "actions.answer", Arguments: arguments})
	if err != nil {
		return err
	}

	if err := insertOperation(op.ctx, op.tx, answer, op.now); err != nil {
		return err
	}

	retained, err := readOperationID(op.ctx, op.tx, answer.ID)
	if err != nil {
		return err
	}

	if retained.ArgumentDigest != answer.ArgumentDigest {
		return errReceiverConflict
	}

	if err := beginActionClaim(op, claim, answer); err != nil {
		return err
	}

	answer.State = OperationApplying
	claim.Answer, claim.Action.AnswerRequestID, claim.Action.State = answer, answer.ID, ActionApplying
	claim.Request.State = OperationApplying

	return checkIntakeCapacity(op.ctx, op.tx, answer.Source.Peer.BotID)
}

func beginActionClaim(op conversationTransaction, claim *ActionClaim, answer OperationRequest) error {
	result, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_actions SET state='applying',consumed_by_request=?,updated_at=? WHERE token=? AND state='pending'`, answer.ID, database.FormatTime(op.now), claim.Action.Token)
	if err != nil {
		return err
	}

	if err := requireChanged(result); err != nil {
		return err
	}

	result, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_requests SET state='applying',updated_at=? WHERE id IN (?,?) AND state IN ('prepared','awaiting_confirmation')`, database.FormatTime(op.now), answer.ID, claim.Request.ID)
	if err != nil {
		return err
	}

	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if changed != 2 {
		return errReceiverConflict
	}

	return nil
}

func supersedeAction(op conversationTransaction, claim *ActionClaim, state ActionState) error {
	_, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_actions SET state=?,updated_at=? WHERE token=? AND state IN ('pending','applying')`, state, database.FormatTime(op.now), claim.Action.Token)
	if err != nil {
		return err
	}

	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_requests SET state='blocked',updated_at=? WHERE id IN (?,?) AND state NOT IN ('confirmed','rejected')`, database.FormatTime(op.now), claim.Request.ID, claim.Action.AnswerRequestID)
	claim.Action.State = state

	if err != nil {
		return err
	}

	return finishHarnessConversation(op.ctx, op.tx, claim.Request, OperationBlocked, op.now)
}

func guardActionClaim(op conversationTransaction, claim *ActionClaim) (ActionState, error) {
	if !claim.Action.ExpiresAt.After(op.now) {
		return ActionExpired, nil
	}

	if !actionDigestMatches(claim.Action, claim.Request) || claim.Action.ConversationRevision != claim.Request.Target.ConversationRevision {
		return ActionSuperseded, nil
	}

	if err := validateActionPresentation(op.ctx, op.tx, claim.Action, claim.Request, op.input.Config.Digest()); err != nil {
		if actionUnavailable(err) {
			return ActionSuperseded, nil
		}

		return "", err
	}

	if err := validateActionTarget(op, claim.Request); err != nil {
		if !actionTargetConflict(err) {
			return "", err
		}

		return ActionSuperseded, nil
	}

	return "", nil
}
