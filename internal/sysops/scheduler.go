package sysops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ScheduledTaskInfo holds info about a scheduled task.
type ScheduledTaskInfo struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Enabled  bool   `json:"enabled"`
	NextRun  string `json:"next_run"`
}

// GetScheduledTasks returns all scheduled tasks (cron on Linux, schtasks on Windows).
func GetScheduledTasks() ([]ScheduledTaskInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return getCronTasks()
	case "windows":
		return getSchtasks()
	default:
		return nil, nil
	}
}

func getCronTasks() ([]ScheduledTaskInfo, error) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var tasks []ScheduledTaskInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			schedule := strings.Join(parts[:5], " ")
			tasks = append(tasks, ScheduledTaskInfo{
				Name:     parts[5],
				Schedule: schedule,
				Command:  strings.Join(parts[5:], " "),
				Enabled:  true,
			})
		}
	}
	return tasks, nil
}

func getSchtasks() ([]ScheduledTaskInfo, error) {
	cmd := common.HiddenCommand("schtasks", "/query", "/fo", "LIST", "/v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var tasks []ScheduledTaskInfo
	blocks := strings.Split(string(output), "\n\n")
	for _, block := range blocks {
		if block == "" {
			continue
		}
		task := ScheduledTaskInfo{}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "TaskName:") {
				task.Name = strings.TrimPrefix(line, "TaskName: ")
			} else if strings.HasPrefix(line, "Next Run Time:") {
				task.NextRun = strings.TrimPrefix(line, "Next Run Time: ")
			} else if strings.HasPrefix(line, "Status:") {
				status := strings.TrimPrefix(line, "Status: ")
				task.Enabled = status == "Ready" || status == "Running"
			} else if strings.HasPrefix(line, "Task To Run:") {
				task.Command = strings.TrimPrefix(line, "Task To Run: ")
			}
		}
		if task.Name != "" {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

// RunScheduledTask executes a scheduled task by name (Windows only).
func RunScheduledTask(name string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("RunScheduledTask is only supported on Windows")
	}
	cmd := common.HiddenCommand("schtasks", "/run", "/tn", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run task %s: %v (output: %s)", name, err, string(output))
	}
	return nil
}
