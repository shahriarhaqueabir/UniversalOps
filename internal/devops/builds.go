package devops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type BuildSystemInfo struct {
	Name    string
	Version string
	Found   bool
	Path    string
}

type BuildTargetInfo struct {
	Name       string
	Type       string
	Path       string
	HasBuild   bool
	HasTest    bool
	HasLint    bool
	HasPackage bool
	HasClean   bool
	HasRun     bool
	DepCount   int
}

func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return strings.TrimSpace(text)
}

func DetectBuildSystems() []BuildSystemInfo {
	specs := []struct {
		name  string
		cmd   string
		args  []string
		parse func(string) string
	}{
		{"npm", "npm", []string{"--version"}, firstLine},
		{"pnpm", "pnpm", []string{"--version"}, firstLine},
		{"yarn", "yarn", []string{"--version"}, firstLine},
		{"Maven", "mvn", []string{"--version"}, func(s string) string {
			for _, line := range strings.Split(s, "\n") {
				if strings.Contains(line, "Apache Maven") {
					fields := strings.Fields(line)
					if len(fields) >= 3 {
						return fields[2]
					}
				}
			}
			return firstLine(s)
		}},
		{"Gradle", "gradle", []string{"--version"}, func(s string) string {
			for _, line := range strings.Split(s, "\n") {
				if strings.Contains(line, "Gradle") && !strings.Contains(line, "Groovy") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						return fields[1]
					}
				}
			}
			return firstLine(s)
		}},
		{"Go", "go", []string{"version"}, func(s string) string {
			fields := strings.Fields(firstLine(s))
			if len(fields) >= 3 {
				return fields[2]
			}
			return firstLine(s)
		}},
		{"Cargo", "cargo", []string{"--version"}, func(s string) string {
			fields := strings.Fields(firstLine(s))
			if len(fields) >= 2 {
				return fields[1]
			}
			return firstLine(s)
		}},
		{"Make", "make", []string{"--version"}, func(s string) string {
			for _, line := range strings.Split(s, "\n") {
				if strings.Contains(line, "GNU Make") {
					fields := strings.Fields(line)
					if len(fields) >= 3 {
						return fields[2]
					}
				}
			}
			return firstLine(s)
		}},
	}

	var results []BuildSystemInfo
	for _, spec := range specs {
		info := BuildSystemInfo{Name: spec.name}
		toolPath, err := exec.LookPath(spec.cmd)
		if err != nil {
			info.Found = false
			results = append(results, info)
			continue
		}
		info.Found = true
		info.Path = toolPath
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		cmd := exec.CommandContext(ctx, spec.cmd, spec.args...)
		var stdout strings.Builder
		cmd.Stdout = &stdout
		_ = cmd.Run()
		cancel()
		info.Version = spec.parse(stdout.String())
		results = append(results, info)
	}
	return results
}

