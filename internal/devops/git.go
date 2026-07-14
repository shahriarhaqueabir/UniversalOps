package devops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// GitRepoStatus holds internal status for a git repository.
type GitRepoStatus struct {
	Path           string
	Branch         string
	ModifiedFiles  int
	UntrackedFiles int
	Ahead          int
	Behind         int
	Clean          bool
}

// GitSummary holds aggregated git repository status.
type GitSummary struct {
	Repositories []GitRepoStatus
	TotalRepos   int
}

// GetGitSummary returns aggregated git repository status.
func GetGitSummary() (GitSummary, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return GitSummary{}, fmt.Errorf("git binary not found")
	}

	paths := FindGitRepos(10)
	repos := []GitRepoStatus{}

	for _, dir := range paths {
		status, err := GetRepoStatus(dir)
		if err != nil {
			common.LogWarn("Failed to get status for repo %s: %v", dir, err)
			continue
		}
		repos = append(repos, status)
	}

	return GitSummary{
		Repositories: repos,
		TotalRepos:   len(repos),
	}, nil
}

// GetRepoStatus returns detailed status for a single repository.
func GetRepoStatus(dir string) (GitRepoStatus, error) {
	branch := gitRun(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if branch == "" {
		return GitRepoStatus{}, fmt.Errorf("could not determine branch")
	}

	statusPorcelain := gitRun(dir, "status", "--porcelain")
	statusUntracked := gitRun(dir, "status", "--porcelain", "-u")

	modified := 0
	for _, line := range strings.Split(statusPorcelain, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "?") {
			modified++
		}
	}

	untracked := 0
	for _, line := range strings.Split(statusUntracked, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "?") {
			untracked++
		}
	}

	ahead, behind := 0, 0
	upstream := gitRun(dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	if upstream != "" {
		counts := gitRun(dir, "rev-list", "--left-right", "--count", "HEAD..."+upstream)
		if counts != "" {
			fmt.Sscanf(counts, "%d %d", &ahead, &behind)
		}
	}

	return GitRepoStatus{
		Path:           dir,
		Branch:         branch,
		ModifiedFiles:  modified,
		UntrackedFiles: untracked,
		Ahead:          ahead,
		Behind:         behind,
		Clean:          modified == 0 && untracked == 0,
	}, nil
}

// GetGitLog returns the recent commit log for a repository.
func GetGitLog(dir string, limit int) (string, error) {
	if limit <= 0 {
		limit = 10
	}
	output := gitRun(dir, "log", "-n", fmt.Sprintf("%d", limit), "--oneline", "--graph", "--decorate")
	if output == "" {
		return "", fmt.Errorf("failed to get git log or empty repo")
	}
	return output, nil
}

// GetGitDiff returns the current uncommitted changes.
func GetGitDiff(dir string) (string, error) {
	output := gitRun(dir, "diff", "HEAD")
	return output, nil
}

// FindGitRepos discovers git repositories in common locations.
func FindGitRepos(maxRepos int) []string {
	var candidates []string

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}

	// Home directory and common subdirectories
	home, err := os.UserHomeDir()
	if err == nil {
		candidates = append(candidates, home)
		for _, sub := range []string{"Documents", "Projects", "src", "dev", "code", "repos", "GitHub", "git"} {
			cand := filepath.Join(home, sub)
			if info, err := os.Stat(cand); err == nil && info.IsDir() {
				candidates = append(candidates, cand)
			}
		}
	}

	found := []string{}
	seen := make(map[string]bool)

	for _, dir := range candidates {
		// Check if dir itself is a git repo
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			if !seen[dir] {
				seen[dir] = true
				found = append(found, dir)
			}
			if len(found) >= maxRepos {
				return found
			}
		}

		// Scan one level deep for git repos
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
				if !seen[path] {
					seen[path] = true
					found = append(found, path)
				}
				if len(found) >= maxRepos {
					return found
				}
			}
		}
	}
	return found
}

// gitRun runs a git command in the given directory with a timeout.
func gitRun(dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	_ = cmd.Run()
	return strings.TrimSpace(stdout.String())
}
