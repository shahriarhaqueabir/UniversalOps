package app

import (
	"fmt"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/secops"
)

// SecOps exposes security operations bindings to the frontend.
type SecOps struct {
	eventBus *common.EventBus
}

// NewSecOps creates a new SecOps facade.
func NewSecOps(eventBus *common.EventBus) *SecOps {
	return &SecOps{eventBus: eventBus}
}

// GetFirewallRules returns current firewall rules.
func (s *SecOps) GetFirewallRules() []FirewallRule {
	rules, err := secops.GetFirewallRules()
	if err != nil {
		common.LogWarn("GetFirewallRules failed: %v", err)
		return []FirewallRule{}
	}
	out := make([]FirewallRule, 0, len(rules))
	for _, r := range rules {
		out = append(out, FirewallRule{
			Name:       r.Name,
			Direction:  r.Direction,
			Action:     r.Action,
			Protocol:   r.Protocol,
			LocalPort:  r.LocalPort,
			RemotePort: r.RemotePort,
			RemoteIP:   r.RemoteIP,
			Profile:    r.Profile,
			Enabled:    r.Enabled,
			IsHighRisk: r.IsHighRisk,
		})
	}
	return out
}

// GetUsers returns local user accounts.
func (s *SecOps) GetUsers() []UserInfo {
	users, err := secops.GetUsers()
	if err != nil {
		common.LogWarn("GetUsers failed: %v", err)
		return []UserInfo{}
	}
	out := make([]UserInfo, 0, len(users))
	for _, u := range users {
		out = append(out, UserInfo{
			Username:  u.Username,
			FullName:  u.FullName,
			SID:       u.SID,
			Group:     u.Group,
			IsAdmin:   u.IsAdmin,
			IsEnabled: u.IsEnabled,
		})
	}
	return out
}

// GetListeningPorts returns all listening network ports.
func (s *SecOps) GetListeningPorts() []ListeningPort {
	ports, err := secops.GetListeningPorts()
	if err != nil {
		common.LogWarn("GetListeningPorts failed: %v", err)
		return []ListeningPort{}
	}
	out := make([]ListeningPort, 0, len(ports))
	for _, p := range ports {
		out = append(out, ListeningPort{
			Port:        p.Port,
			Protocol:    p.Protocol,
			ProcessName: p.ProcessName,
			PID:         p.PID,
			State:       p.State,
			IsExternal:  p.IsExternal,
		})
	}
	return out
}

// GetDefenderStatus returns Windows Defender status.
func (s *SecOps) GetDefenderStatus() DefenderStatus {
	status, err := secops.GetDefenderStatus()
	if err != nil {
		common.LogWarn("GetDefenderStatus failed: %v", err)
		return DefenderStatus{}
	}
	return DefenderStatus{
		Enabled:            status.Enabled,
		UpToDate:           status.UpToDate,
		SignatureAge:       status.SignatureAge,
		LastScan:           status.LastScan,
		RealTimeProtection: status.RealTimeProtection,
		CloudProtection:    status.CloudProtection,
		AMServiceEnabled:   status.AMServiceEnabled,
		AntispywareEnabled: status.AntispywareEnabled,
		NISEnabled:         status.NISEnabled,
		QuickScanAge:       status.QuickScanAge,
		FullScanAge:        status.FullScanAge,
		ThreatsDetected:    status.ThreatsDetected,
	}
}

// GetScheduledTasks returns scheduled tasks.
func (s *SecOps) GetScheduledTasks() []ScheduledTask {
	tasks, err := secops.GetScheduledTasks()
	if err != nil {
		common.LogWarn("GetScheduledTasks failed: %v", err)
		return []ScheduledTask{}
	}
	out := make([]ScheduledTask, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, ScheduledTask{
			Name:    t.Name,
			Status:  t.Status,
			NextRun: t.NextRun,
			LastRun: t.LastRun,
			Author:  t.Author,
			Trigger: t.Trigger,
		})
	}
	return out
}

