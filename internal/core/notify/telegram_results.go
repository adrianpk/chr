package notify

import (
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// TelegramLookup preserves a legacy/unfrozen event separately from read failure.
type TelegramLookup struct {
	Delivery TelegramDelivery
	Frozen   bool
}

type telegramManifestLookup struct {
	Presentation tgmessage.Presentation
	Frozen       bool
}

// DeliveryFailure retains only a safe transport code and its uncertainty.
type DeliveryFailure struct {
	Code      string
	Uncertain bool
}

// TelegramCompletion records the original event's transport outcome and time.
type TelegramCompletion struct {
	ID              string
	Target          TelegramTarget
	Acknowledgement tgmessage.Acknowledgement
	Failure         DeliveryFailure
	At              time.Time
}
