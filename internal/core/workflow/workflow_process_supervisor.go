package workflow

import (
	"context"
	"math"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// The supervisor waits for permission after its birth identity is journaled.
// It keeps fd 3 even when a worker closes inherited descriptors, and waits for
// non-zombie group members, including descendants orphaned by their parent.
const workflowSupervisorScript = `
IFS= read -r permit <&4 || exit 125
exec 4<&-
[ "$permit" = run ] || exit 125
budget=$1
shift
deadline=$((SECONDS + budget))
"$@" <&0 &
worker=$!
# The supervisor must survive the host closing its output pipes.
exec >/dev/null 2>&1
group_alive() {
 for file in /proc/[0-9]*/stat; do
  [ "$file" = /proc/$$/stat ] && continue
  IFS= read -r stat 2>/dev/null <"$file" || continue
  fields=${stat##*) }
  set -- $fields
  [ "$3" = "$$" ] && [ "$1" != Z ] && [ "$1" != X ] && return 0
 done
 return 1
}
while group_alive; do
 [ "$SECONDS" -ge "$deadline" ] && kill -KILL -- -$$
 sleep 0.05
done
wait "$worker"
`

func workflowSupervisedCommand(ctx context.Context, claim *Claim, prepared workflowPreparedTask, command []string, gate *os.File) (*exec.Cmd, *workflowCommandGuard) {
	remaining := max(1, int(math.Ceil(time.Until(prepared.leaseUntil).Seconds())))
	args := []string{"-c", workflowSupervisorScript, "chr-workflow", strconv.Itoa(remaining)}
	args = append(args, command...)
	process := exec.CommandContext(ctx, "bash", args...)
	process.Dir = prepared.input.Lane.Worktree

	process.Env = append(os.Environ(), "CHR_WORKFLOW_INPUT="+filepathForWorkflowInput(claim, prepared))
	if prepared.environment != nil {
		process.Env = prepared.environment
	}

	if prepared.directory != "" {
		process.Dir = prepared.directory
	}

	process.ExtraFiles = []*os.File{claim.reservation, gate, claim.capacityGuard}
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	process.WaitDelay = time.Second
	guard := &workflowCommandGuard{}
	process.Cancel = func() error { return guard.kill(process) }

	return process, guard
}
