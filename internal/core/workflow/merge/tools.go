package merge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Broker messages are the entire model tool surface. Only scoped file operations
// are granted; a model cannot supply shell commands or launch validation code.
type toolRequest struct {
	Operation string                     `json:"operation"`
	Path      string                     `json:"path,omitempty"`
	Content   string                     `json:"content,omitempty"`
	Check     *int                       `json:"check,omitempty"`
	Result    json.RawMessage            `json:"result,omitempty"`
	Usage     *workflowstate.MergerUsage `json:"usage,omitempty"`
}

type toolResponse struct {
	Content string                       `json:"content,omitempty"`
	Check   *workflowstate.CommandResult `json:"check,omitempty"`
	Error   string                       `json:"error,omitempty"`
}

// Tools owns the guarded, contained model file-operation surface.
type Tools struct {
	input  workflowstate.MergerInput
	root   *os.Root
	checks []workflowstate.CommandResult
	guard  func(context.Context) error
}

func (tools *Tools) apply(ctx context.Context, request toolRequest) (toolResponse, error) {
	if err := tools.guard(ctx); err != nil {
		return toolResponse{}, err
	}

	if request.Result != nil || request.Usage != nil {
		return toolResponse{}, workflowstate.ErrConflict
	}

	if request.Operation == "check" {
		// Executing repaired code without a separate command sandbox would bypass
		// the broker's path and credential boundary. Host gates run after diff
		// acceptance; a proposal never grants an unrestricted validation process.
		return toolResponse{}, fmt.Errorf("%w: local validation process is not granted", workflowstate.ErrConflict)
	}

	if request.Check != nil || !PathAllowed(tools.input, request.Path) {
		return toolResponse{}, workflowstate.ErrConflict
	}

	if err := tools.checkPath(request.Path); err != nil {
		return toolResponse{}, err
	}

	return tools.file(request)
}

func (tools *Tools) file(request toolRequest) (toolResponse, error) {
	switch request.Operation {
	case "read":
		if request.Content != "" {
			return toolResponse{}, workflowstate.ErrConflict
		}

		return tools.read(request.Path)
	case "write":
		return toolResponse{}, tools.write(request.Path, request.Content)
	case "delete":
		if request.Content != "" {
			return toolResponse{}, workflowstate.ErrConflict
		}

		return toolResponse{}, tools.root.Remove(request.Path)
	default:
		return toolResponse{}, fmt.Errorf("%w: merger tool is not permitted", workflowstate.ErrConflict)
	}
}

func (tools *Tools) checkPath(path string) error {
	parts := strings.Split(path, "/")
	for i := range parts {
		info, err := tools.root.Lstat(strings.Join(parts[:i+1], "/"))
		if os.IsNotExist(err) && i == len(parts)-1 {
			return nil
		}

		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() || i == len(parts)-1 && !RegularFile(info) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

// RegularFile requires a regular file without another hard-link alias.
func RegularFile(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)

	return info.Mode().IsRegular() && ok && stat.Nlink == 1
}

// NewTools binds a scoped checkout and its exact claimed guard.
// The caller owns root closure; every model operation rechecks the guard.
func NewTools(input workflowstate.MergerInput, root *os.Root, guard func(context.Context) error) *Tools {
	return &Tools{input: input, root: root, guard: guard}
}

// Checks returns the host checks retained during this completed broker exchange.
func (tools *Tools) Checks() []workflowstate.CommandResult {
	return tools.checks
}
