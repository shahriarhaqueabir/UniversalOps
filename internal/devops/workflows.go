package devops

import (
	"fmt"
	"strings"
)

// DevReport is a combined development operations report.
type DevReport struct {
	ShellResults []ShellResult
	LogContent   []string
	Processes    []ProcessEntry
	Services     []ServiceEntry
}

// RunDevDiagnostics runs common dev tasks and returns a combined report.
func RunDevDiagnostics() (*DevReport, error) {
	report := &DevReport{}
	var errs []string

	// Check Go
	goResult, err := RunCommand("go version")
	if err == nil && goResult.ExitCode == 0 {
		report.ShellResults = append(report.ShellResults, *goResult)
	} else {
		errs = append(errs, "Go not found")
	}

	// Check Git
	gitResult, err := RunCommand("git --version")
	if err == nil && gitResult.ExitCode == 0 {
		report.ShellResults = append(report.ShellResults, *gitResult)
	} else {
		errs = append(errs, "Git not found")
	}

	processes, err := ListProcesses(10)
	if err == nil {
		report.Processes = processes
	} else {
		errs = append(errs, fmt.Sprintf("Processes: %v", err))
	}

	services, err := ListServices(10)
	if err == nil {
		report.Services = services
	} else {
		errs = append(errs, fmt.Sprintf("Services: %v", err))
	}

	if len(report.ShellResults) == 0 && len(report.Processes) == 0 && len(report.Services) == 0 {
		return nil, fmt.Errorf("all dev diagnostics failed: %s", strings.Join(errs, "; "))
	}

	return report, nil
}

// String returns a plain-text summary of the dev report.
func (r *DevReport) String() string {
	var b strings.Builder

	b.WriteString("=== Development Operations Report ===\n\n")

	for _, result := range r.ShellResults {
		b.WriteString(fmt.Sprintf("$ %s\n", result.Command))
		b.WriteString(fmt.Sprintf("  exit=%d duration=%s\n", result.ExitCode, result.Duration))
		if result.Output != "" {
			b.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(result.Output)))
		}
		b.WriteString("\n")
	}

	if len(r.Processes) > 0 {
		b.WriteString(fmt.Sprintf("PROCESSES: %d sampled\n", len(r.Processes)))
	}

	if len(r.Services) > 0 {
		b.WriteString(fmt.Sprintf("SERVICES: %d sampled\n", len(r.Services)))
	}

	return b.String()
}

// Markdown returns a markdown-formatted dev report.
func (r *DevReport) Markdown() string {
	var b strings.Builder

	b.WriteString("# ⚙️ Development Operations Report\n\n")

	if len(r.ShellResults) > 0 {
		b.WriteString("## Tools & Environment\n\n")
		b.WriteString("| Command | Exit Code | Duration | Output |\n|---------|-----------|----------|--------|\n")
		for _, result := range r.ShellResults {
			output := strings.TrimSpace(result.Output)
			if len(output) > 60 {
				output = output[:57] + "..."
			}
			b.WriteString(fmt.Sprintf("| `%s` | %d | %s | %s |\n",
				result.Command, result.ExitCode, result.Duration, output))
		}
	}

	if len(r.Processes) > 0 {
		b.WriteString("\n## Processes\n\n")
		b.WriteString("| PID | Name | CPU | Memory |\n|-----|------|-----|--------|\n")
		for _, proc := range r.Processes {
			b.WriteString(fmt.Sprintf("| %d | %s | %.1f | %.1fM |\n",
				proc.PID, proc.Name, proc.CPU, proc.Memory))
		}
	}

	if len(r.Services) > 0 {
		b.WriteString("\n## Services\n\n")
		b.WriteString("| Name | Status | Start Type |\n|------|--------|------------|\n")
		for _, service := range r.Services {
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				service.Name, service.Status, service.StartType))
		}
	}

	return b.String()
}
