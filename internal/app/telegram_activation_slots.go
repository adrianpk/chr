package app

import "forge.adrianpk.com/chr/internal/core"

// These slots account for in-flight adapter calls only. Native workflow leases
// and project execution capacity remain owned by the common application service.
type telegramActivationSlots struct {
	ordinary chan struct{}
	decision chan struct{}
	urgent   chan struct{}
}

func newTelegramActivationSlots() *telegramActivationSlots {
	return &telegramActivationSlots{ordinary: make(chan struct{}, 12), decision: make(chan struct{}, 2), urgent: make(chan struct{}, 2)}
}

func (s *telegramActivationSlots) lane(lane core.TelegramProcessingLane) chan struct{} {
	switch lane {
	case core.TelegramProcessingUrgent:
		return s.urgent
	case core.TelegramProcessingDecision:
		return s.decision
	default:
		return s.ordinary
	}
}

func (s *telegramActivationSlots) acquire(lane core.TelegramProcessingLane) bool {
	if s == nil {
		return true
	}

	select {
	case s.lane(lane) <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *telegramActivationSlots) release(lane core.TelegramProcessingLane) {
	if s != nil {
		<-s.lane(lane)
	}
}
