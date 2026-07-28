package secops

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// FirewallRule represents a firewall rule.
type FirewallRule struct {
	Name       string
	Direction  string // In, Out
	Action     string // Allow, Block
	Protocol   string
	LocalPort  string
	RemotePort string
	RemoteIP   string
	Profile    string // Domain, Private, Public
	Enabled    bool
	IsHighRisk bool
}

// FirewallProfile represents a single firewall profile (Domain/Private/Public).
type FirewallProfile struct {
	Name    string
	Enabled bool
}

// GetFirewallProfiles retrieves firewall profile status (Domain, Private, Public).
func GetFirewallProfiles() ([]FirewallProfile, error) {
	if common.IsWindows() {
		// Use HiddenCommand because Get-NetFirewallProfile needs registry access
		// often restricted by sandboxing.
		cmd := common.HiddenCommand("powershell", "-NoProfile", "-Command",
			"Get-NetFirewallProfile -ErrorAction SilentlyContinue | Select-Object Name,Enabled | ConvertTo-Json -As Array -Depth 1")
		output, err := cmd.Output()
		if err != nil {
			// Fallback: use netsh to check firewall state
			return getFirewallStatusFallback()
		}

		cleaned := common.CleanJSON(string(output))
		if cleaned == "" || cleaned == "null" {
			return nil, fmt.Errorf("empty firewall profile output")
		}

		var raw []map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
			return nil, fmt.Errorf("parse firewall profile json: %w", err)
		}

		profiles := make([]FirewallProfile, 0, len(raw))
		for _, item := range raw {
			var p FirewallProfile
			if name, ok := getJSONString(item, "Name"); ok {
				p.Name = name
			} else {
				continue
			}
			if enabled, ok := getJSONString(item, "Enabled"); ok {
				p.Enabled = strings.EqualFold(enabled, "True")
			}
			profiles = append(profiles, p)
		}

		return profiles, nil
	}

	// On Linux, return default profiles (iptables/nft doesn't have the same concept)
	if common.IsLinux() {
		return []FirewallProfile{
			{Name: "Domain", Enabled: true},
			{Name: "Private", Enabled: true},
			{Name: "Public", Enabled: true},
		}, nil
	}

	return nil, fmt.Errorf("firewall profile query not supported on this platform")
}

// getFirewallStatusFallback uses netsh as a fallback when Get-NetFirewallProfile fails.
func getFirewallStatusFallback() ([]FirewallProfile, error) {
	cmd := common.HiddenCommand("netsh", "advfirewall", "show", "allprofiles", "state")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("netsh firewall fallback failed: %w", err)
	}

	outputStr := string(output)
	profiles := []FirewallProfile{}
	for _, name := range []string{"Domain", "Private", "Public"} {
		// netsh output on English Windows:
		//   Domain Profile Settings:
		//   State                                 ON
		if strings.Contains(outputStr, name+" Profile") {
			enabled := false
			// Find the profile section and check the State line beneath it
			normalized := strings.ReplaceAll(outputStr, "\r\n", "\n")
			lines := strings.Split(normalized, "\n")
			inProfile := false
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, name+" Profile") {
					inProfile = true
					continue
				}
				if inProfile && strings.HasPrefix(trimmed, "State") {
					fields := strings.Fields(trimmed)
					if len(fields) >= 2 && strings.EqualFold(fields[len(fields)-1], "ON") {
						enabled = true
					}
					inProfile = false
				} else if inProfile && trimmed == "" {
					// still in the profile section, keep looking
					continue
				} else if inProfile {
					inProfile = false
				}
			}
			profiles = append(profiles, FirewallProfile{Name: name, Enabled: enabled})
		}
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("could not parse netsh firewall output")
	}
	return profiles, nil
}

