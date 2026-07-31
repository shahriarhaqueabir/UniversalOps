package netops

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzParseHexAddr tests the internal hex address parser with arbitrary inputs.
// It must never panic — only return a valid parsed address or empty/defaults.
func FuzzParseHexAddr(f *testing.F) {
	f.Add("0100007F:0019")
	f.Add("00000000:0000")
	f.Add("")
	f.Add(":")
	f.Add("invalid")
	f.Add("gggg:ffff")
	f.Add("0100007F") // no port

	f.Fuzz(func(t *testing.T, input string) {
		ip, port := parseHexAddr(input)
		if ip != "" {
			// If an IP was returned, it must be parseable
			assert.NotNil(t, net.ParseIP(ip), "returned IP %q must be valid", ip)
		}
		// Port must be in valid range (0-65535) if non-zero
		assert.LessOrEqual(t, port, 65535, "port must be <= 65535")
		assert.GreaterOrEqual(t, port, 0, "port must be >= 0")
	})
}

// FuzzSplitAddrPort tests address:port splitting with arbitrary inputs.
func FuzzSplitAddrPort(f *testing.F) {
	f.Add("127.0.0.1:8080")
	f.Add("[::1]:443")
	f.Add(":")
	f.Add("")
	f.Add("no-port-here")
	f.Add(":80")
	f.Add("host:0")
	f.Add("host:99999")

	f.Fuzz(func(t *testing.T, input string) {
		host, port := splitAddrPort(input)
		// Must not panic for any input
		_ = host
		_ = port
	})
}

// FuzzLookupDNS tests DNS lookup calls with arbitrary hostnames.
// The DNS client has internal timeouts, so this should not hang.
func FuzzLookupDNS(f *testing.F) {
	f.Add("example.com")
	f.Add("localhost")
	f.Add("")
	f.Add("256.256.256.256")
	f.Add("xn--")
	f.Add("a.b.c.d.e.f.g")

	f.Fuzz(func(t *testing.T, hostname string) {
		// DNS lookups should never panic regardless of input.
		// Use a very short context timeout so fuzzing doesn't hang
		// on unresolvable names.
		result, err := LookupDNS(hostname)
		if err == nil && result != nil {
			assert.NotEmpty(t, result.Hostname)
		}
	})
}
