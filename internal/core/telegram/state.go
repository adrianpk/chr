package telegram

// ReceiverHealth preserves the receiver owner's existing stored spellings.
type ReceiverHealth string

// ReceiverHealth values retain their original stored vocabulary.
const (
	HealthReady    ReceiverHealth = "ready"
	HealthDegraded ReceiverHealth = "degraded"
	HealthBlocked  ReceiverHealth = "blocked"
)

// OperationState preserves the receiver owner's existing stored spellings.
type OperationState string

// OperationState values retain their original stored vocabulary.
const (
	OperationPrepared             OperationState = "prepared"
	OperationReady                OperationState = "ready"
	OperationAwaitingConfirmation OperationState = "awaiting_confirmation"
	OperationApplying             OperationState = "applying"
	OperationConfirmed            OperationState = "confirmed"
	OperationRejected             OperationState = "rejected"
	OperationBlocked              OperationState = "blocked"
)

// DeliveryState preserves the receiver owner's existing stored spellings.
type DeliveryState string

// DeliveryState values retain their original stored vocabulary.
const (
	DeliveryPending   DeliveryState = "pending"
	DeliverySending   DeliveryState = "sending"
	DeliveryDelivered DeliveryState = "delivered"
	DeliveryUncertain DeliveryState = "uncertain"
	DeliveryBlocked   DeliveryState = "blocked"
)

// ActionState preserves the receiver owner's existing stored spellings.
type ActionState string

// ActionState values retain their original stored vocabulary.
const (
	ActionPending    ActionState = "pending"
	ActionApplying   ActionState = "applying"
	ActionConsumed   ActionState = "consumed"
	ActionExpired    ActionState = "expired"
	ActionSuperseded ActionState = "superseded"
)

// ConversationPhase preserves the receiver owner's existing stored spellings.
type ConversationPhase string

// ConversationPhase values retain their original stored vocabulary.
const (
	ConversationCollecting           ConversationPhase = "collecting"
	ConversationClarifying           ConversationPhase = "clarifying"
	ConversationPreparing            ConversationPhase = "preparing"
	ConversationPrepared             ConversationPhase = "prepared"
	ConversationAwaitingConfirmation ConversationPhase = "awaiting_confirmation"
	ConversationExecuting            ConversationPhase = "executing"
	ConversationWaiting              ConversationPhase = "waiting"
	ConversationCompleted            ConversationPhase = "completed"
	ConversationCanceled             ConversationPhase = "canceled"
	ConversationFailed               ConversationPhase = "failed"
)

// InputState preserves the receiver owner's existing stored spellings.
type InputState string

// InputState values retain their original stored vocabulary.
const (
	InputPending   InputState = "pending"
	InputProcessed InputState = "processed"
	InputIgnored   InputState = "ignored"
	InputRejected  InputState = "rejected"
)
