package app

import (
	"errors"

	"forge.adrianpk.com/chr/internal/core"
)

var errTelegramLaneChanged = errors.New("telegram input requires another application lane")

func (r telegramReceiver) checkActionLane(claim core.TelegramActionClaim) error {
	if r.processingLane == nil {
		return nil
	}

	lane, err := core.TelegramActionProcessingLane(claim.Request)
	if err != nil {
		return err
	}

	if lane != *r.processingLane {
		return errTelegramLaneChanged
	}

	return nil
}

func (r telegramReceiver) checkCaptureReplyLane(update core.TelegramReceivedUpdate) error {
	if update.ReplyToMessageID <= 0 || r.processingLane == nil {
		return nil
	}

	lane, control := core.TelegramExplicitControlLane(update)
	if control && lane != core.TelegramProcessingOrdinary {
		return nil
	}

	if *r.processingLane != core.TelegramProcessingOrdinary {
		return errTelegramLaneChanged
	}

	return nil
}
