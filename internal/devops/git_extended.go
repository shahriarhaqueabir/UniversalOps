package devops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type GitBranchInfo struct {
	Name       string
	Current    bool
	Upstream   string
	Ahead      int
	Behind     int
	LastCommit string
}

type GitTagInfo struct {
	Name   string
	Commit string
	Date   string
	Msg    string
}

type GitStashEntry struct {
	Index   int
	Branch  string
	Message string
}

type GitRemoteInfo struct {
	Name string
	URL  string
	Type string
}

type GitBlameEntry struct {
	Commit  string
	Author  string
	Date    string
	LineNum int
	Content string
}

type GitCommitInfo struct {
	Hash    string
	Author  string
	Date    string
	Message string
}

func gitRunTimeout(dir string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), err
}

func gitRunStdoutStderr(dir string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func GetGitBranches(dir string) ([]GitBranchInfo, error) {
	output, err := gitRunTimeout(dir, 5*time.Second, "branch", "-vv")
	if err != nil {
		return nil, fmt.Errorf("git branch failed: %w", err)
	}
	if output == "" {
		return nil, fmt.Errorf("no branches or not a git repo")
	}
	var branches []GitBranchInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		current := strings.HasPrefix(line, "*")
		cleaned := strings.TrimLeft(line, "* ")
		parts := strings.Fields(cleaned)
		if len(parts) == 0 {
			continue
		}
		info := GitBranchInfo{Name: parts[0], Current: current}
		if len(parts) > 1 {
			tracking := parts[1]
			if strings.HasPrefix(tracking, "[") && strings.Contains(tracking, ":") {
				info.Upstream = strings.TrimPrefix(tracking, "[")
				info.Upstream = strings.Split(info.Upstream, "]")[0]
				if aheadBehind := strings.Split(info.Upstream, ":"); len(aheadBehind) == 2 {
					fmt.Sscanf(aheadBehind[0], "%d", &info.Ahead)
					fmt.Sscanf(strings.TrimRight(aheadBehind[1], "]"), "%d", &info.Behind)
					info.Upstream = strings.Split(info.Upstream, ":")[0]
				}
			}
		}
		lc, _ := gitRunTimeout(dir, 3*time.Second, "log", "-1", "--oneline", info.Name)
		info.LastCommit = lc
		branches = append(branches, info)
	}
	return branches, nil
}

func GetGitTags(dir string) ([]GitTagInfo, error) {
	output, err := gitRunTimeout(dir, 5*time.Second, "tag", "-l", "--sort=-creatordate")
	if err != nil {
		return nil, fmt.Errorf("git tag failed: %w", err)
	}
	if output == "" {
		return []GitTagInfo{}, nil
	}
	var tags []GitTagInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		details, _ := gitRunTimeout(dir, 3*time.Second, "log", "-1", "--format=%H|%ai|%s", line)
		parts := strings.SplitN(details, "|", 3)
		info := GitTagInfo{Name: line}
		if len(parts) >= 1 {
			info.Commit = parts[0]
		}
		if len(parts) >= 2 {
			info.Date = parts[1]
		}
		if len(parts) >= 3 {
			info.Msg = parts[2]
		}
		tags = append(tags, info)
	}
	return tags, nil
}

func GetGitStash(dir string) ([]GitStashEntry, error) {
	output, err := gitRunTimeout(dir, 5*time.Second, "stash", "list")
	if err != nil {
		return nil, fmt.Errorf("git stash failed: %w", err)
	}
	if output == "" {
		return []GitStashEntry{}, nil
	}
	var entries []GitStashEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var idx int
		var msg string
		if n, err := fmt.Sscanf(line, "stash@{%d}: %s", &idx, &msg); err == nil && n >= 1 {
			entries = append(entries, GitStashEntry{Index: idx, Message: strings.TrimPrefix(line, fmt.Sprintf("stash@{%d}: ", idx))})
		}
	}
	return entries, nil
}