// GetSecurityEvents returns recent security event log entries.
func (s *SecOps) GetSecurityEvents() []SecurityEvent {
	events, err := secops.GetSecurityEvents()
	if err != nil {
		common.LogWarn("GetSecurityEvents failed: %v", err)
		return []SecurityEvent{}
	}
	out := make([]SecurityEvent, 0, len(events))
	for _, e := range events {
		out = append(out, SecurityEvent{
			ID:        e.ID,
			Level:     e.Level,
			Provider:  e.Provider,
			Time:      e.Time,
			Message:   e.Message,
			Important: e.Important,
		})
	}
	return out
}

// SetFirewallRuleHandshake requests a safety handshake for a firewall rule change.
func (s *SecOps) SetFirewallRuleHandshake(name string, enable bool) common.ActionPreview {
	actionName := "Disable Firewall Rule"
	if enable {
		actionName = "Enable Firewall Rule"
	}

	id := common.GetHandshakeRegistry().Register("SetFirewallRule", map[string]interface{}{
		"name":   name,
		"enable": enable,
	})

	return common.ActionPreview{
		HandshakeID: id,
		Action:      actionName,
		Description: fmt.Sprintf("%s: '%s'", actionName, name),
		Risks:       []string{"May disrupt active network connections", "Potential security surface change"},
		Rollback:    "Can be toggled back manually via the same interface.",
	}
}

// executeSetFirewallRuleState enables or disables a firewall rule.
func (s *SecOps) executeSetFirewallRuleState(name string, enable bool) bool {
	common.LogInfo("Setting firewall rule %s enabled=%v", name, enable)
	err := secops.SetFirewallRuleState(name, enable)
	if err != nil {
		common.LogWarn("SetFirewallRuleState failed: %v", err)
		return false
	}

	action := "disabled"
	if enable {
		action = "enabled"
	}
	if s.eventBus != nil {
		s.eventBus.Emit(common.NewEvent(
			common.CatSecurity,
			common.EventInfo,
			"secops",
			"Firewall rule changed",
			fmt.Sprintf("Firewall rule '%s' %s", name, action),
		))
	}
	return true
}

// GetSecurityScore computes a security posture score from existing data sources.
func (s *SecOps) GetSecurityScore() SecurityScore {
	score := secops.ComputeSecurityScore()
	return SecurityScore{
		Score:           score.Score,
		Grade:           score.Grade,
		Breakdown:       score.Breakdown,
		Recommendations: score.Recommendations,
	}
}

// GetFirewallStatus returns the global firewall on/off status and per-profile status.
func (s *SecOps) GetFirewallStatus() FirewallStatus {
	profiles, err := secops.GetFirewallProfiles()
	if err != nil {
		common.LogWarn("GetFirewallStatus failed: %v", err)
		return FirewallStatus{
			Profiles: []FirewallProfile{},
		}
	}
	out := make([]FirewallProfile, 0, len(profiles))
	enabled := false
	for _, p := range profiles {
		out = append(out, FirewallProfile{
			Name:    p.Name,
			Enabled: p.Enabled,
		})
		if p.Enabled {
			enabled = true
		}
	}
	return FirewallStatus{
		Enabled:  enabled,
		Profiles: out,
	}
}

