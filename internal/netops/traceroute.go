package netops

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

const defaultTracerouteMaxHops = 30

// TraceHop holds a single hop from a traceroute run.
type TraceHop struct {
	Number int
	Host   string
	IP     string
	RTTs   []time.Duration
	Timed  bool
}

// TraceRouteResult holds traceroute output parsed into hops.
type TraceRouteResult struct {
	Target string
	Hops   []TraceHop
}

// TraceRoute runs the platform traceroute command and parses its output.
func TraceRoute(target string) (*TraceRouteResult, error) {
	return TraceRouteWithContext(context.Background(), target)
}

// TraceRouteWithContext runs the platform traceroute command with context-based
// cancellation. The context can be used to set a deadline or timeout.
func TraceRouteWithContext(ctx context.Context, target string) (*TraceRouteResult, error) {
	if strings.TrimSpace(target) == "" {
		return nil, fmt.Errorf("target is required")
	}

	name, args := tracerouteCommand(target, defaultTracerouteMaxHops)
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, name, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("traceroute exec: %w", err)
	}

	return parseTracerouteOutput(target, string(output)), nil
}

func tracerouteCommand(target string, maxHops int) (string, []string) {
	if runtime.GOOS == "windows" {
		return "tracert", []string{"-d", "-h", strconv.Itoa(maxHops), target}
	}
	return "traceroute", []string{"-n", "-m", strconv.Itoa(maxHops), target}
}

func parseTracerouteOutput(target, output string) *TraceRouteResult {
	result := &TraceRouteResult{Target: target}
	for _, line := range strings.Split(output, "\n") {
		hop, ok := parseTraceHop(line)
		if ok {
			result.Hops = append(result.Hops, hop)
		}
	}
	return result
}

func parseTraceHop(line string) (TraceHop, bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return TraceHop{}, false
	}

	number, err := strconv.Atoi(strings.TrimSuffix(fields[0], "."))
	if err != nil {
		return TraceHop{}, false
	}

	hop := TraceHop{Number: number}
	for i := 1; i < len(fields); i++ {
		token := fields[i]
		switch {
		case token == "*":
			hop.Timed = true
		case isDurationToken(token) && i+1 < len(fields) && strings.EqualFold(fields[i+1], "ms"):
			if rtt, ok := parseTraceDuration(token); ok {
				hop.RTTs = append(hop.RTTs, rtt)
			}
			i++
		case strings.HasSuffix(strings.ToLower(token), "ms"):
			if rtt, ok := parseTraceDuration(strings.TrimSuffix(strings.ToLower(token), "ms")); ok {
				hop.RTTs = append(hop.RTTs, rtt)
			}
		case token == "<1" && i+1 < len(fields) && strings.EqualFold(fields[i+1], "ms"):
			hop.RTTs = append(hop.RTTs, time.Millisecond)
			i++
		case looksLikeAddress(token):
			if hop.IP == "" {
				hop.IP = strings.Trim(token, "[]")
			}
		case hop.Host == "" && token != "ms":
			hop.Host = strings.Trim(token, "[]")
		}
	}

	return hop, true
}

func parseTraceDuration(value string) (time.Duration, bool) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "<"))
	ms, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	if ms < 1 {
		ms = 1
	}
	return time.Duration(ms * float64(time.Millisecond)), true
}

func isDurationToken(value string) bool {
	_, err := strconv.ParseFloat(strings.TrimPrefix(value, "<"), 64)
	return err == nil
}

func looksLikeAddress(value string) bool {
	value = strings.Trim(value, "[]")
	ipv4 := regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
	return ipv4.MatchString(value) || strings.Contains(value, ":")
}
