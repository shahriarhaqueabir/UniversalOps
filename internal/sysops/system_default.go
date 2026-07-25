//go:build !windows

package sysops

import (
	"github.com/shirou/gopsutil/v4/host"
)

// getLoggedInUsersPlatform returns logged-in users on non-Windows platforms
// by reading utmp/wtmp via gopsutil.
func getLoggedInUsersPlatform() ([]LoggedInUser, error) {
	users, err := host.Users()
	if err != nil {
		return nil, err
	}

	var result []LoggedInUser
	for _, u := range users {
		result = append(result, LoggedInUserFromHost(u))
	}
	return result, nil
}