// GetRisks returns a list of detected security risks.
func (s *SecOps) GetRisks() []RiskInfo {
	var risks []RiskInfo
	dangerousPorts := map[int]string{3389: "RDP", 445: "SMB", 23: "Telnet"}

	// ── Check listening ports for exposed services ──
	ports, pErr := secops.GetListeningPorts()
	if pErr == nil {
		for _, p := range ports {
			if p.IsExternal {
				if svc, ok := dangerousPorts[p.Port]; ok {
					risks = append(risks, RiskInfo{
						Category:       "Network",
						Severity:       "critical",
						Title:          fmt.Sprintf("%s exposed on port %d", svc, p.Port),
						Description:    fmt.Sprintf("%s service is listening on all interfaces (port %d). This is a common attack vector.", svc, p.Port),
						Recommendation: fmt.Sprintf("Close or restrict external access to port %d. Use firewall rules to limit access to trusted IPs only.", p.Port),
					})
				} else {
					risks = append(risks, RiskInfo{
						Category:       "Network",
						Severity:       "medium",
						Title:          fmt.Sprintf("External listening port %d (%s)", p.Port, p.ProcessName),
						Description:    fmt.Sprintf("Port %d is bound to all interfaces via %s. Verify this is expected.", p.Port, p.ProcessName),
						Recommendation: "Review whether this service needs to be externally accessible.",
					})
				}
			}
		}
	}

	// ── Check firewall rules for high-risk entries ──
	rules, fErr := secops.GetFirewallRules()
	if fErr == nil {
		highRisk := 0
		for _, r := range rules {
			if r.IsHighRisk {
				highRisk++
			}
		}
		if highRisk > 0 {
			risks = append(risks, RiskInfo{
				Category:       "Firewall",
				Severity:       "high",
				Title:          fmt.Sprintf("%d high-risk firewall rule(s)", highRisk),
				Description:    "Firewall rules allow traffic from any IP on sensitive ports. This creates an open attack surface.",
				Recommendation: "Tighten firewall rules to restrict access to specific source IPs.",
			})
		}
	}

	// ── Check user accounts ──
	users, uErr := secops.GetUsers()
	if uErr == nil {
		adminCount := 0
		for _, u := range users {
			if u.IsAdmin && u.IsEnabled {
				adminCount++
			}
		}
		if adminCount > 2 {
			risks = append(risks, RiskInfo{
				Category:       "Identity",
				Severity:       "medium",
				Title:          fmt.Sprintf("%d administrator accounts detected", adminCount),
				Description:    "Too many admin accounts increase the risk of privilege escalation.",
				Recommendation: "Reduce to 2 or fewer admin accounts. Apply the Principle of Least Privilege.",
			})
		}
	}

	// ── Check Defender status ──
	defender, dErr := secops.GetDefenderStatus()
	if dErr == nil {
		if !defender.UpToDate {
			risks = append(risks, RiskInfo{
				Category:       "Defender",
				Severity:       "high",
				Title:          "Defender signatures are outdated",
				Description:    fmt.Sprintf("Last signature update was %s. Outdated definitions miss known threats.", defender.SignatureAge),
				Recommendation: "Run Windows Update or trigger a manual Defender signature update.",
			})
		}
		if defender.ThreatsDetected > 0 {
			risks = append(risks, RiskInfo{
				Category:       "Defender",
				Severity:       "critical",
				Title:          fmt.Sprintf("%d threat(s) detected by Defender", defender.ThreatsDetected),
				Description:    "Windows Defender has detected active threats on this system.",
				Recommendation: "Review and remediate detected threats immediately in Windows Security.",
			})
		}
	}

	// ── Check security events for failed logins ──
	events, eErr := secops.GetSecurityEvents()
	if eErr == nil {
		failedLogins := 0
		for _, e := range events {
			if e.ID == 4625 {
				failedLogins++
			}
		}
		if failedLogins > 10 {
			risks = append(risks, RiskInfo{
				Category:       "Identity",
				Severity:       "high",
				Title:          fmt.Sprintf("%d failed login attempts detected", failedLogins),
				Description:    "A spike in failed logins may indicate a brute-force attack.",
				Recommendation: "Investigate the source IPs, enable account lockout policies, and consider enabling MFA.",
			})
		}
	}

	return risks
}

// GetSecuritySummary returns a unified security posture overview.
// It collects data from firewall, defender, ports, users, and events,
// then generates a text summary with categorized risks and recommendations.
func (s *SecOps) GetSecuritySummary() SecuritySummary {
	summary := secops.GetSecuritySummary()
	return SecuritySummary{
		Score:           summary.Score,
		Summary:         summary.Summary,
		Risks:           summary.Risks,
		Recommendations: summary.Recommendations,
		AnalyzedAt:      summary.AnalyzedAt,
	}
}

// ── Phase 2 Bound Methods ─────────────────────────────────────────────────────

// GetPasswordPolicy returns the system password policy.
func (s *SecOps) GetPasswordPolicy() PasswordPolicy {
	p, err := secops.GetPasswordPolicy()
	if err != nil {
		common.LogWarn("GetPasswordPolicy failed: %v", err)
		return PasswordPolicy{}
	}
	return PasswordPolicy{
		MaxAge:           p.MaxAge,
		MinLength:        p.MinLength,
		Complexity:       p.Complexity,
		LockoutThreshold: p.LockoutThreshold,
		LockoutDuration:  p.LockoutDuration,
	}
}

