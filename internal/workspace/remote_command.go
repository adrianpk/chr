package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// RemoteCommandIntent defines a package boundary type.
type RemoteCommandIntent string

const (
	// RemoteCommandIntentStatus defines a package boundary constant.
	RemoteCommandIntentStatus RemoteCommandIntent = "status"
	// RemoteCommandIntentLoopStart defines a package boundary constant.
	RemoteCommandIntentLoopStart RemoteCommandIntent = "loop_start"
	// RemoteCommandIntentTicketStart defines a package boundary constant.
	RemoteCommandIntentTicketStart RemoteCommandIntent = "ticket_start"
	// RemoteCommandIntentTicketComment defines a package boundary constant.
	RemoteCommandIntentTicketComment RemoteCommandIntent = "ticket_comment"
)

// RemoteCommandRequest describes a package boundary value.
type RemoteCommandRequest struct {
	ID              string
	Sender          RemoteCommandSender
	ProjectSelector string
	Intent          RemoteCommandIntent
	TicketKey       string
	Comment         string
	Loop            RemoteCommandLoopOptions
	RawText         string
	ReceivedAt      time.Time
	CreatedAt       time.Time
}

// RemoteCommandSender describes a package boundary value.
type RemoteCommandSender struct {
	Channel     string
	ExternalID  string
	DisplayName string
}

// RemoteCommandLoopOptions describes a package boundary value.
type RemoteCommandLoopOptions struct {
	Statuses      []string
	Priorities    []string
	Severities    []string
	ProjectAreas  []string
	Tags          []string
	Types         []string
	UseCases      []string
	MaxIterations int
}

// RemoteCommandParseOptions describes a package boundary value.
type RemoteCommandParseOptions struct {
	Sender     RemoteCommandSender
	Text       string
	ReceivedAt time.Time
	CreatedAt  time.Time
}

// ParseRemoteCommandRequest parses the corresponding package data.
func ParseRemoteCommandRequest(opts RemoteCommandParseOptions) (RemoteCommandRequest, error) {
	request := RemoteCommandRequest{
		Sender:     opts.Sender,
		RawText:    strings.TrimSpace(opts.Text),
		ReceivedAt: opts.ReceivedAt,
		CreatedAt:  opts.CreatedAt,
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = request.ReceivedAt
	}

	words := strings.Fields(request.RawText)
	if len(words) == 0 {
		return RemoteCommandRequest{}, fmt.Errorf("remote command text is required")
	}

	words = parseRemoteProjectSelector(words, &request)

	err := parseRemoteCommandIntent(words, &request)
	if err != nil {
		return RemoteCommandRequest{}, err
	}

	request.ID = remoteCommandRequestID(request)

	err = ValidateRemoteCommandRequest(request)
	if err != nil {
		return RemoteCommandRequest{}, err
	}

	return request, nil
}

// ValidateRemoteCommandRequest validates the corresponding package boundary value.
func ValidateRemoteCommandRequest(request RemoteCommandRequest) error {
	if request.ID == "" {
		return fmt.Errorf("remote command id is required")
	}

	if err := validateRemoteCommandSender(request.Sender); err != nil {
		return err
	}

	if err := validateRemoteTextField(request.ProjectSelector, "project selector", true); err != nil {
		return err
	}

	if !request.Intent.Valid() {
		return fmt.Errorf("remote command intent is invalid: %s", request.Intent)
	}

	if err := validateRemoteCommandIntentArguments(request); err != nil {
		return err
	}

	if request.RawText == "" {
		return fmt.Errorf("remote command raw text is required")
	}

	if request.ReceivedAt.IsZero() {
		return fmt.Errorf("remote command received time is required")
	}

	if request.CreatedAt.IsZero() {
		return fmt.Errorf("remote command created time is required")
	}

	return nil
}

// Valid reports whether the value is one of the supported constants.
func (i RemoteCommandIntent) Valid() bool {
	switch i {
	case RemoteCommandIntentStatus,
		RemoteCommandIntentLoopStart,
		RemoteCommandIntentTicketStart,
		RemoteCommandIntentTicketComment:
		return true
	default:
		return false
	}
}

