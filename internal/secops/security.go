package secops

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// PasswordPolicy holds password policy configuration.
type PasswordPolicy struct {
	MaxAge           int  `json:"max_age"`
	MinLength        int  `json:"min_length"`
	Complexity       bool `json:"complexity"`
	LockoutThreshold int  `json:"lockout_threshold"`
	LockoutDuration  int  `json:"lockout_duration"`
}

// FailedLogin holds a failed login attempt record.
type FailedLogin struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	SourceIP string `json:"source_ip"`
	Count    int    `json:"count"`
}

// LockedAccount holds a locked account record.
type LockedAccount struct {
	Username    string `json:"username"`
	LockedSince string `json:"locked_since"`
}

// GetPasswordPolicy retrieves the password policy for the current system.
func GetPasswordPolicy() (*PasswordPolicy, error) {
	if common.IsWindows() {
		return getPasswordPolicyWindows()
	}
	return getPasswordPolicyLinux()
}

func getPasswordPolicyWindows() (*PasswordPolicy, error) {
	out, err := exec.Command("net", "accounts").Output()
	if err != nil {
		return nil, fmt.Errorf("net accounts failed: %w", err)
	}
	return parseNetAccounts(string(out)), nil
}

func parseNetAccounts(output string) *PasswordPolicy {
	p := &PasswordPolicy{MaxAge: 42, MinLength: 5, LockoutThreshold: 0, LockoutDuration: 30}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		if strings.Contains(lower, "maximum password age") {
			if v := extractNumber(line); v > 0 {
				p.MaxAge = v
			}
		} else if strings.Contains(lower, "minimum password length") {
			if v := extractNumber(line); v >= 0 {
				p.MinLength = v
			}
		} else if strings.Contains(lower, "lockout threshold") {
			if v := extractNumber(line); v >= 0 {
				p.LockoutThreshold = v
			}
		} else if strings.Contains(lower, "lockout duration") {
			if v := extractNumber(line); v >= 0 {
				p.LockoutDuration = v
			}
		} else if strings.Contains(lower, "password complexity") {
			p.Complexity = strings.Contains(strings.ToLower(line), "enabled")
		}
	}
	return p
}

func extractNumber(s string) int {
	s = strings.TrimSpace(s)
	parts := strings.Fields(s)
	for i := len(parts) - 1; i >= 0; i-- {
		val := strings.TrimRight(parts[i], ".")
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return 0
}

func getPasswordPolicyLinux() (*PasswordPolicy, error) {
	p := &PasswordPolicy{MaxAge: 99999, MinLength: 5, LockoutThreshold: 0, LockoutDuration: 30}

	// Parse /etc/login.defs
	if data, err := os.ReadFile("/etc/login.defs"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			switch fields[0] {
			case "PASS_MAX_DAYS":
				if v, err := strconv.Atoi(fields[1]); err == nil {
					p.MaxAge = v
				}
			case "PASS_MIN_LEN":
				if v, err := strconv.Atoi(fields[1]); err == nil {
					p.MinLength = v
				}
			}
		}
	}
	return p, nil
}

// GetFailedLogins retrieves recent failed login attempts.
func GetFailedLogins() ([]FailedLogin, error) {
	if common.IsWindows() {
		return getFailedLoginsWindows()
	}
	return getFailedLoginsLinux()
}

func getFailedLoginsWindows() ([]FailedLogin, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command",
		`Get-WinEvent -FilterHashtable @{Id=4625} -MaxEvents 50 -ErrorAction SilentlyContinue |
		Select-Object TimeCreated,Message | ConvertTo-Json -As Array -Depth 2`)
	out, err := cmd.Output()
	if err != nil {
		return []FailedLogin{}, nil
	}
	return parseFailedLoginsJSON(string(out))
}

func getFailedLoginsLinux() ([]FailedLogin, error) {
	// Try lastb first
	cmd := exec.Command("lastb", "-F", "-i", "-n", "50")
	out, err := cmd.Output()
	if err != nil {
		return []FailedLogin{}, nil
	}
	return parseLastbOutput(string(out)), nil
}

func parseLastbOutput(output string) []FailedLogin {
	counts := make(map[string]*FailedLogin)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 3 || fields[0] == "" {
			continue
		}
		username := fields[0]
		if username == "wtmp" || username == "" {
			continue
		}
		if fl, ok := counts[username]; ok {
			fl.Count++
		} else {
			fl := &FailedLogin{
				Username: username,
				Count:    1,
			}
			if len(fields) > 3 {
				fl.SourceIP = fields[len(fields)-1]
			}
			counts[username] = fl
		}
	}
	logins := make([]FailedLogin, 0, len(counts))
	for _, fl := range counts {
		logins = append(logins, *fl)
	}
	return logins
}

func parseFailedLoginsJSON(jsonStr string) ([]FailedLogin, error) {
	// Simplified: return empty for now, real implementation parses WinEvent JSON
	return []FailedLogin{}, nil
}

// GetAccountLockouts retrieves currently locked accounts.
func GetAccountLockouts() ([]LockedAccount, error) {
	if common.IsWindows() {
		return getAccountLockoutsWindows()
	}
	return getAccountLockoutsLinux()
}

func getAccountLockoutsWindows() ([]LockedAccount, error) {
	out, err := exec.Command("net", "user").Output()
	if err != nil {
		return nil, fmt.Errorf("net user failed: %w", err)
	}
	var locked []LockedAccount
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "*") && !strings.HasPrefix(line, "---") {
			user := strings.TrimSpace(strings.ReplaceAll(line, "*", ""))
			if user != "" {
				locked = append(locked, LockedAccount{Username: user})
			}
		}
	}
	return locked, nil
}

func getAccountLockoutsLinux() ([]LockedAccount, error) {
	return []LockedAccount{}, nil
}
