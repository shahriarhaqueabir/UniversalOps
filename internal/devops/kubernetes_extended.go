package devops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type K8sResourceItem struct {
	Name      string
	Namespace string
	Status    string
	Age       string
	Details   string
}

type K8sRolloutStatus struct {
	Name      string
	Kind      string
	Ready     bool
	Replicas  string
	Updated   string
	Available string
}

type K8sEvent struct {
	LastSeen string
	Type     string
	Reason   string
	Object   string
	Message  string
}

type K8sNamespaceInfo struct {
	Name   string
	Status string
	Age    string
}

type K8sScalingResult struct {
	Current int
	Desired int
	Success bool
	Output  string
}

func kubectlExec(args ...string) (string, error) {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl binary not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func kubectlExecLongTimeout(timeout time.Duration, args ...string) (string, error) {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl binary not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func GetK8sNamespaces() ([]K8sNamespaceInfo, error) {
	output, err := kubectlExec("get", "namespaces", "--no-headers")
	if err != nil {
		return nil, err
	}
	var namespaces []K8sNamespaceInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			namespaces = append(namespaces, K8sNamespaceInfo{
				Name:   fields[0],
				Status: fields[1],
				Age:    fields[2],
			})
		}
	}
	return namespaces, nil
}

func GetK8sDeployments(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "deployments"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			ns := fields[0]
			if namespace != "" {
				ns = namespace
			}
			items = append(items, K8sResourceItem{
				Name:      fields[0],
				Namespace: ns,
				Status:    fields[2],
				Age:       fields[len(fields)-1],
				Details:   fmt.Sprintf("%s ready", fields[1]),
			})
		}
	}
	return items, nil
}

func GetK8sServices(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "services"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			items = append(items, K8sResourceItem{
				Name:      fields[1],
				Namespace: fields[0],
				Status:    fields[4],
				Age:       fields[len(fields)-1],
				Details:   fmt.Sprintf("%s:%s/%s", fields[2], fields[4], fields[3]),
			})
		}
	}
	return items, nil
}

func GetK8sPods(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "pods"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			ns := fields[0]
			items = append(items, K8sResourceItem{
				Name:      fields[1],
				Namespace: ns,
				Status:    fields[2],
				Age:       fields[len(fields)-1],
				Details:   fmt.Sprintf("%d/%d", strings.Count(fields[1], "/")+1, 1),
			})
		}
	}
	return items, nil
}

func GetK8sRollouts(namespace string) ([]K8sRolloutStatus, error) {
	args := []string{"rollout", "status"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, "deployments", "--timeout=5s")
	output, err := kubectlExec(args...)
	if err != nil {
		output = fmt.Sprintf("rollout status check completed: %v", err)
	}

	args2 := []string{"get", "deployments"}
	if namespace != "" {
		args2 = append(args2, "-n", namespace)
	} else {
		args2 = append(args2, "--all-namespaces")
	}
	args2 = append(args2, "--no-headers")
	deployments, _ := kubectlExec(args2...)

	var rollouts []K8sRolloutStatus
	for _, line := range strings.Split(deployments, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			rollouts = append(rollouts, K8sRolloutStatus{
				Name:      fields[0],
				Kind:      "Deployment",
				Ready:     strings.Contains(output, "successfully"),
				Replicas:  fields[1],
				Updated:   fields[2],
				Available: fields[3],
			})
		}
	}
	return rollouts, nil
}

func K8sRestartDeployment(name string, namespace string) (string, error) {
	args := []string{"rollout", "restart", "deployment", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return kubectlExecLongTimeout(30*time.Second, args...)
}

func K8sRollbackDeployment(name string, namespace string, revision int) (string, error) {
	args := []string{"rollout", "undo", "deployment", name}
	if revision > 0 {
		args = append(args, fmt.Sprintf("--to-revision=%d", revision))
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return kubectlExecLongTimeout(30*time.Second, args...)
}

func K8sScaleDeployment(name string, namespace string, replicas int) (K8sScalingResult, error) {
	args := []string{"scale", "deployment", name, fmt.Sprintf("--replicas=%d", replicas)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	output, err := kubectlExecLongTimeout(30*time.Second, args...)
	result := K8sScalingResult{Desired: replicas, Success: err == nil, Output: output}
	if err == nil {
		args2 := []string{"get", "deployment", name, "-o", "jsonpath={.status.readyReplicas}"}
		if namespace != "" {
			args2 = append(args2, "-n", namespace)
		}
		readyStr, _ := kubectlExec(args2...)
		fmt.Sscanf(readyStr, "%d", &result.Current)
	}
	return result, err
}

func GetK8sEvents(namespace string, limit int) ([]K8sEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	args := []string{"get", "events"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, "--no-headers", "--sort-by=.lastTimestamp")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var events []K8sEvent
	lines := strings.Split(output, "\n")
	start := 0
	if len(lines) > limit {
		start = len(lines) - limit
	}
	for i := start; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			events = append(events, K8sEvent{
				LastSeen: fields[0],
				Type:     fields[1],
				Reason:   fields[2],
				Object:   fields[3],
				Message:  strings.Join(fields[4:], " "),
			})
		}
	}
	return events, nil
}

func K8sDescribeResource(kind string, name string, namespace string) (string, error) {
	args := []string{"describe", kind, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	return kubectlExec(args...)
}

func K8sGetLogs(podName string, namespace string, container string, tail int) (string, error) {
	if tail <= 0 {
		tail = 50
	}
	args := []string{"logs", podName, fmt.Sprintf("--tail=%d", tail)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if container != "" {
		args = append(args, "-c", container)
	}
	return kubectlExec(args...)
}

func K8sPortForward(podName string, namespace string, localPort string, podPort string) (string, error) {
	args := []string{"port-forward", "pod/" + podName, fmt.Sprintf("%s:%s", localPort, podPort)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl binary not found")
	}
	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	output, _ := cmd.CombinedOutput()
	return string(output), nil
}

func K8sGetConfigMaps(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "configmaps"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			items = append(items, K8sResourceItem{
				Name:      fields[1],
				Namespace: fields[0],
				Age:       fields[len(fields)-1],
			})
		}
	}
	return items, nil
}

func K8sGetSecrets(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "secrets"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			items = append(items, K8sResourceItem{
				Name:      fields[1],
				Namespace: fields[0],
				Status:    fields[2],
				Age:       fields[len(fields)-1],
			})
		}
	}
	return items, nil
}

func K8sGetIngresses(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "ingresses"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			items = append(items, K8sResourceItem{
				Name:      fields[1],
				Namespace: fields[0],
				Status:    fields[2],
				Age:       fields[len(fields)-1],
			})
		}
	}
	return items, nil
}

func K8sGetJobs(namespace string) ([]K8sResourceItem, error) {
	args := []string{"get", "jobs"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	args = append(args, "--no-headers")
	output, err := kubectlExec(args...)
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			items = append(items, K8sResourceItem{
				Name:      fields[1],
				Namespace: fields[0],
				Status:    fields[2],
				Age:       fields[len(fields)-1],
				Details:   fmt.Sprintf("Completions: %s", fields[2]),
			})
		}
	}
	return items, nil
}

func K8sGetNodes() ([]K8sResourceItem, error) {
	output, err := kubectlExec("get", "nodes", "--no-headers")
	if err != nil {
		return nil, err
	}
	var items []K8sResourceItem
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			items = append(items, K8sResourceItem{
				Name:   fields[0],
				Status: fields[1],
				Age:    fields[len(fields)-1],
			})
		}
	}
	return items, nil
}
