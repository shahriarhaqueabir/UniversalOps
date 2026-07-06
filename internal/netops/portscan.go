package netops

import (
	"fmt"
	"net"
	"sort"
	"time"
)

// PortResult holds the result of a single port scan.
type PortResult struct {
	Port    int
	Open    bool
	Service string
}

// commonPorts maps well-known port numbers to their service names.
var commonPorts = map[int]string{
	20:    "FTP-data",
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	123:   "NTP",
	143:   "IMAP",
	443:   "HTTPS",
	445:   "SMB",
	993:   "IMAPS",
	995:   "POP3S",
	1433:  "MSSQL",
	1521:  "Oracle DB",
	2049:  "NFS",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	5900:  "VNC",
	6379:  "Redis",
	8080:  "HTTP-Alt",
	8443:  "HTTPS-Alt",
	27017: "MongoDB",
}

// serviceFromPort returns the service name for a port, or "unknown".
func serviceFromPort(port int) string {
	if svc, ok := commonPorts[port]; ok {
		return svc
	}
	return "unknown"
}

// ScanPorts scans specific ports on a host using TCP dial with timeout.
func ScanPorts(host string, ports []int) ([]PortResult, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	var results []PortResult

	for _, port := range ports {
		result := PortResult{
			Port:    port,
			Service: serviceFromPort(port),
		}

		address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err != nil {
			result.Open = false
		} else {
			result.Open = true
			conn.Close()
		}

		results = append(results, result)
	}

	return results, nil
}

// ScanCommonPorts scans a set of well-known ports on a host.
func ScanCommonPorts(host string) ([]PortResult, error) {
	ports := make([]int, 0, len(commonPorts))
	for p := range commonPorts {
		ports = append(ports, p)
	}
	sort.Ints(ports)

	return ScanPorts(host, ports)
}

// DefaultScanPorts returns the default list of common ports for scanning.
func DefaultScanPorts() []int {
	ports := make([]int, 0, len(commonPorts))
	for p := range commonPorts {
		ports = append(ports, p)
	}
	sort.Ints(ports)
	return ports
}
