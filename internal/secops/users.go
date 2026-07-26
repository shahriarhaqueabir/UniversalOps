package secops

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// validUsername matches safe username characters: letters, digits, dot, dash, underscore.
// Max length 64 per POSIX and Windows username limits.
var validUsername = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

// validateUsername rejects usernames that contain shell metacharacters or
// unexpected content before they reach exec.Command (SEC-2).
func validateUsername(name string) error {
	if name == "" {
		return fmt.Errorf("username must not be empty")
	}
	if !validUsername.MatchString(name) {
		return fmt.Errorf("invalid username: contains disallowed characters or exceeds length limit")
	}
	return nil
}

// UserInfo represents a local user account.
type UserInfo struct {
	Username             string
	FullName             string
	SID                  string
	Group                string
	IsAdmin              bool
	IsEnabled            bool
	PasswordNeverExpires bool
	LastLogon            string
}

// GetUsers retrieves local user accounts.
func GetUsers() ([]UserInfo, error) {
	if common.IsWindows() {
		// Use HiddenCommand for Windows system tools because 'net user'
		// requires access often restricted by sandboxing.
		cmd := common.HiddenCommand("cmd", "/c", "net user")
		output, err := cmd.Output()
		if err == nil {
			usernames := parseNetUserList(string(output))
			if len(usernames) > 0 {
				return collectUserDetails(usernames)
			}
			return nil, fmt.Errorf("no users found — check 'net user' output format (non-English locale?)")
		}
		return nil, fmt.Errorf("failed to query Windows users: %w", err)
	}

	if common.IsLinux() {
		return getUsersLinux()
	}

	return nil, fmt.Errorf("user query not supported on this platform")
}

// getUsersLinux parses /etc/passwd for local user accounts.
func getUsersLinux() ([]UserInfo, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("cannot open /etc/passwd: %w", err)
	}
	defer file.Close()

	// Read /etc/group for group names
	groupMap, err := parseGroupFile()
	if err != nil {
		groupMap = make(map[string]string) // non-fatal
	}

	// Read /etc/gshadow or /etc/shadow for account status
	disabledUsers := getDisabledUsers()

	var users []UserInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		username := parts[0]
		fullName := strings.Trim(parts[4], ",")
		uidStr := parts[2]
		gidStr := parts[3]
		// shell := parts[6] // not used currently

		uid, err := strconv.Atoi(uidStr)
		if err != nil || uid < 1000 || uid == 65534 {
			// Skip system users (uid < 1000) and nobody (65534)
			continue
		}

		// Determine group
		gid, _ := strconv.Atoi(gidStr)
		groupName := groupMap[gidStr]
		if groupName == "" {
			groupName = fmt.Sprintf("gid:%d", gid)
		}

		isAdmin := groupName == "root" || groupName == "sudo" || groupName == "wheel" || gid == 0
		isEnabled := !disabledUsers[username]

		// SID equivalent: uid number
		sid := fmt.Sprintf("uid:%d", uid)

		users = append(users, UserInfo{
			Username:  username,
			FullName:  fullName,
			SID:       sid,
			Group:     groupName,
			IsAdmin:   isAdmin,
			IsEnabled: isEnabled,
		})
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no regular users found in /etc/passwd")
	}

	return users, nil
}

// parseGroupFile reads /etc/group and returns a map of GID -> group name.
func parseGroupFile() (map[string]string, error) {
	file, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	groupMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			groupMap[parts[2]] = parts[0]
		}
	}

	return groupMap, nil
}

// getDisabledUsers reads /etc/shadow to find disabled/locked accounts.
// Returns a set of usernames that are disabled.
func getDisabledUsers() map[string]bool {
	disabled := make(map[string]bool)
	file, err := os.Open("/etc/shadow")
	if err != nil {
		return disabled
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		username := parts[0]
		// Password field starts with ! or * for disabled/locked accounts
		pwField := parts[1]
		if pwField == "" || pwField == "*" || pwField == "!" || strings.HasPrefix(pwField, "!") {
			disabled[username] = true
		}
	}

	return disabled
}

