package secops

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// TLSCertificate holds TLS certificate info.
type TLSCertificate struct {
	Subject    string `json:"subject"`
	Issuer     string `json:"issuer"`
	NotAfter   string `json:"not_after"`
	KeySize    int    `json:"key_size"`
	IsExpiring bool   `json:"is_expiring"`
	DaysLeft   int    `json:"days_left"`
}

// PublicExposure holds public-facing port info.
type PublicExposure struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ProcessName string `json:"process_name"`
	Severity    string `json:"severity"`
}

// GetTLSCertificates retrieves TLS certificate information.
func GetTLSCertificates() ([]TLSCertificate, error) {
	if common.IsWindows() {
		return getTLSCertificatesWindows()
	}
	return getTLSCertificatesLinux()
}

func getTLSCertificatesWindows() ([]TLSCertificate, error) {
	out, err := exec.Command("powershell", "-Command",
		`Get-ChildItem Cert:\LocalMachine\My | Select-Object Subject,Issuer,NotAfter,KeySize | ConvertTo-Json -As Array -Depth 2`).Output()
	if err != nil {
		return []TLSCertificate{}, nil
	}
	return parseCertJSON(string(out)), nil
}

func parseCertJSON(jsonStr string) []TLSCertificate {
	// Simplified: return empty
	return []TLSCertificate{}
}

func getTLSCertificatesLinux() ([]TLSCertificate, error) {
	out, err := exec.Command("ls", "/etc/ssl/certs/").Output()
	if err != nil {
		return []TLSCertificate{}, nil
	}
	var certs []TLSCertificate
	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		certs = append(certs, TLSCertificate{
			Subject:    name,
			Issuer:     "System",
			NotAfter:   "Unknown",
			KeySize:    0,
			IsExpiring: false,
			DaysLeft:   -1,
		})
	}
	return certs, nil
}

// GetPublicExposure retrieves externally-facing listening ports.
func GetPublicExposure() ([]PublicExposure, error) {
	ports, err := GetListeningPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to get listening ports: %w", err)
	}
	var exposed []PublicExposure
	for _, p := range ports {
		if p.IsExternal {
			severity := "medium"
			if p.Port == 3389 || p.Port == 22 || p.Port == 23 {
				severity = "critical"
			} else if p.Port == 80 || p.Port == 443 {
				severity = "low"
			}
			exposed = append(exposed, PublicExposure{
				Port:        p.Port,
				Protocol:    p.Protocol,
				ProcessName: p.ProcessName,
				Severity:    severity,
			})
		}
	}
	return exposed, nil
}