func parseRemoteProjectSelector(words []string, request *RemoteCommandRequest) []string {
	first := words[0]
	if !strings.HasPrefix(first, "@") || len(first) == 1 {
		return words
	}

	request.ProjectSelector = strings.TrimPrefix(first, "@")

	return words[1:]
}

func parseRemoteCommandIntent(words []string, request *RemoteCommandRequest) error {
	if len(words) == 0 {
		return fmt.Errorf("remote command intent is required")
	}

	switch words[0] {
	case "status":
		return parseRemoteStatusCommand(words, request)
	case "loop":
		return parseRemoteLoopCommand(words, request)
	case "ticket":
		return parseRemoteTicketCommand(words, request)
	default:
		return fmt.Errorf("unsupported remote command: %s", words[0])
	}
}

func parseRemoteStatusCommand(words []string, request *RemoteCommandRequest) error {
	if len(words) != 1 {
		return fmt.Errorf("status remote command does not accept arguments")
	}

	request.Intent = RemoteCommandIntentStatus

	return nil
}

func parseRemoteLoopCommand(words []string, request *RemoteCommandRequest) error {
	if len(words) < 2 || words[1] != "start" {
		return fmt.Errorf("unsupported loop remote command")
	}

	request.Intent = RemoteCommandIntentLoopStart

	return parseRemoteLoopFlags(words[2:], request)
}

func parseRemoteTicketCommand(words []string, request *RemoteCommandRequest) error {
	if len(words) < 3 {
		return fmt.Errorf("ticket remote command requires an action and ticket key")
	}

	switch words[1] {
	case "start":
		return parseRemoteTicketStartCommand(words, request)
	case "comment":
		return parseRemoteTicketCommentCommand(words, request)
	default:
		return fmt.Errorf("unsupported ticket remote command: %s", words[1])
	}
}

func parseRemoteTicketStartCommand(words []string, request *RemoteCommandRequest) error {
	if len(words) != 3 {
		return fmt.Errorf("ticket start remote command requires exactly one ticket key")
	}

	request.Intent = RemoteCommandIntentTicketStart
	request.TicketKey = words[2]

	return nil
}

func parseRemoteTicketCommentCommand(words []string, request *RemoteCommandRequest) error {
	if len(words) < 4 {
		return fmt.Errorf("ticket comment remote command requires comment text")
	}

	request.Intent = RemoteCommandIntentTicketComment
	request.TicketKey = words[2]
	request.Comment = strings.Join(words[3:], " ")

	return nil
}

