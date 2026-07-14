package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/secops"
)

// SecOps exposes security operations bindings to the frontend.
type SecOps struct {
	app *App
}

// NewSecOps creates a new SecOps facade.
func NewSecOps(app *App) *SecOps {
	return &SecOps{app: app}
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

// SetFirewallRuleState enables or disables a firewall rule.
func (s *SecOps) SetFirewallRuleState(name string, enable bool) bool {
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
	s.app.eventBus.Emit(common.NewEvent(
		common.CatSecurity,
		common.EventInfo,
		"secops",
		"Firewall rule changed",
		fmt.Sprintf("Firewall rule '%s' %s", name, action),
	))
	return true
}

// GetSecurityScore computes a security posture score from existing data sources.
func (s *SecOps) GetSecurityScore() SecurityScore {
	score := 15 // base score
	breakdown := make(map[string]int)
	var recommendations []string

	// ── Defender ──
	defenderScore := 0
	defender, dErr := secops.GetDefenderStatus()
	if dErr == nil {
		if defender.Enabled {
			defenderScore += 20
		} else {
			recommendations = append(recommendations, "Enable Windows Defender for endpoint protection.")
		}
		if defender.UpToDate {
			defenderScore += 10
		} else {
			recommendations = append(recommendations, "Update Defender signatures — definitions are outdated.")
		}
		if defender.LastScan != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", defender.LastScan); err == nil {
				if time.Since(t) <= 7*24*time.Hour {
					defenderScore += 5
				} else {
					recommendations = append(recommendations, "Run a Defender scan — last scan was over 7 days ago.")
				}
			} else if t, err := time.Parse(time.RFC3339, defender.LastScan); err == nil {
				if time.Since(t) <= 7*24*time.Hour {
					defenderScore += 5
				} else {
					recommendations = append(recommendations, "Run a Defender scan — last scan was over 7 days ago.")
				}
			}
		}
	}
	breakdown["Defender"] = defenderScore
	score += defenderScore

	// ── Firewall ──
	firewallScore := 0
	rules, fErr := secops.GetFirewallRules()
	if fErr == nil {
		if len(rules) > 0 {
			allDisabled := true
			highRiskCount := 0
			for _, r := range rules {
				if r.Enabled {
					allDisabled = false
				}
				if r.IsHighRisk {
					highRiskCount++
				}
			}
			if !allDisabled {
				firewallScore += 15
			} else {
				recommendations = append(recommendations, "Enable firewall rules — no active rules detected.")
			}
			if highRiskCount == 0 {
				firewallScore += 5
			} else {
				recommendations = append(recommendations, fmt.Sprintf("Review %d high-risk firewall rule(s) and tighten access policies.", highRiskCount))
			}
		} else {
			recommendations = append(recommendations, "Configure firewall rules — no rules found.")
		}
	}
	breakdown["Firewall"] = firewallScore
	score += firewallScore

	// ── Users ──
	userScore := 10
	users, uErr := secops.GetUsers()
	if uErr == nil {
		adminCount := 0
		for _, u := range users {
			if u.IsAdmin && u.IsEnabled {
				adminCount++
			}
		}
		if adminCount <= 2 {
			userScore = 10
		} else {
			userScore = 10 - 3*(adminCount-2)
			if userScore < -10 {
				userScore = -10
			}
			recommendations = append(recommendations, fmt.Sprintf("Reduce admin accounts — %d enabled admins found (recommend ≤2).", adminCount))
		}
	}
	breakdown["Users"] = userScore
	score += userScore

	// ── Listening Ports ──
	portScore := 10
	dangerousPorts := map[int]string{3389: "RDP", 445: "SMB", 23: "Telnet"}
	ports, pErr := secops.GetListeningPorts()
	if pErr == nil {
		dangerousCount := 0
		for _, p := range ports {
			if p.IsExternal {
				if svc, ok := dangerousPorts[p.Port]; ok {
					dangerousCount++
					if dangerousCount == 1 {
						recommendations = append(recommendations, fmt.Sprintf("Close or restrict external-facing %s (port %d).", svc, p.Port))
					}
				}
			}
		}
		if dangerousCount > 0 {
			portScore = 10 - 5*dangerousCount
			if portScore < -10 {
				portScore = -10
			}
		}
	}
	breakdown["Ports"] = portScore
	score += portScore

	// ── Security Events ──
	eventScore := 10
	events, eErr := secops.GetSecurityEvents()
	if eErr == nil {
		failedLogins := 0
		for _, e := range events {
			if strings.Contains(strings.ToLower(e.Message), "failed") ||
				strings.Contains(strings.ToLower(e.Message), "logon") && strings.Contains(strings.ToLower(e.Message), "fail") {
				failedLogins++
			}
		}
		if failedLogins > 10 {
			penalty := failedLogins - 10
			if penalty > 10 {
				penalty = 10
			}
			eventScore = 10 - penalty
			recommendations = append(recommendations, fmt.Sprintf("Investigate %d failed login(s) in recent security events.", failedLogins))
		}
	}
	breakdown["Events"] = eventScore
	score += eventScore

	// ── Cap & Grade ──
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	grade := "F"
	switch {
	case score >= 90:
		grade = "A"
	case score >= 75:
		grade = "B"
	case score >= 60:
		grade = "C"
	case score >= 40:
		grade = "D"
	}

	// Keep top 3 recommendations
	if len(recommendations) > 3 {
		recommendations = recommendations[:3]
	}

	return SecurityScore{
		Score:           score,
		Grade:           grade,
		Breakdown:       breakdown,
		Recommendations: recommendations,
	}
}

