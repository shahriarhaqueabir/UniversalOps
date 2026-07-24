package sysops

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// ── Parser unit tests ────────────────────────────────────────────────────────
// Each parse* function is a pure function — no exec calls, trivially testable.

func TestParseAptOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"empty", "", 0},
		{"single install line", "curl\t\tinstall\n", 1},
		{"multiple install lines", "bash\t\tinstall\ngit\t\tinstall\nzsh\t\tinstall\n", 3},
		{"skips non-install lines", "vim\t\tinstall\ndpkg\t\tdeinstall\nnano\t\tpurge\n", 1},
		{"skips header noise", "Desired=Unknown/Install\nlinux-image\tinstall\n", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseAptOutput(tt.input)
			if len(pkgs) != tt.expect {
				t.Errorf("len = %d, want %d", len(pkgs), tt.expect)
			}
		})
	}

	t.Run("package names are preserved", func(t *testing.T) {
		pkgs := parseAptOutput("curl\t\tinstall\n")
		if len(pkgs) != 1 || pkgs[0].Name != "curl" {
			t.Errorf("got %+v, want [{Name:curl}]", pkgs)
		}
	})
}

func TestParseDnfOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"empty", "", 0},
		{"only header", "Installed Packages\n", 0},
		{"single package", "Installed Packages\ncurl.x86_64\t7.76.1\n", 1},
		{"multiple packages", "Installed Packages\ncurl.x86_64\t7.76.1\ngit.x86_64\t2.35.1\nvim.x86_64\t8.2\n", 3},
		{"skips blank lines", "Installed Packages\n\ncurl.x86_64\t7.76.1\n\n\n", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseDnfOutput(tt.input)
			if len(pkgs) != tt.expect {
				t.Errorf("len = %d, want %d", len(pkgs), tt.expect)
			}
		})
	}

	t.Run("parses name and version", func(t *testing.T) {
		pkgs := parseDnfOutput("Installed Packages\ncurl.x86_64\t7.76.1\n")
		if len(pkgs) != 1 || pkgs[0].Name != "curl.x86_64" || pkgs[0].Version != "7.76.1" {
			t.Errorf("got %+v", pkgs)
		}
	})
}

func TestParsePacmanOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"empty", "", 0},
		{"single package", "curl 7.76.1\n", 1},
		{"multiple packages", "curl 7.76.1\ngit 2.35.1\nvim 8.2\n", 3},
		{"skips blank lines", "curl 7.76.1\n\n\nvim 8.2\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parsePacmanOutput(tt.input)
			if len(pkgs) != tt.expect {
				t.Errorf("len = %d, want %d", len(pkgs), tt.expect)
			}
		})
	}

	t.Run("parses name and version", func(t *testing.T) {
		pkgs := parsePacmanOutput("curl 7.76.1\n")
		if len(pkgs) != 1 || pkgs[0].Name != "curl" || pkgs[0].Version != "7.76.1" {
			t.Errorf("got %+v", pkgs)
		}
	})
}

func TestParseWingetOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"empty", "", 0},
		{"header only", "Name Id Version\n--- -- -------\n\n", 0},
		{"single package", "Name                Id                Version\n--------------------------------------------------\ncurl                Curl.Curl         7.76.1\n", 1},
		{"multiple packages", "Name                Id                Version\n--------------------------------------------------\ncurl                Curl.Curl         7.76.1\ngit                 Git.Git           2.35.1\n", 2},
		{"skips short rows", "Name                Id                Version\n--------------------------------------------------\ncurl                Curl.Curl         7.76.1\nfoo\n", 1},
		{"handles preamble and source column", "\nName                         Id                         Version        Source\n--------------------------------------------------------------------------------\nMicrosoft Edge               Microsoft.Edge              138.0.3351.83 winget\n", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseWingetOutput(tt.input)
			if len(pkgs) != tt.expect {
				t.Errorf("len = %d, want %d", len(pkgs), tt.expect)
			}
		})
	}

	t.Run("parses name and version", func(t *testing.T) {
		pkgs := parseWingetOutput("Name                Id                Version\n--------------------------------------------------\ncurl                Curl.Curl         7.76.1\n")
		if len(pkgs) != 1 || pkgs[0].Name != "curl" || pkgs[0].Version != "7.76.1" {
			t.Errorf("got %+v", pkgs)
		}
	})

	t.Run("preserves names containing spaces", func(t *testing.T) {
		pkgs := parseWingetOutput("Name                         Id                         Version        Source\n--------------------------------------------------------------------------------\nMicrosoft Edge               Microsoft.Edge              138.0.3351.83 winget\n")
		if len(pkgs) != 1 || pkgs[0].Name != "Microsoft Edge" || pkgs[0].Version != "138.0.3351.83" {
			t.Errorf("got %+v", pkgs)
		}
	})
}

func TestParseChocoOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"empty", "", 0},
		{"header only", "Chocolatey v2.3.0\n", 0},
		{"single package", "Chocolatey v2.3.0\n---\ncurl 7.76.1\n", 1},
		{"multiple packages", "Chocolatey v2.3.0\n---\ncurl 7.76.1\ngit 2.35.1\nvim 8.2\n", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseChocoOutput(tt.input)
			if len(pkgs) != tt.expect {
				t.Errorf("len = %d, want %d", len(pkgs), tt.expect)
			}
		})
	}

	t.Run("parses name and version", func(t *testing.T) {
		pkgs := parseChocoOutput("Chocolatey v2.3.0\n---\ncurl 7.76.1\n")
		if len(pkgs) != 1 || pkgs[0].Name != "curl" || pkgs[0].Version != "7.76.1" {
			t.Errorf("got %+v", pkgs)
		}
	})
}

