package secops

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
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
	// Removed -As Array for PS 5.1 compatibility.
	out, err := common.HiddenCommand("powershell", "-NoProfile", "-Command",
		`Get-ChildItem Cert:\LocalMachine\My -ErrorAction SilentlyContinue | Select-Object Subject,Issuer,NotAfter,KeySize | ConvertTo-Json -Depth 2`).Output()
	if err != nil {
		return []TLSCertificate{}, nil
	}
	return parseCertJSON(string(out)), nil
}

type certEntry struct {
	Subject  string `json:"Subject"`
	Issuer   string `json:"Issuer"`
	NotAfter string `json:"NotAfter"`
	KeySize  int    `json:"KeySize"`
}

func parseCertJSON(jsonStr string) []TLSCertificate {
	var certs []certEntry
	if err := json.Unmarshal([]byte(jsonStr), &certs); err != nil {
		var single certEntry
		if err2 := json.Unmarshal([]byte(jsonStr), &single); err2 != nil {
			return []TLSCertificate{}
		}
		certs = []certEntry{single}
	}

	now := time.Now()
	var result []TLSCertificate
	for _, c := range certs {
		if c.Subject == "" {
			continue
		}
		tlsCert := TLSCertificate{
			Subject:  c.Subject,
			Issuer:   c.Issuer,
			NotAfter: c.NotAfter,
			KeySize:  c.KeySize,
		}
		// Calculate days left and expiry status
		if parsed, err := time.Parse(time.RFC3339, c.NotAfter); err == nil {
			days := int(parsed.Sub(now).Hours() / 24)
			tlsCert.DaysLeft = days
			tlsCert.IsExpiring = days < 30
		} else if parsed, err := time.Parse("2006-01-02T15:04:05Z", c.NotAfter); err == nil {
			days := int(parsed.Sub(now).Hours() / 24)
			tlsCert.DaysLeft = days
			tlsCert.IsExpiring = days < 30
		} else {
			// Try other common formats
			for _, fmt := range []string{"2006-01-02 15:04:05", "01/02/2006", "2006-01-02"} {
				if parsed, err := time.Parse(fmt, c.NotAfter); err == nil {
					days := int(parsed.Sub(now).Hours() / 24)
					tlsCert.DaysLeft = days
					tlsCert.IsExpiring = days < 30
					break
				}
			}
		}
		result = append(result, tlsCert)
	}
	return result
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
