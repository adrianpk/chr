package merge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckWorkflowMergerJSON rejects invalid encoding, duplicate keys, nesting and trailing output.
func CheckWorkflowMergerJSON(data []byte) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("%w: merger JSON is not UTF-8", workflowstate.ErrConflict)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := workflowMergerJSONValue(decoder, 0); err != nil {
		return fmt.Errorf("%w: malformed merger JSON: %w", workflowstate.ErrConflict, err)
	}

	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("%w: merger result contains trailing output", workflowstate.ErrConflict)
	}

	return nil
}

func workflowMergerJSONValue(decoder *json.Decoder, depth int) error {
	if depth > 32 {
		return fmt.Errorf("merger JSON nesting exceeds bound")
	}

	token, err := decoder.Token()
	if err != nil {
		return err
	}

	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}

	if delimiter == '{' {
		return workflowMergerJSONObject(decoder, depth+1)
	}

	if delimiter != '[' {
		return fmt.Errorf("invalid JSON delimiter")
	}

	for decoder.More() {
		if err := workflowMergerJSONValue(decoder, depth+1); err != nil {
			return err
		}
	}

	_, err = decoder.Token()

	return err
}

func workflowMergerJSONObject(decoder *json.Decoder, depth int) error {
	seen := map[string]bool{}

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		key, valid := token.(string)
		if !valid || seen[key] || strings.ToLower(key) != key {
			return fmt.Errorf("duplicate or invalid JSON field")
		}

		seen[key] = true

		if err := workflowMergerJSONValue(decoder, depth); err != nil {
			return err
		}
	}

	_, err := decoder.Token()

	return err
}
