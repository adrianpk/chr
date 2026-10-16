package core

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// TelegramOperationTarget freezes an authorized registry root and namespace.
type TelegramOperationTarget struct {
	ProjectID            string `json:"project_id"`
	ProjectBindingDigest string `json:"project_binding_digest"`
	Namespace            string `json:"namespace"`
	ConversationID       string `json:"conversation_id,omitempty"`
	ConversationRevision int64  `json:"conversation_revision,omitempty"`
}

// ResolveTelegramOperationTarget uses the existing exact reply and project precedence.
// The transaction ends before a caller reads or mutates a project owner.
func (s *TelegramReceiverStore) ResolveTelegramOperationTarget(ctx context.Context, input TelegramConversationInput, now time.Time) (TelegramOperationTarget, error) {
	var target TelegramOperationTarget
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return target, ErrWorkflowConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		conversation, err := op.explicitTarget()
		if err != nil {
			return err
		}

		target, err = op.operationTarget(conversation)

		return err
	})

	return target, err
}

// ValidateTelegramOperationTarget rechecks current actor, root and conversation bindings.
func (s *TelegramReceiverStore) ValidateTelegramOperationTarget(ctx context.Context, input TelegramConversationInput, target TelegramOperationTarget) error {
	user, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	if !ok || user.OperatorID == "" {
		return ErrWorkflowConflict
	}

	op := telegramConversationTransaction{ctx: ctx, input: input}

	project, err := exactTelegramProject(op.allowedProjects(), target.ProjectID)
	if err != nil {
		return err
	}

	binding, err := telegramProjectBinding(project)
	if err != nil {
		return err
	}

	if binding != target.ProjectBindingDigest {
		return ErrWorkflowConflict
	}

	if target.ConversationID == "" {
		return nil
	}

	conversation, err := scanTelegramConversation(s.db.QueryRowContext(ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, target.ConversationID, input.Update.Peer.BotID, input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrWorkflowConflict
	}

	if err != nil {
		return err
	}

	return validateTelegramBoundConversation(conversation, target, binding)
}

func validateTelegramBoundConversation(conversation TelegramConversation, target TelegramOperationTarget, binding string) error {
	if conversation.ProjectID != target.ProjectID || conversation.Namespace != target.Namespace || conversation.Revision != target.ConversationRevision || conversation.ProjectBindingDigest != binding {
		return ErrWorkflowConflict
	}

	return nil
}

func (op telegramConversationTransaction) operationTarget(conversation *TelegramConversation) (TelegramOperationTarget, error) {
	project, err := op.resolveProject(conversation)
	if err != nil {
		return TelegramOperationTarget{}, err
	}

	binding, err := telegramProjectBinding(project)
	if err != nil {
		return TelegramOperationTarget{}, err
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: project.Path, Namespace: op.input.Namespace})
	if err != nil {
		return TelegramOperationTarget{}, err
	}

	target := TelegramOperationTarget{ProjectID: project.ID, ProjectBindingDigest: binding, Namespace: layout.Namespace()}
	if conversation == nil {
		return target, nil
	}

	if op.input.Namespace != "" && op.input.Namespace != conversation.Namespace {
		return target, ErrWorkflowConflict
	}

	if err := op.validateTarget(conversation); err != nil {
		return target, err
	}

	target.Namespace, target.ConversationID, target.ConversationRevision = conversation.Namespace, conversation.ID, conversation.Revision

	return target, nil
}