// GetFailedLogins returns recent failed login attempts.
func (s *SecOps) GetFailedLogins() []FailedLogin {
	logins, err := secops.GetFailedLogins()
	if err != nil {
		common.LogWarn("GetFailedLogins failed: %v", err)
		return []FailedLogin{}
	}
	out := make([]FailedLogin, 0, len(logins))
	for _, l := range logins {
		out = append(out, FailedLogin{
			Time:     l.Time,
			Username: l.Username,
			SourceIP: l.SourceIP,
			Count:    l.Count,
		})
	}
	return out
}

// GetAccountLockouts returns locked accounts.
func (s *SecOps) GetAccountLockouts() []LockedAccount {
	locked, err := secops.GetAccountLockouts()
	if err != nil {
		common.LogWarn("GetAccountLockouts failed: %v", err)
		return []LockedAccount{}
	}
	out := make([]LockedAccount, 0, len(locked))
	for _, l := range locked {
		out = append(out, LockedAccount{
			Username:    l.Username,
			LockedSince: l.LockedSince,
		})
	}
	return out
}

// GetDiskEncryptionStatus returns disk encryption status.
func (s *SecOps) GetDiskEncryptionStatus() []DiskEncryption {
	disks, err := secops.GetDiskEncryptionStatus()
	if err != nil {
		common.LogWarn("GetDiskEncryptionStatus failed: %v", err)
		return []DiskEncryption{}
	}
	out := make([]DiskEncryption, 0, len(disks))
	for _, d := range disks {
		out = append(out, DiskEncryption{
			Volume:    d.Volume,
			Encrypted: d.Encrypted,
			Method:    d.Method,
			Status:    d.Status,
		})
	}
	return out
}

// GetSecureBootStatus returns secure boot status.
func (s *SecOps) GetSecureBootStatus() SecureBoot {
	sb, err := secops.GetSecureBootStatus()
	if err != nil {
		common.LogWarn("GetSecureBootStatus failed: %v", err)
		return SecureBoot{}
	}
	return SecureBoot{
		Enabled: sb.Enabled,
		State:   sb.State,
	}
}

// GetRunningServices returns running system services.
func (s *SecOps) GetRunningServices() []SystemService {
	services, err := secops.GetRunningServices()
	if err != nil {
		common.LogWarn("GetRunningServices failed: %v", err)
		return []SystemService{}
	}
	out := make([]SystemService, 0, len(services))
	for _, sv := range services {
		out = append(out, SystemService{
			Name:        sv.Name,
			DisplayName: sv.DisplayName,
			Status:      sv.Status,
			StartupType: sv.StartupType,
		})
	}
	return out
}

// GetTLSCertificates returns TLS certificate info.
func (s *SecOps) GetTLSCertificates() []TLSCertificate {
	certs, err := secops.GetTLSCertificates()
	if err != nil {
		common.LogWarn("GetTLSCertificates failed: %v", err)
		return []TLSCertificate{}
	}
	out := make([]TLSCertificate, 0, len(certs))
	for _, c := range certs {
		out = append(out, TLSCertificate{
			Subject:    c.Subject,
			Issuer:     c.Issuer,
			NotAfter:   c.NotAfter,
			KeySize:    c.KeySize,
			IsExpiring: c.IsExpiring,
			DaysLeft:   c.DaysLeft,
		})
	}
	return out
}

// GetPublicExposure returns externally-facing ports.
func (s *SecOps) GetPublicExposure() []PublicExposure {
	exposed, err := secops.GetPublicExposure()
	if err != nil {
		common.LogWarn("GetPublicExposure failed: %v", err)
		return []PublicExposure{}
	}
	out := make([]PublicExposure, 0, len(exposed))
	for _, e := range exposed {
		out = append(out, PublicExposure{
			Port:        e.Port,
			Protocol:    e.Protocol,
			ProcessName: e.ProcessName,
			Severity:    e.Severity,
		})
	}
	return out
}

