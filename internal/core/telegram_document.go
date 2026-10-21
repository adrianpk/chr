package core

import (
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// TelegramDocument contains a complete authorized source image or an explicitly
// identified export. It never asks the uploader to open a mutable filesystem path.
type TelegramDocument struct {
	ProjectID            string
	Namespace            string
	ProjectBindingDigest string
	SourceRef            string
	SourceHash           string
	Filename             string
	MediaType            string
	Payload              []byte
}

// TelegramDocumentResponse pairs a short view with one complete frozen source.
// This is presentation data, not an operation or confirmation authorization.
type TelegramDocumentResponse struct {
	Message  TelegramMessageBlocks
	Document TelegramDocument
}

const (
	telegramDocumentLimit = 2 << 20
	telegramDocumentQuota = 64 << 20
)

// ErrTelegramDocumentUnavailable includes invalid images and exhausted storage.
// A completed operation keeps its original receipt when presentation fails.
var ErrTelegramDocumentUnavailable = errors.New("complete telegram document unavailable")

func validateTelegramDocument(document TelegramDocument) error {
	if len(document.Payload) == 0 || len(document.Payload) > telegramDocumentLimit || !utf8.Valid(document.Payload) {
		return ErrTelegramDocumentUnavailable
	}
	for _, err := range []error{validateTelegramDocumentTarget(document), validateTelegramDocumentSource(document), validateTelegramDocumentFilename(document.Filename)} {
		if err != nil {
			return err
		}
	}
	switch document.MediaType {
	case "text/markdown", "application/json", "text/plain":
		return nil
	default:
		return ErrTelegramDocumentUnavailable
	}
}

func validateTelegramDocumentTarget(document TelegramDocument) error {
	if document.ProjectID == "" || !IsValidOpsNamespace(document.Namespace) {
		return ErrTelegramDocumentUnavailable
	}
	decoded, err := hex.DecodeString(document.ProjectBindingDigest)
	if err != nil || len(decoded) != 32 {
		return ErrTelegramDocumentUnavailable
	}
	return nil
}

func validateTelegramDocumentSource(document TelegramDocument) error {
	if len(document.SourceRef) == 0 || len(document.SourceRef) > 4096 || !filepath.IsLocal(document.SourceRef) {
		return ErrTelegramDocumentUnavailable
	}
	decoded, err := hex.DecodeString(document.SourceHash)
	if err != nil || len(decoded) != 32 {
		return ErrTelegramDocumentUnavailable
	}
	return nil
}

func validateTelegramDocumentFilename(filename string) error {
	if filename == "" || len(filename) > 128 || strings.ContainsAny(filename, "/\\\r\n\x00\"") || strings.HasPrefix(filename, ".") {
		return ErrTelegramDocumentUnavailable
	}
	return nil
}

// TelegramDocumentExcerpt returns direct opening body lines without YAML front
// matter. The shortening marker is included in its 600 UTF-16 unit allowance.
func TelegramDocumentExcerpt(payload []byte) string {
	body := strings.TrimPrefix(string(payload), "\ufeff")
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	lines = telegramDocumentBodyLines(lines)

	var selected []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			selected = append(selected, line)
		}
	}

	shortened := len(selected) > 6
	if shortened {
		selected = selected[:6]
	}

	excerpt := strings.Join(selected, "\n")
	if telegramVisibleUnits(excerpt) > 600 {
		shortened = true
	}

	if shortened {
		const marker = "\n[excerpt shortened]"

		excerpt = telegramResponsePrefix(excerpt, 600-len(marker)) + marker
	}

	return excerpt
}

func telegramDocumentBodyLines(lines []string) []string {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return lines
	}

	for index, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
			return lines[index+2:]
		}
	}

	return nil
}

// RenderTelegramDocument uses the same optional blocks within the 900-unit
// caption limit. The complete source image remains outside the shortened view.
func RenderTelegramDocument(message TelegramMessageBlocks) TelegramPresentation {
	return renderTelegramMessage(message, 900, "document-v1")
}
