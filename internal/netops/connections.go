package netops

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ConnectionInfo holds information about a network connection.
type ConnectionInfo struct {
	LocalAddr   string `json:"local_addr"`
	RemoteAddr  string `json:"remote_addr"`
	LocalPort   int    `json:"local_port"`
	RemotePort  int    `json:"remote_port"`
	Protocol    string `json:"protocol"`
	State       string `json:"state"`
	ProcessName string `json:"process_name"`
	PID         int    `json:"pid"`
	IsLocal     bool   `json:"is_local"`
	IsExternal  bool   `json:"is_external"`
}

// SubnetTraffic aggregates connection counts by remote subnet.
type SubnetTraffic struct {
	Subnet  string `json:"subnet"`
	Conns   int    `json:"connections"`
	IsLocal bool   `json:"is_local"`
}

// GetSubnetTraffic groups active connections by remote subnet.
func GetSubnetTraffic() []SubnetTraffic {
	conns, _ := GetConnections()
	counts := make(map[string]int)

	localSubnet := GetLocalSubnet()
	_, localNet, _ := net.ParseCIDR(localSubnet)

	for _, c := range conns {
		if c.State == "LISTEN" { continue }
		host, _, _ := net.SplitHostPort(c.RemoteAddr)
		ip := net.ParseIP(host)
		if ip == nil { continue }

		subnet := "external"
		if localNet != nil && localNet.Contains(ip) {
			subnet = localSubnet
		} else if ip.IsLoopback() {
			subnet = "localhost"
		} else if ip.IsPrivate() {
			subnet = "private_other"
		}

		counts[subnet]++
	}

	var res []SubnetTraffic
	for s, count := range counts {
		res = append(res, SubnetTraffic{
			Subnet: s,
			Conns: count,
			IsLocal: s == localSubnet || s == "localhost",
		})
	}
	return res
}

// GetConnections returns network connection information.
func GetConnections() ([]ConnectionInfo, error) {
	if common.IsWindows() {
		return getConnectionsWindows()
	}
	if common.IsLinux() {
		return getConnectionsLinux()
	}
	return nil, fmt.Errorf("connections not supported on this platform")
}

// getConnectionsWindows runs netstat -ano and parses the output.
func getConnectionsWindows() ([]ConnectionInfo, error) {
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false // netstat queries system network state
	cmd := common.SandboxedCommandWithConfig(cfg, "netstat", "-ano")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	connections, err := parseNetstatOutput(string(output))
	if err != nil {
		return nil, err
	}

	if err := resolveProcessNames(connections); err != nil {
		// Non-fatal: we still return connections with empty process names
		common.LogInfo("failed to resolve process names: %v", err)
	}

	return connections, nil
}

// getConnectionsLinux parses /proc/net/tcp and /proc/net/udp.
func getConnectionsLinux() ([]ConnectionInfo, error) {
	var connections []ConnectionInfo

	tcpConns, err := parseProcNet("/proc/net/tcp", "TCP")
	if err == nil {
		connections = append(connections, tcpConns...)
	}

	tcp6Conns, err := parseProcNet("/proc/net/tcp6", "TCP")
	if err == nil {
		connections = append(connections, tcp6Conns...)
	}

	udpConns, err := parseProcNet("/proc/net/udp", "UDP")
	if err == nil {
		connections = append(connections, udpConns...)
	}

	if len(connections) == 0 {
		return nil, fmt.Errorf("no connections found in /proc/net")
	}

	return connections, nil
}

