package devops

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CICDStatus struct {
	Platform string
	Enabled  bool
	ConfigFound bool
	Pipelines   []CICDPipelineInfo
}

type CICDPipelineInfo struct {
	Name       string
	Status     string
	Branch     string
	Commit     string
	Duration   string
	UpdatedAt  string
	URL        string
}

type CICDConfig struct {
	Platform     string
	ConfigFiles  []string
	Detected     bool
}

func DetectCICDConfigs(rootDir string) []CICDConfig {
	configs := []CICDConfig{
		{
			Platform: "GitHub Actions",
			ConfigFiles: []string{
				".github/workflows",
				".github/workflows/main.yml",
				".github/workflows/ci.yml",
				".github/workflows/test.yml",
				".github/workflows/release.yml",
				".github/workflows/deploy.yml",
			},
		},
		{
			Platform: "GitLab CI",
			ConfigFiles: []string{".gitlab-ci.yml"},
		},
		{
			Platform: "Jenkins",
			ConfigFiles: []string{"Jenkinsfile", "Jenkinsfile.groovy"},
		},
		{
			Platform: "Azure DevOps",
			ConfigFiles: []string{"azure-pipelines.yml", ".azure-pipelines", "azure-pipelines.yaml"},
		},
		{
			Platform: "CircleCI",
			ConfigFiles: []string{".circleci/config.yml"},
		},
	}

	var results []CICDConfig
	for _, cfg := range configs {
		detected := false
		for _, cf := range cfg.ConfigFiles {
			cfPath := cf
			if !filepath.IsAbs(cf) {
				cfPath = filepath.Join(rootDir, cf)
			}
			if info, err := os.Stat(cfPath); err == nil {
				detected = true
				_ = info
				break
			}
		}
		cfg.Detected = detected
		results = append(results, cfg)
	}
	return results
}

func GetCICDStatus(rootDir string) CICDStatus {
	configs := DetectCICDConfigs(rootDir)

	combined := CICDStatus{Enabled: false, ConfigFound: false}

	for _, cfg := range configs {
		if cfg.Detected {
			combined.Enabled = true
			combined.ConfigFound = true
			combined.Platform = cfg.Platform
			break
		}
	}

	if combined.Enabled {
		pipelines := checkGithubActionsStatus(rootDir)
		if len(pipelines) == 0 {
			pipelines = checkGitLabCILocal(rootDir)
		}
		combined.Pipelines = pipelines
	}

	return combined
}

func checkGithubActionsStatus(rootDir string) []CICDPipelineInfo {
	ghDir := filepath.Join(rootDir, ".github", "workflows")
	info, err := os.Stat(ghDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	entries, err := os.ReadDir(ghDir)
	if err != nil {
		return nil
	}

	var pipelines []CICDPipelineInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		pipeline := CICDPipelineInfo{
			Name:   strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml"),
			Status: "detected",
		}

		data, err := os.ReadFile(filepath.Join(ghDir, name))
		if err == nil {
			content := string(data)
			for _, line := range strings.Split(content, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "on:") {
					pipeline.Branch = "various triggers"
				}
				if strings.HasPrefix(line, "branches:") || strings.Contains(line, "branches:") {
					brParts := strings.Split(line, ":")
					if len(brParts) >= 2 {
						pipeline.Branch = strings.TrimSpace(brParts[1])
					}
				}
			}
		}

		pipeline.Status = "configured"
		pipelines = append(pipelines, pipeline)
	}
	return pipelines
}

func checkGitLabCILocal(rootDir string) []CICDPipelineInfo {
	gitlabFile := filepath.Join(rootDir, ".gitlab-ci.yml")
	if _, err := os.Stat(gitlabFile); err != nil {
		return nil
	}

	data, err := os.ReadFile(gitlabFile)
	if err != nil {
		return nil
	}

	pipeline := CICDPipelineInfo{
		Name:   "GitLab CI Pipeline",
		Status: "configured",
	}

	content := string(data)
	stages := 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "stage:") || strings.HasPrefix(trimmed, "-") {
			stages++
		}
	}
	if stages > 0 {
		pipeline.Duration = fmt.Sprintf("%d stages", stages)
	}

	return []CICDPipelineInfo{pipeline}
}

func CheckGithubAPI(repoOwner string, repoName string, token string) ([]CICDPipelineInfo, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("no GitHub token available")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?per_page=10", repoOwner, repoName)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return nil, fmt.Errorf("GitHub API check requires token: %s", resp.Status)
}

func RunScript(scriptPath string) (string, error) {
	info, err := os.Stat(scriptPath)
	if err != nil {
		return "", fmt.Errorf("script not found: %s", scriptPath)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", scriptPath)
	}

	ext := strings.ToLower(filepath.Ext(scriptPath))
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	switch ext {
	case ".sh":
		cmd = exec.CommandContext(ctx, "sh", scriptPath)
	case ".ps1":
		shell := "pwsh"
		if _, err := exec.LookPath("pwsh"); err != nil {
			shell = "powershell"
		}
		cmd = exec.CommandContext(ctx, shell, "-NoProfile", "-File", scriptPath)
	case ".bat", ".cmd":
		cmd = exec.CommandContext(ctx, "cmd", "/c", scriptPath)
	case ".py":
		cmd = exec.CommandContext(ctx, "python", scriptPath)
	default:
		return "", fmt.Errorf("unsupported script extension: %s", ext)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	output := strings.TrimSpace(stdout.String())
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += strings.TrimSpace(stderr.String())
	}
	if err != nil {
		return output, err
	}
	return output, nil
}
