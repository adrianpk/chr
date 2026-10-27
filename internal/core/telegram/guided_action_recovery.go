package telegram

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

func telegramGuidedActionRecovering(claim *TelegramActionClaim) bool {
	return claim.Request.OperationID == "guided.step" && claim.Action.State == "applying" && claim.Request.State == "applying"
}

// Recovery of an accepted guided choice keeps its immutable request and answer.
// Expiry and display revisions fence new decisions, not an already applying
// command. The shared owner still requires its original operation and revisions.
func claimTelegramGuidedRecovery(op telegramConversationTransaction, claim *TelegramActionClaim) error {
	if !telegramActionDigestMatches(claim.Action, claim.Request) || claim.Action.ConversationRevision != claim.Request.Target.ConversationRevision {
		return supersedeTelegramAction(op, claim, "superseded")
	}

	if err := validateTelegramActionPresentation(op.ctx, op.tx, claim.Action, claim.Request, op.input.Config.Digest()); err != nil {
		if telegramActionUnavailable(err) {
			return supersedeTelegramAction(op, claim, "superseded")
		}

		return err
	}

	peer := claim.Request.Source.Peer

	current, err := scanTelegramConversation(op.tx.QueryRowContext(op.ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, claim.Request.Target.ConversationID, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID))
	if err != nil {
		return err
	}

	if current.GuidedRunID == "" || current.Revision < claim.Request.Target.ConversationRevision {
		return workflowstate.ErrConflict
	}

	if err := validateTelegramGuidedRecoveryTarget(op, current, claim.Request); err != nil {
		if telegramActionTargetConflict(err) {
			return supersedeTelegramAction(op, claim, "superseded")
		}

		return err
	}

	return readTelegramActionClaimAnswer(op, claim)
}

func validateTelegramGuidedRecoveryTarget(op telegramConversationTransaction, current TelegramConversation, request TelegramOperationRequest) error {
	if request.Source.Peer != op.input.Update.Peer {
		return workflowstate.ErrConflict
	}

	target := request.Target

	target.ConversationRevision = current.Revision
	if err := validateTelegramBoundConversation(current, target, target.ProjectBindingDigest); err != nil {
		return err
	}

	return op.validateTarget(&current)
}
