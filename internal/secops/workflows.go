package secops

import (
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// SecurityReport is a combined security audit report.
type SecurityReport struct {
	FirewallRules  []FirewallRule
	Users          []UserInfo
	Groups         []string
	ListeningPorts []ListeningPort
	Defender       *DefenderStatus
	ScheduledTasks []ScheduledTask
	SecurityEvents []SecurityEvent
}

// RunSecurityAudit collects all security data and returns a combined report.
func RunSecurityAudit() (*SecurityReport, error) {
	report := &SecurityReport{}
	var errs []string

	rules, err := GetFirewallRules()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Firewall: %v", err))
	} else {
		report.FirewallRules = rules
	}

	users, err := GetUsers()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Users: %v", err))
	} else {
		report.Users = users
	}

	groups, err := GetGroups()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Groups: %v", err))
	} else {
		report.Groups = groups
	}

	ports, err := GetListeningPorts()
	if err != nil {
		errs = append(errs, fmt.Sprintf("ListeningPorts: %v", err))
	} else {
		report.ListeningPorts = ports
	}

	status, err := GetDefenderStatus()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Defender: %v", err))
	} else {
		report.Defender = status
	}

	tasks, err := GetScheduledTasks()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Tasks: %v", err))
	} else {
		report.ScheduledTasks = tasks
	}

	events, err := GetSecurityEvents()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Events: %v", err))
	} else {
		report.SecurityEvents = events
	}

	if len(errs) > 0 && len(report.FirewallRules) == 0 && len(report.Users) == 0 {
		return nil, fmt.Errorf("all security checks failed: %s", strings.Join(errs, "; "))
	}

	return report, nil
}

// String returns a plain-text summary of the security report.
func (r *SecurityReport) String() string {
	var b strings.Builder

	b.WriteString("=== Security Audit Report ===\n\n")

	b.WriteString(fmt.Sprintf("FIREWALL: %d rules\n", len(r.FirewallRules)))
	b.WriteString(fmt.Sprintf("USERS: %d accounts, %d groups\n", len(r.Users), len(r.Groups)))

	if r.Defender != nil {
		rt := "disabled"
		if r.Defender.RealTimeProtection {
			rt = "enabled"
		}
		b.WriteString(fmt.Sprintf("DEFENDER: %s, real-time=%s, signatures=%s\n",
			boolStr(r.Defender.Enabled), rt, r.Defender.SignatureAge))
	}

	b.WriteString(fmt.Sprintf("LISTENING PORTS: %d\n", len(r.ListeningPorts)))
	b.WriteString(fmt.Sprintf("SCHEDULED TASKS: %d\n", len(r.ScheduledTasks)))
	b.WriteString(fmt.Sprintf("SECURITY EVENTS: %d\n", len(r.SecurityEvents)))

	// Warning for disabled firewall rules
	disabledCount := 0
	for _, rule := range r.FirewallRules {
		if !rule.Enabled {
			disabledCount++
		}
	}
	if disabledCount > 0 {
		b.WriteString(fmt.Sprintf("\n⚠ WARNING: %d firewall rules are disabled\n", disabledCount))
	}

	// Admin users
	adminCount := 0
	for _, u := range r.Users {
		if u.IsAdmin {
			adminCount++
		}
	}
	if adminCount > 0 {
		b.WriteString(fmt.Sprintf("⚠ %d users have administrator privileges\n", adminCount))
	}

	importantEvents := 0
	for _, event := range r.SecurityEvents {
		if event.Important {
			importantEvents++
		}
	}
	if importantEvents > 0 {
		b.WriteString(fmt.Sprintf("⚠ %d important security events found\n", importantEvents))
	}

	return b.String()
}

