package netops

import (
	"testing"
)

func TestIsVPNActive(t *testing.T) {
	// This just verifies it runs without crashing.
	// Actual VPN detection is environment dependent.
	_ = IsVPNActive()
}