func FindBuildTargets(rootDir string) ([]BuildTargetInfo, error) {
	var targets []BuildTargetInfo

	if rootDir == "" {
		rootDir = "."
	}

	info, err := os.Stat(rootDir)
	if err != nil || !info.IsDir() {
		return targets, fmt.Errorf("invalid directory: %s", rootDir)
	}

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return targets, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(rootDir, name)

		switch {
		case name == "package.json":
			target := BuildTargetInfo{Name: filepath.Base(rootDir), Type: "Node.js", Path: path}
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				target.HasBuild = strings.Contains(content, "\"build\"")
				target.HasTest = strings.Contains(content, "\"test\"")
				target.HasLint = strings.Contains(content, "\"lint\"")
				target.HasRun = strings.Contains(content, "\"start\"")
				target.DepCount = strings.Count(content, "\"dependencies\"") + strings.Count(content, "\"devDependencies\"")
			}
			targets = append(targets, target)

		case name == "pom.xml":
			target := BuildTargetInfo{Name: filepath.Base(rootDir), Type: "Maven", Path: path}
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				target.HasBuild = strings.Contains(content, "<build>")
				target.HasTest = strings.Contains(content, "<test>")
				target.HasPackage = strings.Contains(content, "<packaging>")
				target.DepCount = strings.Count(content, "<dependency>")
			}
			targets = append(targets, target)

		case name == "build.gradle" || name == "build.gradle.kts":
			target := BuildTargetInfo{Name: filepath.Base(rootDir), Type: "Gradle", Path: path}
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				target.HasBuild = strings.Contains(content, "build")
				target.HasTest = strings.Contains(content, "test")
				target.DepCount = strings.Count(content, "implementation") + strings.Count(content, "compile")
			}
			targets = append(targets, target)

		case name == "go.mod":
			target := BuildTargetInfo{Name: filepath.Base(rootDir), Type: "Go", Path: path}
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				target.HasBuild = true
				target.HasTest = true
				target.DepCount = strings.Count(content, "\t") - 1
				if target.DepCount < 0 {
					target.DepCount = 0
				}
			}
			targets = append(targets, target)

		case name == "Cargo.toml":
			target := BuildTargetInfo{Name: filepath.Base(rootDir), Type: "Cargo", Path: path}
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				target.HasBuild = true
				target.HasTest = true
				target.DepCount = strings.Count(content, "[dependencies]")
			}
			targets = append(targets, target)

		case name == "Makefile" || name == "makefile":
			target := BuildTargetInfo{Name: filepath.Base(rootDir), Type: "Make", Path: path}
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				target.HasBuild = strings.Contains(content, "build:")
				target.HasTest = strings.Contains(content, "test:")
				target.HasLint = strings.Contains(content, "lint:")
				target.HasClean = strings.Contains(content, "clean:")
				target.DepCount = strings.Count(content, ":")
			}
			targets = append(targets, target)
		}
	}

	return targets, nil
}

func RunBuildCommand(target BuildTargetInfo, action string) (string, error) {
	dir := filepath.Dir(target.Path)

	switch target.Type {
	case "Node.js":
		return runNpxCommand(dir, action)
	case "Maven":
		return runMavenCommand(dir, action)
	case "Gradle":
		return runGradleCommand(dir, action)
	case "Go":
		return runGoCommand(dir, action)
	case "Cargo":
		return runCargoCommand(dir, action)
	case "Make":
		return runMakeCommand(dir, action)
	default:
		return "", fmt.Errorf("unsupported build type: %s", target.Type)
	}
}

func runNpxCommand(dir string, action string) (string, error) {
	var args []string
	switch action {
	case "build":
		args = []string{"run", "build"}
	case "test":
		args = []string{"test"}
	case "lint":
		args = []string{"run", "lint"}
	case "clean":
		args = []string{"run", "clean"}
	case "start":
		args = []string{"start"}
	default:
		args = []string{"run", action}
	}
	return runCommandInDir(dir, "npm", args...)
}

func runMavenCommand(dir string, action string) (string, error) {
	var args []string
	switch action {
	case "build":
		args = []string{"clean", "compile"}
	case "test":
		args = []string{"test"}
	case "package":
		args = []string{"package"}
	case "clean":
		args = []string{"clean"}
	default:
		args = []string{action}
	}
	return runCommandInDir(dir, "mvn", args...)
}

func runGradleCommand(dir string, action string) (string, error) {
	var args []string
	switch action {
	case "build":
		args = []string{"build"}
	case "test":
		args = []string{"test"}
	case "clean":
		args = []string{"clean"}
	default:
		args = []string{action}
	}
	return runCommandInDir(dir, "gradle", args...)
}

func runGoCommand(dir string, action string) (string, error) {
	var args []string
	switch action {
	case "build":
		args = []string{"build", "./..."}
	case "test":
		args = []string{"test", "./..."}
	case "clean":
		args = []string{"clean"}
	case "vet":
		args = []string{"vet", "./..."}
	default:
		args = []string{action}
	}
	return runCommandInDir(dir, "go", args...)
}

func runCargoCommand(dir string, action string) (string, error) {
	var args []string
	switch action {
	case "build":
		args = []string{"build"}
	case "test":
		args = []string{"test"}
	case "clean":
		args = []string{"clean"}
	case "check":
		args = []string{"check"}
	default:
		args = []string{action}
	}
	return runCommandInDir(dir, "cargo", args...)
}

func runMakeCommand(dir string, action string) (string, error) {
	return runCommandInDir(dir, "make", action)
}

func runCommandInDir(dir string, command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
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
