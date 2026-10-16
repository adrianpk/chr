package app

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

func decodeOperationJSON(text string, target any) error {
	if len(text) > 65536 || !strings.HasPrefix(strings.TrimSpace(text), "{") {
		return errors.New("operation payload must be a bounded JSON object")
	}

	if err := validateOperationJSONKeys(json.NewDecoder(strings.NewReader(text)), 0); err != nil {
		return err
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("trailing operation payload")
	}

	return nil
}

func validateOperationJSONKeys(decoder *json.Decoder, depth int) error {
	if depth > 16 {
		return errors.New("operation payload nesting exceeds limit")
	}

	token, err := decoder.Token()
	if err != nil {
		return err
	}

	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	if delimiter != '{' && delimiter != '[' {
		return errors.New("invalid JSON delimiter")
	}

	seen := map[string]bool{}

	for decoder.More() {
		if delimiter == '{' {
			if err := readOperationJSONKey(decoder, seen); err != nil {
				return err
			}
		}

		if err := validateOperationJSONKeys(decoder, depth+1); err != nil {
			return err
		}
	}

	_, err = decoder.Token()

	return err
}

func readOperationJSONKey(decoder *json.Decoder, seen map[string]bool) error {
	key, err := decoder.Token()
	if err != nil {
		return err
	}

	name, ok := key.(string)
	if !ok || seen[name] {
		return errors.New("duplicate or invalid operation field")
	}

	seen[name] = true

	return nil
}

func requiredOperationFields(output json.RawMessage, names []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(output, &fields); err != nil {
		return nil, err
	}

	if len(fields) != len(names) {
		return nil, errors.New("missing or null operation field")
	}

	for _, name := range names {
		value, found := fields[name]
		if !found || string(value) == "null" {
			return nil, errors.New("missing or null operation field")
		}
	}

	return fields, nil
}
