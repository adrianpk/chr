package core

import projectcore "forge.adrianpk.com/chr/internal/core/project"

import (
	"context"
	"database/sql"
	"time"
)

// ResolveTelegramHarnessTarget observes the explicitly selected registered
// project. Configuration inspection never allocates or advances a conversation.
func (s *TelegramReceiverStore) ResolveTelegramHarnessTarget(ctx context.Context, input TelegramConversationInput, now time.Time) (TelegramOperationTarget, error) {
	var target TelegramOperationTarget

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: now}

		project, err := op.selectedHarnessProject()
		if err != nil {
			return err
		}

		op.input.Project = project.ID
		target, err = op.operationTarget(nil)

		return err
	})

	return target, err
}

func (op telegramConversationTransaction) selectedHarnessProject() (projectcore.Project, error) {
	_, allowed := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)
	if !allowed || op.input.Project != "" || op.input.Conversation != "" || op.input.Namespace != "" || op.input.Update.ReplyToMessageID != 0 {
		return projectcore.Project{}, ErrWorkflowConflict
	}

	if err := op.ensurePeer(); err != nil {
		return projectcore.Project{}, err
	}

	selected, _, err := op.peerSelection()
	if err != nil {
		return projectcore.Project{}, err
	}

	if selected == "" {
		return projectcore.Project{}, ErrWorkflowConflict
	}

	return exactTelegramProject(op.allowedProjects(), selected)
}
