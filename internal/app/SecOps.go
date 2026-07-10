package app

import (
	"fmt"

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
		return nil
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
		return nil
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
		return nil
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
	}
}

// GetScheduledTasks returns scheduled tasks.
func (s *SecOps) GetScheduledTasks() []ScheduledTask {
	tasks, err := secops.GetScheduledTasks()
	if err != nil {
		common.LogWarn("GetScheduledTasks failed: %v", err)
		return nil
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
		return nil
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
