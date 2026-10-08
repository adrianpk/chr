package core

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WorkflowProcessIdentity binds a PID to one Linux boot and process birth.
type WorkflowProcessIdentity struct {
	PID        int    `json:"pid"`
	BootID     string `json:"boot_id"`
	StartTicks string `json:"start_ticks"`
}

// WorkflowProcess retains each supervised worker or validation command.
type WorkflowProcess struct {
	Identity   WorkflowProcessIdentity `json:"identity"`
	GroupID    int                     `json:"group_id"`
	Command    []string                `json:"command"`
	State      string                  `json:"state"`
	StdoutRef  string                  `json:"stdout_ref"`
	StderrRef  string                  `json:"stderr_ref"`
	StdoutHash string                  `json:"stdout_hash"`
	StderrHash string                  `json:"stderr_hash"`
	Result     *WorkflowCommandResult  `json:"result"`
}

type workflowProcStat struct {
	state, birth string
	group        int
}

func workflowProcessIdentity(pid int) (WorkflowProcessIdentity, error) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return WorkflowProcessIdentity{}, err
	}

	stat, err := workflowReadProcStat(pid)
	if err != nil {
		return WorkflowProcessIdentity{}, err
	}

	return WorkflowProcessIdentity{PID: pid, BootID: strings.TrimSpace(string(boot)), StartTicks: stat.birth}, nil
}

func workflowReadProcStat(pid int) (workflowProcStat, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return workflowProcStat{}, err
	}

	tail := strings.LastIndex(string(data), ") ")
	if tail < 0 {
		return workflowProcStat{}, ErrWorkflowConflict
	}

	fields := strings.Fields(string(data)[tail+2:])
	if len(fields) < 20 {
		return workflowProcStat{}, ErrWorkflowConflict
	}

	group, err := strconv.Atoi(fields[2])

	return workflowProcStat{state: fields[0], birth: fields[19], group: group}, err
}

func workflowProcessAlive(identity WorkflowProcessIdentity) (bool, error) {
	observed, err := workflowProcessIdentity(identity.PID)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if observed != identity {
		return false, nil
	}

	stat, err := workflowReadProcStat(identity.PID)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return stat.state != "Z" && stat.state != "X", err
}

func workflowProcessGroupAlive(process WorkflowProcess) (bool, error) {
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
		if errors.Is(readErr, os.ErrNotExist) {
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

func workflowAttemptQuiescent(attempt WorkflowAttempt) error {
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
