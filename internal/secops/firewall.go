package secops

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
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
}

// GetFirewallRules retrieves firewall rules from the current platform.
func GetFirewallRules() ([]FirewallRule, error) {
	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "cmd", "/c", "netsh advfirewall firewall show rule name=all verbose")
		output, err := cmd.Output()
		if err == nil {
			rules, parseErr := parseFirewallRules(string(output))
			if parseErr == nil && len(rules) > 0 {
				return rules, nil
			}
		}
		return nil, fmt.Errorf("no firewall rules parsed — check netsh output format (non-English locale?): %w", err)
	}

	if common.IsLinux() {
		return getFirewallRulesLinux()
	}

	return nil, fmt.Errorf("firewall query not supported on this platform")
}

// SetFirewallRuleState enables or disables a firewall rule by name.
func SetFirewallRuleState(name string, enable bool) error {
	state := "no"
	if enable {
		state = "yes"
	}

	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "cmd", "/c", "netsh advfirewall firewall set rule name=\""+name+"\" new enable="+state)
		return cmd.Run()
	}

	return fmt.Errorf("firewall state modification not supported on this platform")
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
			rules = append(rules, rule)
		}
	}

	if len(rules) > common.MaxFirewallRules {
		rules = rules[:common.MaxFirewallRules]
	}

	return rules, nil
}
