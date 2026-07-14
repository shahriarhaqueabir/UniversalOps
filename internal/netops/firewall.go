package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// FirewallRule holds a single firewall rule.
type FirewallRule struct {
	Name        string `json:"name"`
	Direction   string `json:"direction"`
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	Ports       string `json:"ports"`
	Enabled     bool   `json:"enabled"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// GetFirewallRules retrieves the system firewall rules.
func GetFirewallRules() ([]FirewallRule, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all", "dir=in", "status=all")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseNetshOutput(string(output))
	case "linux":
		cmd := exec.Command("sudo", "ufw", "status", "verbose")
		output, err := cmd.CombinedOutput()
		if err == nil && strings.Contains(string(output), "Status: active") {
			return parseUFWOutput(string(output))
		}
		cmd = exec.Command("sudo", "iptables", "-L", "-n", "-v")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseIptablesOutput(string(output))
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func parseNetshOutput(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	var current *FirewallRule
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Rule Name:") {
			if current != nil {
				rules = append(rules, *current)
			}
			current = &FirewallRule{Name: strings.TrimPrefix(trimmed, "Rule Name: "), Enabled: true}
		} else if current != nil {
			if strings.HasPrefix(trimmed, "Enabled:") {
				current.Enabled = strings.ToLower(strings.TrimPrefix(trimmed, "Enabled: ")) == "yes"
			} else if strings.HasPrefix(trimmed, "Direction:") {
				current.Direction = strings.TrimPrefix(trimmed, "Direction: ")
			} else if strings.HasPrefix(trimmed, "Action:") {
				current.Action = strings.TrimPrefix(trimmed, "Action: ")
			} else if strings.HasPrefix(trimmed, "Protocol:") {
				current.Protocol = strings.TrimPrefix(trimmed, "Protocol: ")
			} else if strings.HasPrefix(trimmed, "LocalPort:") {
				current.Ports = strings.TrimPrefix(trimmed, "LocalPort: ")
			}
		}
	}
	if current != nil {
		rules = append(rules, *current)
	}
	return rules, nil
}

func parseUFWOutput(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "[") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 4 {
			rules = append(rules, FirewallRule{Action: fields[0], Protocol: fields[1], Ports: fields[2], Direction: "in", Enabled: true})
		}
	}
	return rules, nil
}

func parseIptablesOutput(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 8 && fields[0] != "pkts" {
			rules = append(rules, FirewallRule{Action: fields[7], Protocol: fields[4], Direction: "in", Enabled: true})
		}
	}
	return rules, nil
}

// ManageFirewallRules adds or deletes a firewall rule.
func ManageFirewallRules(action string, rule FirewallRule) error {
	switch runtime.GOOS {
	case "windows":
		switch action {
		case "add":
			_, err := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
				"name="+rule.Name, "dir="+rule.Direction, "action="+rule.Action,
				"protocol="+rule.Protocol, "localport="+rule.Ports).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+rule.Name).CombinedOutput()
			return err
		}
	case "linux":
		switch action {
		case "add":
			_, err := exec.Command("sudo", "ufw", "allow", rule.Ports).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("sudo", "ufw", "delete", "allow", rule.Ports).CombinedOutput()
			return err
		}
	}
	return fmt.Errorf("unknown action: %s", action)
}
