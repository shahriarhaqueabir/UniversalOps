package app

import (
"testing"
)

func TestSecOps_GetFirewallRules(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
rules := s.GetFirewallRules()
if rules == nil {
t.Fatal("GetFirewallRules returned nil, expected non-nil slice")
}
}

func TestSecOps_GetUsers(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
users := s.GetUsers()
if users == nil {
t.Fatal("GetUsers returned nil, expected non-nil slice")
}
if len(users) > 0 && users[0].Username == "" {
t.Error("First user has empty Username")
}
}

func TestSecOps_GetListeningPorts(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
ports := s.GetListeningPorts()
if ports == nil {
t.Fatal("GetListeningPorts returned nil, expected non-nil slice")
}
}

func TestSecOps_GetDefenderStatus(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
status := s.GetDefenderStatus()
if status.Enabled {
t.Log("Defender is enabled")
}
}

func TestSecOps_GetScheduledTasks(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
tasks := s.GetScheduledTasks()
if tasks == nil {
t.Fatal("GetScheduledTasks returned nil, expected non-nil slice")
}
}

func TestSecOps_GetSecurityEvents(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
events := s.GetSecurityEvents()
if events == nil {
t.Fatal("GetSecurityEvents returned nil, expected non-nil slice")
}
}

func TestSecOps_SetFirewallRuleHandshake(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
preview := s.SetFirewallRuleHandshake("nonexistent-rule", true)
if preview.HandshakeID == "" {
t.Error("SetFirewallRuleHandshake returned empty HandshakeID")
}
}

func TestSecOps_GetFirewallStatus(t *testing.T) {
a := NewApp()
so := NewSecOps(a.eventBus)
status := so.GetFirewallStatus()
if status.Profiles == nil {
t.Log("GetFirewallStatus.Profiles is nil (may fail on some systems)")
}
}

func TestSecOps_GetSecurityScore(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
score := s.GetSecurityScore()
if score.Score < 0 || score.Score > 100 {
t.Errorf("SecurityScore out of range: %d", score.Score)
}
if score.Recommendations == nil {
t.Error("SecurityScore.Recommendations is nil, expected non-nil slice")
}
if score.Breakdown == nil {
t.Log("SecurityScore.Breakdown is nil (expected when no data available)")
}
}

func TestSecOps_GetRisks(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
risks := s.GetRisks()
if risks == nil {
t.Fatal("GetRisks returned nil, expected non-nil slice")
}
}

func TestSecOps_GetSecuritySummary(t *testing.T) {
a := NewApp()
s := NewSecOps(a.eventBus)
summary := s.GetSecuritySummary()
if summary.Score < 0 || summary.Score > 100 {
t.Errorf("SecuritySummary.Score out of range: %d", summary.Score)
}
if summary.Risks == nil {
t.Error("SecuritySummary.Risks is nil, expected non-nil slice")
}
if summary.Recommendations == nil {
t.Error("SecuritySummary.Recommendations is nil, expected non-nil slice")
}
}
