package secops

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// AuditCheckItem holds a single audit check result.
type AuditCheckItem struct {
	Category    string `json:"category"`
	Check       string `json:"check"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

// SecurityAuditResult holds the full audit result.
type SecurityAuditResult struct {
	Score     int              `json:"score"`
	Total     int              `json:"total"`
	Passed    int              `json:"passed"`
	Failed    int              `json:"failed"`
	Items     []AuditCheckItem `json:"items"`
	Timestamp string           `json:"timestamp"`
}

// RunSecurityAuditChecklist runs a one-click security audit.
func RunSecurityAuditChecklist() (*SecurityAuditResult, error) {
	var items []AuditCheckItem

	// 1. Firewall enabled
	firewallStatus, _ := GetFirewallProfiles()
	firewallOK := false
	for _, p := range firewallStatus {
		if p.Enabled {
			firewallOK = true
			break
		}
	}
	items = append(items, AuditCheckItem{
		Category: "Firewall", Check: "Firewall enabled", Passed: firewallOK,
		Description: "Windows Firewall should be enabled on all profiles",
		Remediation: "Enable firewall in Control Panel > System and Security > Windows Defender Firewall",
	})

	// 2. Secure Boot
	sb, _ := GetSecureBootStatus()
	sbOK := sb != nil && sb.Enabled
	items = append(items, AuditCheckItem{
		Category: "Boot Security", Check: "Secure Boot enabled", Passed: sbOK,
		Description: "Secure Boot prevents unauthorized bootloaders",
		Remediation: "Enable Secure Boot in UEFI/BIOS settings",
	})

	// 3. Disk encryption
	disks, _ := GetDiskEncryptionStatus()
	encrypted := false
	for _, d := range disks {
		if d.Encrypted {
			encrypted = true
			break
		}
	}
	items = append(items, AuditCheckItem{
		Category: "Data Protection", Check: "Disk encrypted", Passed: encrypted,
		Description: "Full disk encryption protects data at rest",
		Remediation: "Enable BitLocker (Windows) or LUKS (Linux)",
	})

	// 4. Exposed RDP/SSH
	exposed, _ := GetPublicExposure()
	exposedDangerous := false
	for _, e := range exposed {
		if e.Port == 3389 || e.Port == 22 {
			exposedDangerous = true
			break
		}
	}
	items = append(items, AuditCheckItem{
		Category: "Network", Check: "No exposed RDP/SSH", Passed: !exposedDangerous,
		Description: "RDP (3389) and SSH (22) should not be publicly exposed",
		Remediation: "Restrict RDP/SSH access via firewall rules or VPN",
	})

	// 5. Password policy
	policy, _ := GetPasswordPolicy()
	policyOK := policy != nil && policy.MinLength >= 8 && policy.Complexity
	items = append(items, AuditCheckItem{
		Category: "Identity", Check: "Password policy compliant", Passed: policyOK,
		Description: "Minimum 8 characters with complexity requirements",
		Remediation: "Configure: net accounts /minpwlen:8 /complexity:yes",
	})

	// 6. Certificates valid
	certs, _ := GetTLSCertificates()
	certsOK := true
	for _, c := range certs {
		if c.IsExpiring {
			certsOK = false
			break
		}
	}
	items = append(items, AuditCheckItem{
		Category: "Certificates", Check: "Certificates valid", Passed: certsOK,
		Description: "No certificates expiring within 30 days",
		Remediation: "Renew expiring certificates before expiry",
	})

	// 7. Scheduled tasks anomaly
	tasks, _ := GetScheduledTasks()
	anomalyCount := 0
	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Trigger), "logon") ||
			strings.Contains(strings.ToLower(t.Trigger), "startup") {
			anomalyCount++
		}
	}
	items = append(items, AuditCheckItem{
		Category: "Persistence", Check: "No suspicious persistence", Passed: anomalyCount < 5,
		Description: fmt.Sprintf("%d startup/logon tasks detected", anomalyCount),
		Remediation: "Review scheduled tasks for unauthorized entries",
	})

	// Count results
	passed := 0
	for _, item := range items {
		if item.Passed {
			passed++
		}
	}
	total := len(items)
	score := 0
	if total > 0 {
		score = (passed * 100) / total
	}

	_ = runtime.GOOS // used for platform detection

	return &SecurityAuditResult{
		Score:     score,
		Total:     total,
		Passed:    passed,
		Failed:    total - passed,
		Items:     items,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
