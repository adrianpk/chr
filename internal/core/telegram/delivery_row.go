package telegram

const receiverDeliveryColumns = `o.id,o.chat_id,o.topic_id,o.user_id,o.text,o.state,o.attempts,o.config_digest,COALESCE(o.action_token,''),COALESCE((SELECT allowed_choices FROM telegram_actions WHERE token=o.action_token),'[]'),o.parse_mode,o.plain_text,(o.plain_fallback OR o.presentation_fallback),COALESCE(o.presentation,''),COALESCE(o.presentation_hash,''),COALESCE(o.document_id,'')`

type receiverDeliveryRow struct {
	delivery                                ReceiverDelivery
	choices, presentation, presentationHash string
}

func (row *receiverDeliveryRow) scanTargets() []any {
	return []any{
		&row.delivery.ID,
		&row.delivery.Peer.ChatID,
		&row.delivery.Peer.TopicID,
		&row.delivery.Peer.UserID,
		&row.delivery.Text,
		&row.delivery.State,
		&row.delivery.Attempts,
		&row.delivery.ConfigDigest,
		&row.delivery.ActionToken,
		&row.choices,
		&row.delivery.ParseMode,
		&row.delivery.PlainText,
		&row.delivery.PlainFallback,
		&row.presentation,
		&row.presentationHash,
		&row.delivery.DocumentID,
	}
}
