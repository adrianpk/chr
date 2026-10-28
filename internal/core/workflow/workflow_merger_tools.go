package workflow

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
type workflowMergerToolRequest struct {
	Operation string          `json:"operation"`
	Path      string          `json:"path,omitempty"`
	Content   string          `json:"content,omitempty"`
	Check     *int            `json:"check,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
	Usage     *MergerUsage    `json:"usage,omitempty"`
}

type workflowMergerToolResponse struct {
	Content string         `json:"content,omitempty"`
	Check   *CommandResult `json:"check,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type workflowMergerTools struct {
	input  MergerInput
	root   *os.Root
	checks []CommandResult
	guard  func(context.Context) error
}

func (tools *workflowMergerTools) apply(ctx context.Context, request workflowMergerToolRequest) (workflowMergerToolResponse, error) {
	if err := tools.guard(ctx); err != nil {
		return workflowMergerToolResponse{}, err
	}

	if request.Result != nil || request.Usage != nil {
		return workflowMergerToolResponse{}, workflowstate.ErrConflict
	}

	if request.Operation == "check" {
		// Executing repaired code without a separate command sandbox would bypass
		// the broker's path and credential boundary. Host gates run after diff
		// acceptance; a proposal never grants an unrestricted validation process.
		return workflowMergerToolResponse{}, fmt.Errorf("%w: local validation process is not granted", workflowstate.ErrConflict)
	}

	if request.Check != nil || !workflowMergerPathAllowed(tools.input, request.Path) {
		return workflowMergerToolResponse{}, workflowstate.ErrConflict
	}

	if err := tools.checkPath(request.Path); err != nil {
		return workflowMergerToolResponse{}, err
	}

	return tools.file(request)
}

func (tools *workflowMergerTools) file(request workflowMergerToolRequest) (workflowMergerToolResponse, error) {
	switch request.Operation {
	case "read":
		if request.Content != "" {
			return workflowMergerToolResponse{}, workflowstate.ErrConflict
		}

		return tools.read(request.Path)
	case "write":
		return workflowMergerToolResponse{}, tools.write(request.Path, request.Content)
	case "delete":
		if request.Content != "" {
			return workflowMergerToolResponse{}, workflowstate.ErrConflict
		}

		return workflowMergerToolResponse{}, tools.root.Remove(request.Path)
	default:
		return workflowMergerToolResponse{}, fmt.Errorf("%w: merger tool is not permitted", workflowstate.ErrConflict)
	}
}

func (tools *workflowMergerTools) checkPath(path string) error {
	parts := strings.Split(path, "/")
	for i := range parts {
		info, err := tools.root.Lstat(strings.Join(parts[:i+1], "/"))
		if os.IsNotExist(err) && i == len(parts)-1 {
			return nil
		}

		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() || i == len(parts)-1 && !workflowMergerRegularFile(info) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func workflowMergerRegularFile(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)

	return info.Mode().IsRegular() && ok && stat.Nlink == 1
}
