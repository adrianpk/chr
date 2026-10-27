package telegram

import (
	"io"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// ActionSendRequest sends plain text with the retained opaque action choices.
type ActionSendRequest struct {
	Peer              PeerIdentity
	Text, ActionToken string
	Choices           []string
}

// PresentationSendRequest sends the frozen text and explicit formatting entities.
type PresentationSendRequest struct {
	Peer     PeerIdentity
	Text     string
	Entities []tgmessage.Entity
}

// ActionPresentationRequest retains exact action labels alongside rich text.
type ActionPresentationRequest struct {
	Peer            PeerIdentity
	Text            string
	Entities        []tgmessage.Entity
	ActionToken     string
	Choices, Labels []string
}

// DocumentSendRequest uploads frozen bytes and the retained caption and keyboard.
// Labels preserves omission versus one explicit label list, including an empty one.
type DocumentSendRequest struct {
	Peer        PeerIdentity
	Document    tgmessage.Document
	Caption     string
	Entities    []tgmessage.Entity
	ActionToken string
	Choices     []string
	Labels      [][]string
}

type messageSend struct {
	Peer                         PeerIdentity
	Text, ParseMode, ActionToken string
	Choices                      []string
	Entities                     []tgmessage.Entity
	Labels                       [][]string
}
type requestBody struct {
	Method      string
	Body        io.Reader
	ContentType string
	Timeout     time.Duration
}
type documentUpload struct {
	Destination string
	Topic       int64
	Document    tgmessage.Document
	Caption     string
	Entities    []tgmessage.Entity
	Keyboard    *inlineKeyboard
}
