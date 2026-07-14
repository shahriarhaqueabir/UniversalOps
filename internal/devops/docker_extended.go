package devops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type DockerStatsEntry struct {
	ContainerID   string
	Name          string
	CPUPercent    string
	MemoryUsage   string
	MemoryLimit   string
	MemoryPercent string
	NetIO         string
	BlockIO       string
	PIDCount      string
}

type DockerComposeInfo struct {
	Project      string
	Status       string
	WorkDir      string
	Services     []DockerComposeService
}

type DockerComposeService struct {
	Name    string
	State   string
	Ports   string
}

type DockerNetworkInfo struct {
	ID       string
	Name     string
	Driver   string
	Scope    string
	Subnet   string
	Gateway  string
	Containers int
}

type DockerVolumeInfo struct {
	Driver     string
	Name       string
	Mountpoint string
	Size       string
}

type DockerBuildResult struct {
	Success bool
	Output  string
	ImageID string
}

func dockerExec(args ...string) (string, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("docker binary not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, dockerPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func dockerExecLongTimeout(timeout time.Duration, args ...string) (string, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("docker binary not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, dockerPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func GetDockerStats() ([]DockerStatsEntry, error) {
	output, err := dockerExec("stats", "--no-stream", "--format", "{{.ID}}|{{.Name}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}|{{.PIDs}}")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	var entries []DockerStatsEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 8 {
			continue
		}
		memParts := strings.SplitN(parts[3], "/", 2)
		memUsage := strings.TrimSpace(memParts[0])
		memLimit := ""
		if len(memParts) > 1 {
			memLimit = strings.TrimSpace(memParts[1])
		}
		entries = append(entries, DockerStatsEntry{
			ContainerID:   parts[0],
			Name:          strings.TrimPrefix(parts[1], "/"),
			CPUPercent:    parts[2],
			MemoryUsage:   memUsage,
			MemoryLimit:   memLimit,
			MemoryPercent: parts[4],
			NetIO:         parts[5],
			BlockIO:       parts[6],
			PIDCount:      parts[7],
		})
	}
	return entries, nil
}

func GetDockerLogs(containerID string, tail int) (string, error) {
	if tail <= 0 {
		tail = 50
	}
	output, err := dockerExec("logs", "--tail", fmt.Sprintf("%d", tail), containerID)
	if err != nil {
		return "", err
	}
	return output, nil
}

func DockerExec(containerID string, cmdArgs []string) (string, error) {
	args := append([]string{"exec", containerID}, cmdArgs...)
	output, err := dockerExecLongTimeout(30*time.Second, args...)
	if err != nil {
		return "", err
	}
	return output, nil
}

func DockerPull(image string) (string, error) {
	return dockerExecLongTimeout(120*time.Second, "pull", image)
}

func DockerBuild(path string, tag string, dockerfile string) (DockerBuildResult, error) {
	args := []string{"build"}
	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}
	if tag != "" {
		args = append(args, "-t", tag)
	}
	args = append(args, path)
	output, err := dockerExecLongTimeout(300*time.Second, args...)
	result := DockerBuildResult{Output: output, Success: err == nil}
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Successfully built") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					result.ImageID = fields[len(fields)-1]
				}
			}
		}
	}
	return result, err
}

func DockerComposeList() ([]DockerComposeInfo, error) {
	output, err := dockerExec("compose", "ls", "--format", "{{.Name}}|{{.Status}}|{{.WorkingDirectory}}")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	var projects []DockerComposeInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			proj := DockerComposeInfo{
				Project: parts[0],
				Status:  parts[1],
			}
			if len(parts) >= 3 {
				proj.WorkDir = parts[2]
			}
			projects = append(projects, proj)
		}
	}
	return projects, nil
}

func DockerComposePS(projectDir string) ([]DockerComposeService, error) {
	args := []string{"compose"}
	if projectDir != "" {
		args = append(args, "--project-directory", projectDir)
	}
	args = append(args, "ps", "--format", "{{.Name}}|{{.State}}|{{.Ports}}")
	output, err := dockerExec(args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	var services []DockerComposeService
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		svc := DockerComposeService{Name: parts[0]}
		if len(parts) >= 2 {
			svc.State = parts[1]
		}
		if len(parts) >= 3 {
			svc.Ports = parts[2]
		}
		services = append(services, svc)
	}
	return services, nil
}

