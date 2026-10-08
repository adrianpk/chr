package core

import (
	"os"
	"os/exec"
	"sync"
)

type workflowCommandGuard struct {
	sync.Mutex
	identity *WorkflowProcessIdentity
}

func (guard *workflowCommandGuard) remember(pid int) error {
	identity, err := workflowProcessIdentity(pid)
	if err != nil {
		return err
	}

	guard.Lock()
	guard.identity = &identity
	guard.Unlock()

	return nil
}

func (guard *workflowCommandGuard) kill(command *exec.Cmd) error {
	guard.Lock()
	identity := guard.identity
	guard.Unlock()
	// Before birth registration the child cannot pass its closed launch gate.
	if identity == nil {
		return os.ErrProcessDone
	}

	alive, err := workflowProcessAlive(*identity)
	if err != nil {
		return err
	}

	if !alive || command.Process == nil || command.Process.Pid != identity.PID {
		return os.ErrProcessDone
	}

	return workflowKillCommandGroup(command)
}
