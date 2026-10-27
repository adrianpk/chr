package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ActionResult contains confirmed owner facts or an explicit blocked outcome.
type ActionResult struct {
	Delivery          *tgmessage.DocumentResponse
	Guided            *GuidedProgress
	Message           *tgmessage.Blocks
	Invocation        *InvocationBinding
	InvocationPhase   string
	State             string
	ReceiptRef        string
	Facts             json.RawMessage
	Text              string
	ConfigDigest      string
	ConversationPhase string
}

// FinishAction commits both requests, action consumption, input acknowledgement
// and result delivery atomically. A blocked result supersedes rather than consumes.
func (s *ReceiverStore) FinishAction(ctx context.Context, claim ActionClaim, result ActionResult, now time.Time) error {
	if result.State != "confirmed" && result.State != "blocked" {
		return errReceiverConflict
	}

	if !validOperationFinish(result.State, result.ReceiptRef, result.Facts) {
		return errReceiverConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		retained, err := retainedActionClaim(ctx, tx, claim)
		if err != nil {
			return err
		}

		sourceResolution, err := actionResponseSource(ctx, tx, retained, result)

		source, pending := sourceResolution.Update, sourceResolution.Pending
		if err != nil || !pending {
			return err
		}

		if retained.Action.State != "applying" && !confirmedControlReceipt(retained, result) {
			return errReceiverConflict
		}

		retained.ResponseSource = source

		return finishActionResult(ctx, tx, retained, result, now)
	})
}

func retainedActionClaim(ctx context.Context, tx *sql.Tx, claim ActionClaim) (ActionClaim, error) {
	action, err := readAction(ctx, tx, claim.Action.Token, claim.Action.Peer)
	if err != nil {
		return claim, err
	}

	request, err := readOperationID(ctx, tx, action.RequestID)
	if err != nil {
		return claim, err
	}

	if action.AnswerRequestID != claim.Answer.ID || request.ID != claim.Request.ID || request.ArgumentDigest != claim.Request.ArgumentDigest || !actionDigestMatches(action, request) {
		return claim, errReceiverConflict
	}

	retained := ActionClaim{Action: action, Request: request, Choice: claim.Choice, ResponseSource: claim.ResponseSource}
	err = readActionClaimAnswer(conversationTransaction{ctx: ctx, tx: tx}, &retained)

	return retained, err
}

func finishActionRequests(ctx context.Context, tx *sql.Tx, claim ActionClaim, result ActionResult, now time.Time) error {
	recovery := confirmedControlReceipt(claim, result)
	for _, request := range []OperationRequest{claim.Request, claim.Answer} {
		changed, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET state=?,domain_receipt_ref=NULLIF(?,''),result=?,updated_at=? WHERE id=? AND argument_digest=? AND (state='applying' OR (? AND state='blocked'))`, result.State, result.ReceiptRef, string(result.Facts), database.FormatTime(now), request.ID, request.ArgumentDigest, recovery)
		if err != nil {
			return err
		}

		if err := requireChanged(changed); err != nil {
			return err
		}
	}

	return nil
}

func finishActionState(ctx context.Context, tx *sql.Tx, claim ActionClaim, result ActionResult, now time.Time) error {
	actionState := "consumed"
	if result.State != "confirmed" {
		actionState = "superseded"
	}

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_actions SET state=?,updated_at=? WHERE token=? AND (state='applying' OR (? AND state IN ('expired','superseded'))) AND consumed_by_request=?`, actionState, database.FormatTime(now), claim.Action.Token, confirmedControlReceipt(claim, result), claim.Answer.ID)
	if err != nil {
		return err
	}

	if err := requireChanged(changed); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_replies SET invalidated_at=? WHERE action_token=? AND invalidated_at IS NULL`, database.FormatTime(now), claim.Action.Token)

	return err
}

func finishActionResult(ctx context.Context, tx *sql.Tx, retained ActionClaim, result ActionResult, now time.Time) error {
	if err := finishHarnessConversation(ctx, tx, retained.Request, result.State, now); err != nil {
		return err
	}

	if err := projectWorkBinding(ctx, tx, retained.Request, result, database.FormatTime(now)); err != nil {
		return err
	}

	if result.ConversationPhase != "" && result.State != "confirmed" {
		return workflowstate.ErrConflict
	}

	if err := projectAuthoringWorkPhase(ctx, tx, retained.Request, result.ConversationPhase, database.FormatTime(now)); err != nil {
		return err
	}

	if err := finishActionRequests(ctx, tx, retained, result, now); err != nil {
		return err
	}

	if err := finishActionState(ctx, tx, retained, result, now); err != nil {
		return err
	}

	return finishActionResponse(ctx, tx, retained, result, now)
}

func finishActionResponse(ctx context.Context, tx *sql.Tx, retained ActionClaim, result ActionResult, now time.Time) error {
	var err error

	retained, err = projectGuidedAction(ctx, tx, retained, result, now)
	if err != nil {
		return err
	}

	if err := queueActionResultResponse(ctx, tx, retained, result, now); err != nil {
		return err
	}

	response := retained.Request

	response.Source = retained.ResponseSource
	if err := bindOperationResponse(ctx, tx, response); err != nil {
		return err
	}

	if result.Guided != nil {
		_, err := tx.ExecContext(ctx, `UPDATE telegram_outbox SET question_binding=? WHERE request_id=? AND action_token IS NULL`, result.Guided.Binding, retained.Request.ID)

		return err
	}

	return nil
}

func actionResponseSource(ctx context.Context, tx *sql.Tx, claim ActionClaim, result ActionResult) (responseSource, error) {
	original := claim.Answer.Source

	pending, err := readResponsePending(ctx, tx, original)
	if err != nil || pending {
		return responseSource{Update: original, Pending: pending}, err
	}

	if !confirmedControlReceipt(claim, result) || claim.ResponseSource.Peer != original.Peer || claim.ResponseSource.UpdateID == 0 || claim.Action.State == "consumed" {
		return responseSource{Update: original, Pending: false}, nil
	}

	pending, err = readResponsePending(ctx, tx, claim.ResponseSource)

	return responseSource{Update: claim.ResponseSource, Pending: pending}, err
}

func confirmedControlReceipt(claim ActionClaim, result ActionResult) bool {
	return claim.Request.OperationID == "work.control" && claim.Answer.ID != "" && result.State == "confirmed" && result.ReceiptRef == workflowstate.MutationReceiptPath(claim.Request.ID)
}

func queueActionResultResponse(ctx context.Context, tx *sql.Tx, retained ActionClaim, result ActionResult, now time.Time) error {
	if result.Delivery != nil {
		response := retained.Request
		response.Source = retained.ResponseSource
		response.Delivery = result.Delivery

		return queueOperationResult(ctx, tx, operationReply{Request: response, Text: result.Text, ReceiptRef: result.ReceiptRef, ConfigDigest: result.ConfigDigest, At: now})
	}

	if result.Message != nil {
		return queuePresentation(ctx, tx, presentationReply{Source: retained.ResponseSource, Message: *result.Message, ConfigDigest: result.ConfigDigest, At: now})
	}

	return queueReferencedResponse(ctx, tx, ReferencedReply{Source: retained.ResponseSource, Text: result.Text, Reference: retained.Request.ID, ConfigDigest: result.ConfigDigest, At: now})
}
