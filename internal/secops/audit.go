package secops

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
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
	isAdmin := common.IsAdmin()

	// 1. Firewall enabled
	firewallStatus, fwErr := GetFirewallProfiles()
	firewallOK := false
	for _, p := range firewallStatus {
		if p.Enabled {
			firewallOK = true
			break
		}
	}
	fwDesc := "Windows Firewall should be enabled on all profiles"
	if fwErr != nil {
		fwDesc = fmt.Sprintf("Error checking firewall: %v", fwErr)
	} else if len(firewallStatus) == 0 {
		fwDesc = "No firewall profiles detected"
	}
	items = append(items, AuditCheckItem{
		Category: "Firewall", Check: "Firewall enabled", Passed: firewallOK,
		Description: fwDesc,
		Remediation: "Enable firewall in Control Panel > System and Security > Windows Defender Firewall",
	})

	// 2. Secure Boot
	sb, _ := GetSecureBootStatus()
	sbOK := sb != nil && sb.Enabled
	sbDesc := "Secure Boot prevents unauthorized bootloaders"
	if !isAdmin && !sbOK {
		sbDesc = "Secure Boot status unavailable (Run as Administrator to check)"
	}
	items = append(items, AuditCheckItem{
		Category: "Boot Security", Check: "Secure Boot enabled", Passed: sbOK,
		Description: sbDesc,
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
	diskDesc := "Full disk encryption protects data at rest"
	if !isAdmin && !encrypted {
		diskDesc = "Disk encryption status unavailable (Run as Administrator to check)"
	}
	items = append(items, AuditCheckItem{
		Category: "Data Protection", Check: "Disk encrypted", Passed: encrypted,
		Description: diskDesc,
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
	policyOK := policy != nil && (policy.MinLength >= 8 || policy.Complexity)
	passDesc := "Minimum 8 characters with complexity requirements"
	if policy != nil {
		passDesc = fmt.Sprintf("Current: MinLength=%d, Complexity=%v", policy.MinLength, policy.Complexity)
	}
	items = append(items, AuditCheckItem{
		Category: "Identity", Check: "Password policy compliant", Passed: policyOK,
		Description: passDesc,
		Remediation: "Configure: net accounts /minpwlen:8 /complexity:yes",
	})

	// 6. Certificates valid
	certs, _ := GetTLSCertificates()
	certsOK := true
	expiringCount := 0
	for _, c := range certs {
		if c.IsExpiring {
			certsOK = false
			expiringCount++
		}
	}
	certDesc := "No certificates expiring within 30 days"
	if expiringCount > 0 {
		certDesc = fmt.Sprintf("%d certificate(s) expiring soon", expiringCount)
	}
	items = append(items, AuditCheckItem{
		Category: "Certificates", Check: "Certificates valid", Passed: certsOK,
		Description: certDesc,
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
		Category: "Persistence", Check: "No suspicious persistence", Passed: anomalyCount < 10,
		Description: fmt.Sprintf("%d startup/logon tasks detected", anomalyCount),
		Remediation: "Review scheduled tasks for unauthorized entries",
	})

	// 8. Windows Defender Status
	defender, _ := GetDefenderStatus()
	defenderOK := defender != nil && defender.Enabled && defender.RealTimeProtection
	defDesc := "Real-time protection should be active"
	if defender != nil {
		if !defender.Enabled {
			defDesc = "Windows Defender is disabled"
		} else if !defender.RealTimeProtection {
			defDesc = "Real-time protection is disabled"
		}
	}
	items = append(items, AuditCheckItem{
		Category: "Endpoint", Check: "Defender active", Passed: defenderOK,
		Description: defDesc,
		Remediation: "Enable Windows Defender and Real-time protection in Settings",
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
