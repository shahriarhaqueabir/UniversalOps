package devops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ContainerInfo holds information about a single container.
type ContainerInfo struct {
	ID     string
	Name   string
	Image  string
	State  string
	Status string
	Ports  string
}

// ContainerSummary holds Docker container status overview.
type ContainerSummary struct {
	Running    int
	Stopped    int
	Failed     int
	Total      int
	Containers []ContainerInfo
}

// DockerStatus holds Docker daemon and container overview.
type DockerStatus struct {
	Installed bool
	Running   bool
	Version   string
	Summary   ContainerSummary
}

// GetContainers returns Docker container status and summary.
func GetContainers() (ContainerSummary, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return ContainerSummary{}, fmt.Errorf("docker binary not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, dockerPath, "ps", "-a", "--format", `{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}`)
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ContainerSummary{}, err
	}

	var containers []ContainerInfo
	running, stopped, failed := 0, 0, 0

	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 6)
		if len(parts) < 5 {
			continue
		}
		c := ContainerInfo{
			ID:     parts[0],
			Name:   parts[1],
			Image:  parts[2],
			State:  parts[3],
			Status: parts[4],
		}
		if len(parts) == 6 {
			c.Ports = parts[5]
		}
		containers = append(containers, c)

		state := strings.ToLower(c.State)
		switch state {
		case "running":
			running++
		case "exited":
			stopped++
			if !strings.Contains(c.Status, "(0)") && strings.Contains(c.Status, "Exited") {
				failed++
			}
		default:
			stopped++
		}
	}

	return ContainerSummary{
		Running:    running,
		Stopped:    stopped,
		Failed:     failed,
		Total:      running + stopped,
		Containers: containers,
	}, nil
}

// GetDockerStatus checks Docker installation and daemon status.
func GetDockerStatus() (DockerStatus, error) {
	var status DockerStatus

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return status, nil // Not installed
	}
	status.Installed = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	verCmd := exec.CommandContext(ctx, dockerPath, "--version")
	if output, err := verCmd.Output(); err == nil {
		fields := strings.Fields(strings.TrimSpace(string(output)))
		if len(fields) >= 3 {
			status.Version = strings.TrimRight(fields[2], ",")
		}
	}

	infoCmd := exec.CommandContext(ctx, dockerPath, "info", "--format", `{{.ServerVersion}}`)
	if err := infoCmd.Run(); err == nil {
		status.Running = true
	}

	if summary, err := GetContainers(); err == nil {
		status.Summary = summary
	}

	return status, nil
}

// ControlContainer performs actions on a container.
func ControlContainer(id string, action string) error {
	if !common.ValidContainerID(id) {
		return fmt.Errorf("invalid container ID: %q", id)
	}
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker binary not found")
	}

	var cmdArg string
	switch strings.ToLower(action) {
	case "start":
		cmdArg = "start"
	case "stop":
		cmdArg = "stop"
	case "restart":
		cmdArg = "restart"
	case "remove":
		cmdArg = "rm"
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, dockerPath, cmdArg, id)
	return cmd.Run()
}