func parseRemoteLoopFlags(words []string, request *RemoteCommandRequest) error {
	for index := 0; index < len(words); index++ {
		flag := words[index]
		if !strings.HasPrefix(flag, "--") {
			return fmt.Errorf("loop remote command argument must be a flag: %s", flag)
		}

		name, value, consumed, err := remoteCommandFlagValue(words, index)
		if err != nil {
			return err
		}

		index += consumed

		err = applyRemoteLoopFlag(request, name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func remoteCommandFlagValue(words []string, index int) (string, string, int, error) {
	flag := strings.TrimPrefix(words[index], "--")

	name, value, ok := strings.Cut(flag, "=")
	if ok {
		if value == "" {
			return "", "", 0, fmt.Errorf("remote command flag value is required: --%s", name)
		}

		return name, value, 0, nil
	}

	if index+1 >= len(words) {
		return "", "", 0, fmt.Errorf("remote command flag value is required: --%s", name)
	}

	return name, words[index+1], 1, nil
}

func applyRemoteLoopFlag(request *RemoteCommandRequest, name, value string) error {
	switch name {
	case "status":
		request.Loop.Statuses = append(request.Loop.Statuses, value)
	case "priority":
		request.Loop.Priorities = append(request.Loop.Priorities, value)
	case "severity":
		request.Loop.Severities = append(request.Loop.Severities, value)
	case "project-area":
		request.Loop.ProjectAreas = append(request.Loop.ProjectAreas, value)
	case "tag":
		request.Loop.Tags = append(request.Loop.Tags, value)
	case "type":
		request.Loop.Types = append(request.Loop.Types, value)
	case "use-case":
		request.Loop.UseCases = append(request.Loop.UseCases, value)
	case "max-iterations":
		return applyRemoteLoopMaxIterations(request, value)
	default:
		return fmt.Errorf("unsupported loop remote command flag: --%s", name)
	}

	return nil
}

func applyRemoteLoopMaxIterations(request *RemoteCommandRequest, value string) error {
	maxIterations, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid loop max iterations: %s", value)
	}

	request.Loop.MaxIterations = maxIterations

	return nil
}

func validateRemoteCommandSender(sender RemoteCommandSender) error {
	if err := validateRemoteTextField(sender.Channel, "sender channel", false); err != nil {
		return err
	}

	if err := validateRemoteTextField(sender.ExternalID, "sender external id", false); err != nil {
		return err
	}

	return validateRemoteTextField(sender.DisplayName, "sender display name", true)
}

func validateRemoteCommandIntentArguments(request RemoteCommandRequest) error {
	switch request.Intent {
	case RemoteCommandIntentStatus:
		return validateRemoteStatusArguments(request)
	case RemoteCommandIntentLoopStart:
		return validateRemoteLoopArguments(request.Loop)
	case RemoteCommandIntentTicketStart:
		return validateRemoteTicketKey(request.TicketKey)
	case RemoteCommandIntentTicketComment:
		return validateRemoteTicketCommentArguments(request)
	default:
		return nil
	}
}

func validateRemoteStatusArguments(request RemoteCommandRequest) error {
	if request.TicketKey != "" || request.Comment != "" || !request.Loop.empty() {
		return fmt.Errorf("status remote command does not accept ticket or loop arguments")
	}

	return nil
}

func validateRemoteLoopArguments(loop RemoteCommandLoopOptions) error {
	if loop.MaxIterations < 0 {
		return fmt.Errorf("loop max iterations cannot be negative")
	}

	for _, value := range remoteLoopFreeformValues(loop) {
		if err := validateRemoteTextField(value, "loop filter value", false); err != nil {
			return err
		}
	}

	return nil
}

func validateRemoteTicketCommentArguments(request RemoteCommandRequest) error {
	if err := validateRemoteTicketKey(request.TicketKey); err != nil {
		return err
	}

	return validateRemoteTextField(request.Comment, "ticket comment", false)
}

func validateRemoteTicketKey(ticketKey string) error {
	return validateRemoteTextField(ticketKey, "ticket key", false)
}

func validateRemoteTextField(value, name string, optional bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if optional {
			return nil
		}

		return fmt.Errorf("remote command %s is required", name)
	}

	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("remote command %s contains control characters", name)
		}
	}

	return nil
}

func (loop RemoteCommandLoopOptions) empty() bool {
	return len(loop.Statuses) == 0 &&
		len(loop.Priorities) == 0 &&
		len(loop.Severities) == 0 &&
		len(loop.ProjectAreas) == 0 &&
		len(loop.Tags) == 0 &&
		len(loop.Types) == 0 &&
		len(loop.UseCases) == 0 &&
		loop.MaxIterations == 0
}

func remoteLoopFreeformValues(loop RemoteCommandLoopOptions) []string {
	values := []string{}
	values = append(values, loop.Statuses...)
	values = append(values, loop.Priorities...)
	values = append(values, loop.Severities...)
	values = append(values, loop.ProjectAreas...)
	values = append(values, loop.Tags...)
	values = append(values, loop.Types...)
	values = append(values, loop.UseCases...)

	return values
}

func remoteCommandRequestID(request RemoteCommandRequest) string {
	key := request.Sender.Channel + "\x00" +
		request.Sender.ExternalID + "\x00" +
		request.RawText + "\x00" +
		request.ReceivedAt.Format(time.RFC3339Nano)
	sum := sha256.Sum256([]byte(key))

	return "remote-" + hex.EncodeToString(sum[:8])
}
