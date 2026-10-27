package telegram

import "time"

// ActionAnswerResolution distinguishes recognized input from an unrelated update.
type ActionAnswerResolution struct {
	Answer     ActionAnswer
	Recognized bool
}

// ConversationBinding freezes the routing record and operation target together.
type ConversationBinding struct {
	Conversation Conversation
	Target       OperationTarget
}

// OperationLookup distinguishes a missing retained request from a read failure.
type OperationLookup struct {
	Request OperationRequest
	Found   bool
}

// ConversationLookup distinguishes a missing peer-owned binding from a read failure.
type ConversationLookup struct {
	Conversation Conversation
	Found        bool
}

type responseSource struct {
	Update  ReceivedUpdate
	Pending bool
}
type conversationReply struct {
	Text         string
	Conversation *Conversation
}
type (
	peerSelection struct{ ProjectID, RootDigest string }
	multipartBody struct {
		Payload     []byte
		ContentType string
	}
)
type guidedProjection struct {
	Conversation Conversation
	Ready        bool
}
type interpretationReservation struct {
	Interpretation Interpretation
	Call           bool
}
type operationModelReservation struct {
	Model OperationModel
	Call  bool
}
type laneResolution struct {
	Lane       ProcessingLane
	Determined bool
}
type actionTokenResolution struct {
	Token string
	Found bool
}
type (
	yamlField     struct{ Key, Value string }
	failurePolicy struct {
		Code      string
		Permanent bool
		Delay     time.Duration
	}
)
type workReferencePage struct {
	References []WorkReference
	IDs        []string
}
