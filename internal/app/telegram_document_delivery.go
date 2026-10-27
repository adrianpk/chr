package app

import (
	"context"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramDeliveryEntities(delivery core.TelegramReceiverDelivery, text string) []core.TelegramTextEntity {
	if delivery.Presentation == nil || delivery.PlainFallback {
		return nil
	}

	entities := slices.Clone(delivery.Presentation.Entities)
	prefixUnits := len([]rune(text)) - len([]rune(delivery.Text))
	// The retained retry notice is ASCII. Its length is also its UTF-16 length.
	for index := range entities {
		entities[index].Offset += prefixUnits
	}

	return entities
}

func (r telegramReceiver) sendDocumentDelivery(ctx context.Context, delivery core.TelegramReceiverDelivery, text string) error {
	document, err := r.store.ReadDeliveryDocument(ctx, delivery)
	if err == nil {
		err = r.validateDocumentDestination(document, delivery)
	}

	if err != nil {
		return r.failedDelivery(ctx, delivery, &core.TelegramAPIError{Code: "document_unavailable_or_revoked", Permanent: true})
	}

	acknowledgement, err := r.api.SendDocument(ctx, r.token, delivery.Peer, document, text, telegramDeliveryEntities(delivery, text), delivery.ActionToken, delivery.ActionChoices, delivery.Presentation.ActionLabels)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return r.failedDelivery(ctx, delivery, err)
	}

	return r.store.DeliveryDocumentSent(ctx, delivery.ID, acknowledgement, time.Now())
}

func (r telegramReceiver) validateDocumentDestination(document core.TelegramDocument, delivery core.TelegramReceiverDelivery) error {
	config, err := r.currentConfig()
	if err != nil || config.Digest() != delivery.ConfigDigest {
		return workflowstate.ErrConflict
	}

	chatType := "supergroup"
	if delivery.Peer.ChatID == delivery.Peer.UserID {
		chatType = "private"
	}

	user, allowed := config.AuthorizedUser(delivery.Peer.ChatID, delivery.Peer.TopicID, delivery.Peer.UserID, chatType)
	if !allowed || !slices.Contains(user.ProjectIDs, document.ProjectID) {
		return workflowstate.ErrConflict
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return err
	}

	for _, project := range projects {
		if project.ID == document.ProjectID && project.Enabled {
			return core.ValidateTelegramProjectTarget(project, core.TelegramOperationTarget{ProjectID: document.ProjectID, Namespace: document.Namespace, ProjectBindingDigest: document.ProjectBindingDigest})
		}
	}

	return workflowstate.ErrConflict
}