// GetFirewallRules retrieves firewall rules from the current platform.
func GetFirewallRules() ([]FirewallRule, error) {
	if common.IsWindows() {
		// Approach 1: netsh verbose (English locale) — pass args separately to avoid cmd/c injection
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "netsh", "advfirewall", "firewall", "show", "rule", "name=all", "verbose")
		output, err := cmd.Output()
		if err == nil {
			rules, parseErr := parseFirewallRules(string(output))
			if parseErr == nil && len(rules) > 0 {
				return rules, nil
			}
		}

		// Approach 2: PowerShell Get-NetFirewallRule (locale-independent)
		if rules, fallbackErr := getFirewallRulesPowerShell(); fallbackErr == nil && len(rules) > 0 {
			return rules, nil
		}

		// Both approaches failed; don't wrap nil err
		if err != nil {
			return nil, fmt.Errorf("netsh failed: %w", err)
		}
		return nil, fmt.Errorf("no firewall rules parsed — check netsh output format (non-English locale?)")
	}

	if common.IsLinux() {
		return getFirewallRulesLinux()
	}

	return nil, fmt.Errorf("firewall query not supported on this platform")
}

// SetFirewallRuleState enables or disables a firewall rule by name.
func SetFirewallRuleState(name string, enable bool) error {
	if !common.ValidFirewallRuleName(name) {
		return fmt.Errorf("invalid firewall rule name: %q", name)
	}

	if common.IsWindows() {
		// Use PowerShell Set-NetFirewallRule — avoids cmd/c string injection risk
		enabledVal := "Disabled"
		if enable {
			enabledVal = "Enabled"
		}
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-NoProfile", "-Command",
			fmt.Sprintf("Set-NetFirewallRule -Name '%s' -Enabled %s", name, enabledVal))
		return cmd.Run()
	}

	return fmt.Errorf("firewall state modification not supported on this platform")
}

// IsolateHost enables or disables host-wide network isolation.
func IsolateHost(isolate bool, autoExpireSeconds int) (*common.SecActionResult, error) {
	if common.IsWindows() {
		// Use netsh with direct args (no cmd/c string concat)
		var policy string
		if isolate {
			policy = "blockinbound,blockoutbound"
		} else {
			policy = "blockinbound,allowoutbound"
		}

		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "netsh", "advfirewall", "set", "allprofiles", "firewallpolicy", policy)
		if err := cmd.Run(); err != nil {
			return nil, err
		}

		msg := "Host network isolation DISABLED"
		if isolate {
			msg = "Host network isolation ENABLED"
		}

		return &common.SecActionResult{
			Success: true,
			Message: msg,
		}, nil
	}

	if common.IsLinux() {
		if isolate {
			// Self-exclusion: ensure the app's own connections survive isolation.
			// Allow loopback and established connections so the GUI remains responsive.
			common.HiddenCommand("iptables", "-I", "INPUT", "1", "-i", "lo", "-j", "ACCEPT").Run()
			common.HiddenCommand("iptables", "-I", "OUTPUT", "1", "-o", "lo", "-j", "ACCEPT").Run()
			common.HiddenCommand("iptables", "-I", "INPUT", "1", "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT").Run()
			common.HiddenCommand("iptables", "-I", "OUTPUT", "1", "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT").Run()

			// Only append DROP rule if not already present
			if common.HiddenCommand("iptables", "-C", "INPUT", "-j", "DROP").Run() != nil {
				_ = common.HiddenCommand("iptables", "-A", "INPUT", "-j", "DROP").Run()
			}
			if common.HiddenCommand("iptables", "-C", "OUTPUT", "-j", "DROP").Run() != nil {
				_ = common.HiddenCommand("iptables", "-A", "OUTPUT", "-j", "DROP").Run()
			}

			// Auto-expire: schedule removal of isolation rules after the specified duration.
			// This prevents permanent lockout if the user forgets to disable isolation.
			if autoExpireSeconds > 0 {
				go func() {
					time.Sleep(time.Duration(autoExpireSeconds) * time.Second)
					// Remove all DROP rules (reverse order)
					for common.HiddenCommand("iptables", "-D", "INPUT", "-j", "DROP").Run() == nil {
					}
					for common.HiddenCommand("iptables", "-D", "OUTPUT", "-j", "DROP").Run() == nil {
					}
					common.LogInfo("Host isolation auto-expired after %d seconds", autoExpireSeconds)
				}()
			}
		} else {
			// Remove ALL matching DROP rules (loop until none left)
			for common.HiddenCommand("iptables", "-D", "INPUT", "-j", "DROP").Run() == nil {
			}
			for common.HiddenCommand("iptables", "-D", "OUTPUT", "-j", "DROP").Run() == nil {
			}
			// Clean up self-exclusion rules added during isolation
			for common.HiddenCommand("iptables", "-D", "INPUT", "-i", "lo", "-j", "ACCEPT").Run() == nil {
			}
			for common.HiddenCommand("iptables", "-D", "OUTPUT", "-o", "lo", "-j", "ACCEPT").Run() == nil {
			}
		}

		return &common.SecActionResult{
			Success: true,
			Message: "Host isolation state changed via iptables",
		}, nil
	}

	return nil, fmt.Errorf("host isolation not supported on this platform")
}

