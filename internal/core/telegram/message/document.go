package message

import (
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// Document contains a complete authorized source image or an explicitly
// identified export. It never asks the uploader to open a mutable filesystem path.
type Document struct {
	ProjectID            string
	Namespace            string
	ProjectBindingDigest string
	SourceRef            string
	SourceHash           string
	Filename             string
	MediaType            string
	Payload              []byte
}

// DocumentResponse pairs a short view with one complete frozen source.
// This is presentation data, not an operation or confirmation authorization.
type DocumentResponse struct {
	Message  Blocks
	Document Document
}

const (
	// DocumentLimit bounds a complete document payload in bytes.
	DocumentLimit = 2 << 20
	// DocumentQuota bounds total retained receiver and notification document bytes.
	DocumentQuota = 64 << 20
)

// ErrDocumentUnavailable includes invalid images and exhausted storage.
// A completed operation keeps its original receipt when presentation fails.
var ErrDocumentUnavailable = errors.New("complete telegram document unavailable")

// ValidateDocument checks frozen bytes, target, source identity and filename.
func ValidateDocument(document Document) error {
	if len(document.Payload) == 0 || len(document.Payload) > DocumentLimit || !utf8.Valid(document.Payload) {
		return ErrDocumentUnavailable
	}

	for _, err := range []error{validateDocumentTarget(document), validateDocumentSource(document), validateDocumentFilename(document.Filename)} {
		if err != nil {
			return err
		}
	}

	switch document.MediaType {
	case "text/markdown", "application/json", "text/plain":
		return nil
	default:
		return ErrDocumentUnavailable
	}
}

func validateDocumentTarget(document Document) error {
	if document.ProjectID == "" || !opsnamespace.Valid(document.Namespace) {
		return ErrDocumentUnavailable
	}

	decoded, err := hex.DecodeString(document.ProjectBindingDigest)
	if err != nil || len(decoded) != 32 {
		return ErrDocumentUnavailable
	}

	return nil
}

func validateDocumentSource(document Document) error {
	if len(document.SourceRef) == 0 || len(document.SourceRef) > 4096 || !filepath.IsLocal(document.SourceRef) {
		return ErrDocumentUnavailable
	}

	decoded, err := hex.DecodeString(document.SourceHash)
	if err != nil || len(decoded) != 32 {
		return ErrDocumentUnavailable
	}

	return nil
}

func validateDocumentFilename(filename string) error {
	if filename == "" || len(filename) > 128 || strings.ContainsAny(filename, "/\\\r\n\x00\"") || strings.HasPrefix(filename, ".") {
		return ErrDocumentUnavailable
	}

	return nil
}

// DocumentExcerpt returns direct opening body lines without YAML front
// matter. The shortening marker is included in its 600 UTF-16 unit allowance.
func DocumentExcerpt(payload []byte) string {
	body := strings.TrimPrefix(string(payload), "\ufeff")
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	lines = documentBodyLines(lines)

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
	if VisibleUnits(excerpt) > 600 {
		shortened = true
	}

	if shortened {
		const marker = "\n[excerpt shortened]"

		excerpt = Prefix(excerpt, 600-len(marker)) + marker
	}

	return excerpt
}

func documentBodyLines(lines []string) []string {
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

// RenderDocument uses the same optional blocks within the 900-unit
// caption limit. The complete source image remains outside the shortened view.
func RenderDocument(message Blocks) Presentation {
	return render(message, 900, "document-v1")
}
