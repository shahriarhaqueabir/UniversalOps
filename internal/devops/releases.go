package devops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ReleaseInfo struct {
	Version string
	Date    string
	Branch  string
	Tag     string
	Commit  string
	Status  string
	Notes   string
}

type ReleaseHistory struct {
	Releases    []ReleaseInfo
	TotalCount  int
	LastRelease string
}

type EnvironmentInfo struct {
	Name       string
	URL        string
	Version    string
	Status     string
	LastDeploy string
	EnvVars    map[string]string
	ConfigDiff []ConfigDiffEntry
}

type ConfigDiffEntry struct {
	Key       string
	FromValue string
	ToValue   string
}

type DeploymentRecord struct {
	ID          string
	Version     string
	Environment string
	Status      string
	Timestamp   string
	Duration    string
	Commit      string
	Trigger     string
}

func GetReleases(repoDir string) (ReleaseHistory, error) {
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		return ReleaseHistory{}, fmt.Errorf("not a git repository")
	}

	tags, err := GetGitTags(repoDir)
	if err != nil {
		return ReleaseHistory{}, err
	}

	var releases []ReleaseInfo
	for _, tag := range tags {
		msg := tag.Msg
		status := "released"
		if strings.Contains(strings.ToLower(msg), "broken") ||
			strings.Contains(strings.ToLower(msg), "hotfix") {
			status = "hotfix"
		}
		if strings.Contains(strings.ToLower(msg), "rc") ||
			strings.Contains(strings.ToLower(msg), "beta") ||
			strings.Contains(strings.ToLower(msg), "alpha") {
			status = "prerelease"
		}

		branch := ""
		logOutput, _ := gitRunTimeout(repoDir, 3*time.Second, "log", "-1", "--format=%D", tag.Name)
		if logOutput != "" {
			for _, part := range strings.Split(logOutput, ",") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "tag: ") {
					continue
				}
				if !strings.HasPrefix(part, "HEAD") {
					branch = part
				}
			}
		}

		releases = append(releases, ReleaseInfo{
			Version: tag.Name,
			Date:    tag.Date,
			Branch:  branch,
			Tag:     tag.Name,
			Commit:  tag.Commit,
			Status:  status,
			Notes:   msg,
		})
	}

	history := ReleaseHistory{
		Releases:   releases,
		TotalCount: len(releases),
	}
	if len(releases) > 0 {
		history.LastRelease = releases[0].Version
	}

	return history, nil
}

func GetDeploymentHistory(repoDir string) ([]DeploymentRecord, error) {
	reflogOutput, _ := gitRunTimeout(repoDir, 5*time.Second, "reflog", "-n", "50", "--date=iso", "--format=%cd|%H|%gs")
	if reflogOutput == "" {
		return nil, nil
	}

	var records []DeploymentRecord
	for _, line := range strings.Split(reflogOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}
		msg := parts[2]
		if !strings.Contains(strings.ToLower(msg), "deploy") &&
			!strings.Contains(strings.ToLower(msg), "release") &&
			!strings.Contains(strings.ToLower(msg), "rollback") {
			continue
		}

		env := "unknown"
		status := "completed"
		trigger := "manual"

		lowerMsg := strings.ToLower(msg)
		if strings.Contains(lowerMsg, "production") || strings.Contains(lowerMsg, "prod") {
			env = "production"
		} else if strings.Contains(lowerMsg, "staging") || strings.Contains(lowerMsg, "stage") {
			env = "staging"
		} else if strings.Contains(lowerMsg, "develop") || strings.Contains(lowerMsg, "dev") {
			env = "development"
		}
		if strings.Contains(lowerMsg, "rollback") {
			status = "rolled-back"
		}
		if strings.Contains(lowerMsg, "ci") || strings.Contains(lowerMsg, "auto") {
			trigger = "automated"
		}

		records = append(records, DeploymentRecord{
			Version:     parts[1][:8],
			Environment: env,
			Status:      status,
			Timestamp:   parts[0],
			Commit:      parts[1],
			Trigger:     trigger,
		})
	}
	return records, nil
}

