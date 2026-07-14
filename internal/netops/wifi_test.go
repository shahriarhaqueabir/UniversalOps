package netops

import "testing"

func TestGetWiFiInfo(t *testing.T) {
	info, err := GetWiFiInfo()
	if err != nil {
		t.Logf("GetWiFiInfo error (may be expected without WiFi): %v", err)
		return
	}
	t.Logf("Connected to: %s, Signal: %d%%, Speed: %s", info.SSID, info.Signal, info.Speed)
}
