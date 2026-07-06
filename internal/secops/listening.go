package secops

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ListeningPort represents a network port in LISTENING state.
type ListeningPort struct {
	Port        int
	Protocol    string // TCP, UDP
	ProcessName string
	PID         int
	State       string
}

// GetListeningPorts retrieves all ports in LISTENING state.
func GetListeningPorts() ([]ListeningPort, error) {
	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "cmd", "/c", "netstat -ano")
		output, err := cmd.Output()
		if err == nil {
			ports := parseListeningPorts(string(output))
			if len(ports) > 0 {
				ports = resolveProcessNames(ports)
				return ports, nil
			}
		}
		return nil, fmt.Errorf("failed to query listening ports on Windows: %w", err)
	}

	if common.IsLinux() {
		return getListeningPortsLinux()
	}

	return nil, fmt.Errorf("listening port query not supported on this platform")
}

// getListeningPortsLinux uses ss -tulnp to list listening ports.
func getListeningPortsLinux() ([]ListeningPort, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "ss", "-tulnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ss command failed: %w", err)
	}

	return parseSSOutput(string(output))
}

// parseSSOutput parses the output of "ss -tulnp".
// Format: Netid State  Recv-Q Send-Q  Local Address:Port  Peer Address:Port  Process
func parseSSOutput(output string) ([]ListeningPort, error) {
	var ports []ListeningPort
	seen := make(map[string]bool)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Netid") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		netid := strings.ToUpper(fields[0])
		if netid != "TCP" && netid != "UDP" && netid != "TCP6" && netid != "UDP6" {
			continue
		}

		proto := "TCP"
		if strings.HasPrefix(netid, "UDP") {
			proto = "UDP"
		}

		localField := fields[4]
		// Extract port from local address: [::]:port or 0.0.0.0:port or *:port
		portStr := extractPortFromSS(localField)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		// Extract PID and process name from the process field (last field)
		pid := 0
		procName := ""
		if len(fields) > 5 {
			procField := fields[len(fields)-1]
			if strings.HasPrefix(procField, "users:((") {
				// users:(("process_name",pid=1234,...))
				inner := strings.TrimPrefix(procField, "users:((")
				inner = strings.TrimSuffix(inner, "))")
				parts := strings.Split(inner, ",")
				if len(parts) >= 2 {
					procName = strings.Trim(parts[0], "\"")
					pidStr := strings.TrimPrefix(parts[1], "pid=")
					pid, _ = strconv.Atoi(pidStr)
				}
			}
		}

		key := fmt.Sprintf("%s:%d", proto, port)
		if seen[key] {
			continue
		}
		seen[key] = true

		ports = append(ports, ListeningPort{
			Port:        port,
			Protocol:    proto,
			ProcessName: procName,
			PID:         pid,
			State:       "LISTENING",
		})
	}

	return ports, nil
}

// extractPortFromSS extracts the port from an ss address field.
func extractPortFromSS(addr string) string {
	// Handle [::]:port
	if strings.HasPrefix(addr, "[") {
		closeBracket := strings.LastIndex(addr, "]")
		if closeBracket >= 0 && closeBracket+1 < len(addr) && addr[closeBracket+1] == ':' {
			return addr[closeBracket+2:]
		}
		return ""
	}
	// Handle *:port
	if strings.HasPrefix(addr, "*:") {
		return addr[2:]
	}
	// Handle ip:port
	colonIdx := strings.LastIndex(addr, ":")
	if colonIdx >= 0 {
		return addr[colonIdx+1:]
	}
	return ""
}

// parseListeningPorts parses netstat -ano output for LISTENING ports.
func parseListeningPorts(output string) []ListeningPort {
	var ports []ListeningPort
	seen := make(map[string]bool)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		if proto != "TCP" && proto != "UDP" {
			continue
		}

		localAddr := fields[1]
		state := ""
		var pidStr string

		if proto == "TCP" {
			if len(fields) < 5 {
				continue
			}
			state = fields[3]
			pidStr = fields[4]

			if !strings.EqualFold(state, "LISTENING") {
				continue
			}
		} else {
			if len(fields) < 4 {
				continue
			}
			state = "LISTENING"
			pidStr = fields[3]
		}

		portStr := extractPort(localAddr)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			pid = 0
		}

		key := fmt.Sprintf("%s:%d:%d", proto, port, pid)
		if seen[key] {
			continue
		}
		seen[key] = true

		ports = append(ports, ListeningPort{
			Port:     port,
			Protocol: proto,
			PID:      pid,
			State:    state,
		})
	}

	return ports
}

// extractPort extracts the port number from an address string.
func extractPort(addr string) string {
	if strings.HasPrefix(addr, "[") {
		closeBracket := strings.LastIndex(addr, "]")
		if closeBracket < 0 {
			return ""
		}
		if closeBracket+1 < len(addr) && addr[closeBracket+1] == ':' {
			return addr[closeBracket+2:]
		}
		return ""
	}

	colonIdx := strings.LastIndex(addr, ":")
	if colonIdx < 0 {
		return ""
	}
	return addr[colonIdx+1:]
}

// resolveProcessNames resolves PID to process name using tasklist.
func resolveProcessNames(ports []ListeningPort) []ListeningPort {
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false // tasklist communicates with the system process table
	cmd := common.SandboxedCommandWithConfig(cfg, "cmd", "/c", "tasklist /FO CSV /NH")
	output, err := cmd.Output()
	if err != nil {
		for i := range ports {
			if ports[i].PID > 0 {
				ports[i].ProcessName = getProcessNameByPID(ports[i].PID)
			}
		}
		return ports
	}

	pidMap := parseTasklistCSV(string(output))
	for i := range ports {
		if name, ok := pidMap[ports[i].PID]; ok {
			ports[i].ProcessName = name
		} else if ports[i].PID > 0 {
			ports[i].ProcessName = getProcessNameByPID(ports[i].PID)
		}
	}

	return ports
}

// parseTasklistCSV parses the CSV output of tasklist /FO CSV /NH.
func parseTasklistCSV(output string) map[int]string {
	pidMap := make(map[int]string)
	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		return pidMap
	}

	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		name := strings.TrimSpace(record[0])
		pidStr := strings.TrimSpace(record[1])
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		name = strings.TrimSuffix(name, ".exe")
		pidMap[pid] = name
	}

	return pidMap
}

// getProcessNameByPID resolves a single PID to a process name.
func getProcessNameByPID(pid int) string {
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	cmd := common.SandboxedCommandWithConfig(cfg, "cmd", "/c", fmt.Sprintf("tasklist /FI \"PID eq %d\" /FO CSV /NH", pid))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("pid:%d", pid)
	}

	reader := csv.NewReader(strings.NewReader(string(output)))
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 || len(records[0]) < 1 {
		return fmt.Sprintf("pid:%d", pid)
	}

	name := strings.TrimSpace(records[0][0])
	name = strings.TrimSuffix(name, ".exe")
	return name
}
