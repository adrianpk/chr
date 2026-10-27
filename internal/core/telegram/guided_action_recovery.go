package telegram

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

func guidedActionRecovering(claim *ActionClaim) bool {
	return claim.Request.OperationID == "guided.step" && claim.Action.State == "applying" && claim.Request.State == "applying"
}

// Recovery of an accepted guided choice keeps its immutable request and answer.
// Expiry and display revisions fence new decisions, not an already applying
// command. The shared owner still requires its original operation and revisions.
func claimGuidedRecovery(op conversationTransaction, claim *ActionClaim) error {
	if !actionDigestMatches(claim.Action, claim.Request) || claim.Action.ConversationRevision != claim.Request.Target.ConversationRevision {
		return supersedeAction(op, claim, "superseded")
	}

	if err := validateActionPresentation(op.ctx, op.tx, claim.Action, claim.Request, op.input.Config.Digest()); err != nil {
		if actionUnavailable(err) {
			return supersedeAction(op, claim, "superseded")
		}

		return err
	}

	peer := claim.Request.Source.Peer

	current, err := scanConversation(op.tx.QueryRowContext(op.ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, claim.Request.Target.ConversationID, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID))
	if err != nil {
		return err
	}

	if current.GuidedRunID == "" || current.Revision < claim.Request.Target.ConversationRevision {
		return workflowstate.ErrConflict
	}

	if err := validateGuidedRecoveryTarget(op, current, claim.Request); err != nil {
		if actionTargetConflict(err) {
			return supersedeAction(op, claim, "superseded")
		}

		return err
	}

	return readActionClaimAnswer(op, claim)
}

func validateGuidedRecoveryTarget(op conversationTransaction, current Conversation, request OperationRequest) error {
	if request.Source.Peer != op.input.Update.Peer {
		return workflowstate.ErrConflict
	}

	target := request.Target

	target.ConversationRevision = current.Revision
	if err := validateBoundConversation(current, target, target.ProjectBindingDigest); err != nil {
		return err
	}

	return op.validateTarget(&current)
}
