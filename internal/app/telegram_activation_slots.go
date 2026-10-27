package app

import telegramcore "forge.adrianpk.com/chr/internal/core/telegram"

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

func (s *telegramActivationSlots) lane(lane telegramcore.ProcessingLane) chan struct{} {
	switch lane {
	case telegramcore.ProcessingUrgent:
		return s.urgent
	case telegramcore.ProcessingDecision:
		return s.decision
	default:
		return s.ordinary
	}
}

func (s *telegramActivationSlots) acquire(lane telegramcore.ProcessingLane) bool {
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

func (s *telegramActivationSlots) release(lane telegramcore.ProcessingLane) {
	if s != nil {
		<-s.lane(lane)
	}
}
