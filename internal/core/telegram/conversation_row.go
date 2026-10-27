package telegram

const conversationColumns = `id,bot_id,chat_id,topic_id,user_id,project_id,project_binding_digest,namespace,intent,phase,topic,capture,question_binding,COALESCE(authoring_session_ref,''),COALESCE(authorization_ref,''),revision,COALESCE(invocation_id,''),COALESCE(last_observed_revision,0),COALESCE(guided_run_id,'')`

// conversationRow aligns the retained projection and its nullable-default scan targets.
type conversationRow struct {
	conversation Conversation
}

func (row *conversationRow) scanTargets() []any {
	return []any{
		&row.conversation.ID,
		&row.conversation.Peer.BotID,
		&row.conversation.Peer.ChatID,
		&row.conversation.Peer.TopicID,
		&row.conversation.Peer.UserID,
		&row.conversation.ProjectID,
		&row.conversation.ProjectBindingDigest,
		&row.conversation.Namespace,
		&row.conversation.Intent,
		&row.conversation.Phase,
		&row.conversation.Topic,
		&row.conversation.Capture,
		&row.conversation.QuestionBinding,
		&row.conversation.AuthoringSessionRef,
		&row.conversation.AuthorizationRef,
		&row.conversation.Revision,
		&row.conversation.InvocationID,
		&row.conversation.LastObservedRevision,
		&row.conversation.GuidedRunID,
	}
}
