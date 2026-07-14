package netops

import "testing"

func TestLookupVendor(t *testing.T) {
	tests := []struct{ mac, expected string }{
		{"00:00:0C:01:02:03", "Cisco"},
		{"00-03-FF-AA-BB-CC", "Microsoft"},
		{"b8:27:eb:12:34:56", "Raspberry Pi"},
		{"AA:BB:CC:DD:EE:FF", ""},
	}
	for _, tt := range tests {
		if got := LookupVendor(tt.mac); got != tt.expected {
			t.Errorf("LookupVendor(%q) = %q, want %q", tt.mac, got, tt.expected)
		}
	}
}
