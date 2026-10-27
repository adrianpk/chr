package app

import (
	"errors"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

var errTelegramLaneChanged = errors.New("telegram input requires another application lane")

func (r telegramReceiver) checkActionLane(claim telegramcore.TelegramActionClaim) error {
	if r.processingLane == nil {
		return nil
	}

	lane, err := telegramcore.TelegramActionChoiceProcessingLane(claim.Request, claim.Choice)
	if err != nil {
		return err
	}

	if lane != *r.processingLane {
		return errTelegramLaneChanged
	}

	return nil
}

func (r telegramReceiver) checkCaptureReplyLane(update telegramcore.TelegramReceivedUpdate) error {
	if update.ReplyToMessageID <= 0 || r.processingLane == nil {
		return nil
	}

	lane, control := telegramcore.TelegramExplicitControlLane(update)
	if control && lane != telegramcore.TelegramProcessingOrdinary {
		return nil
	}

	if *r.processingLane != telegramcore.TelegramProcessingOrdinary {
		return errTelegramLaneChanged
	}

	return nil
}
