package workflow

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTaskCommand(t *testing.T, mode string) []string {
	t.Helper()

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	return []string{program, "-test.run=^TestWorkflowTaskProcess$", "--", mode}
}

// TestWorkflowTaskProcess is a real worker/gate fixture with frozen stdin and cwd.
func TestWorkflowTaskProcess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-2] != "--" {
		return
	}

	input := workflowProcessInput(t)
	mode := os.Args[len(os.Args)-1]

	path := filepath.Join(input.Lane.Worktree, input.TaskKey+".txt")
	if strings.HasPrefix(mode, "gate") {
		workflowProcessGate(t, input, path, mode)

		return
	}

	if strings.HasPrefix(mode, "edit") {
		workflowProcessEdit(t, input, mode)

		return
	}

	workflowProcessWorker(t, input, path, mode)
}

func workflowProcessWorker(t *testing.T, input TaskInput, path, mode string) {
	t.Helper()

	switch mode {
	case "worker", "instructed worker":
		workflowProcessWrite(t, path, input.TaskKey+"\n")
		// Model a worker that edits successfully but needs explicit commit authority.
		if mode == "instructed worker" && !strings.Contains(input.Instructions, "Commit the task changes") {
			return
		}

		workflowProcessCommit(t, input.Lane.Worktree)
	case "repair worker":
		workflowProcessWrite(t, path+".repair", "Repaired\n")
		workflowProcessCommit(t, input.Lane.Worktree)
	case "dirty":
		workflowProcessWrite(t, path, "Uncommitted\n")
	case "fail":
		os.Exit(7)
	case "flood":
		workflowProcessWrite(t, path, input.TaskKey+"\n")
		workflowProcessCommit(t, input.Lane.Worktree)

		workflowProcessFlood(t)
	default:
		t.Fatal("unknown process mode")
	}
}

func workflowProcessEdit(t *testing.T, input TaskInput, mode string) {
	t.Helper()

	path := os.Getenv("CHR_WORKFLOW_INPUT")
	if mode == "edit input" {
		workflowProcessWrite(t, path, "Changed\n")

		return
	}

	path = filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(path))), "workflow.json")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var run workflowInvocationFile
	if err := json.Unmarshal(data, &run); err != nil {
		t.Fatal(err)
	}

	if mode == "edit process" {
		run.Attempts[0].Processes[0].Identity.StartTicks += "0"
	} else {
		run.Lane.Worktree = filepath.Join(input.Lane.Worktree, "different")
	}

	data, err = json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func workflowProcessFlood(t *testing.T) {
	t.Helper()

	if _, err := os.Stdout.WriteString(strings.Repeat("x", 100000)); err != nil {
		t.Fatal(err)
	}
}

func workflowProcessGate(t *testing.T, input TaskInput, path, mode string) {
	t.Helper()

	switch mode {
	case "gate":
		data, err := os.ReadFile(path)
		if err != nil || string(data) != input.TaskKey+"\n" {
			t.Fatalf("worker output missing: %v", err)
		}
	case "gate aggregate":
		workflowProcessAggregateGate(t, input)
	case "gate dirty":
		workflowProcessWrite(t, filepath.Join(input.Lane.Worktree, "gate.tmp"), "Gate mutation\n")
	case "gate commit":
		workflowProcessWrite(t, path, "Gate mutation\n")
		workflowProcessCommit(t, input.Lane.Worktree)
	case "gate sleep":
		time.Sleep(3 * time.Second)
	default:
		t.Fatal("unknown process mode")
	}
}

func workflowProcessAggregateGate(t *testing.T, input TaskInput) {
	t.Helper()

	if input.ExecutionMode != "aggregate" || input.Instructions != "" || input.Slice == nil || len(input.Slice.Tasks) != 3 {
		t.Fatal("aggregate gate received partial task input")
	}

	for _, proof := range input.Slice.Tasks {
		data, err := os.ReadFile(filepath.Join(input.Lane.Worktree, proof.Key+".txt"))
		if err != nil || string(data) != proof.Key+"\n" {
			t.Fatal("aggregate output missing")
		}
	}
}

func workflowProcessInput(t *testing.T) TaskInput {
	t.Helper()

	data, err := io.ReadAll(io.LimitReader(os.Stdin, workflowstate.FileLimit+1))
	if err != nil {
		t.Fatal(err)
	}

	var input TaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}

	frozen, err := os.ReadFile(os.Getenv("CHR_WORKFLOW_INPUT"))
	if err != nil || string(frozen) != string(data) {
		t.Fatal("input file and stdin differ")
	}

	cwd, err := os.Getwd()
	if err != nil || cwd != input.Lane.Worktree || validateWorkflowTaskHarnessSchema(input) != nil {
		t.Fatal("worker received a mutable or shared checkout")
	}

	if _, err := os.Stat("/proc/self/fd/3"); err != nil {
		t.Fatal("worker has no inherited capacity reservation")
	}

	return input
}

func workflowProcessWrite(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func workflowProcessCommit(t *testing.T, dir string) {
	t.Helper()

	for _, argv := range [][]string{{"add", "."}, {"-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "Worker result"}} {
		output, err := exec.Command("git", append([]string{"-C", dir}, argv...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git: %v: %s", err, output)
		}
	}
}
