package file

import (
	"bytes"
	"errors"
	"os"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ProviderOutput retains the original file-bounded provider bytes and overflow fact.
type ProviderOutput struct {
	Buffer   bytes.Buffer
	Overflow bool
}

// Write retains the original provider byte bound and full input count.
func (output *ProviderOutput) Write(data []byte) (int, error) {
	length := len(data)

	remaining := workflowstate.FileLimit - output.Buffer.Len()
	if length > remaining {
		data, output.Overflow = data[:remaining], true
	}

	_, err := output.Buffer.Write(data)

	return length, err
}

// CommandOutputLimit is the original retained command prefix bound.
const CommandOutputLimit = 16 << 10

// CommandOutput retains the original command prefix and optional synced file.
type CommandOutput struct {
	Buffer    bytes.Buffer
	Truncated bool
	File      *os.File
}

// Write retains the original prefix bound and optional per-write durable sync.
func (output *CommandOutput) Write(data []byte) (int, error) {
	length := len(data)

	remaining := CommandOutputLimit - output.Buffer.Len()
	if length > remaining {
		data = data[:remaining]
		output.Truncated = true
	}

	_, err := output.Buffer.Write(data)
	if output.File != nil {
		_, writeErr := output.File.Write(data)
		err = errors.Join(err, writeErr, output.File.Sync())
	}

	return length, err
}
