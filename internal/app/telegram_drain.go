package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

const telegramUncertainDeliveryText = "Delivery confirmation was lost; this notice may be repeated."

const telegramRequestCapacityText = "The request queue is full. This request was not accepted. Retry after pending requests finish; existing work is unchanged."

func (r telegramReceiver) drain(ctx context.Context) error {
	updates, err := r.store.PendingUpdates(ctx, r.botID)
	if err != nil {
		return err
	}

	for _, update := range updates {
		if err := r.respond(ctx, update); err != nil {
			return err
		}
	}

	return r.drainDeliveries(ctx)
}

func (r telegramReceiver) drainDeliveries(ctx context.Context) error {
	deliveries, err := r.store.DueDeliveries(ctx, r.botID, time.Now())
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		if err := r.deliver(ctx, delivery); err != nil {
			return err
		}
	}

	return nil
}

func (r telegramReceiver) respond(ctx context.Context, update core.TelegramReceivedUpdate) error {
	config, err := r.currentConfig()
	if err != nil {
		return err
	}

	user, allowed := config.AuthorizedUser(update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID, update.ChatType)
	if !allowed {
		return r.store.RejectPendingUpdate(ctx, update, time.Now())
	}

	if reply, ok := telegramHelpRequest(update); ok {
		return r.queueReadReply(ctx, update, config, reply)
	}

	for _, handler := range []func(context.Context, core.TelegramAllowedUser, core.TelegramReceiverConfig, core.TelegramReceivedUpdate) (bool, error){r.respondGuidedStart, r.respondAction, r.respondGuided, r.respondHarness, r.respondControl, r.respondWork, r.respondRecord, r.respondPreparedWork, r.respondExplanation, r.respondAuthoring, r.respondIntent, r.respondConversation} {
		if handled, err := handler(ctx, user, config, update); handled || err != nil {
			if errors.Is(err, core.ErrTelegramRequestCapacity) {
				return r.store.QueuePresentationResponse(ctx, update, core.TelegramMessageBlocks{Family: "42", InitialText: telegramRequestCapacityText}, config.Digest(), time.Now())
			}

			return err
		}
	}

	reply, err := telegramReadResponse(r.operations, user, update)
	if err != nil {
		reply = telegramReadReply{Text: "The requested read is unavailable. Saved state is unchanged. Run chr telegram doctor locally before retrying."}
	}

	return r.queueReadReply(ctx, update, config, reply)
}

func (r telegramReceiver) deliver(ctx context.Context, delivery core.TelegramReceiverDelivery) error {
	config, err := r.currentConfig()
	if err != nil {
		return err
	}

	chatType := "supergroup"
	if delivery.Peer.ChatID == delivery.Peer.UserID {
		chatType = "private"
	}

	if _, allowed := config.AuthorizedUser(delivery.Peer.ChatID, delivery.Peer.TopicID, delivery.Peer.UserID, chatType); !allowed || config.Digest() != delivery.ConfigDigest {
		return r.store.BlockDelivery(ctx, delivery.ID, time.Now())
	}

	reserved, err := r.store.TryReserveDelivery(ctx, delivery.ID, time.Now())
	if err != nil {
		return err
	}

	if !reserved {
		return nil
	}

	return r.sendReservedDelivery(ctx, delivery)
}

func (r telegramReceiver) sendReservedDelivery(ctx context.Context, delivery core.TelegramReceiverDelivery) error {
	if err := r.store.ValidateTelegramActionDelivery(ctx, delivery); err != nil {
		return r.failedDelivery(ctx, delivery, &core.TelegramAPIError{Code: "action_presentation_unavailable", Permanent: true})
	}

	text := delivery.Text
	if delivery.PlainFallback {
		text = delivery.PlainText
	}

	if delivery.State == "sending" || delivery.State == "uncertain" {
		text = core.BoundTelegramReceiverText(telegramUncertainDeliveryText + "\n" + text)
	}

	if delivery.DocumentID != "" {
		return r.sendDocumentDelivery(ctx, delivery, text)
	}

	messageID, err := r.sendDeliveryBody(ctx, delivery, text)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return r.failedDelivery(ctx, delivery, err)
	}

	return r.store.DeliverySent(ctx, delivery.ID, messageID, time.Now())
}

func (r telegramReceiver) sendDeliveryBody(ctx context.Context, delivery core.TelegramReceiverDelivery, text string) (int64, error) {
	if delivery.Presentation != nil && !delivery.PlainFallback {
		entities := telegramDeliveryEntities(delivery, text)

		return r.api.SendActionPresentation(ctx, r.token, delivery.Peer, text, entities, delivery.ActionToken, delivery.ActionChoices, delivery.Presentation.ActionLabels)
	}

	if delivery.ParseMode == "Markdown" && !delivery.PlainFallback {
		return r.api.SendMarkdown(ctx, r.token, delivery.Peer, text)
	}

	if delivery.Presentation != nil {
		return r.api.SendActionPresentation(ctx, r.token, delivery.Peer, text, nil, delivery.ActionToken, delivery.ActionChoices, delivery.Presentation.ActionLabels)
	}

	return r.api.SendAction(ctx, r.token, delivery.Peer, text, delivery.ActionToken, delivery.ActionChoices)
}

func (r telegramReceiver) failedDelivery(ctx context.Context, delivery core.TelegramReceiverDelivery, failure error) error {
	var apiErr *core.TelegramAPIError
	if (delivery.ParseMode == "Markdown" || delivery.Presentation != nil) && !delivery.PlainFallback && errors.As(failure, &apiErr) && apiErr.Code == "format_rejected" {
		selected, err := r.store.SelectTelegramPlainFallback(ctx, delivery.ID, time.Now())
		if err != nil {
			return err
		}

		if selected {
			delivery.PlainFallback, delivery.State = true, "pending"
			delivery.Attempts++

			return r.deliver(ctx, delivery)
		}
	}

	return r.store.DeliveryFailed(ctx, delivery, failure, time.Now())
}
