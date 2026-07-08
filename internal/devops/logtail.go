package devops

import (
	"bufio"
	"os"
	"strings"
)

// LogEntry represents a single log line with metadata.
type LogEntry struct {
	Line      string
	Timestamp string
	Source    string
}

// TailLog reads the last n lines from a file.
func TailLog(path string, n int) ([]string, error) {
	if err := isPathSafe(path); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}

// SearchLog searches a file for lines containing the given pattern.
func SearchLog(path string, pattern string) ([]string, error) {
	if err := isPathSafe(path); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			matches = append(matches, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}
