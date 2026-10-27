package core

import (
	"encoding/json"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// ValidateTelegramNotificationDelivery validates a frozen notification view and attachment.
func ValidateTelegramNotificationDelivery(delivery TelegramNotificationDelivery) error {
	p := delivery.Presentation
	if p.PlainFallback == "" || len(p.InitialText)+len(p.Content)+len(p.SecondaryText)+len(p.PlainFallback) > 64<<10 || len(p.Entities) > 32 {
		return tgmessage.ErrDocumentUnavailable
	}

	if err := validateTelegramNotificationActions(p); err != nil {
		return err
	}

	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	if _, err := tgmessage.DecodeRetained(tgmessage.RetainedInput{Payload: string(payload), Hash: contenthash.Sum(payload), Text: p.PlainFallback, DocumentID: p.DocumentID}); err != nil {
		return err
	}

	return validateTelegramNotificationAttachment(p, delivery.Document)
}

func validateTelegramNotificationAttachment(p tgmessage.Presentation, document *tgmessage.Document) error {
	if document == nil {
		if p.DocumentID != "" || p.DocumentHash != "" {
			return tgmessage.ErrDocumentUnavailable
		}

		return nil
	}

	if p.TemplateRevision != "document-v1" || p.DocumentID == "" || p.DocumentHash != contenthash.Sum(document.Payload) {
		return tgmessage.ErrDocumentUnavailable
	}

	return tgmessage.ValidateDocument(*document)
}
