package review

import (
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// InputBytes encodes the original bounded, indented review input including its trailing newline.
func InputBytes(input workflowstate.ReviewInput) ([]byte, error) {
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, fmt.Errorf("review input exceeds file bounds")
	}

	return append(data, '\n'), nil
}

// InputHash hashes the unchanged retained review input bytes.
func InputHash(input workflowstate.ReviewInput) (string, error) {
	data, err := InputBytes(input)

	return contenthash.Sum(data), err
}

// CheckInputHash rejects a changed candidate or review input.
func CheckInputHash(input workflowstate.ReviewInput, expected string) error {
	hash, err := InputHash(input)
	if err != nil || hash != expected {
		return fmt.Errorf("%w: review inputs changed during execution", workflowstate.ErrConflict)
	}

	return nil
}