// GetFirewallStatus returns the global firewall on/off status and per-profile status.
func (s *SecOps) GetFirewallStatus() FirewallStatus {
	profiles, err := secops.GetFirewallProfiles()
	if err != nil {
		common.LogWarn("GetFirewallStatus failed: %v", err)
		return FirewallStatus{}
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
	var risks []string
	var recommendations []string
	score := 15 // base score
	var summaryParts []string

	// ── Firewall ──
	firewallScore := 0
	rules, fErr := secops.GetFirewallRules()
	if fErr == nil {
		if len(rules) > 0 {
			allDisabled := true
			highRiskCount := 0
			highRiskNames := make([]string, 0)
			for _, r := range rules {
				if r.Enabled {
					allDisabled = false
				}
				if r.IsHighRisk {
					highRiskCount++
					highRiskNames = append(highRiskNames, r.Name)
				}
			}
			if !allDisabled {
				firewallScore += 15
			} else {
				risks = append(risks, "Firewall rules are all disabled — no active filtering in place.")
				recommendations = append(recommendations, "Enable firewall rules to filter inbound/outbound traffic.")
			}
			if highRiskCount == 0 {
				firewallScore += 5
			} else {
				risks = append(risks, fmt.Sprintf("%d high-risk firewall rule(s) allow unrestricted access on sensitive ports.", highRiskCount))
				recommendations = append(recommendations, fmt.Sprintf("Tighten access policies on %d high-risk rule(s).", highRiskCount))
			}
		} else {
			risks = append(risks, "No firewall rules configured — system has no network filtering.")
			recommendations = append(recommendations, "Configure firewall rules to control network access.")
		}
	}
	breakdown := firewallScore

	// ── Defender ──
	defenderScore := 0
	defender, dErr := secops.GetDefenderStatus()
	if dErr == nil {
		if defender.Enabled {
			defenderScore += 20
		} else {
			risks = append(risks, "Windows Defender is disabled — no endpoint protection active.")
			recommendations = append(recommendations, "Enable Windows Defender for endpoint protection.")
		}
		if defender.UpToDate {
			defenderScore += 10
		} else {
			risks = append(risks, fmt.Sprintf("Defender signatures are outdated (age: %s) — known threats may go undetected.", defender.SignatureAge))
			recommendations = append(recommendations, "Update Defender signatures to cover the latest threat definitions.")
		}
		if defender.ThreatsDetected > 0 {
			risks = append(risks, fmt.Sprintf("%d active threat(s) detected by Windows Defender.", defender.ThreatsDetected))
			recommendations = append(recommendations, "Review and remediate detected threats immediately.")
		}
		if defender.LastScan != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", defender.LastScan); err == nil {
				if time.Since(t) <= 7*24*time.Hour {
					defenderScore += 5
				} else {
					risks = append(risks, "Last Defender scan was over 7 days ago.")
					recommendations = append(recommendations, "Schedule a Defender scan to ensure recent file coverage.")
				}
			} else if t, err := time.Parse(time.RFC3339, defender.LastScan); err == nil {
				if time.Since(t) <= 7*24*time.Hour {
					defenderScore += 5
				} else {
					risks = append(risks, "Last Defender scan was over 7 days ago.")
					recommendations = append(recommendations, "Schedule a Defender scan to ensure recent file coverage.")
				}
			}
		}
	}
	breakdown += defenderScore

	// ── Users ──
	userScore := 10
	users, uErr := secops.GetUsers()
	if uErr == nil {
		adminCount := 0
		for _, u := range users {
			if u.IsAdmin && u.IsEnabled {
				adminCount++
			}
		}
		if adminCount > 2 {
			userScore = 10 - 3*(adminCount-2)
			if userScore < -10 {
				userScore = -10
			}
			risks = append(risks, fmt.Sprintf("%d administrator accounts detected — excessive admin access increases attack surface.", adminCount))
			recommendations = append(recommendations, fmt.Sprintf("Reduce admin accounts to 2 or fewer (found %d).", adminCount))
		}
	}
	breakdown += userScore

	// ── Listening Ports ──
	portScore := 10
	dangerousPorts := map[int]string{3389: "RDP", 445: "SMB", 23: "Telnet"}
	ports, pErr := secops.GetListeningPorts()
	if pErr == nil {
		dangerousCount := 0
		var exposedServices []string
		for _, p := range ports {
			if p.IsExternal {
				if svc, ok := dangerousPorts[p.Port]; ok {
					dangerousCount++
					exposedServices = append(exposedServices, fmt.Sprintf("%s (port %d)", svc, p.Port))
				}
			}
		}
		if dangerousCount > 0 {
			portScore = 10 - 5*dangerousCount
			if portScore < -10 {
				portScore = -10
			}
			risks = append(risks, fmt.Sprintf("Dangerous services exposed externally: %s.", strings.Join(exposedServices, ", ")))
			recommendations = append(recommendations, "Close or restrict external access on sensitive ports (RDP, SMB, Telnet).")
		}
	}
	breakdown += portScore

	// ── Security Events ──
	eventScore := 10
	events, eErr := secops.GetSecurityEvents()
	if eErr == nil {
		failedLogins := 0
		for _, e := range events {
			if e.ID == 4625 {
				failedLogins++
			}
		}
		if failedLogins > 10 {
			eventScore = 10 - (failedLogins - 10)
			if eventScore < -10 {
				eventScore = -10
			}
			risks = append(risks, fmt.Sprintf("%d failed login attempts detected — possible brute-force activity.", failedLogins))
			recommendations = append(recommendations, "Investigate failed logons and consider enabling account lockout policies.")
		}
	}
	breakdown += eventScore

	// ── Cap score ──
	score += breakdown
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	// ── Build summary text ──
	summaryParts = append(summaryParts, fmt.Sprintf("Security posture score: %d/100.", score))
	if len(risks) == 0 {
		summaryParts = append(summaryParts, "No significant risks detected across firewall, endpoint protection, user accounts, or network exposure.")
	} else {
		summaryParts = append(summaryParts, fmt.Sprintf("%d risk(s) identified across analyzed subsystems.", len(risks)))
	}
	if len(recommendations) > 0 {
		summaryParts = append(summaryParts, "Review the recommendations below to improve your security posture.")
	}

	// Keep top 5 risks and 5 recommendations for the summary panel
	if len(risks) > 5 {
		risks = risks[:5]
	}
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return SecuritySummary{
		Score:           score,
		Summary:         strings.Join(summaryParts, " "),
		Risks:           risks,
		Recommendations: recommendations,
		AnalyzedAt:      time.Now().Format(time.RFC3339),
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

// IsolateHost isolates the host from the network.
// confirm: requires explicit confirmation (default true for backward compat in binding)
// autoExpireSeconds: auto-remove isolation rule after N seconds (0 = no auto-expiry)
func (s *SecOps) IsolateHost(confirm bool, autoExpireSeconds int) SecActionResult {
	result, err := secops.IsolateHost(confirm, autoExpireSeconds)
	if err != nil {
		common.LogWarn("IsolateHost failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// KillProcess force-kills a process by PID.
func (s *SecOps) KillProcess(pid int) SecActionResult {
	result, err := secops.KillProcess(pid)
	if err != nil {
		common.LogWarn("KillProcess failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// BlockIP blocks an IP address via firewall.
func (s *SecOps) BlockIP(ip string) SecActionResult {
	result, err := secops.BlockIP(ip)
	if err != nil {
		common.LogWarn("BlockIP failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// DisableAccount disables a local user account.
func (s *SecOps) DisableAccount(username string) SecActionResult {
	result, err := secops.DisableAccount(username)
	if err != nil {
		common.LogWarn("DisableAccount failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// CaptureEvidence collects forensic evidence into a summary.
func (s *SecOps) CaptureEvidence() SecActionResult {
	result, err := secops.CaptureEvidence()
	if err != nil {
		common.LogWarn("CaptureEvidence failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}

// ExportForensicBundle exports evidence to a file.
func (s *SecOps) ExportForensicBundle() SecActionResult {
	result, err := secops.ExportForensicBundle()
	if err != nil {
		common.LogWarn("ExportForensicBundle failed: %v", err)
		return SecActionResult{Success: false, Error: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message, Error: result.Error}
}