// getFirewallRulesPowerShell retrieves firewall rules via PowerShell Get-NetFirewallRule.
// This is locale-independent and serves as a fallback when netsh verbose parsing fails.
func getFirewallRulesPowerShell() ([]FirewallRule, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-NoProfile", "-Command",
		"Get-NetFirewallRule -All | Select-Object Name,Direction,Action,Enabled,Profile | ConvertTo-Json -As Array -Depth 1")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Get-NetFirewallRule failed: %w", err)
	}

	cleaned := common.CleanJSON(string(output))
	if cleaned == "" || cleaned == "null" {
		return nil, fmt.Errorf("empty firewall rule output from PowerShell")
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		return nil, fmt.Errorf("parse firewall rules json: %w", err)
	}

	rules := make([]FirewallRule, 0, len(raw))
	for _, item := range raw {
		var rule FirewallRule

		if name, ok := getJSONString(item, "Name"); ok {
			rule.Name = name
		} else {
			continue
		}

		if dir, ok := getJSONString(item, "Direction"); ok {
			if strings.EqualFold(dir, "Inbound") {
				rule.Direction = "In"
			} else if strings.EqualFold(dir, "Outbound") {
				rule.Direction = "Out"
			} else {
				rule.Direction = dir
			}
		}

		if action, ok := getJSONString(item, "Action"); ok {
			if strings.EqualFold(action, "Allow") {
				rule.Action = "Allow"
			} else {
				rule.Action = "Block"
			}
		}

		if enabled, ok := getJSONString(item, "Enabled"); ok {
			rule.Enabled = strings.EqualFold(enabled, "True")
		}

		if profile, ok := getJSONString(item, "Profile"); ok {
			rule.Profile = profile
		}

		rule.Protocol = "Any"
		rule.IsHighRisk = false

		rules = append(rules, rule)
	}

	if len(rules) > common.MaxFirewallRules {
		rules = rules[:common.MaxFirewallRules]
	}

	return rules, nil
}

// getFirewallRulesLinux retrieves firewall rules using iptables-save or nft.
func getFirewallRulesLinux() ([]FirewallRule, error) {
	// Try iptables-save first
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "iptables-save")
	output, err := cmd.Output()
	if err == nil {
		return parseIPTablesSave(string(output))
	}

	// Fallback: try nft list ruleset
	cmd2 := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "nft", "list", "ruleset")
	output2, err2 := cmd2.Output()
	if err2 == nil {
		return parseNFTList(string(output2))
	}

	return nil, fmt.Errorf("no firewall tool available (tried iptables-save, nft)")
}

