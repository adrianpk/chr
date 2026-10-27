package state

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// MutationBinding identifies a caller-neutral, immutable mutation payload.
// Routing identities belong to adapters and must never enter domain receipts.
type MutationBinding struct {
	RequestID string `json:"request_id"`
	Operator  string `json:"operator"`
}

// Validate checks the opaque request and configured operator identities.
func (binding MutationBinding) Validate() error {
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`).MatchString(binding.RequestID) || binding.Operator == "" || len(binding.Operator) > 128 || strings.ContainsAny(binding.Operator, "\x00\r\n") {
		return errors.New("opaque mutation request and operator identities are required")
	}

	return nil
}

// MutationReceiptPath derives the canonical reference without reading or writing it.
func MutationReceiptPath(requestID string) string {
	return filepath.ToSlash(filepath.Join("run", "operation-receipts", contenthash.Sum([]byte(requestID))+".json"))
}
