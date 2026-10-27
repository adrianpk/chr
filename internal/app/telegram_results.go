package app

import (
	"encoding/json"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// Parser results retain recognition independently of malformed recognized input.
type telegramAuthoringParsing struct {
	Command    telegramAuthoringCommand
	Recognized bool
}
type telegramControlParsing struct {
	Command    telegramControlCommand
	Recognized bool
}
type telegramExplanationParsing struct {
	Command    telegramExplanationCommand
	Recognized bool
}
type telegramWorkParsing struct {
	Command    telegramWorkCommand
	Recognized bool
}
type telegramRecordParsing struct {
	Command    telegramRecordCommand
	Recognized bool
}
type telegramHarnessParsing struct {
	Choices    telegramHarnessChoices
	Recognized bool
}
type telegramInputParsing struct {
	Input      telegramcore.ConversationInput
	Recognized bool
}
type telegramConversationParsing struct {
	Input      telegramcore.ConversationInput
	Recognized bool
	Diagnostic string
}
type telegramCommandParts struct{ Command, Body string }

// Needed only indicates that the retained intent requires its existing target guard.
type telegramTargetResolution struct {
	Target telegramcore.OperationTarget
	Needed bool
}
type telegramControlReceipt struct {
	Status WorkflowWorkStatus
	Found  bool
}
type telegramGuidedPreparation struct {
	Input  telegramGuidedInput
	Config telegramcore.ReceiverConfig
}
type telegramGuidedOutcome struct {
	Progress telegramcore.GuidedProgress
	Facts    json.RawMessage
}
type (
	telegramGuidedText    struct{ Action, Content string }
	telegramGuidedRouting struct {
		Input telegramcore.ConversationInput
		Start bool
	}
)

type telegramHarnessRetention struct {
	Items   []WorkflowWorkStatus
	More    bool
	Problem string
}
type telegramIntentObservation struct {
	Context ConversationIntentContext
	Digest  string
}
type telegramReceiverHosts struct {
	Finished <-chan error
	Quiesced <-chan struct{}
}
