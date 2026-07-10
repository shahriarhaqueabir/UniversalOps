package netops

import (
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// PingResult holds the results of a ping operation.
type PingResult struct {
	Target   string
	IP       string
	Sent     int
	Received int
	Lost     int
	Min      time.Duration
	Max      time.Duration
	Avg      time.Duration
	Jitter   time.Duration
	TTL      int
}

// Ping sends ICMP echo requests to a target host.
// On Windows, it falls back to using the system ping.exe.
func Ping(target string, count int) (*PingResult, error) {
	// Windows does not allow raw ICMP sockets without Administrator
	// privileges, so skip directly to the system ping.exe fallback.
	if runtime.GOOS == "windows" {
		return pingExec(target, count)
	}

	// Try ICMP-based ping first (requires elevated privileges on Linux/macOS)
	result, err := pingICMP(target, count)
	if err == nil && result != nil {
		return result, nil
	}

	// Fallback to system ping command
	return pingExec(target, count)
}

// pingICMP uses golang.org/x/net/icmp to send ping requests.
// This requires elevated privileges on most systems.
func pingICMP(target string, count int) (*PingResult, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("icmp listen: %w", err)
	}
	defer conn.Close()

	// Resolve target
	ipAddr, err := resolveTarget(target)
	if err != nil {
		return nil, fmt.Errorf("resolve target: %w", err)
	}

	result := &PingResult{
		Target:   target,
		IP:       ipAddr,
		Sent:     count,
		Received: 0,
	}

	var rtts []time.Duration

	for i := 0; i < count; i++ {
		// Create ICMP echo message
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   i,
				Seq:  i,
				Data: []byte("HawkwardNetOps"),
			},
		}

		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			continue
		}

		start := time.Now()

		// Set deadline for write/read
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		_, err = conn.WriteTo(msgBytes, &netIPAddr{addr: ipAddr})
		if err != nil {
			continue
		}

		reply := make([]byte, 1500)
		n, _, err := conn.ReadFrom(reply)
		if err != nil {
			continue
		}
		rtt := time.Since(start)

		// Parse reply
		parsedMsg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
		if err != nil {
			continue
		}

		switch parsedMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			result.Received++
			rtts = append(rtts, rtt)
		}
	}

	result.Lost = result.Sent - result.Received

	if len(rtts) > 0 {
		result.Min = rtts[0]
		result.Max = rtts[0]
		var total time.Duration
		for _, rtt := range rtts {
			total += rtt
			if rtt < result.Min {
				result.Min = rtt
			}
			if rtt > result.Max {
				result.Max = rtt
			}
		}
		result.Avg = total / time.Duration(len(rtts))

		// Calculate Jitter: average difference between consecutive RTTs
		if len(rtts) > 1 {
			var jitterTotal time.Duration
			for i := 1; i < len(rtts); i++ {
				diff := rtts[i] - rtts[i-1]
				if diff < 0 {
					diff = -diff
				}
				jitterTotal += diff
			}
			result.Jitter = jitterTotal / time.Duration(len(rtts)-1)
		}
	}

	return result, nil
}

// pingExec uses the system ping command as a fallback.
func pingExec(target string, count int) (*PingResult, error) {
	// Platform-specific ping flags
	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"-n", fmt.Sprintf("%d", count), target}
	} else {
		args = []string{"-c", fmt.Sprintf("%d", count), "-W", "5", target}
	}

	// Ping needs network access, so use custom config with DenyNetworkAccess: false
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	cmd := common.SandboxedCommandWithConfig(cfg, "ping", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ping exec: %w", err)
	}

	return parsePingOutput(target, string(output), count)
}

// parsePingOutput parses the output from the system ping command.
func parsePingOutput(target, output string, count int) (*PingResult, error) {
	result := &PingResult{
		Target: target,
	}

	// Extract IP address
	ipRe := regexp.MustCompile(`\[([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\]|Pinging\s+\S+\s+\[([0-9.]+)\]|Pinging\s+([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\s+with|from\s+([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`)
	if matches := ipRe.FindStringSubmatch(output); len(matches) > 0 {
		for _, m := range matches[1:] {
			if m != "" {
				result.IP = m
				break
			}
		}
	}
	// Also try "Reply from" pattern
	if result.IP == "" {
		replyRe := regexp.MustCompile(`Reply from\s+([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`)
		if matches := replyRe.FindStringSubmatch(output); len(matches) > 1 {
			result.IP = matches[1]
		}
	}

	// Extract sent/received/lost from "Packets: Sent = X, Received = Y, Lost = Z" (Windows)
	// or "Sent = X, Received = Y, Lost = Z" (Linux).
	statsRe := regexp.MustCompile(`(?:Packets:\s*)?Sent\s*=\s*(\d+)[^=]*Received\s*=\s*(\d+)[^=]*Lost\s*=\s*(\d+)`)
	if matches := statsRe.FindStringSubmatch(output); len(matches) > 3 {
		result.Sent, _ = strconv.Atoi(matches[1])
		result.Received, _ = strconv.Atoi(matches[2])
		result.Lost, _ = strconv.Atoi(matches[3])
	}

	// Fallback: count "Reply from" lines for received
	if result.Received == 0 {
		replyCount := strings.Count(output, "Reply from")
		if replyCount > 0 {
			result.Received = replyCount
			result.Sent = count
			result.Lost = count - replyCount
		}
	}

	// Extract timing statistics
	// Pattern: "Minimum = Xms, Maximum = Yms, Average = Zms"
	timingRe := regexp.MustCompile(`Minimum\s*=\s*(\d+)ms[^=]*Maximum\s*=\s*(\d+)ms[^=]*Average\s*=\s*(\d+)ms`)
	if matches := timingRe.FindStringSubmatch(output); len(matches) > 3 {
		minVal, _ := strconv.Atoi(matches[1])
		maxVal, _ := strconv.Atoi(matches[2])
		avgVal, _ := strconv.Atoi(matches[3])
		result.Min = time.Duration(minVal) * time.Millisecond
		result.Max = time.Duration(maxVal) * time.Millisecond
		result.Avg = time.Duration(avgVal) * time.Millisecond
	}

	// Extract TTL
	ttlRe := regexp.MustCompile(`TTL[=:](\d+)`)
	if matches := ttlRe.FindStringSubmatch(output); len(matches) > 1 {
		result.TTL, _ = strconv.Atoi(matches[1])
	}

	return result, nil
}

// resolveTarget resolves a hostname to an IP address string.
func resolveTarget(target string) (string, error) {
	ips, err := net.LookupHost(target)
	if err != nil {
		return "", err
	}
	if len(ips) > 0 {
		return ips[0], nil
	}
	return target, nil
}

// netIPAddr is a helper to wrap a string IP address as a net.Addr.
type netIPAddr struct {
	addr string
}

func (a *netIPAddr) Network() string { return "ip4" }
func (a *netIPAddr) String() string  { return a.addr }