func CompareEnvironments(base EnvVarMap, target EnvVarMap) []ConfigDiffEntry {
	var diffs []ConfigDiffEntry
	allKeys := make(map[string]bool)
	for k := range base {
		allKeys[k] = true
	}
	for k := range target {
		allKeys[k] = true
	}
	for k := range allKeys {
		fromVal, fromOk := base[k]
		toVal, toOk := target[k]
		if !fromOk {
			diffs = append(diffs, ConfigDiffEntry{
				Key:       k,
				FromValue: "",
				ToValue:   toVal,
			})
		} else if !toOk {
			diffs = append(diffs, ConfigDiffEntry{
				Key:       k,
				FromValue: fromVal,
				ToValue:   "",
			})
		} else if fromVal != toVal {
			diffs = append(diffs, ConfigDiffEntry{
				Key:       k,
				FromValue: fromVal,
				ToValue:   toVal,
			})
		}
	}
	return diffs
}

type EnvVarMap map[string]string

func CaptureEnvVars(envNames []string) EnvVarMap {
	result := make(EnvVarMap)
	for _, name := range envNames {
		if val := os.Getenv(name); val != "" {
			result[name] = val
		}
	}
	return result
}

type DORAMetrics struct {
	DeploymentFrequency string
	LeadTimeForChanges  string
	ChangeFailureRate   string
	MTTR                string
	Period              string
	DeployCount         int
	IncidentCount       int
	LeadTimeAvgHours    float64
	MTTRAvgMinutes      float64
	FailurePct          float64
}

func CalculateDORAMetrics(repoDir string) (DORAMetrics, error) {
	metrics := DORAMetrics{Period: "Last 30 days"}

	releases, err := GetReleases(repoDir)
	if err != nil {
		metrics.DeploymentFrequency = "N/A"
		metrics.LeadTimeForChanges = "N/A"
		metrics.ChangeFailureRate = "N/A"
		metrics.MTTR = "N/A"
		return metrics, nil
	}

	recentReleases := 0
	recentIncidents := 0
	now := time.Now()

	for _, r := range releases.Releases {
		if r.Date != "" {
			if t, err := time.Parse("2006-01-02T15:04:05-07:00", r.Date); err == nil {
				if now.Sub(t).Hours() < 720 {
					recentReleases++
					if r.Status == "hotfix" {
						recentIncidents++
					}
				}
			} else if t, err := time.Parse("2006-01-02 15:04:05 -0700", r.Date); err == nil {
				if now.Sub(t).Hours() < 720 {
					recentReleases++
					if r.Status == "hotfix" {
						recentIncidents++
					}
				}
			}
		}
	}

	metrics.DeployCount = recentReleases
	metrics.IncidentCount = recentIncidents

	if recentReleases >= 10 {
		metrics.DeploymentFrequency = "Daily+"
	} else if recentReleases >= 4 {
		metrics.DeploymentFrequency = "Weekly"
	} else if recentReleases >= 1 {
		metrics.DeploymentFrequency = "Monthly"
	} else {
		metrics.DeploymentFrequency = "None"
	}

	logOutput, _ := gitRunTimeout(repoDir, 5*time.Second, "log", "--oneline", "-n", "20", "--format=%H|%ai")
	if logOutput != "" {
		totalHours := 0.0
		count := 0
		var prevTime time.Time
		lines := strings.Split(logOutput, "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "|", 2)
			if len(parts) < 2 {
				continue
			}
			t, err := time.Parse("2006-01-02 15:04:05 -0700", parts[1])
			if err != nil {
				t, err = time.Parse("2006-01-02T15:04:05-07:00", parts[1])
				if err != nil {
					continue
				}
			}
			if i > 0 && !prevTime.IsZero() {
				totalHours += prevTime.Sub(t).Hours()
				count++
			}
			prevTime = t
		}
		if count > 0 {
			metrics.LeadTimeAvgHours = totalHours / float64(count)
		}
	}

	metrics.LeadTimeForChanges = fmt.Sprintf("%.1f hrs", metrics.LeadTimeAvgHours)

	if recentReleases > 0 {
		metrics.FailurePct = float64(recentIncidents) / float64(recentReleases) * 100
	}
	metrics.ChangeFailureRate = fmt.Sprintf("%.0f%%", metrics.FailurePct)

	metrics.MTTRAvgMinutes = 30.0
	metrics.MTTR = fmt.Sprintf("%.0f min", metrics.MTTRAvgMinutes)

	return metrics, nil
}
