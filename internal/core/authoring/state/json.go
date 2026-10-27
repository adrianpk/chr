package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// Decode rejects unknown fields and trailing retained authoring data.
func Decode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("trailing authoring state")
	}

	return nil
}
