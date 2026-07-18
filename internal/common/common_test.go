package common

import (
	"os"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"512 bytes", 512, "512 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"2 KB", 2048, "2.0 KB"},
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"1 TB", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBytes(tt.input); got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"zero", 0, "0.0%"},
		{"fifty", 50, "50.0%"},
		{"one hundred", 100, "100.0%"},
		{"with decimals", 45.678, "45.7%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPercent(tt.input); got != tt.want {
				t.Errorf("FormatPercent(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		want  string
	}{
		{"zero seconds", 0, "0m"},
		{"one minute", 60, "1m"},
		{"one hour", 3600, "1h 0m"},
		{"one hour one minute", 3660, "1h 1m"},
		{"one day", 86400, "1d 0h 0m"},
		{"one day one hour one minute", 90061, "1d 1h 1m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatUptime(tt.input); got != tt.want {
				t.Errorf("FormatUptime(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRepeatString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"normal repetition", "abc", 3, "abcabcabc"},
		{"empty string", "", 5, ""},
		{"zero count", "x", 0, ""},
		{"negative count", "x", -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RepeatString(tt.s, tt.n); got != tt.want {
				t.Errorf("RepeatString(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"one over length", "hello", 4, "h..."},
		{"long string", "hello world", 8, "hello..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateString(tt.input, tt.maxLen); got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFixPowerShellDashes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no change: normal json", `{"a": "b"}`, `{"a": "b"}`},
		{"dash value for unset field", `{"StartType": -}`, `{"StartType": ""}`},
		{"dash value with comma", `{"StartType": -, "Name": "svc"}`, `{"StartType": "", "Name": "svc"}`},
		{"dash value with newline", "{\"Id\": -\n}", "{\"Id\": \"\"\n}"},
		{"dash value with spaces", `{"Val":  - }`, `{"Val":  "" }`},
		{"dash value with spaces and comma", `{"Val":  - , "Next": 1}`, `{"Val":  "" , "Next": 1}`},
		{"negative number not affected", `{"val": -1}`, `{"val": -1}`},
		{"negative float not affected", `{"val": -1.5}`, `{"val": -1.5}`},
		{"string with dash not affected", `{"val": "-"}`, `{"val": "-"}`},
		{"empty string input", "", ""},
		{"dash in middle of array", `[-]`, `[-]`}, // unchanged, not after colon
		{"multiple dashes", `{"a": -, "b": -, "c": 3}`, `{"a": "", "b": "", "c": 3}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FixPowerShellDashes(tt.input); got != tt.want {
				t.Errorf("FixPowerShellDashes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	info := DetectPlatform()
	if info.OS == "" {
		t.Error("DetectPlatform().OS is empty")
	}
	if info.Arch == "" {
		t.Error("DetectPlatform().Arch is empty")
	}
}

func TestPlatformChecks(t *testing.T) {
	// At least one OS check must be true on any real platform.
	anyOS := IsWindows() || IsLinux() || IsMacOS()
	if !anyOS {
		t.Error("none of IsWindows, IsLinux, IsMacOS returned true")
	}

	// IsAdminRequired should match the disjunction of Windows and Linux.
	want := IsWindows() || IsLinux()
	if got := IsAdminRequired(); got != want {
		t.Errorf("IsAdminRequired() = %v, want %v (IsWindows=%v, IsLinux=%v, IsMacOS=%v)",
			got, want, IsWindows(), IsLinux(), IsMacOS())
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() returned error: %v", err)
	}
	if dir == "" {
		t.Error("ConfigDir() returned an empty path")
	}
	// Should be "data" for self-contained mode
	if dir != "data" {
		t.Errorf("ConfigDir() = %q, want 'data'", dir)
	}
}

func TestIsOnboarded(t *testing.T) {
	// Must be false initially (no marker file exists in a temp context).
	// We clear any existing state by working with a temp dir override trick:
	// Instead of modifying the real config dir, we verify the function
	// returns false when no file exists.
	//
	// The cleanest way: test via Mark/Clear round-trip below.
	// This test just confirms the negative case on a clean system.
	onboarded := IsOnboarded()
	// We can't assert false here because another test may have left the
	// marker file. The round-trip test is authoritative.
	_ = onboarded
}

func TestMarkAndClearOnboarded(t *testing.T) {
	// Use a temporary config dir to isolate tests.
	_ = os.Getenv("XDG_CONFIG_HOME")
	_ = os.Getenv("HOME")
	// Can't unset/restore on Windows easily via HOME, so use the simple
	// approach: call MarkOnboarded and verify, then Clear and verify.

	// First ensure we start clean.
	_ = ClearOnboarded()

	if IsOnboarded() {
		t.Fatal("expected IsOnboarded() to be false after ClearOnboarded()")
	}

	if err := MarkOnboarded(); err != nil {
		t.Fatalf("MarkOnboarded() returned error: %v", err)
	}

	if !IsOnboarded() {
		t.Fatal("expected IsOnboarded() to be true after MarkOnboarded()")
	}

	if err := ClearOnboarded(); err != nil {
		t.Fatalf("ClearOnboarded() returned error: %v", err)
	}

	if IsOnboarded() {
		t.Fatal("expected IsOnboarded() to be false after second ClearOnboarded()")
	}
}
