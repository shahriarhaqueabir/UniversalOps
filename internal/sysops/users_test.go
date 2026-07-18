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
		// If query user failed, it should have been caught by implementation
		t.Fatalf("GetLoggedInUsers returned error: %v", err)
	}
	t.Logf("Found %d logged-in users", len(users))
	for _, u := range users {
		t.Logf("User: %s, Terminal: %s, Host: %s, Started: %s", u.User, u.Terminal, u.Host, u.Started)
	}
}