// Markdown returns a markdown-formatted security report.
func (r *SecurityReport) Markdown() string {
	var b strings.Builder

	b.WriteString("# 🔒 Security Audit Report\n\n")

	// Summary
	b.WriteString("## Summary\n\n")
	b.WriteString("| Category | Count |\n|----------|-------|\n")
	b.WriteString(fmt.Sprintf("| Firewall Rules | %d |\n", len(r.FirewallRules)))
	b.WriteString(fmt.Sprintf("| User Accounts | %d |\n", len(r.Users)))
	b.WriteString(fmt.Sprintf("| Security Groups | %d |\n", len(r.Groups)))
	b.WriteString(fmt.Sprintf("| Listening Ports | %d |\n", len(r.ListeningPorts)))
	if r.Defender != nil {
		b.WriteString(fmt.Sprintf("| Defender | %s |\n", boolStr(r.Defender.Enabled)))
	}
	b.WriteString(fmt.Sprintf("| Scheduled Tasks | %d |\n", len(r.ScheduledTasks)))
	b.WriteString(fmt.Sprintf("| Security Events | %d |\n", len(r.SecurityEvents)))

	// Firewall rules
	if len(r.FirewallRules) > 0 {
		b.WriteString("\n## Firewall Rules\n\n")
		b.WriteString("| Name | Direction | Action | Protocol | Port | Enabled |\n|------|-----------|--------|----------|------|---------|\n")
		for i, rule := range r.FirewallRules {
			if i >= common.MaxFirewallRulesDisplay {
				b.WriteString(fmt.Sprintf("| ... and %d more | | | | |\n", len(r.FirewallRules)-common.MaxFirewallRulesDisplay))
				break
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %v |\n",
				common.TruncateString(rule.Name, 30),
				rule.Direction, rule.Action, rule.Protocol, rule.LocalPort, rule.Enabled))
		}
	}

	// Users
	if len(r.Users) > 0 {
		b.WriteString("\n## User Accounts\n\n")
		b.WriteString("| Username | Admin | Enabled | Group |\n|----------|-------|---------|-------|\n")
		for _, u := range r.Users {
			b.WriteString(fmt.Sprintf("| %s | %v | %v | %s |\n",
				u.Username, u.IsAdmin, u.IsEnabled, u.Group))
		}
	}

	// Listening ports
	if len(r.ListeningPorts) > 0 {
		b.WriteString("\n## Listening Ports\n\n")
		b.WriteString("| Protocol | Port | Process | PID |\n|----------|------|---------|-----|\n")
		for _, p := range r.ListeningPorts {
			procName := p.ProcessName
			if procName == "" {
				procName = fmt.Sprintf("pid:%d", p.PID)
			}
			b.WriteString(fmt.Sprintf("| %s | %d | %s | %d |\n",
				p.Protocol, p.Port, procName, p.PID))
		}
	}

	// Defender
	if r.Defender != nil {
		b.WriteString("\n## Windows Defender\n\n")
		b.WriteString("| Setting | Status |\n|---------|--------|\n")
		b.WriteString(fmt.Sprintf("| Enabled | %s |\n", boolStr(r.Defender.Enabled)))
		b.WriteString(fmt.Sprintf("| Real-Time Protection | %s |\n", boolStr(r.Defender.RealTimeProtection)))
		b.WriteString(fmt.Sprintf("| Signatures | %s |\n", r.Defender.SignatureAge))
		b.WriteString(fmt.Sprintf("| Last Scan | %s |\n", r.Defender.LastScan))
	}

	// Tasks
	if len(r.ScheduledTasks) > 0 {
		b.WriteString("\n## Scheduled Tasks\n\n")
		b.WriteString("| Name | Status | Next Run |\n|------|--------|----------|\n")
		for i, t := range r.ScheduledTasks {
			if i >= common.MaxScheduledTasks {
				b.WriteString(fmt.Sprintf("| ... and %d more | | |\n", len(r.ScheduledTasks)-common.MaxScheduledTasks))
				break
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				common.TruncateString(t.Name, 35), t.Status, t.NextRun))
		}
	}

	if len(r.SecurityEvents) > 0 {
		b.WriteString("\n## Security Events\n\n")
		b.WriteString("| Event ID | Level | Time | Provider | Message |\n|----------|-------|------|----------|---------|\n")
		for i, event := range r.SecurityEvents {
			if i >= common.MaxScheduledTasks {
				b.WriteString(fmt.Sprintf("| ... and %d more | | | | |\n", len(r.SecurityEvents)-common.MaxScheduledTasks))
				break
			}
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
				event.ID,
				event.Level,
				event.Time,
				common.TruncateString(event.Provider, 30),
				common.TruncateString(strings.ReplaceAll(event.Message, "\n", " "), 50)))
		}
	}

	return b.String()
}
