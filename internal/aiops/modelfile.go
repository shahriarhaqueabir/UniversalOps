package aiops

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ModelfileConfig represents the configuration extracted from a Modelfile.
type ModelfileConfig struct {
	From       string
	System     string
	Parameters map[string]any
}

// ParseModelfile reads an Ollama Modelfile and extracts key configurations.
func ParseModelfile(path string) (*ModelfileConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Modelfile: %w", err)
	}
	defer file.Close()

	config := &ModelfileConfig{
		Parameters: make(map[string]any),
	}

	scanner := bufio.NewScanner(file)
	var inSystemBlock bool
	var systemLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle SYSTEM block
		if inSystemBlock {
			if strings.HasSuffix(line, `"""`) {
				content := strings.TrimSuffix(line, `"""`)
				if content != "" {
					systemLines = append(systemLines, content)
				}
				inSystemBlock = false
				config.System = strings.Join(systemLines, "\n")
			} else {
				systemLines = append(systemLines, line)
			}
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		command := strings.ToUpper(parts[0])
		value := strings.Join(parts[1:], " ")

		switch command {
		case "FROM":
			config.From = value
		case "SYSTEM":
			if strings.HasPrefix(value, `"""`) {
				inSystemBlock = true
				content := strings.TrimPrefix(value, `"""`)
				if strings.HasSuffix(content, `"""`) {
					// Single line block: SYSTEM """text"""
					config.System = strings.TrimSuffix(content, `"""`)
					inSystemBlock = false
				} else if content != "" {
					systemLines = append(systemLines, content)
				}
			} else {
				config.System = strings.Trim(value, `"'`)
			}
		case "PARAMETER":
			if len(parts) >= 3 {
				key := parts[1]
				val := strings.Join(parts[2:], " ")
				val = strings.Trim(val, `"'`)

				if key == "stop" {
					if existing, ok := config.Parameters["stop"].([]string); ok {
						config.Parameters["stop"] = append(existing, val)
					} else {
						config.Parameters["stop"] = []string{val}
					}
				} else {
					// Try numeric conversion
					if f, err := strconv.ParseFloat(val, 64); err == nil {
						config.Parameters[key] = f
					} else {
						config.Parameters[key] = val
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Modelfile: %w", err)
	}

	return config, nil
}
