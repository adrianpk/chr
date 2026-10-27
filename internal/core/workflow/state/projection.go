package state

import (
	"errors"
	"fmt"
	"strings"
)

// ErrProjectionStale preserves domain progress when a derived export fails.
var ErrProjectionStale = errors.New("projection_stale")

// TrackerMetadataPath validates the original set identity and returns its projection path.
func TrackerMetadataPath(setID string) (string, error) {
	if setID == "" || setID == "." || setID == ".." || strings.ContainsAny(setID, "/\\") {
		return "", fmt.Errorf("invalid tracker set identity")
	}

	return "tracker/projections/" + setID + ".json", nil
}