// GetGroups retrieves local security groups.
func GetGroups() ([]string, error) {
	if common.IsWindows() {
		// Use HiddenCommand.
	cmd := common.HiddenCommand("cmd", "/c", "net localgroup")
		output, err := cmd.Output()
		if err == nil {
			groups := parseNetLocalGroup(string(output))
			if len(groups) > 0 {
				return groups, nil
			}
			return nil, fmt.Errorf("no groups found — check 'net localgroup' output format (non-English locale?)")
		}
		return nil, fmt.Errorf("failed to query Windows groups: %w", err)
	}

	if common.IsLinux() {
		return getGroupsLinux()
	}

	return nil, fmt.Errorf("group query not supported on this platform")
}

// getGroupsLinux parses /etc/group for security groups.
func getGroupsLinux() ([]string, error) {
	file, err := os.Open("/etc/group")
	if err != nil {
		return nil, fmt.Errorf("cannot open /etc/group: %w", err)
	}
	defer file.Close()

	var groups []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 1 && parts[0] != "" {
			groups = append(groups, parts[0])
		}
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("no groups found in /etc/group")
	}

	return groups, nil
}

// parseNetUserList parses the listing output of "net user".
func parseNetUserList(output string) []string {
	var users []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	inTable := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "---") {
			inTable = true
			continue
		}
		if trimmed == "The command completed successfully." {
			break
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "user accounts for") {
			continue
		}
		if !inTable || trimmed == "" {
			continue
		}

		cols := splitBySpaces(trimmed)
		users = append(users, cols...)
	}

	return users
}

// collectUserDetails runs "net user <username>" for each user and parses details.
func collectUserDetails(usernames []string) ([]UserInfo, error) {
	var users []UserInfo

	for _, name := range usernames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if strings.EqualFold(name, "DefaultAccount") ||
			strings.EqualFold(name, "WDAGUtilityAccount") ||
			strings.EqualFold(name, "Guest") {
			continue
		}

		// SEC-2: validate username before use in exec.Command
		if err := validateUsername(name); err != nil {
			common.LogWarn("collectUserDetails: skipping user %q: %v", name, err)
			continue
		}

		// Use HiddenCommand.
	cmd := common.HiddenCommand("cmd", "/c", "net user", name)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		user := parseNetUserDetail(string(output), name)
		users = append(users, user)
	}

	return users, nil
}

// parseNetUserDetail parses the detailed output of "net user <username>".
func parseNetUserDetail(output, username string) UserInfo {
	user := UserInfo{
		Username:  username,
		IsEnabled: true,
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Windows "net user <name>" uses label-based lines without colons.
		// Match known label prefixes and extract the value portion.
		switch {
		case strings.HasPrefix(trimmed, "Full Name"):
			user.FullName = strings.TrimSpace(strings.TrimPrefix(trimmed, "Full Name"))
		case strings.HasPrefix(trimmed, "Account active"):
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "Account active"))
			user.IsEnabled = strings.EqualFold(val, "Yes")
		case strings.HasPrefix(trimmed, "Password expires"):
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "Password expires"))
			user.PasswordNeverExpires = strings.EqualFold(val, "Never")
		case strings.HasPrefix(trimmed, "Last logon"):
			user.LastLogon = strings.TrimSpace(strings.TrimPrefix(trimmed, "Last logon"))
		case strings.HasPrefix(trimmed, "Local Group Memberships"):
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "Local Group Memberships"))
			user.IsAdmin = strings.Contains(val, "*Administrators")
			if strings.Contains(val, "*Administrators") {
				user.Group = "Administrators"
			} else if strings.Contains(val, "*Users") {
				user.Group = "Users"
			} else if strings.Contains(val, "*Guests") {
				user.Group = "Guests"
			} else if strings.Contains(val, "*Remote Desktop") {
				user.Group = "Remote Desktop Users"
			} else {
				user.Group = "Users"
			}
		}
	}

	return user
}

// parseNetLocalGroup parses the output of "net localgroup".
func parseNetLocalGroup(output string) []string {
	var groups []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	inTable := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "---") {
			inTable = true
			continue
		}
		if line == "The command completed successfully." {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "aliases for") {
			continue
		}
		if !inTable || line == "" {
			continue
		}

		cols := splitBySpaces(line)
		groups = append(groups, cols...)
	}

	return groups
}

// splitBySpaces splits a string by whitespace fields.
func splitBySpaces(s string) []string {
	var result []string
	fields := strings.Fields(s)
	for _, f := range fields {
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}
