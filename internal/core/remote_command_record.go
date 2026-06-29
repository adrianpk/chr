package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// RemoteCommandRecordStatus defines a package boundary type.
type RemoteCommandRecordStatus string

const (
	// RemoteCommandRecordAccepted defines a package boundary constant.
	RemoteCommandRecordAccepted RemoteCommandRecordStatus = "accepted"
	// RemoteCommandRecordRejected defines a package boundary constant.
	RemoteCommandRecordRejected RemoteCommandRecordStatus = "rejected"
)

// RemoteCommandRecord describes a package boundary value.
type RemoteCommandRecord struct {
	Request         RemoteCommandRequest
	Status          RemoteCommandRecordStatus
	HandoffCommand  string
	RejectionReason string
}

// NewAcceptedRemoteCommandRecord constructs the corresponding package service or value.
func NewAcceptedRemoteCommandRecord(request RemoteCommandRequest) (RemoteCommandRecord, error) {
	record := RemoteCommandRecord{
		Request:        request,
		Status:         RemoteCommandRecordAccepted,
		HandoffCommand: RemoteCommandHandoffCommand(request),
	}

	err := ValidateRemoteCommandRecord(record)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	return record, nil
}

// NewRejectedRemoteCommandRecord constructs the corresponding package service or value.
func NewRejectedRemoteCommandRecord(
	request RemoteCommandRequest,
	rawText string,
	sender RemoteCommandSender,
	reason string,
	receivedAt time.Time,
	createdAt time.Time,
) (RemoteCommandRecord, error) {
	if request.ID == "" {
		request = rejectedRemoteCommandRequest(rawText, sender, receivedAt, createdAt)
	}

	record := RemoteCommandRecord{
		Request:         request,
		Status:          RemoteCommandRecordRejected,
		RejectionReason: strings.TrimSpace(reason),
	}

	err := ValidateRemoteCommandRecord(record)
	if err != nil {
		return RemoteCommandRecord{}, err
	}

	return record, nil
}

// ValidateRemoteCommandRecord validates the corresponding package boundary value.
func ValidateRemoteCommandRecord(record RemoteCommandRecord) error {
	if !record.Status.Valid() {
		return fmt.Errorf("remote command record status is invalid: %s", record.Status)
	}

	if err := ValidateRemoteCommandRequest(record.Request); err != nil {
		return err
	}

	if record.Status == RemoteCommandRecordAccepted && record.HandoffCommand == "" {
		return fmt.Errorf("remote command handoff command is required")
	}

	if record.Status == RemoteCommandRecordRejected && record.RejectionReason == "" {
		return fmt.Errorf("remote command rejection reason is required")
	}

	return nil
}

// Valid reports whether the value is one of the supported constants.
func (s RemoteCommandRecordStatus) Valid() bool {
	switch s {
	case RemoteCommandRecordAccepted,
		RemoteCommandRecordRejected:
		return true
	default:
		return false
	}
}

// RemoteCommandHandoffCommand executes the corresponding package operation.
func RemoteCommandHandoffCommand(request RemoteCommandRequest) string {
	switch request.Intent {
	case RemoteCommandIntentStatus:
		return "chr daemon status"
	case RemoteCommandIntentLoopStart:
		return remoteLoopHandoffCommand(request)
	case RemoteCommandIntentTicketStart:
		return "chr ticket start " + request.TicketKey
	case RemoteCommandIntentTicketComment:
		return "chr ticket comment " + request.TicketKey + " " + shellQuoteRemoteArgument(request.Comment)
	default:
		return ""
	}
}

func remoteLoopHandoffCommand(request RemoteCommandRequest) string {
	parts := []string{"chr", "ticket", "loop"}
	parts = appendRemoteLoopFlags(parts, "--status", request.Loop.Statuses)
	parts = appendRemoteLoopFlags(parts, "--priority", request.Loop.Priorities)
	parts = appendRemoteLoopFlags(parts, "--severity", request.Loop.Severities)
	parts = appendRemoteLoopFlags(parts, "--project-area", request.Loop.ProjectAreas)
	parts = appendRemoteLoopFlags(parts, "--tag", request.Loop.Tags)
	parts = appendRemoteLoopFlags(parts, "--type", request.Loop.Types)
	parts = appendRemoteLoopFlags(parts, "--use-case", request.Loop.UseCases)

	if request.Loop.MaxIterations > 0 {
		parts = append(parts, "--max-iterations", fmt.Sprintf("%d", request.Loop.MaxIterations))
	}

	return strings.Join(parts, " ")
}

func appendRemoteLoopFlags(parts []string, name string, values []string) []string {
	for _, value := range values {
		parts = append(parts, name, shellQuoteRemoteArgument(value))
	}

	return parts
}

func shellQuoteRemoteArgument(value string) string {
	if value == "" {
		return "''"
	}

	if !strings.ContainsAny(value, " \t\n'\"") {
		return value
	}

	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func rejectedRemoteCommandRequest(
	rawText string,
	sender RemoteCommandSender,
	receivedAt time.Time,
	createdAt time.Time,
) RemoteCommandRequest {
	if receivedAt.IsZero() {
		receivedAt = time.Unix(0, 0).UTC()
	}

	if createdAt.IsZero() {
		createdAt = receivedAt
	}

	request := RemoteCommandRequest{
		ID:         remoteRejectedCommandID(sender, rawText, receivedAt),
		Sender:     sender,
		Intent:     RemoteCommandIntentStatus,
		RawText:    strings.TrimSpace(rawText),
		ReceivedAt: receivedAt,
		CreatedAt:  createdAt,
	}
	if request.RawText == "" {
		request.RawText = "<empty>"
	}

	return request
}

func remoteRejectedCommandID(sender RemoteCommandSender, rawText string, receivedAt time.Time) string {
	key := sender.Channel + "\x00" + sender.ExternalID + "\x00rejected\x00" + rawText + "\x00" +
		receivedAt.Format(time.RFC3339Nano)
	sum := sha256.Sum256([]byte(key))

	return "remote-" + hex.EncodeToString(sum[:8])
}