// GetHardeningChecks returns hardening check results.
func (s *SecOps) GetHardeningChecks() []HardeningCheck {
	checks, err := secops.GetHardeningChecks()
	if err != nil {
		common.LogWarn("GetHardeningChecks failed: %v", err)
		return []HardeningCheck{}
	}
	out := make([]HardeningCheck, 0, len(checks))
	for _, c := range checks {
		out = append(out, HardeningCheck{
			Category:    c.Category,
			Check:       c.Check,
			Passed:      c.Passed,
			Severity:    c.Severity,
			Remediation: c.Remediation,
		})
	}
	return out
}

// GetSSHConfig returns SSH configuration.
func (s *SecOps) GetSSHConfig() SSHConfig {
	sc, err := secops.GetSSHConfig()
	if err != nil {
		common.LogWarn("GetSSHConfig failed: %v", err)
		return SSHConfig{}
	}
	return SSHConfig{
		PermitRootLogin:        sc.PermitRootLogin,
		PasswordAuthentication: sc.PasswordAuthentication,
		PubkeyAuthentication:   sc.PubkeyAuthentication,
		X11Forwarding:          sc.X11Forwarding,
		MaxAuthTries:           sc.MaxAuthTries,
	}
}

// GetPrivilegeEvents returns privilege escalation events.
func (s *SecOps) GetPrivilegeEvents() []PrivilegeEvent {
	events, err := secops.GetPrivilegeEvents()
	if err != nil {
		common.LogWarn("GetPrivilegeEvents failed: %v", err)
		return []PrivilegeEvent{}
	}
	out := make([]PrivilegeEvent, 0, len(events))
	for _, e := range events {
		out = append(out, PrivilegeEvent{
			Time:      e.Time,
			Username:  e.Username,
			Privilege: e.Privilege,
			Process:   e.Process,
		})
	}
	return out
}

// GetEventTimeline returns merged chronological security events.
func (s *SecOps) GetEventTimeline() []SecTimelineEvent {
	events, err := secops.GetEventTimeline()
	if err != nil {
		common.LogWarn("GetEventTimeline failed: %v", err)
		return []SecTimelineEvent{}
	}
	out := make([]SecTimelineEvent, 0, len(events))
	for _, e := range events {
		out = append(out, SecTimelineEvent{
			Time:     e.Time,
			Type:     e.Type,
			Detail:   e.Detail,
			Severity: e.Severity,
		})
	}
	return out
}

// RunSecurityAuditChecklist runs a one-click security audit.
func (s *SecOps) RunSecurityAuditChecklist() SecurityAuditResult {
	result, err := secops.RunSecurityAuditChecklist()
	if err != nil {
		common.LogWarn("RunSecurityAuditChecklist failed: %v", err)
		return SecurityAuditResult{}
	}
	items := make([]AuditCheckItem, 0, len(result.Items))
	for _, i := range result.Items {
		items = append(items, AuditCheckItem{
			Category:    i.Category,
			Check:       i.Check,
			Passed:      i.Passed,
			Description: i.Description,
			Remediation: i.Remediation,
		})
	}
	return SecurityAuditResult{
		Score:     result.Score,
		Total:     result.Total,
		Passed:    result.Passed,
		Failed:    result.Failed,
		Items:     items,
		Timestamp: result.Timestamp,
	}
}

// IsolateHost requests a safety handshake for isolating the host.
func (s *SecOps) IsolateHost(confirm bool, autoExpireSeconds int) common.ActionPreview {
	id := common.GetHandshakeRegistry().Register("IsolateHost", map[string]interface{}{
		"confirm":           confirm,
		"autoExpireSeconds": autoExpireSeconds,
	})

	return common.ActionPreview{
		HandshakeID: id,
		Action:      "Isolate Host",
		Description: "Sever all inbound and outbound network connectivity except for essential management traffic.",
		Risks:       []string{"Loss of network connectivity", "May disrupt active SSH or remote management sessions", "Potential for local lockout if not carefully configured"},
		Rollback:    fmt.Sprintf("Isolation rule will automatically expire in %d seconds if configured.", autoExpireSeconds),
	}
}