// parseProcNet parses /proc/net/tcp or /proc/net/udp format.
// Format:
//
//	sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
//	0: 0100007F:0019 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1 0000000000000000 100 0 0 10 0
func parseProcNet(path string, proto string) ([]ConnectionInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Build inode->pid mapping once for all connections
	pidMap := buildProcInodeMap()

	var connections []ConnectionInfo
	scanner := bufio.NewScanner(file)
	firstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if firstLine {
			firstLine = false
			continue // skip header
		}
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 12 {
			continue
		}

		localAddrHex := fields[1]
		remoteAddrHex := fields[2]
		stHex := fields[3]
		inodeStr := fields[9]

		localIP, localPort := parseHexAddr(localAddrHex)
		remoteIP, remotePort := parseHexAddr(remoteAddrHex)

		state := tcpStateToString(stHex)

		pid := 0
		procName := ""
		if pidInfo, ok := pidMap[inodeStr]; ok {
			pid = pidInfo.pid
			procName = pidInfo.name
		}

		conn := ConnectionInfo{
			LocalAddr:   fmt.Sprintf("%s:%d", localIP, localPort),
			RemoteAddr:  fmt.Sprintf("%s:%d", remoteIP, remotePort),
			LocalPort:   localPort,
			RemotePort:  remotePort,
			Protocol:    proto,
			State:       state,
			PID:         pid,
			ProcessName: procName,
		}

		if proto == "UDP" && state == "" {
			conn.State = "UDP"
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// parseHexAddr parses a hex-encoded IP:port from /proc/net.
// Format: "0100007F:0019" → 127.0.0.1:25
func parseHexAddr(hexAddr string) (ip string, port int) {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return "", 0
	}

	ipHex := parts[0]
	portHex := parts[1]

	// Parse port (big-endian hex)
	p, err := strconv.ParseInt(portHex, 16, 32)
	if err != nil {
		return "", 0
	}
	port = int(p)

	// Parse IP (little-endian hex bytes, network byte order)
	// 0100007F → reverse bytes: 7F 00 00 01 → 127.0.0.1
	if len(ipHex) == 8 {
		// IPv4
		ip = fmt.Sprintf("%d.%d.%d.%d",
			hexToUint8(ipHex[6:8]),
			hexToUint8(ipHex[4:6]),
			hexToUint8(ipHex[2:4]),
			hexToUint8(ipHex[0:2]))
	} else if len(ipHex) == 32 {
		// IPv6 - parse 8 groups of 4 hex bytes, reversed
		var groups []string
		for i := 0; i < 32; i += 8 {
			part := ipHex[i : i+8]
			groups = append(groups, fmt.Sprintf("%s%s:%s%s",
				string(part[6:8]), string(part[4:6]),
				string(part[2:4]), string(part[0:2])))
		}
		ip = strings.Join(groups, ":")
	}

	return ip, port
}

// hexToUint8 converts a 2-char hex string to a byte.
func hexToUint8(s string) uint8 {
	var v uint8
	for i := 0; i < 2; i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v = v*16 + uint8(c-'0')
		case c >= 'a' && c <= 'f':
			v = v*16 + uint8(c-'a'+10)
		case c >= 'A' && c <= 'F':
			v = v*16 + uint8(c-'A'+10)
		}
	}
	return v
}

// buildProcInodeMap maps socket inodes to PID/process name.
func buildProcInodeMap() map[string]struct {
	pid  int
	name string
} {
	result := make(map[string]struct {
		pid  int
		name string
	})

	procDirs, err := os.ReadDir("/proc")
	if err != nil {
		return result
	}

	for _, dir := range procDirs {
		if !dir.IsDir() {
			continue
		}
		pidStr := dir.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", pidStr, "fd")
		fdEntries, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		// Get process name
		procName := ""
		cmdline, err := os.ReadFile(filepath.Join("/proc", pidStr, "comm"))
		if err == nil {
			procName = strings.TrimSpace(string(cmdline))
		}

		for _, fd := range fdEntries {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			// Socket links look like: socket:[12345]
			if strings.HasPrefix(link, "socket:[") {
				inode := strings.TrimPrefix(link, "socket:[")
				inode = strings.TrimSuffix(inode, "]")
				result[inode] = struct {
					pid  int
					name string
				}{pid: pid, name: procName}
			}
		}
	}

	return result
}

// tcpStateToString converts a TCP state hex value to a human-readable string.
func tcpStateToString(stHex string) string {
	st, err := strconv.ParseInt(stHex, 16, 8)
	if err != nil {
		return ""
	}

	states := map[int]string{
		0x01: "ESTABLISHED",
		0x02: "SYN_SENT",
		0x03: "SYN_RECV",
		0x04: "FIN_WAIT1",
		0x05: "FIN_WAIT2",
		0x06: "TIME_WAIT",
		0x07: "CLOSE",
		0x08: "CLOSE_WAIT",
		0x09: "LAST_ACK",
		0x0A: "LISTEN",
		0x0B: "CLOSING",
	}

	if state, ok := states[int(st)]; ok {
		return state
	}
	return fmt.Sprintf("STATE(0x%X)", st)
}

