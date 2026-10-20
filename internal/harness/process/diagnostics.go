// Package process supervises bounded, independently owned subprocess calls.
package process

import (
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"forge.adrianpk.com/chr/internal/harness"
)

var (
	secretField    = regexp.MustCompile(`(?i)(token|secret|password|authorization|api[_-]?key)`)
	bearerValue    = regexp.MustCompile(`(?i)bearer[ \t]+[^\s"']+`)
	assignedSecret = regexp.MustCompile(`(?i)((?:api[_-]?key|token|password|secret)["']?\s*[:=]\s*["']?)[^\s,"'}]+`)
)

// Sanitize removes environment secrets and common credential diagnostics.
func Sanitize(data []byte) []byte {
	value := string(data)

	for _, entry := range os.Environ() {
		name, secret, ok := strings.Cut(entry, "=")
		if ok && secret != "" && secretField.MatchString(name) {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}

	if !needsCredentialPatterns(value) {
		return []byte(value)
	}

	value = bearerValue.ReplaceAllString(value, "Bearer [REDACTED]")
	value = assignedSecret.ReplaceAllString(value, "${1}[REDACTED]")

	return []byte(value)
}

func needsCredentialPatterns(value string) bool {
	// Keep regexp Unicode folding unchanged; the fast path is ASCII-only.
	for _, char := range value {
		if char >= utf8.RuneSelf {
			return true
		}
	}

	lower := strings.ToLower(value)
	for _, marker := range []string{"bearer", "api", "token", "password", "secret"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	return false
}

// Diagnostic retains a sanitized excerpt within the diagnostic ceiling.
func Diagnostic(data []byte) []byte {
	clean := Sanitize(data)

	return clean[:min(len(clean), harness.MaxDiagnostic)]
}

func (e Executor) diagnostic(data []byte) []byte {
	return DiagnosticWithSecrets(data, e.Secrets...)
}

// DiagnosticWithSecrets redacts selected file credentials before excerpt truncation.
func DiagnosticWithSecrets(data []byte, secrets ...string) []byte {
	value := string(data)

	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}

	return Diagnostic([]byte(value))
}