// executeIsolateHost isolates the host from the network.
func (s *SecOps) executeIsolateHost(confirm bool, autoExpireSeconds int) SecActionResult {
	result, err := secops.IsolateHost(confirm, autoExpireSeconds)
	if err != nil {
		common.LogWarn("IsolateHost failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// KillProcess requests a safety handshake for terminating a process.
func (s *SecOps) KillProcess(pid int) common.ActionPreview {
	id := common.GetHandshakeRegistry().Register("KillProcess", map[string]interface{}{"pid": pid})

	return common.ActionPreview{
		HandshakeID: id,
		Action:      "Kill Process",
		Description: fmt.Sprintf("Forcefully terminate process with PID %d", pid),
		Risks:       []string{"Unsaved data in the application may be lost", "System instability if a critical service is terminated"},
		Rollback:    "Cannot be undone. Application must be restarted manually.",
	}
}

// executeKillProcess force-kills a process by PID.
func (s *SecOps) executeKillProcess(pid int) SecActionResult {
	result, err := secops.KillProcess(pid)
	if err != nil {
		common.LogWarn("KillProcess failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// BlockIP requests a safety handshake for blocking an IP address.
func (s *SecOps) BlockIP(ip string) common.ActionPreview {
	id := common.GetHandshakeRegistry().Register("BlockIP", map[string]interface{}{"ip": ip})

	return common.ActionPreview{
		HandshakeID: id,
		Action:      "Block IP Address",
		Description: fmt.Sprintf("Add a host-level firewall rule to block all traffic from: %s", ip),
		Risks:       []string{"May block legitimate traffic if IP is spoofed or shared", "Increases firewall rule complexity"},
		Rollback:    "Can be removed manually via the Firewall management tab.",
	}
}

// executeBlockIP blocks an IP address via firewall.
func (s *SecOps) executeBlockIP(ip string) SecActionResult {
	result, err := secops.BlockIP(ip)
	if err != nil {
		common.LogWarn("BlockIP failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// DisableAccount requests a safety handshake for disabling a local user account.
func (s *SecOps) DisableAccount(username string) common.ActionPreview {
	id := common.GetHandshakeRegistry().Register("DisableAccount", map[string]interface{}{"username": username})

	return common.ActionPreview{
		HandshakeID: id,
		Action:      "Disable Account",
		Description: fmt.Sprintf("Instantly disable the local user account: '%s'", username),
		Risks:       []string{"User will be unable to log in", "May disrupt active sessions or scheduled tasks running under this account"},
		Rollback:    "Can be re-enabled via the Identity & Access tab or Computer Management.",
	}
}

// executeDisableAccount disables a local user account.
func (s *SecOps) executeDisableAccount(username string) SecActionResult {
	result, err := secops.DisableAccount(username)
	if err != nil {
		common.LogWarn("DisableAccount failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// CaptureEvidence requests a safety handshake for collecting forensic evidence.
func (s *SecOps) CaptureEvidence() common.ActionPreview {
	id := common.GetHandshakeRegistry().Register("CaptureEvidence", nil)

	return common.ActionPreview{
		HandshakeID: id,
		Action:      "Capture Evidence",
		Description: "Perform a high-density snapshot of volatile system state (processes, connections, memory strings).",
		Risks:       []string{"Temporary CPU spike during collection", "Privacy risk: will collect metadata from all running processes"},
		Rollback:    "Non-destructive. Collected evidence can be deleted manually.",
	}
}

// executeCaptureEvidence collects forensic evidence into a summary.
func (s *SecOps) executeCaptureEvidence() SecActionResult {
	result, err := secops.CaptureEvidence()
	if err != nil {
		common.LogWarn("CaptureEvidence failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// ExportForensicBundle requests a safety handshake for exporting evidence.
func (s *SecOps) ExportForensicBundle() common.ActionPreview {
	id := common.GetHandshakeRegistry().Register("ExportForensicBundle", nil)

	return common.ActionPreview{
		HandshakeID: id,
		Action:      "Export Forensic Bundle",
		Description: "Package and compress all captured evidence into a portable .zip archive.",
		Risks:       []string{"Disk space consumption", "Exposure risk if bundle is stored in an unencrypted location"},
		Rollback:    "Non-destructive. Delete the resulting bundle manually.",
	}
}

// executeExportForensicBundle exports evidence to a file.
func (s *SecOps) executeExportForensicBundle() SecActionResult {
	result, err := secops.ExportForensicBundle()
	if err != nil {
		common.LogWarn("ExportForensicBundle failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}
