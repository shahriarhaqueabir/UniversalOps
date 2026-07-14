package devops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// KubernetesStatus holds kubectl and cluster connectivity info.
type KubernetesStatus struct {
	Installed bool
	Connected bool
	Cluster   string
	Nodes     int
	Pods      int
}

// GetKubernetesStatus checks kubectl availability and cluster connectivity.
func GetKubernetesStatus() (KubernetesStatus, error) {
	var status KubernetesStatus

	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return status, nil // Not installed
	}
	status.Installed = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clusterCmd := exec.CommandContext(ctx, kubectlPath, "cluster-info")
	output, err := clusterCmd.CombinedOutput()
	if err == nil {
		status.Connected = true
		for _, line := range strings.Split(string(output), "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "control plane") || strings.Contains(line, "Kubernetes master") {
				if idx := strings.Index(line, "http"); idx >= 0 {
					url := line[idx:]
					if start := strings.Index(url, "//"); start >= 0 {
						host := url[start+2:]
						if end := strings.Index(host, ":"); end >= 0 {
							host = host[:end]
						}
						if end := strings.Index(host, "/"); end >= 0 {
							host = host[:end]
						}
						status.Cluster = host
					}
				}
				break
			}
		}
	}

	if !status.Connected {
		return status, nil
	}

	// Count nodes
	nodeCmd := exec.CommandContext(ctx, kubectlPath, "get", "nodes", "--no-headers")
	if nodeOut, err := nodeCmd.Output(); err == nil {
		for _, line := range strings.Split(string(nodeOut), "\n") {
			if strings.TrimSpace(line) != "" {
				status.Nodes++
			}
		}
	}

	// Count pods across all namespaces
	podCmd := exec.CommandContext(ctx, kubectlPath, "get", "pods", "--all-namespaces", "--no-headers")
	if podOut, err := podCmd.Output(); err == nil {
		for _, line := range strings.Split(string(podOut), "\n") {
			if strings.TrimSpace(line) != "" {
				status.Pods++
			}
		}
	}

	return status, nil
}

// GetK8sResources retrieves basic info about resources in a namespace.
func GetK8sResources(namespace string, resourceType string) (string, error) {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl binary not found")
	}

	args := []string{"get", resourceType}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(output))
	}
	return string(output), nil
}
