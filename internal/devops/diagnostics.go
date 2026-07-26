package devops

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

type DiagnosticCheck struct {
	Name    string
	Status  string
	Message string
	Value   string
}

type DiagnosticResult struct {
	Checks    []DiagnosticCheck
	Score     int
	Timestamp string
}

func RunDevOpsDiagnostics() DiagnosticResult {
	var checks []DiagnosticCheck
	passed := 0
	total := 0

	gitCheck := func() DiagnosticCheck {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cmd := common.HiddenCommandContext(ctx, "git", "status")
		if err := cmd.Run(); err != nil {
			return DiagnosticCheck{Name: "Git", Status: "fail", Message: "Git not available or not in a repo", Value: "not-found"}
		}
		return DiagnosticCheck{Name: "Git", Status: "pass", Message: "Git available", Value: "ok"}
	}
	c := gitCheck()
	total++
	if c.Status == "pass" {
		passed++
	}
	checks = append(checks, c)

	dockerCheck := func() DiagnosticCheck {
		if _, err := exec.LookPath("docker"); err != nil {
			return DiagnosticCheck{Name: "Docker CLI", Status: "fail", Message: "Docker binary not found in PATH", Value: "not-found"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := common.HiddenCommandContext(ctx, "docker", "info")
		if err := cmd.Run(); err != nil {
			return DiagnosticCheck{Name: "Docker Daemon", Status: "fail", Message: "Docker installed but daemon not running", Value: "stopped"}
		}
		return DiagnosticCheck{Name: "Docker Daemon", Status: "pass", Message: "Docker daemon running", Value: "running"}
	}
	c = dockerCheck()
	total++
	if c.Status == "pass" {
		passed++
	}
	checks = append(checks, c)

	k8sCheck := func() DiagnosticCheck {
		if _, err := exec.LookPath("kubectl"); err != nil {
			return DiagnosticCheck{Name: "kubectl CLI", Status: "fail", Message: "kubectl binary not found", Value: "not-found"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := common.HiddenCommandContext(ctx, "kubectl", "cluster-info")
		if err := cmd.Run(); err != nil {
			return DiagnosticCheck{Name: "Kubernetes Cluster", Status: "warn", Message: "kubectl installed but cannot connect to cluster", Value: "disconnected"}
		}
		return DiagnosticCheck{Name: "Kubernetes Cluster", Status: "pass", Message: "Connected to cluster", Value: "connected"}
	}
	c = k8sCheck()
	total++
	if c.Status == "pass" {
		passed++
	}
	checks = append(checks, c)

	nodeCheck := func() DiagnosticCheck {
		if _, err := exec.LookPath("node"); err != nil {
			return DiagnosticCheck{Name: "Node.js", Status: "fail", Message: "Node.js not found", Value: "not-found"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cmd := common.HiddenCommandContext(ctx, "node", "--version")
		output, err := cmd.Output()
		if err != nil {
			return DiagnosticCheck{Name: "Node.js", Status: "fail", Message: "Node.js found but version check failed", Value: "error"}
		}
		return DiagnosticCheck{Name: "Node.js", Status: "pass", Message: "Node.js available", Value: strings.TrimSpace(string(output))}
	}
	c = nodeCheck()
	total++
	if c.Status == "pass" {
		passed++
	}
	checks = append(checks, c)

	goCheck := func() DiagnosticCheck {
		if _, err := exec.LookPath("go"); err != nil {
			return DiagnosticCheck{Name: "Go", Status: "fail", Message: "Go not found", Value: "not-found"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cmd := common.HiddenCommandContext(ctx, "go", "version")
		output, err := cmd.Output()
		if err != nil {
			return DiagnosticCheck{Name: "Go", Status: "fail", Message: "Go found but version check failed", Value: "error"}
		}
		return DiagnosticCheck{Name: "Go", Status: "pass", Message: "Go available", Value: strings.TrimSpace(string(output))}
	}
	c = goCheck()
	total++
	if c.Status == "pass" {
		passed++
	}
	checks = append(checks, c)

	depCheck := func() DiagnosticCheck {
		modCheck, _ := exec.LookPath("go")
		if modCheck != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := common.HiddenCommandContext(ctx, "go", "env", "GOMODCACHE")
			if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
				return DiagnosticCheck{Name: "Module Cache", Status: "pass", Message: "Go module cache configured", Value: "ok"}
			}
		}
		return DiagnosticCheck{Name: "Module Cache", Status: "warn", Message: "No dependency cache detected", Value: "not-configured"}
	}
	c = depCheck()
	total++
	if c.Status == "pass" {
		passed++
	}
	checks = append(checks, c)

	score := 0
	if total > 0 {
		score = (passed * 100) / total
	}

	return DiagnosticResult{
		Checks:    checks,
		Score:     score,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
