package telegram

import (
	"context"
	"database/sql"
	"errors"
	"time"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// OperationTarget freezes an authorized registry root and namespace.
type OperationTarget struct {
	ProjectID            string `json:"project_id"`
	ProjectBindingDigest string `json:"project_binding_digest"`
	Namespace            string `json:"namespace"`
	ConversationID       string `json:"conversation_id,omitempty"`
	ConversationRevision int64  `json:"conversation_revision,omitempty"`
}

// ValidateProjectTarget checks the resolved registry root immediately
// before an adapter enters a project-scoped owner or configured provider.
func ValidateProjectTarget(project projectcore.Project, target OperationTarget) error {
	if project.ID != target.ProjectID || !project.Enabled {
		return workflowstate.ErrConflict
	}

	binding, err := projectBinding(project)
	if err != nil {
		return err
	}

	if binding != target.ProjectBindingDigest {
		return workflowstate.ErrConflict
	}

	return nil
}

// ResolveOperationTarget uses the existing exact reply and project precedence.
// The transaction ends before a caller reads or mutates a project owner.
func (s *ReceiverStore) ResolveOperationTarget(ctx context.Context, input ConversationInput, now time.Time) (OperationTarget, error) {
	var target OperationTarget
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return target, workflowstate.ErrConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
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

// ValidateOperationTarget rechecks current actor, root and conversation bindings.
func (s *ReceiverStore) ValidateOperationTarget(ctx context.Context, input ConversationInput, target OperationTarget) error {
	return s.validateOperationTarget(ctx, input, target, false)
}

// ValidateControlReceiptTarget checks current actor, peer, registry root
// and namespace while allowing the original conversation revision to be historical.
// This check authorizes only reading an already committed native receipt.
func (s *ReceiverStore) ValidateControlReceiptTarget(ctx context.Context, input ConversationInput, request OperationRequest) error {
	user, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	if !ok || user.OperatorID != request.Operator || input.Update.Peer != request.Source.Peer || request.OperationID != "work.control" || request.Target.ConversationID == "" {
		return workflowstate.ErrConflict
	}

	return s.validateOperationTarget(ctx, input, request.Target, true)
}

func (s *ReceiverStore) validateOperationTarget(ctx context.Context, input ConversationInput, target OperationTarget, historical bool) error {
	user, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	if !ok || user.OperatorID == "" {
		return workflowstate.ErrConflict
	}

	op := conversationTransaction{ctx: ctx, input: input}

	project, err := exactProject(op.allowedProjects(), target.ProjectID)
	if err != nil {
		return err
	}

	binding, err := projectBinding(project)
	if err != nil {
		return err
	}

	if binding != target.ProjectBindingDigest {
		return workflowstate.ErrConflict
	}

	if target.ConversationID == "" {
		return nil
	}

	conversation, err := scanConversation(s.db.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, target.ConversationID, input.Update.Peer.BotID, input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return workflowstate.ErrConflict
	}

	if err != nil {
		return err
	}

	if historical {
		target.ConversationRevision = conversation.Revision
	}

	return validateBoundConversation(conversation, target, binding)
}

func validateBoundConversation(conversation Conversation, target OperationTarget, binding string) error {
	if conversation.ProjectID != target.ProjectID || conversation.Namespace != target.Namespace || conversation.Revision != target.ConversationRevision || conversation.ProjectBindingDigest != binding {
		return workflowstate.ErrConflict
	}

	return nil
}

func (op conversationTransaction) operationTarget(conversation *Conversation) (OperationTarget, error) {
	project, err := op.resolveProject(conversation)
	if err != nil {
		return OperationTarget{}, err
	}

	binding, err := projectBinding(project)
	if err != nil {
		return OperationTarget{}, err
	}

	if _, err := projectcore.ResolveDirectory(project.Path); err != nil {
		return OperationTarget{}, err
	}

	namespace, err := opsnamespace.Resolve(op.input.Namespace, nil)
	if err != nil {
		return OperationTarget{}, err
	}

	target := OperationTarget{ProjectID: project.ID, ProjectBindingDigest: binding, Namespace: namespace}
	if conversation == nil {
		return target, nil
	}

	if op.input.Namespace != "" && op.input.Namespace != conversation.Namespace {
		return target, workflowstate.ErrConflict
	}

	if err := op.validateTarget(conversation); err != nil {
		return target, err
	}

	target.Namespace, target.ConversationID, target.ConversationRevision = conversation.Namespace, conversation.ID, conversation.Revision

	return target, nil
}
