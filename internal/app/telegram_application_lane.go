package app

import (
	"errors"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

var errTelegramLaneChanged = errors.New("telegram input requires another application lane")

func (r telegramReceiver) checkActionLane(claim telegramcore.ActionClaim) error {
	if r.processingLane == nil {
		return nil
	}

	lane, err := telegramcore.ActionChoiceProcessingLane(claim.Request, claim.Choice)
	if err != nil {
		return err
	}

	if lane != *r.processingLane {
		return errTelegramLaneChanged
	}

	return nil
}

func (r telegramReceiver) checkCaptureReplyLane(update telegramcore.ReceivedUpdate) error {
	if update.ReplyToMessageID <= 0 || r.processingLane == nil {
		return nil
	}

	lane, control := telegramcore.ExplicitControlLane(update)
	if control && lane != telegramcore.ProcessingOrdinary {
		return nil
	}

	if *r.processingLane != telegramcore.ProcessingOrdinary {
		return errTelegramLaneChanged
	}

	return nil
}
