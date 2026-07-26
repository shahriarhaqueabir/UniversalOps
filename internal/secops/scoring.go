package secops

import (
	"fmt"
	"strings"
	"time"
)

// SecurityScore holds a computed security posture score.
type SecurityScore struct {
	Score           int            `json:"score"`
	Grade           string         `json:"grade"`
	Breakdown       map[string]int `json:"breakdown"`
	Recommendations []string       `json:"recommendations"`
}

// ComputeSecurityScore computes a security posture score from live data sources.
func ComputeSecurityScore() SecurityScore {
	score := 15 // base score
	breakdown := make(map[string]int)
	var recommendations []string

	// ── Defender ──
	defenderScore := 0
	defender, dErr := GetDefenderStatus()
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
	rules, fErr := GetFirewallRules()
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
	userScore := 0
	users, uErr := GetUsers()
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
	portScore := 0
	dangerousPorts := map[int]string{3389: "RDP", 445: "SMB", 23: "Telnet"}
	ports, pErr := GetListeningPorts()
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
	eventScore := 0
	events, eErr := GetSecurityEvents()
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

// SecuritySummary holds a unified security posture overview.
type SecuritySummary struct {
	Score           int      `json:"score"`
	Summary         string   `json:"summary"`
	Risks           []string `json:"risks"`
	Recommendations []string `json:"recommendations"`
	AnalyzedAt      string   `json:"analyzedAt"`
}

// GetSecuritySummary generates a summary of the system security posture.
func GetSecuritySummary() SecuritySummary {
	var risks []string
	var recommendations []string
	score := 15 // base score
	var summaryParts []string

	// ── Firewall ──
	firewallScore := 0
	rules, fErr := GetFirewallRules()
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
	defender, dErr := GetDefenderStatus()
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
	users, uErr := GetUsers()
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
	ports, pErr := GetListeningPorts()
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
	events, eErr := GetSecurityEvents()
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

	// Keep top 5 risks and 5 recommendations
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
