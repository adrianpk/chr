package telegram

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func queueActionView(op conversationTransaction, request OperationRequest, offer ActionOffer, token, legacy string) error {
	if offer.Message != nil {
		return queuePresentation(op.ctx, op.tx, request.Source, *offer.Message, offer.Input.Config.Digest(), op.now)
	}

	if offer.Delivery == nil {
		return queueReferencedResponse(op.ctx, op.tx, request.Source, legacy, request.ID, offer.Input.Config.Digest(), op.now)
	}

	document := offer.Delivery.Document
	if document.ProjectID != request.Target.ProjectID || document.Namespace != request.Target.Namespace || document.ProjectBindingDigest != request.Target.ProjectBindingDigest {
		return tgmessage.ErrDocumentUnavailable
	}

	response := *offer.Delivery
	response.Message.SecondaryText = "Shortened view; complete scope attached. Reply with " + offer.Choices[0] + " or use /answer " + token + " <choice>."

	return queueDocument(op.ctx, op.tx, request.Source, response, offer.Input.Config.Digest(), op.now)
}

func sealActionPresentation(op conversationTransaction, request OperationRequest, token string, choices []string) error {
	var data string
	if err := op.tx.QueryRowContext(op.ctx, "SELECT presentation FROM telegram_outbox WHERE action_token=?", token).Scan(&data); err != nil {
		return err
	}

	var presentation tgmessage.Presentation
	if err := json.Unmarshal([]byte(data), &presentation); err != nil {
		return err
	}

	presentation.ActionRef, presentation.ScopeDigest = token, actionGuard(request)

	presentation.ActionChoices = slices.Clone(choices)
	for _, choice := range choices {
		presentation.ActionLabels = append(presentation.ActionLabels, actionLabel(choice))
	}

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	hash := contenthash.Sum(payload)
	if _, err := op.tx.ExecContext(op.ctx, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE action_token=?", string(payload), hash, token); err != nil {
		return err
	}

	_, err = op.tx.ExecContext(op.ctx, "UPDATE telegram_actions SET guard_version=1,presentation_hash=?,guard_digest=? WHERE token=?", hash, actionPresentationDigest(request, hash), token)

	return err
}

func actionPresentationDigest(request OperationRequest, hash string) string {
	return contenthash.Sum([]byte("telegram-action/v1:" + actionGuard(request) + ":" + hash))
}

func actionDigestMatches(action Action, request OperationRequest) bool {
	switch action.GuardVersion {
	case 0:
		return action.GuardDigest == actionGuard(request)
	case 1:
		return len(action.PresentationHash) == 64 && action.GuardDigest == actionPresentationDigest(request, action.PresentationHash)
	default:
		return false
	}
}

func validateActionPresentation(ctx context.Context, reader conversationReader, action Action, request OperationRequest, configDigest string) error {
	if action.GuardVersion == 0 {
		return nil
	}

	if action.GuardVersion != 1 || !actionDigestMatches(action, request) {
		return ErrActionUnavailable
	}

	delivery := ReceiverDelivery{Peer: action.Peer}

	var payload, hash string
	if err := reader.QueryRowContext(ctx, `SELECT id,text,config_digest,presentation,presentation_hash,COALESCE(document_id,'') FROM telegram_outbox WHERE action_token=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, action.Token, action.Peer.BotID, action.Peer.ChatID, action.Peer.TopicID, action.Peer.UserID).Scan(&delivery.ID, &delivery.Text, &delivery.ConfigDigest, &payload, &hash, &delivery.DocumentID); err != nil {
		return err
	}

	if !actionManifestValid(payload, hash, configDigest, &delivery, action, request) {
		return ErrActionUnavailable
	}

	if delivery.DocumentID == "" {
		return nil
	}

	return validateActionDocument(ctx, reader, delivery, request)
}

func actionManifestValid(payload, hash, configDigest string, delivery *ReceiverDelivery, action Action, request OperationRequest) bool {
	if delivery.ConfigDigest != configDigest || hash != action.PresentationHash || decodePresentationHash(payload, hash, delivery) != nil || delivery.Presentation == nil {
		return false
	}

	presentation := delivery.Presentation

	return presentation.ActionRef == action.Token && presentation.ScopeDigest == actionGuard(request) && slices.Equal(presentation.ActionChoices, action.Choices) && len(presentation.ActionLabels) == len(action.Choices)
}

func validateActionDocument(ctx context.Context, reader conversationReader, delivery ReceiverDelivery, request OperationRequest) error {
	var (
		hash     string
		document tgmessage.Document
	)

	p := request.Source.Peer
	if err := reader.QueryRowContext(ctx, `SELECT project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,payload FROM telegram_document_snapshots WHERE id=? AND owner_id=? AND owner_kind='receiver' AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND config_digest=? AND payload IS NOT NULL`, delivery.DocumentID, delivery.ID, p.BotID, p.ChatID, p.TopicID, p.UserID, delivery.ConfigDigest).Scan(&document.ProjectID, &document.Namespace, &document.ProjectBindingDigest, &document.SourceRef, &document.SourceHash, &hash, &document.Filename, &document.MediaType, &document.Payload); err != nil {
		return err
	}

	if tgmessage.ValidateDocument(document) != nil || document.ProjectID != request.Target.ProjectID || document.Namespace != request.Target.Namespace || document.ProjectBindingDigest != request.Target.ProjectBindingDigest || hash != delivery.Presentation.DocumentHash || contenthash.Sum(document.Payload) != hash {
		return ErrActionUnavailable
	}

	return nil
}

// ValidateActionDelivery refuses a changed or future action image before
// sending its keyboard. It does not consume an answer or grant authority.
func (s *ReceiverStore) ValidateActionDelivery(ctx context.Context, delivery ReceiverDelivery) error {
	if delivery.ActionToken == "" {
		return nil
	}

	action, err := readAction(ctx, s.db, delivery.ActionToken, delivery.Peer)
	if err != nil {
		return err
	}

	if action.GuardVersion == 1 && !action.ExpiresAt.After(time.Now()) {
		return ErrActionUnavailable
	}

	request, err := readOperationID(ctx, s.db, action.RequestID)
	if err != nil {
		return err
	}

	return validateActionPresentation(ctx, s.db, action, request, delivery.ConfigDigest)
}

func sealOfferedAction(op conversationTransaction, request OperationRequest, offer ActionOffer, token string) error {
	if offer.Delivery == nil && offer.Message == nil {
		return nil
	}

	return sealActionPresentation(op, request, token, offer.Choices)
}