// ── Windows registry JSON parser ─────────────────────────────────────────────

func TestParseWindowsInstalledApps(t *testing.T) {
	t.Run("valid JSON array", func(t *testing.T) {
		json := `[{"DisplayName":"7-Zip","DisplayVersion":"22.01"},{"DisplayName":"Firefox","DisplayVersion":"120.0"}]`
		result := parseWindowsInstalledApps(json)
		if !result.Found {
			t.Fatal("expected Found=true")
		}
		if len(result.Packages) != 2 {
			t.Fatalf("got %d packages, want 2", len(result.Packages))
		}
		if result.Packages[0].Name != "7-Zip" || result.Packages[0].Version != "22.01" {
			t.Errorf("first package = %+v", result.Packages[0])
		}
		if result.Packages[1].Name != "Firefox" || result.Packages[1].Version != "120.0" {
			t.Errorf("second package = %+v", result.Packages[1])
		}
	})

	t.Run("single object wrapped in array", func(t *testing.T) {
		json := `[{"DisplayName":"Python","DisplayVersion":"3.12"}]`
		result := parseWindowsInstalledApps(json)
		if !result.Found {
			t.Fatal("expected Found=true")
		}
		if len(result.Packages) != 1 {
			t.Fatalf("got %d packages, want 1", len(result.Packages))
		}
		if result.Packages[0].Name != "Python" {
			t.Errorf("name = %q, want Python", result.Packages[0].Name)
		}
	})

	t.Run("single object without array wrapper", func(t *testing.T) {
		json := `{"DisplayName":"Python","DisplayVersion":"3.12"}`
		result := parseWindowsInstalledApps(json)
		if !result.Found {
			t.Fatal("expected Found=true")
		}
		if len(result.Packages) != 1 {
			t.Fatalf("got %d packages, want 1", len(result.Packages))
		}
	})

	t.Run("empty array", func(t *testing.T) {
		result := parseWindowsInstalledApps(`[]`)
		if !result.Found {
			t.Fatal("expected Found=true")
		}
		if len(result.Packages) != 0 {
			t.Errorf("got %d packages, want 0", len(result.Packages))
		}
	})

	t.Run("skips entries with empty DisplayName", func(t *testing.T) {
		json := `[{"DisplayName":"","DisplayVersion":"1.0"},{"DisplayName":"VLC","DisplayVersion":"3.0"}]`
		result := parseWindowsInstalledApps(json)
		if len(result.Packages) != 1 || result.Packages[0].Name != "VLC" {
			t.Errorf("got %+v, want VLC only", result.Packages)
		}
	})

	t.Run("null version", func(t *testing.T) {
		json := `[{"DisplayName":"VLC","DisplayVersion":null}]`
		result := parseWindowsInstalledApps(json)
		if len(result.Packages) != 1 {
			t.Fatalf("got %d packages", len(result.Packages))
		}
		if result.Packages[0].Version != "" {
			t.Errorf("version = %q, want empty", result.Packages[0].Version)
		}
	})

	t.Run("malformed JSON returns Found=false", func(t *testing.T) {
		result := parseWindowsInstalledApps(`{bad json`)
		if result.Found {
			t.Error("expected Found=false for malformed JSON")
		}
	})

	t.Run("empty string returns Found=false", func(t *testing.T) {
		result := parseWindowsInstalledApps(``)
		if result.Found {
			t.Error("expected Found=false for empty string")
		}
	})
}

// ── Orchestration / fallback tests ───────────────────────────────────────────
// These override execLookPath / execCommand to simulate different system states.

func TestGetWindowsInstalledApps_powershellNotFound(t *testing.T) {
	defer func(old func(string) (string, error)) { execLookPath = old }(execLookPath)
	execLookPath = func(name string) (string, error) {
		if name == "powershell" {
			return "", errors.New("not found")
		}
		return exec.LookPath(name)
	}

	result := getWindowsInstalledApps()
	if result.Found {
		t.Error("expected Found=false when powershell is missing")
	}
	if len(result.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(result.Packages))
	}
}

func TestGetWindowsInstalledApps_powershellFails(t *testing.T) {
	defer func(old func(string, ...string) *exec.Cmd) { execCommand = old }(execCommand)
	execCommand = func(name string, args ...string) *exec.Cmd {
		// execLookPath returns the full path, e.g. C:\Windows\...\powershell.exe
		if strings.Contains(name, "powershell") {
			return exec.Command("cmd", "/c", "exit 1")
		}
		return exec.Command(name, args...)
	}

	result := getWindowsInstalledApps()
	if result.Found {
		t.Error("expected Found=false when powershell exits non-zero")
	}
}

func TestGetInstalledPackages_platformDispatch(t *testing.T) {
	// Verify the function runs without error on whatever platform we're on.
	managers := GetInstalledPackages()
	if len(managers) == 0 {
		t.Fatal("expected at least one manager entry")
	}
	// Every entry must have a name
	for _, m := range managers {
		if m.Name == "" {
			t.Error("manager with empty name")
		}
	}
	// Log what we got for manual inspection
	t.Logf("runtime.GOOS=%q", runtime.GOOS)
	for _, m := range managers {
		t.Logf("  %s: found=%v count=%d", m.Name, m.Found, len(m.Packages))
	}
}
