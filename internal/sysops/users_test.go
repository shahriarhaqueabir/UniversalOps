package sysops

import (
	"strings"
	"testing"
)

func TestGetLoggedInUsers(t *testing.T) {
	users, err := GetLoggedInUsers()
	if err != nil {
		// host.Users() may not be implemented on all platforms
		if strings.Contains(err.Error(), "not implemented") {
			t.Skip("host.Users() not implemented on this platform")
		}
		t.Fatalf("GetLoggedInUsers returned error: %v", err)
	}
	t.Logf("Found %d logged-in users", len(users))
}
