package metadata

import (
	"fmt"
	"strconv"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// OptionalTicketInt parses an optional ticket integer with the original empty-value contract.
func OptionalTicketInt(value string) (int, error) {
	value = opsmetadata.Unquote(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse ticket integer: %w", err)
	}

	return parsed, nil
}
