package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

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

	for _, handler := range []func(context.Context, core.TelegramAllowedUser, core.TelegramReceiverConfig, core.TelegramReceivedUpdate) (bool, error){r.respondAction, r.respondControl, r.respondWork, r.respondRecord, r.respondPreparedWork, r.respondExplanation, r.respondAuthoring, r.respondIntent, r.respondConversation} {
		if handled, err := handler(ctx, user, config, update); handled || err != nil {
			return err
		}
	}

	text, err := telegramReadResponse(r.operations, user, update)
	if err != nil {
		text = "The requested read is unavailable. Use local diagnostics to inspect project readiness."
	}

	return r.store.QueueReadResponse(ctx, update, core.BoundTelegramReceiverText(text), config.Digest(), time.Now())
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
	text := delivery.Text
	if delivery.State == "sending" || delivery.State == "uncertain" {
		text = core.BoundTelegramReceiverText("Uncertain delivery retry; this notice may repeat.\n" + text)
	}

	messageID, err := r.api.SendAction(ctx, r.token, delivery.Peer, text, delivery.ActionToken, delivery.ActionChoices)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return r.store.DeliveryFailed(ctx, delivery, err, time.Now())
	}

	return r.store.DeliverySent(ctx, delivery.ID, messageID, time.Now())
}
