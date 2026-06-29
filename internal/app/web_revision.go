package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func webStableRevision(value any) string {
	content, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:])
}