func DockerComposeUp(projectDir string, service string) (string, error) {
	args := []string{"compose"}
	if projectDir != "" {
		args = append(args, "--project-directory", projectDir)
	}
	args = append(args, "up", "-d")
	if service != "" {
		args = append(args, service)
	}
	return dockerExecLongTimeout(120*time.Second, args...)
}

func DockerComposeDown(projectDir string) (string, error) {
	args := []string{"compose"}
	if projectDir != "" {
		args = append(args, "--project-directory", projectDir)
	}
	args = append(args, "down")
	return dockerExecLongTimeout(60*time.Second, args...)
}

func DockerComposeLogs(projectDir string, service string, tail int) (string, error) {
	if tail <= 0 {
		tail = 50
	}
	args := []string{"compose"}
	if projectDir != "" {
		args = append(args, "--project-directory", projectDir)
	}
	args = append(args, "logs", "--tail", fmt.Sprintf("%d", tail))
	if service != "" {
		args = append(args, service)
	}
	return dockerExec(args...)
}

func GetDockerNetworks() ([]DockerNetworkInfo, error) {
	output, err := dockerExec("network", "ls", "--format", "{{.ID}}|{{.Name}}|{{.Driver}}|{{.Scope}}")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	var networks []DockerNetworkInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		netInfo := DockerNetworkInfo{
			ID:     parts[0],
			Name:   parts[1],
			Driver: parts[2],
			Scope:  parts[3],
		}
		detail, _ := dockerExec("network", "inspect", parts[1], "--format", "{{range .IPAM.Config}}{{.Subnet}}|{{.Gateway}}{{end}}")
		if subGateway := strings.SplitN(detail, "|", 2); len(subGateway) >= 1 {
			netInfo.Subnet = subGateway[0]
			if len(subGateway) >= 2 {
				netInfo.Gateway = subGateway[1]
			}
		}
		containerCount, _ := dockerExec("network", "inspect", parts[1], "--format", "{{len .Containers}}")
		fmt.Sscanf(containerCount, "%d", &netInfo.Containers)
		networks = append(networks, netInfo)
	}
	return networks, nil
}

func GetDockerVolumes() ([]DockerVolumeInfo, error) {
	output, err := dockerExec("volume", "ls", "--format", "{{.Driver}}|{{.Name}}|{{.Mountpoint}}")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	var volumes []DockerVolumeInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		v := DockerVolumeInfo{Driver: parts[0], Name: parts[1]}
		if len(parts) >= 3 {
			v.Mountpoint = parts[2]
		}
		size, _ := dockerExec("run", "--rm", "-v", parts[1]+":/vol", "alpine", "du", "-sh", "/vol")
		v.Size = strings.Fields(size)[0]
		volumes = append(volumes, v)
	}
	return volumes, nil
}

func DockerPrune() (string, error) {
	output, err := dockerExecLongTimeout(60*time.Second, "system", "prune", "-f")
	if err != nil {
		return "", err
	}
	return output, nil
}

func DockerKill(containerID string) (string, error) {
	return dockerExec("kill", containerID)
}

func DockerPause(containerID string) (string, error) {
	return dockerExec("pause", containerID)
}

func DockerUnpause(containerID string) (string, error) {
	return dockerExec("unpause", containerID)
}

func DockerRename(containerID string, newName string) (string, error) {
	return dockerExec("rename", containerID, newName)
}

func DockerContainerInspect(containerID string) (string, error) {
	return dockerExec("inspect", containerID)
}

func DockerImageList() (string, error) {
	output, err := dockerExec("images", "--format", "{{.Repository}}|{{.Tag}}|{{.ID}}|{{.Size}}|{{.CreatedSince}}")
	if err != nil {
		return "", err
	}
	return output, nil
}

func DockerRemoveImage(imageID string) (string, error) {
	return dockerExec("rmi", imageID)
}
