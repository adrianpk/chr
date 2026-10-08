package core

import (
	"fmt"
	"os"
	"strings"
)

// Legacy writers remain available for unadopted tickets. Adopted records must
// pass the repository lock, image and revision guards through WorkflowStore.
func readLegacyTicketFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	front, _, err := splitWorkflowMarkdown(content)
	if err == nil && strings.Contains("\n"+front, "\nworkflow_schema:") {
		return nil, fmt.Errorf("adopted workflow ticket requires a guarded record write")
	}

	return content, nil
}
