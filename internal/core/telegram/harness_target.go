package telegram

import (
	"context"
	"database/sql"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ResolveHarnessTarget observes the explicitly selected registered
// project. Configuration inspection never allocates or advances a conversation.
func (s *ReceiverStore) ResolveHarnessTarget(ctx context.Context, input ConversationInput, now time.Time) (OperationTarget, error) {
	var target OperationTarget

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}

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

func (op conversationTransaction) selectedHarnessProject() (projectcore.Project, error) {
	_, allowed := op.input.Config.AuthorizedUser(op.input.Update.Peer.ChatID, op.input.Update.Peer.TopicID, op.input.Update.Peer.UserID, op.input.Update.ChatType)
	if !allowed || op.input.Project != "" || op.input.Conversation != "" || op.input.Namespace != "" || op.input.Update.ReplyToMessageID != 0 {
		return projectcore.Project{}, workflowstate.ErrConflict
	}

	if err := op.ensurePeer(); err != nil {
		return projectcore.Project{}, err
	}

	selection, err := op.peerSelection()
	selected := selection.ProjectID

	if err != nil {
		return projectcore.Project{}, err
	}

	if selected == "" {
		return projectcore.Project{}, workflowstate.ErrConflict
	}

	return exactProject(op.allowedProjects(), selected)
}
