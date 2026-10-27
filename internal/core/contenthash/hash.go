// Package contenthash owns canonical SHA-256 content identity.
package contenthash

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sum returns the existing lowercase hexadecimal SHA-256 digest.
func Sum(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}
