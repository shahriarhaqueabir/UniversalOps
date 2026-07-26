package devops

import (
	"fmt"
	"os"
	"sort"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shirou/gopsutil/v4/process"
)

// validatePID rejects PIDs that would be dangerous or nonsensical to target:
// non-positive values, PID 1 (init/systemd — killing it can take down the
// whole system on Linux), and the current process's own PID (this app).
// (SEC-3)
func validatePID(pid int32) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d: must be positive", pid)
	}
	if pid == 1 {
		return fmt.Errorf("refusing to target pid 1 (init)")
	}
	if int(pid) == os.Getpid() {
		return fmt.Errorf("refusing to target our own process (pid %d)", pid)
	}
	return nil
}

// ProcessEntry holds process information for the DevOps process manager.
type ProcessEntry struct {
	PID     int32
	Name    string
	CPU     float64
	Memory  float32
	Status  string
	Command string
}

// ListProcesses returns processes sorted by CPU usage.
func ListProcesses(limit int) ([]ProcessEntry, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	entries := make([]ProcessEntry, 0, len(procs))
	for _, proc := range procs {
		entry, ok := processEntryFromProcess(proc)
		if ok {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].CPU == entries[j].CPU {
			return entries[i].Memory > entries[j].Memory
		}
		return entries[i].CPU > entries[j].CPU
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

func processEntryFromProcess(proc *process.Process) (ProcessEntry, bool) {
	name, err := proc.Name()
	if err != nil || name == "" {
		return ProcessEntry{}, false
	}

	cpu, err := proc.CPUPercent()
	if err != nil {
		cpu = 0
	}

	mem := float32(0)
	if memInfo, err := proc.MemoryInfo(); err == nil && memInfo != nil {
		mem = float32(memInfo.RSS) / 1024 / 1024
	}

	status := ""
	if states, err := proc.Status(); err == nil && len(states) > 0 {
		status = states[0]
	}

	command := ""
	if cmdline, err := proc.Cmdline(); err == nil {
		command = cmdline
	}

	return ProcessEntry{
		PID:     proc.Pid,
		Name:    name,
		CPU:     cpu,
		Memory:  mem,
		Status:  status,
		Command: command,
	}, true
}

// KillProcess terminates a process by PID.
func KillProcess(pid int32) error {
	if err := validatePID(pid); err != nil {
		return err
	}
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("kill process %d: %w", pid, err)
	}
	return nil
}

// RestartProcess attempts to stop and relaunch a process by PID.
func RestartProcess(pid int32) error {
	if err := validatePID(pid); err != nil {
		return err
	}
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	exe, err := proc.Exe()
	if err != nil || exe == "" {
		return fmt.Errorf("resolve executable for process %d: %w", pid, err)
	}

	cmdline, _ := proc.CmdlineSlice()
	args := []string{}
	if len(cmdline) > 1 {
		args = cmdline[1:]
	}

	if err := KillProcess(pid); err != nil {
		return err
	}

	cmd := common.HiddenCommand(exe, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("restart process %d: %w", pid, err)
	}
	return nil
}