// parseNetstatOutput parses the output of netstat -ano.
func parseNetstatOutput(output string) ([]ConnectionInfo, error) {
	lines := strings.Split(output, "\n")
	var connections []ConnectionInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Proto") ||
			strings.HasPrefix(line, "Active") ||
			strings.HasPrefix(line, "  ") {
			continue
		}

		fields := splitNetstatLine(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		if proto != "TCP" && proto != "UDP" {
			continue
		}

		localAddr := fields[1]
		remoteAddr := fields[2]

		_, localPortStr := splitAddrPort(localAddr)
		_, remotePortStr := splitAddrPort(remoteAddr)

		localPort, _ := strconv.Atoi(localPortStr)

		state := ""
		pid := 0

		if proto == "TCP" && len(fields) >= 5 {
			state = fields[3]
			pidStr := fields[len(fields)-1]
			pid, _ = strconv.Atoi(pidStr)
		} else if proto == "UDP" {
			state = "UDP"
			if len(fields) >= 5 {
				pidStr := fields[len(fields)-1]
				pid, _ = strconv.Atoi(pidStr)
			}
		}

		remotePort := 0
		if remotePortStr != "*" && remotePortStr != "" {
			remotePort, _ = strconv.Atoi(remotePortStr)
		}

		conn := ConnectionInfo{
			LocalAddr:  localAddr,
			RemoteAddr: remoteAddr,
			LocalPort:  localPort,
			RemotePort: remotePort,
			Protocol:   proto,
			State:      state,
			PID:        pid,
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// resolveProcessNames enriches connection entries with process names
// using PowerShell Get-Process (locale-independent) instead of tasklist CSV.
func resolveProcessNames(connections []ConnectionInfo) error {
	pidMap, err := getPidMapViaPowerShell()
	if err != nil {
		common.LogInfo("Get-Process failed, trying wmic fallback: %v", err)
		pidMap, err = getPidMapViaWmic()
		if err != nil {
			common.LogInfo("wmic fallback also failed: %v", err)
			return nil // non-fatal — connections still work without process names
		}
	}

	for i := range connections {
		if name, ok := pidMap[connections[i].PID]; ok {
			connections[i].ProcessName = name
		}
	}

	return nil
}

// getPidMapViaPowerShell uses Get-Process (locale-independent).
func getPidMapViaPowerShell() (map[int]string, error) {
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	// #nosec G204 - only querying process list, user input is not involved
	cmd := common.SandboxedCommandWithConfig(cfg, "powershell", "-NoProfile", "-NonInteractive",
		"-Command", "Get-Process | Select-Object Id,ProcessName | ConvertTo-Csv -NoTypeInformation")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Get-Process failed: %w", err)
	}

	pidMap := make(map[int]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "\"Id\"") {
			continue
		}
		// CSV format: "Id","ProcessName"
		parts := strings.Split(line, ",\"")
		if len(parts) < 2 {
			continue
		}
		idStr := strings.Trim(parts[0], "\"")
		nameStr := strings.Trim(parts[len(parts)-1], "\"\r\n")
		pid, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		nameStr = strings.TrimSuffix(nameStr, ".exe")
		pidMap[pid] = nameStr
	}

	return pidMap, nil
}

// getPidMapViaWmic uses wmic as a fallback (locale-independent field selection).
func getPidMapViaWmic() (map[int]string, error) {
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	// #nosec G204 - only querying process list, user input is not involved
	cmd := common.SandboxedCommandWithConfig(cfg, "wmic", "process", "get", "ProcessId,Name", "/FORMAT:CSV")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("wmic failed: %w", err)
	}

	pidMap := make(map[int]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Node") || strings.Contains(line, "ProcessId") {
			continue
		}
		// CSV format: "Computer","Name","ProcessId"
		reader := csv.NewReader(strings.NewReader(line))
		record, err := reader.Read()
		if err != nil || len(record) < 3 {
			continue
		}
		nameStr := strings.TrimSpace(record[1])
		pidStr := strings.TrimSpace(record[2])
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		nameStr = strings.TrimSuffix(nameStr, ".exe")
		pidMap[pid] = nameStr
	}

	return pidMap, nil
}

// splitNetstatLine splits a netstat output line by whitespace.
func splitNetstatLine(line string) []string {
	return strings.Fields(line)
}

// splitAddrPort splits an address:port string.
func splitAddrPort(addrPort string) (host, port string) {
	if strings.HasPrefix(addrPort, "[") {
		closeBracket := strings.Index(addrPort, "]")
		if closeBracket > 0 && closeBracket+1 < len(addrPort) && addrPort[closeBracket+1] == ':' {
			return addrPort[:closeBracket+1], addrPort[closeBracket+2:]
		}
	}

	colonIdx := strings.LastIndex(addrPort, ":")
	if colonIdx >= 0 {
		return addrPort[:colonIdx], addrPort[colonIdx+1:]
	}

	return addrPort, ""
}
