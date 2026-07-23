package netops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// DoHResult holds the result of a DNS-over-HTTPS test.
type DoHResult struct {
	Server     string  `json:"server"`
	LatencyMs  float64 `json:"latency_ms"`
	Success    bool    `json:"success"`
	ResolvedIP string  `json:"resolved_ip"`
}

// FlushDNSCache flushes the system DNS resolver cache.
func FlushDNSCache() error {
	switch runtime.GOOS {
	case "windows":
		_, err := common.HiddenCommand("ipconfig", "/flushdns").CombinedOutput()
		return err
	case "linux":
		_, err := exec.Command("sudo", "resolvectl", "flush-caches").CombinedOutput()
		if err != nil {
			_, err = exec.Command("sudo", "systemd-resolve", "--flush-caches").CombinedOutput()
		}
		return err
	default:
		return fmt.Errorf("unsupported platform")
	}
}

// ReverseLookup performs a PTR lookup for an IP address.
func ReverseLookup(ip string) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	v4 := parsed.To4()
	if v4 == nil {
		return "", fmt.Errorf("only IPv4 is supported for reverse lookup")
	}
	arpa := fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa.", v4[3], v4[2], v4[1], v4[0])
	msg := new(dns.Msg)
	msg.SetQuestion(arpa, dns.TypePTR)
	msg.RecursionDesired = true
	client := &dns.Client{Timeout: 5 * time.Second}
	resp, _, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	for _, rr := range resp.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			return strings.TrimSuffix(ptr.Ptr, "."), nil
		}
	}
	return "", fmt.Errorf("no PTR record found")
}

// TestDoH tests DNS-over-HTTPS connectivity to a given server.
func TestDoH(server string) DoHResult {
	result := DoHResult{Server: server}
	type dohQuery struct {
		Name string `json:"name"`
		Type int    `json:"type"`
	}
	type dohAnswer struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	type dohMsg struct {
		Status int         `json:"Status"`
		Answer []dohAnswer `json:"Answer"`
	}
	body, _ := json.Marshal(dohQuery{Name: "google.com", Type: 1})
	start := time.Now()
	dohClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, server+"/dns-query", bytes.NewReader(body))
	if err != nil {
		return result
	}
	req.Header.Set("Content-Type", "application/dns-json")
	resp, err := dohClient.Do(req)
	if err != nil {
		return result
	}
	defer resp.Body.Close()
	result.LatencyMs = float64(time.Since(start).Milliseconds())
	var msg dohMsg
	json.NewDecoder(resp.Body).Decode(&msg)
	result.Success = msg.Status == 0 && len(msg.Answer) > 0
	if result.Success {
		result.ResolvedIP = msg.Answer[0].Data
	}
	return result
}
