package telegram

import (
	"context"
	"database/sql"
	"errors"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func bindInitialTelegramAuthoringWork(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest) error {
	if request.OperationID != "authoring.work" {
		return nil
	}

	p := request.Source.Peer

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET first_authoring_work_request=COALESCE(first_authoring_work_request,?) WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND project_id=? AND namespace=? AND project_binding_digest=? AND revision=?`, request.ID, request.Target.ConversationID, p.BotID, p.ChatID, p.TopicID, p.UserID, request.Target.ProjectID, request.Target.Namespace, request.Target.ProjectBindingDigest, request.Target.ConversationRevision)
	if err != nil {
		return err
	}

	return telegramRequireChanged(changed)
}

// IsInitialTelegramAuthoringWork identifies the one immutable first reservation.
// It supplies no authority; the host must also validate the original explicit
// instruction and current operator, project, source images and prepared scope.
func (s *TelegramReceiverStore) IsInitialTelegramAuthoringWork(ctx context.Context, request TelegramOperationRequest) (bool, error) {
	var first string

	err := s.db.QueryRowContext(ctx, `SELECT c.first_authoring_work_request FROM telegram_conversations c JOIN telegram_requests r ON r.conversation_id=c.id WHERE r.id=? AND r.argument_digest=? AND r.operation_id='authoring.work' AND r.bot_id=? AND r.update_id=?`, request.ID, request.ArgumentDigest, request.Source.Peer.BotID, request.Source.UpdateID).Scan(&first)

	return first == request.ID, err
}

// LatestTelegramAuthoringWork reads the retained preview/receipt for an exact
// currently authorized conversation. It does not grant or refresh a prior scope.
func (s *TelegramReceiverStore) LatestTelegramAuthoringWork(ctx context.Context, input TelegramConversationInput, target TelegramOperationTarget) (TelegramOperationRequest, bool, error) {
	if target.ConversationID == "" {
		return TelegramOperationRequest{}, false, workflowstate.ErrConflict
	}

	if err := s.ValidateTelegramOperationTarget(ctx, input, target); err != nil {
		return TelegramOperationRequest{}, false, err
	}

	var id string

	err := s.db.QueryRowContext(ctx, `SELECT id FROM telegram_requests WHERE conversation_id=? AND operation_id='authoring.work' ORDER BY rowid DESC LIMIT 1`, target.ConversationID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return TelegramOperationRequest{}, false, nil
	}

	if err != nil {
		return TelegramOperationRequest{}, false, err
	}

	request, err := readTelegramOperationID(ctx, s.db, id)

	return request, err == nil, err
}

func projectTelegramAuthoringWorkPhase(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, phase, now string) error {
	if phase == "" {
		return nil
	}

	if request.OperationID != "authoring.work" || (phase != "prepared" && phase != "awaiting_confirmation") {
		return workflowstate.ErrConflict
	}

	_, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET phase=?,updated_at=? WHERE id=? AND revision=? AND project_binding_digest=? AND phase IN ('preparing','prepared','awaiting_confirmation')`, phase, now, request.Target.ConversationID, request.Target.ConversationRevision, request.Target.ProjectBindingDigest)

	return err
}
