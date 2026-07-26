package secops

import (
	"fmt"
	"os"
	"strings"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// HardeningCheck holds a single hardening check result.
type HardeningCheck struct {
	Category    string `json:"category"`
	Check       string `json:"check"`
	Passed      bool   `json:"passed"`
	Severity    string `json:"severity"`
	Remediation string `json:"remediation"`
}

// SSHConfig holds SSH configuration.
type SSHConfig struct {
	PermitRootLogin        string `json:"permit_root_login"`
	PasswordAuthentication string `json:"password_authentication"`
	PubkeyAuthentication   string `json:"pubkey_authentication"`
	X11Forwarding          string `json:"x11_forwarding"`
	MaxAuthTries           string `json:"max_auth_tries"`
}

// GetHardeningChecks retrieves hardening check results.
func GetHardeningChecks() ([]HardeningCheck, error) {
	if common.IsWindows() {
		return getHardeningChecksWindows()
	}
	return getHardeningChecksLinux()
}

func getHardeningChecksWindows() ([]HardeningCheck, error) {
	var checks []HardeningCheck

	// Firewall check
	firewallStatus, _ := GetFirewallProfiles()
	firewallEnabled := false
	for _, p := range firewallStatus {
		if p.Enabled {
			firewallEnabled = true
			break
		}
	}
	checks = append(checks, HardeningCheck{
		Category: "Firewall", Check: "Firewall enabled", Passed: firewallEnabled,
		Severity: "high", Remediation: "Enable Windows Firewall in Control Panel",
	})

	// Defender check
	defender, _ := GetDefenderStatus()
	defenderOK := defender != nil && defender.Enabled && defender.RealTimeProtection
	checks = append(checks, HardeningCheck{
		Category: "Antivirus", Check: "Windows Defender enabled", Passed: defenderOK,
		Severity: "high", Remediation: "Enable Windows Defender in Settings",
	})

	// SMBv1 check
	out, err := common.HiddenCommand("powershell", "-NoProfile", "-Command",
		"Get-SmbServerConfiguration | Select-Object -ExpandProperty EnableSMB1Protocol").Output()
	smbv1Disabled := err != nil || !strings.Contains(strings.TrimSpace(string(out)), "True")
	checks = append(checks, HardeningCheck{
		Category: "Protocol", Check: "SMBv1 disabled", Passed: smbv1Disabled,
		Severity: "medium", Remediation: "Disable SMBv1: Set-SmbServerConfiguration -EnableSMB1Protocol $false",
	})

	// Guest account check
	guestOut, err := common.HiddenCommand("net", "user", "Guest").Output()
	guestDisabled := true // default: assume disabled
	if err == nil {
		for _, line := range strings.Split(string(guestOut), "\n") {
			if strings.Contains(strings.ToLower(line), "account active") {
				guestDisabled = !strings.Contains(strings.ToLower(line), "yes")
				break
			}
		}
	}
	checks = append(checks, HardeningCheck{
		Category: "Account", Check: "Guest account disabled", Passed: guestDisabled,
		Severity: "medium", Remediation: "Disable guest account: net user Guest /active:no",
	})

	// BitLocker check
	bitLockerOut, err := common.HiddenCommand("powershell", "-NoProfile", "-Command",
		"(Get-BitLockerVolume -MountPoint C:).ProtectionStatus").Output()
	bitLockerOK := err == nil && strings.TrimSpace(string(bitLockerOut)) == "1"
	checks = append(checks, HardeningCheck{
		Category: "Encryption", Check: "BitLocker enabled", Passed: bitLockerOK,
		Severity: "high", Remediation: "Enable BitLocker: Enable-BitLocker -MountPoint C: -TpmProtector",
	})

	return checks, nil
}

func getHardeningChecksLinux() ([]HardeningCheck, error) {
	var checks []HardeningCheck

	// UFW/firewall check
	ufwOut, err := common.HiddenCommand("ufw", "status").Output()
	firewallActive := err == nil && strings.Contains(strings.ToLower(string(ufwOut)), "active")
	checks = append(checks, HardeningCheck{
		Category: "Firewall", Check: "Firewall active", Passed: firewallActive,
		Severity: "high", Remediation: "Enable UFW: sudo ufw enable",
	})

	// SSH root login check
	sshConfig, _ := GetSSHConfig()
	rootDisabled := sshConfig.PermitRootLogin == "no"
	checks = append(checks, HardeningCheck{
		Category: "SSH", Check: "Root login disabled", Passed: rootDisabled,
		Severity: "high", Remediation: "Set PermitRootLogin no in /etc/ssh/sshd_config",
	})

	// World-writable files in /etc
	out, err := common.HiddenCommand("find", "/etc", "-maxdepth", "2", "-perm", "-002", "-type", "f").Output()
	noWorldWritable := err == nil && strings.TrimSpace(string(out)) == ""
	checks = append(checks, HardeningCheck{
		Category: "Files", Check: "No world-writable in /etc", Passed: noWorldWritable,
		Severity: "medium", Remediation: "Run: sudo find /etc -perm -002 -type f",
	})

	// SELinux/AppArmor check
	selinuxOut, selinuxErr := common.HiddenCommand("getenforce").Output()
	apparmorOut, apparmorErr := common.HiddenCommand("aa-status").Output()
	macEnabled := (selinuxErr == nil && strings.Contains(strings.ToLower(string(selinuxOut)), "enforcing")) ||
		(apparmorErr == nil && strings.Contains(strings.ToLower(string(apparmorOut)), "enabled"))
	checks = append(checks, HardeningCheck{
		Category: "Access Control", Check: "SELinux/AppArmor enforcing", Passed: macEnabled,
		Severity: "high", Remediation: "Enable SELinux: setenforce 1; or AppArmor: systemctl enable apparmor",
	})

	// SSH password auth check
	passAuthOK := sshConfig.PasswordAuthentication == "no" || sshConfig.PasswordAuthentication == "prohibit-password"
	checks = append(checks, HardeningCheck{
		Category: "SSH", Check: "Password auth disabled", Passed: passAuthOK,
		Severity: "high", Remediation: "Set PasswordAuthentication no in /etc/ssh/sshd_config",
	})

	return checks, nil
}

// GetSSHConfig retrieves SSH configuration.
func GetSSHConfig() (*SSHConfig, error) {
	data, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return &SSHConfig{
			PermitRootLogin:        "unknown",
			PasswordAuthentication: "unknown",
			PubkeyAuthentication:   "unknown",
			X11Forwarding:          "unknown",
			MaxAuthTries:           "unknown",
		}, nil
	}
	config := &SSHConfig{}
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
		case "PermitRootLogin":
			config.PermitRootLogin = fields[1]
		case "PasswordAuthentication":
			config.PasswordAuthentication = fields[1]
		case "PubkeyAuthentication":
			config.PubkeyAuthentication = fields[1]
		case "X11Forwarding":
			config.X11Forwarding = fields[1]
		case "MaxAuthTries":
			config.MaxAuthTries = fields[1]
		}
	}
	return config, nil
}

// GetFirewallProfiles is defined in firewall.go — used by hardening checks.
// GetDefenderStatus is defined in defender.go — used by hardening checks.

func init() {
	// Ensure hardening.go can reference firewall/defender types
	_ = fmt.Sprintf("firewall=%v", false)
}
