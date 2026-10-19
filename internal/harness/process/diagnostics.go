// Package process supervises bounded, independently owned subprocess calls.
package process

import (
	"os"
	"regexp"
	"strings"

	"forge.adrianpk.com/chr/internal/harness"
)

var secretField = regexp.MustCompile(`(?i)(token|secret|password|authorization|api[_-]?key)`)
var bearerValue = regexp.MustCompile(`(?i)bearer[ \t]+[^\s"']+`)
var assignedSecret = regexp.MustCompile(`(?i)((?:api[_-]?key|token|password|secret)["']?\s*[:=]\s*["']?)[^\s,"'}]+`)

// Sanitize removes environment secrets and common credential diagnostics.
func Sanitize(data []byte) []byte {
	value := string(data)
	for _, entry := range os.Environ() {
		name, secret, ok := strings.Cut(entry, "=")
		if ok && secret != "" && secretField.MatchString(name) {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}

	value = bearerValue.ReplaceAllString(value, "Bearer [REDACTED]")
	value = assignedSecret.ReplaceAllString(value, "${1}[REDACTED]")

	return []byte(value)
}

// Diagnostic retains a sanitized excerpt within the diagnostic ceiling.
func Diagnostic(data []byte) []byte {
	clean := Sanitize(data)
	return clean[:min(len(clean), harness.MaxDiagnostic)]
}