func GetGitRemotes(dir string) ([]GitRemoteInfo, error) {
	output, err := gitRunTimeout(dir, 5*time.Second, "remote", "-v")
	if err != nil {
		return nil, fmt.Errorf("git remote failed: %w", err)
	}
	if output == "" {
		return []GitRemoteInfo{}, nil
	}
	var remotes []GitRemoteInfo
	seen := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remoteType := ""
			if len(parts) >= 3 {
				remoteType = strings.TrimRight(parts[2], ")")
				remoteType = strings.TrimLeft(remoteType, "(")
			}
			key := parts[0] + "|" + parts[1]
			if !seen[key] {
				seen[key] = true
				remotes = append(remotes, GitRemoteInfo{
					Name: parts[0],
					URL:  parts[1],
					Type: remoteType,
				})
			}
		}
	}
	return remotes, nil
}

func GetGitBlame(dir string, filePath string) ([]GitBlameEntry, error) {
	stdout, _, _ := gitRunStdoutStderr(dir, "blame", "--porcelain", filePath)
	if stdout == "" {
		return nil, fmt.Errorf("could not blame file %s", filePath)
	}
	var entries []GitBlameEntry
	commits := make(map[string]struct{ author, date string })
	lineNum := 0
	lines := strings.Split(stdout, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		commit := parts[0]
		if commit == "" {
			continue
		}
		if len(commit) < 40 {
			continue
		}
		lineNumStr := parts[2]
		fmt.Sscanf(lineNumStr, "%d", &lineNum)

		for i+1 < len(lines) {
			next := lines[i+1]
			if strings.HasPrefix(next, "author ") {
				author := strings.TrimPrefix(next, "author ")
				c := commits[commit]
				c.author = author
				commits[commit] = c
			} else if strings.HasPrefix(next, "author-time ") {
				ts := strings.TrimPrefix(next, "author-time ")
				c := commits[commit]
				c.date = ts
				commits[commit] = c
			} else if strings.HasPrefix(next, "\t") {
				content := strings.TrimPrefix(next, "\t")
				c := commits[commit]
				entries = append(entries, GitBlameEntry{
					Commit:  commit,
					Author:  c.author,
					Date:    c.date,
					LineNum: lineNum,
					Content: content,
				})
				break
			}
			i++
		}
	}
	return entries, nil
}

func GetGitReflog(dir string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	output, err := gitRunTimeout(dir, 5*time.Second, "reflog", "-n", fmt.Sprintf("%d", limit), "--date=iso")
	if err != nil {
		return nil, fmt.Errorf("git reflog failed: %w", err)
	}
	if output == "" {
		return nil, fmt.Errorf("no reflog entries")
	}
	return strings.Split(output, "\n"), nil
}

func GitCheckout(dir string, branch string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "checkout", branch)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("checkout failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("checkout failed: %s", stderr)
	}
	return stdout, nil
}

func GitCreateBranch(dir string, branch string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "checkout", "-b", branch)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("branch creation failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("branch creation failed: %s", stderr)
	}
	return stdout, nil
}

func GitDeleteBranch(dir string, branch string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "branch", "-d", branch)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("branch deletion failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("branch deletion failed: %s", stderr)
	}
	return stdout, nil
}

func GitMerge(dir string, branch string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "merge", branch)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("merge failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("merge failed: %s", stderr)
	}
	return stdout, nil
}

func GitRebase(dir string, branch string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "rebase", branch)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("rebase failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("rebase failed: %s", stderr)
	}
	return stdout, nil
}

func GitStashPush(dir string, message string) (string, error) {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("stash failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("stash failed: %s", stderr)
	}
	return stdout, nil
}

func GitStashPop(dir string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "stash", "pop")
	if err != nil && stderr == "" {
		return "", fmt.Errorf("stash pop failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("stash pop failed: %s", stderr)
	}
	return stdout, nil
}

