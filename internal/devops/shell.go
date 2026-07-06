package devops

import (
	"bufio"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ShellResult holds the result of a shell command execution.
type ShellResult struct {
	Command  string
	Output   string
	ExitCode int
	Duration time.Duration
}

// RunCommand executes a shell command and returns the combined output.
// It uses cmd /c on Windows and sh -c on Unix.
func RunCommand(cmd string) (*ShellResult, error) {
	start := time.Now()

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = common.SandboxedCommand("cmd", "/c", cmd)
	} else {
		c = common.SandboxedCommand("sh", "-c", cmd)
	}

	output, err := c.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &ShellResult{
		Command:  cmd,
		Output:   string(output),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}

// RunCommandWithLiveOutput executes a command and streams each line of output
// through the provided channel. The channel is closed when the command finishes.
func RunCommandWithLiveOutput(cmd string, output chan string) (*ShellResult, error) {
	start := time.Now()
	var stdoutBuf, stderrBuf strings.Builder

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = common.SandboxedCommand("cmd", "/c", cmd)
	} else {
		c = common.SandboxedCommand("sh", "-c", cmd)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := c.Start(); err != nil {
		return nil, err
	}

	// Read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	// Read stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	err = c.Wait()
	close(output)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &ShellResult{
		Command:  cmd,
		Output:   stdoutBuf.String() + stderrBuf.String(),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}
