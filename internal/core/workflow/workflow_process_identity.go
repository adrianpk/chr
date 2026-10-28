package workflow

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowProcStat struct {
	state, birth string
	group        int
}

func workflowProcessIdentity(pid int) (workflowstate.ProcessIdentity, error) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return workflowstate.ProcessIdentity{}, err
	}

	stat, err := workflowReadProcStat(pid)
	if err != nil {
		return workflowstate.ProcessIdentity{}, err
	}

	return workflowstate.ProcessIdentity{PID: pid, BootID: strings.TrimSpace(string(boot)), StartTicks: stat.birth}, nil
}

func workflowReadProcStat(pid int) (workflowProcStat, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return workflowProcStat{}, err
	}

	tail := strings.LastIndex(string(data), ") ")
	if tail < 0 {
		return workflowProcStat{}, workflowstate.ErrConflict
	}

	fields := strings.Fields(string(data)[tail+2:])
	if len(fields) < 20 {
		return workflowProcStat{}, workflowstate.ErrConflict
	}

	group, err := strconv.Atoi(fields[2])

	return workflowProcStat{state: fields[0], birth: fields[19], group: group}, err
}

func workflowProcessAlive(identity workflowstate.ProcessIdentity) (bool, error) {
	observed, err := workflowProcessIdentity(identity.PID)
	if workflowProcessMissing(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if observed != identity {
		return false, nil
	}

	stat, err := workflowReadProcStat(identity.PID)
	if workflowProcessMissing(err) {
		return false, nil
	}

	return stat.state != "Z" && stat.state != "X", err
}

func workflowProcessGroupAlive(process workflowstate.Process) (bool, error) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(string(boot)) != process.Identity.BootID {
		return false, nil
	}

	return workflowGroupMembersAlive(process.GroupID)
}

func workflowGroupMembersAlive(groupID int) (bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		pid, parseErr := strconv.Atoi(entry.Name())
		if parseErr != nil {
			continue
		}

		stat, readErr := workflowReadProcStat(pid)
		if workflowProcessMissing(readErr) {
			continue
		}

		if readErr != nil {
			return false, readErr
		}

		if stat.group == groupID && stat.state != "Z" && stat.state != "X" {
			return true, nil
		}
	}

	return false, nil
}

// A process can disappear after its proc entry is opened. Linux then reports
// ESRCH instead of ENOENT; other observation failures must still block recovery.
func workflowProcessMissing(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH)
}

func workflowAttemptQuiescent(attempt Attempt) error {
	for _, process := range attempt.Processes {
		alive, err := workflowProcessGroupAlive(process)
		if err != nil {
			return err
		}

		if alive {
			return ErrWorkflowCapacity
		}
	}

	return nil
}
