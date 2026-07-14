package netops

import "testing"

func TestGetVPNStatus(t *testing.T) {
	status := GetVPNStatus()
	t.Logf("VPN Active=%v Type=%s Interface=%s LocalIP=%s", status.Active, status.Type, status.Interface, status.LocalIP)
}
