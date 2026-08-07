package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// GitHubRelease represents a release object from the GitHub API.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// UpdateInfo holds the results of an update check.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseNotes   string `json:"release_notes"`
	DownloadURL    string `json:"download_url"`
	Size           int64  `json:"size"`
}

const githubLatestReleaseURL = "https://api.github.com/repos/shahriarhaqueabir/UniversalOps/releases/latest"

// CheckForUpdates queries the GitHub API for the latest release.
func CheckForUpdates() (*UpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", githubLatestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create update check request: %w", err)
	}

	// Identify as UniversalOps for GitHub API best practices
	req.Header.Set("User-Agent", "UniversalOps-Update-Checker")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode github release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	info := &UpdateInfo{
		CurrentVersion: Version,
		LatestVersion:  release.TagName,
		ReleaseNotes:   release.Body,
		Available:      isVersionNewer(current, latest),
	}

	// Find the correct asset for this OS/Arch
	// Asset naming pattern: universal-ops-1.6.0-windows-amd64.exe
	searchPattern := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(searchPattern)) {
			info.DownloadURL = asset.BrowserDownloadURL
			info.Size = asset.Size
			break
		}
	}

	return info, nil
}

// isVersionNewer returns true if latest is higher than current using simple semver rules.
func isVersionNewer(current, latest string) bool {
	if current == latest {
		return false
	}

	cParts := strings.Split(current, ".")
	lParts := strings.Split(latest, ".")

	for i := 0; i < len(cParts) && i < len(lParts); i++ {
		var c, l int
		fmt.Sscanf(cParts[i], "%d", &c)
		fmt.Sscanf(lParts[i], "%d", &l)

		if l > c {
			return true
		}
		if c > l {
			return false
		}
	}

	return len(lParts) > len(cParts)
}
