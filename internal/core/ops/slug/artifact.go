package slug

import (
	"strings"
)

// ArtifactName keeps the existing normalized artifact basename within 48 bytes.
func ArtifactName(value string) string {
	const maxLength = 48

	slug := Normalize(value)
	if len(slug) <= maxLength {
		return slug
	}

	return strings.Trim(slug[:maxLength], "-")
}