// parseIPTablesSave parses the output of "iptables-save".
func parseIPTablesSave(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	scanner := bufio.NewScanner(strings.NewReader(output))
	currentChain := ""
	currentTable := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Table definition
		if strings.HasPrefix(line, "*") {
			currentTable = strings.TrimPrefix(line, "*")
			continue
		}

		// Chain definition
		if strings.HasPrefix(line, ":") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				currentChain = strings.TrimPrefix(parts[0], ":")
			}
			continue
		}

		// Rule
		if strings.HasPrefix(line, "-A") || strings.HasPrefix(line, "-I") {
			rule := FirewallRule{
				Enabled:   true,
				Direction: "In",
			}

			if strings.Contains(currentChain, "OUTPUT") ||
				strings.Contains(currentChain, "FORWARD") {
				rule.Direction = "Out"
			}

			rule.Name = fmt.Sprintf("%s/%s", currentTable, currentChain)
			rule.Action = "Allow" // default
			rule.Profile = currentTable

			// Parse rule options
			fields := strings.Fields(line)
			for i, f := range fields {
				switch f {
				case "-j", "--jump":
					if i+1 < len(fields) {
						target := fields[i+1]
						if target == "DROP" || target == "REJECT" {
							rule.Action = "Block"
						} else if target == "ACCEPT" {
							rule.Action = "Allow"
						}
					}
				case "-p", "--protocol":
					if i+1 < len(fields) {
						rule.Protocol = strings.ToUpper(fields[i+1])
					}
				case "--dport":
					if i+1 < len(fields) {
						rule.LocalPort = fields[i+1]
					}
				case "--sport":
					if i+1 < len(fields) {
						rule.RemotePort = fields[i+1]
					}
				case "-s", "--source":
					if i+1 < len(fields) {
						rule.RemoteIP = fields[i+1]
					}
				case "-d", "--destination":
					if i+1 < len(fields) && rule.RemoteIP == "" {
						rule.RemoteIP = fields[i+1]
					}
				}
			}

			rules = append(rules, rule)
		}
	}

	if len(rules) > common.MaxFirewallRules {
		rules = rules[:common.MaxFirewallRules]
	}

	return rules, nil
}

// parseNFTList parses the output of "nft list ruleset".
func parseNFTList(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "table") {
			continue
		}
		if strings.HasPrefix(line, "chain") {
			continue
		}
		if strings.HasPrefix(line, "}") {
			continue
		}

		// Parse rule lines
		if strings.Contains(line, "drop") || strings.Contains(line, "reject") {
			rule := FirewallRule{
				Enabled:   true,
				Action:    "Block",
				Direction: "In",
				Protocol:  "any",
			}
			rule.Name = fmt.Sprintf("nft-%d", len(rules)+1)
			rules = append(rules, rule)
		} else if strings.Contains(line, "accept") {
			rule := FirewallRule{
				Enabled:   true,
				Action:    "Allow",
				Direction: "In",
				Protocol:  "any",
			}
			rule.Name = fmt.Sprintf("nft-%d", len(rules)+1)
			rules = append(rules, rule)
		}
	}

	if len(rules) > common.MaxFirewallRules {
		rules = rules[:common.MaxFirewallRules]
	}

	return rules, nil
}

// parseFirewallRules parses the output of "netsh advfirewall firewall show rule name=all verbose".
func parseFirewallRules(output string) ([]FirewallRule, error) {
	var rules []FirewallRule

	blocks := strings.Split(output, "\n\n")

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" || !strings.HasPrefix(block, "Rule Name:") {
			continue
		}

		var rule FirewallRule
		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "---") {
				continue
			}

			colonIdx := strings.Index(line, ":")
			if colonIdx < 0 {
				continue
			}

			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])

			switch key {
			case "Rule Name":
				rule.Name = value
			case "Enabled":
				rule.Enabled = strings.EqualFold(value, "Yes")
			case "Direction":
				rule.Direction = value
			case "Action":
				rule.Action = value
			case "Protocol":
				rule.Protocol = value
			case "LocalPort":
				rule.LocalPort = value
			case "RemotePort":
				rule.RemotePort = value
			case "RemoteIP":
				rule.RemoteIP = value
			case "Profiles":
				rule.Profile = value
			}
		}

		if rule.Name != "" {
			// Vulnerability intelligence: Flag "Allow" rules with "Any" IP on sensitive ports
			if rule.Action == "Allow" && rule.Enabled && (rule.RemoteIP == "Any" || rule.RemoteIP == "*") {
				sensitivePorts := []string{"22", "3389", "445", "139", "21", "23", "3306", "5432", "1433"}
				for _, p := range sensitivePorts {
					if strings.Contains(rule.LocalPort, p) {
						rule.IsHighRisk = true
						break
					}
				}
			}
			rules = append(rules, rule)
		}
	}

	if len(rules) > common.MaxFirewallRules {
		rules = rules[:common.MaxFirewallRules]
	}

	return rules, nil
}