func GitStashDrop(dir string, index int) (string, error) {
	args := []string{"stash", "drop"}
	if index >= 0 {
		args = append(args, fmt.Sprintf("stash@{%d}", index))
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("stash drop failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("stash drop failed: %s", stderr)
	}
	return stdout, nil
}

func GitPush(dir string, remote string, branch string) (string, error) {
	args := []string{"push"}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("push failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("push failed: %s", stderr)
	}
	return stdout, nil
}

func GitPull(dir string, remote string, branch string) (string, error) {
	args := []string{"pull"}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("pull failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("pull failed: %s", stderr)
	}
	return stdout, nil
}

func GitFetch(dir string, remote string) (string, error) {
	args := []string{"fetch"}
	if remote != "" {
		args = append(args, remote)
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("fetch failed: %s", stderr)
	}
	return stdout, nil
}

func GitCommit(dir string, message string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "commit", "-m", message)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("commit failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("commit failed: %s", stderr)
	}
	return stdout, nil
}

func GitAdd(dir string, filespec string) (string, error) {
	if filespec == "" {
		filespec = "."
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, "add", filespec)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("add failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("add failed: %s", stderr)
	}
	return stdout, nil
}

func GitClean(dir string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "clean", "-fd")
	if err != nil && stderr == "" {
		return "", fmt.Errorf("clean failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("clean failed: %s", stderr)
	}
	return stdout, nil
}

func GitTagCreate(dir string, tag string, msg string) (string, error) {
	args := []string{"tag", "-a", tag, "-m", msg}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("tag creation failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("tag creation failed: %s", stderr)
	}
	return stdout, nil
}

func GitTagDelete(dir string, tag string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "tag", "-d", tag)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("tag deletion failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("tag deletion failed: %s", stderr)
	}
	return stdout, nil
}

func GitCherryPick(dir string, commits []string) (string, error) {
	args := append([]string{"cherry-pick"}, commits...)
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("cherry-pick failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("cherry-pick failed: %s", stderr)
	}
	return stdout, nil
}

func GitBisectStart(dir string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "bisect", "start")
	if err != nil && stderr == "" {
		return "", fmt.Errorf("bisect start failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("bisect start failed: %s", stderr)
	}
	return stdout, nil
}

func GitBisectGood(dir string, rev string) (string, error) {
	args := []string{"bisect", "good"}
	if rev != "" {
		args = append(args, rev)
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("bisect good failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("bisect good failed: %s", stderr)
	}
	return stdout, nil
}

func GitBisectBad(dir string, rev string) (string, error) {
	args := []string{"bisect", "bad"}
	if rev != "" {
		args = append(args, rev)
	}
	stdout, stderr, err := gitRunStdoutStderr(dir, args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("bisect bad failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("bisect bad failed: %s", stderr)
	}
	return stdout, nil
}

func GitBisectReset(dir string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "bisect", "reset")
	if err != nil && stderr == "" {
		return "", fmt.Errorf("bisect reset failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("bisect reset failed: %s", stderr)
	}
	return stdout, nil
}

func GitStatus(dir string) (string, error) {
	stdout, stderr, err := gitRunStdoutStderr(dir, "status")
	if err != nil && stderr == "" {
		return "", fmt.Errorf("status failed: %w", err)
	}
	if stderr != "" {
		return stderr, fmt.Errorf("status failed: %s", stderr)
	}
	return stdout, nil
}

func GitLogExtended(dir string, limit int, branch string) (string, error) {
	if limit <= 0 {
		limit = 10
	}
	args := []string{"log", "-n", fmt.Sprintf("%d", limit), "--oneline", "--graph", "--decorate", "--date=relative"}
	if branch != "" {
		args = append(args, branch)
	}
	output, err := gitRunTimeout(dir, 5*time.Second, args...)
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}
	if output == "" {
		return "", fmt.Errorf("no commits found")
	}
	return output, nil
}
