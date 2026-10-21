package core

import (
	"context"
	"encoding/json"
	"slices"
	"time"
)

func queueTelegramActionView(op telegramConversationTransaction, request TelegramOperationRequest, offer TelegramActionOffer, token, legacy string) error {
	if offer.Delivery == nil {
		return queueTelegramReferencedResponse(op.ctx, op.tx, request.Source, legacy, request.ID, offer.Input.Config.Digest(), op.now)
	}

	document := offer.Delivery.Document
	if document.ProjectID != request.Target.ProjectID || document.Namespace != request.Target.Namespace || document.ProjectBindingDigest != request.Target.ProjectBindingDigest {
		return ErrTelegramDocumentUnavailable
	}

	response := *offer.Delivery
	response.Message.SecondaryText = "Shortened view; complete scope attached. Reply with " + offer.Choices[0] + " or use /answer " + token + " <choice>."

	return queueTelegramDocument(op.ctx, op.tx, request.Source, response, offer.Input.Config.Digest(), op.now)
}

func sealTelegramActionPresentation(op telegramConversationTransaction, request TelegramOperationRequest, token string, choices []string) error {
	var data string
	if err := op.tx.QueryRowContext(op.ctx, "SELECT presentation FROM telegram_outbox WHERE action_token=?", token).Scan(&data); err != nil {
		return err
	}

	var presentation TelegramPresentation
	if err := json.Unmarshal([]byte(data), &presentation); err != nil {
		return err
	}

	presentation.ActionRef, presentation.ScopeDigest = token, telegramActionGuard(request)

	presentation.ActionChoices = slices.Clone(choices)
	for _, choice := range choices {
		presentation.ActionLabels = append(presentation.ActionLabels, telegramActionLabel(choice))
	}

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	hash := workflowHash(payload)
	if _, err := op.tx.ExecContext(op.ctx, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE action_token=?", string(payload), hash, token); err != nil {
		return err
	}

	_, err = op.tx.ExecContext(op.ctx, "UPDATE telegram_actions SET guard_version=1,presentation_hash=?,guard_digest=? WHERE token=?", hash, telegramActionPresentationDigest(request, hash), token)

	return err
}

func telegramActionPresentationDigest(request TelegramOperationRequest, hash string) string {
	return workflowHash([]byte("telegram-action/v1:" + telegramActionGuard(request) + ":" + hash))
}

func telegramActionDigestMatches(action TelegramAction, request TelegramOperationRequest) bool {
	switch action.GuardVersion {
	case 0:
		return action.GuardDigest == telegramActionGuard(request)
	case 1:
		return len(action.PresentationHash) == 64 && action.GuardDigest == telegramActionPresentationDigest(request, action.PresentationHash)
	default:
		return false
	}
}

func validateTelegramActionPresentation(ctx context.Context, reader telegramConversationReader, action TelegramAction, request TelegramOperationRequest, configDigest string) error {
	if action.GuardVersion == 0 {
		return nil
	}

	if action.GuardVersion != 1 || !telegramActionDigestMatches(action, request) {
		return ErrTelegramActionUnavailable
	}

	delivery := TelegramReceiverDelivery{Peer: action.Peer}

	var payload, hash string
	if err := reader.QueryRowContext(ctx, `SELECT id,text,config_digest,presentation,presentation_hash,COALESCE(document_id,'') FROM telegram_outbox WHERE action_token=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, action.Token, action.Peer.BotID, action.Peer.ChatID, action.Peer.TopicID, action.Peer.UserID).Scan(&delivery.ID, &delivery.Text, &delivery.ConfigDigest, &payload, &hash, &delivery.DocumentID); err != nil {
		return err
	}

	if !telegramActionManifestValid(payload, hash, configDigest, &delivery, action, request) {
		return ErrTelegramActionUnavailable
	}

	if delivery.DocumentID == "" {
		return nil
	}

	return validateTelegramActionDocument(ctx, reader, delivery, request)
}

func telegramActionManifestValid(payload, hash, configDigest string, delivery *TelegramReceiverDelivery, action TelegramAction, request TelegramOperationRequest) bool {
	if delivery.ConfigDigest != configDigest || hash != action.PresentationHash || decodeTelegramPresentationHash(payload, hash, delivery) != nil || delivery.Presentation == nil {
		return false
	}

	presentation := delivery.Presentation

	return presentation.ActionRef == action.Token && presentation.ScopeDigest == telegramActionGuard(request) && slices.Equal(presentation.ActionChoices, action.Choices) && len(presentation.ActionLabels) == len(action.Choices)
}

func validateTelegramActionDocument(ctx context.Context, reader telegramConversationReader, delivery TelegramReceiverDelivery, request TelegramOperationRequest) error {
	var (
		hash     string
		document TelegramDocument
	)

	p := request.Source.Peer
	if err := reader.QueryRowContext(ctx, `SELECT project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,payload FROM telegram_document_snapshots WHERE id=? AND owner_id=? AND owner_kind='receiver' AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND config_digest=? AND payload IS NOT NULL`, delivery.DocumentID, delivery.ID, p.BotID, p.ChatID, p.TopicID, p.UserID, delivery.ConfigDigest).Scan(&document.ProjectID, &document.Namespace, &document.ProjectBindingDigest, &document.SourceRef, &document.SourceHash, &hash, &document.Filename, &document.MediaType, &document.Payload); err != nil {
		return err
	}

	if validateTelegramDocument(document) != nil || document.ProjectID != request.Target.ProjectID || document.Namespace != request.Target.Namespace || document.ProjectBindingDigest != request.Target.ProjectBindingDigest || hash != delivery.Presentation.DocumentHash || workflowHash(document.Payload) != hash {
		return ErrTelegramActionUnavailable
	}

	return nil
}

// ValidateTelegramActionDelivery refuses a changed or future action image before
// sending its keyboard. It does not consume an answer or grant authority.
func (s *TelegramReceiverStore) ValidateTelegramActionDelivery(ctx context.Context, delivery TelegramReceiverDelivery) error {
	if delivery.ActionToken == "" {
		return nil
	}

	action, err := readTelegramAction(ctx, s.db, delivery.ActionToken, delivery.Peer)
	if err != nil {
		return err
	}

	if action.GuardVersion == 1 && !action.ExpiresAt.After(time.Now()) {
		return ErrTelegramActionUnavailable
	}

	request, err := readTelegramOperationID(ctx, s.db, action.RequestID)
	if err != nil {
		return err
	}

	return validateTelegramActionPresentation(ctx, s.db, action, request, delivery.ConfigDigest)
}

func sealTelegramOfferedAction(op telegramConversationTransaction, request TelegramOperationRequest, offer TelegramActionOffer, token string) error {
	if offer.Delivery == nil {
		return nil
	}

	return sealTelegramActionPresentation(op, request, token, offer.Choices)
}
