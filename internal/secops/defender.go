package secops

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// DefenderStatus represents the Windows Defender health status.
type DefenderStatus struct {
	Enabled            bool
	UpToDate           bool
	SignatureAge       string
	LastScan           string
	RealTimeProtection bool
	CloudProtection    bool
	AMServiceEnabled   bool
	AntispywareEnabled bool
	NISEnabled         bool
	QuickScanAge       int
	FullScanAge        int
}

// GetDefenderStatus retrieves Windows Defender status via PowerShell.
// On non-Windows systems it returns an error indicating the feature is unavailable.
func GetDefenderStatus() (*DefenderStatus, error) {
	if !common.IsWindows() {
		return nil, fmt.Errorf("Windows Defender is not available on this platform")
	}

	// Pre-check: verify Defender is installed by looking for MpCmdRun.exe
	defenderPath := filepath.Join(os.Getenv("ProgramFiles"), "Windows Defender", "MpCmdRun.exe")
	if _, err := os.Stat(defenderPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Windows Defender not installed on this system")
	}

	var errs []error

	// Approach 1: Standard Get-MpComputerStatus
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-NoProfile", "-Command",
		"$r=Get-MpComputerStatus -ErrorAction SilentlyContinue; if($r){$r|ConvertTo-Json -Depth 2}else{echo '{}'}")
	output, err := cmd.Output()
	if err != nil {
		errs = append(errs, fmt.Errorf("Get-MpComputerStatus: %w", err))
	} else {
		status, parseErr := parseDefenderJSON(string(output))
		if parseErr != nil {
			errs = append(errs, fmt.Errorf("Get-MpComputerStatus parse: %w", parseErr))
		} else if status != nil && (status.AMServiceEnabled || status.AntispywareEnabled) {
			return status, nil
		} else {
			errs = append(errs, fmt.Errorf("Get-MpComputerStatus returned empty or inactive status"))
		}
	}

	// Approach 2: Try Get-MpPreference as alternative
	cmd3 := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-NoProfile", "-Command",
		"Get-MpPreference -ErrorAction SilentlyContinue | Select-Object DisableRealtimeMonitoring,DisableIOAVProtection,DisableBehaviorMonitoring,DisableScriptScanning | ConvertTo-Json -Depth 2")
	output3, err3 := cmd3.Output()
	if err3 != nil {
		errs = append(errs, fmt.Errorf("Get-MpPreference: %w", err3))
	} else {
		status, parseErr := parseDefenderPreferenceJSON(string(output3))
		if parseErr != nil {
			errs = append(errs, fmt.Errorf("Get-MpPreference parse: %w", parseErr))
		} else if status != nil {
			return status, nil
		}
	}

	// Approach 3: Try WMIC (available in locked-down environments)
	status, wmicErr := defenderWMICFallback()
	if wmicErr != nil {
		errs = append(errs, fmt.Errorf("WMIC: %w", wmicErr))
	} else {
		return status, nil
	}

	return nil, fmt.Errorf("failed to query Windows Defender: all approaches failed: %v", errs)
}

// defenderWMICFallback queries Windows Defender via WMIC as a fallback
// when PowerShell is unavailable.
func defenderWMICFallback() (*DefenderStatus, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "wmic",
		"/namespace:\\\\root\\Microsoft\\Windows\\Defender",
		"path", "MSFT_MpComputerStatus",
		"get", "AntivirusEnabled,AMServiceEnabled,AntispywareEnabled,NISEnabled,RealTimeProtectionEnabled,CloudProtectionEnabled,SignatureAge,QuickScanAge,FullScanAge,QuickScanEndTime,FullScanEndTime",
		"/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("WMIC Defender query failed: %w", err)
	}
	return parseDefenderWMIC(string(output))
}

// parseDefenderWMIC parses the CSV output from WMIC MSFT_MpComputerStatus.
// WMIC CSV format: first row is headers (Node,FieldName,...), second row is data.
func parseDefenderWMIC(csvStr string) (*DefenderStatus, error) {
	reader := csv.NewReader(strings.NewReader(csvStr))
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, fmt.Errorf("failed to parse WMIC CSV output")
	}

	// Build header index (skip first column "Node")
	headers := make([]string, 0, len(records[0]))
	for _, h := range records[0] {
		headers = append(headers, strings.TrimSpace(h))
	}

	// Find the first data row (non-empty node name)
	var dataRow []string
	for i := 1; i < len(records); i++ {
		if len(records[i]) > 0 && strings.TrimSpace(records[i][0]) != "" {
			dataRow = records[i]
			break
		}
	}
	if dataRow == nil {
		return nil, fmt.Errorf("no data row found in WMIC output")
	}

	// Build field map from headers -> values
	data := make(map[string]string)
	for i, h := range headers {
		if i < len(dataRow) {
			data[h] = strings.TrimSpace(dataRow[i])
		}
	}

	status := &DefenderStatus{}

	// Parse boolean fields
	status.Enabled = wmicBool(data, "AntivirusEnabled") || wmicBool(data, "AMServiceEnabled")
	status.AMServiceEnabled = wmicBool(data, "AMServiceEnabled")
	status.AntispywareEnabled = wmicBool(data, "AntispywareEnabled")
	status.NISEnabled = wmicBool(data, "NISEnabled")
	status.RealTimeProtection = wmicBool(data, "RealTimeProtectionEnabled")
	status.CloudProtection = wmicBool(data, "CloudProtectionEnabled")

	// Parse signature age
	if age, ok := wmicInt(data, "SignatureAge"); ok {
		status.SignatureAge = formatAge(age)
		status.UpToDate = age <= 7
	} else {
		status.SignatureAge = "Unknown"
		status.UpToDate = false
	}

	// Parse scan times
	if t := data["QuickScanEndTime"]; t != "" && t != "N/A" {
		status.LastScan = "Quick: " + formatTimeStr(t)
	} else if t := data["FullScanEndTime"]; t != "" && t != "N/A" {
		status.LastScan = "Full: " + formatTimeStr(t)
	} else {
		status.LastScan = "Unknown"
	}

	if age, ok := wmicInt(data, "QuickScanAge"); ok {
		status.QuickScanAge = age
	}
	if age, ok := wmicInt(data, "FullScanAge"); ok {
		status.FullScanAge = age
	}

	return status, nil
}

// wmicBool extracts a boolean from a WMIC field map.
func wmicBool(data map[string]string, key string) bool {
	v, ok := data[key]
	if !ok {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(v))
	return lower == "true" || lower == "1"
}

// wmicInt extracts an integer from a WMIC field map.
func wmicInt(data map[string]string, key string) (int, bool) {
	v, ok := data[key]
	if !ok || v == "" {
		return 0, false
	}
	// WMIC may return floats as strings like "2.0000000000"
	if idx := strings.Index(v, "."); idx >= 0 {
		v = v[:idx]
	}
	var val int
	_, err := fmt.Sscanf(v, "%d", &val)
	if err != nil {
		return 0, false
	}
	return val, true
}

// parseDefenderJSON parses the JSON output from Get-MpComputerStatus.
func parseDefenderJSON(jsonStr string) (*DefenderStatus, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse defender JSON: %w", err)
	}

	status := &DefenderStatus{}

	status.Enabled = getJSONBool(data, "AntivirusEnabled") || getJSONBool(data, "AMServiceEnabled")
	status.AMServiceEnabled = getJSONBool(data, "AMServiceEnabled")
	status.AntispywareEnabled = getJSONBool(data, "AntispywareEnabled")
	status.NISEnabled = getJSONBool(data, "NISEnabled")
	status.RealTimeProtection = getJSONBool(data, "RealTimeProtectionEnabled")
	status.CloudProtection = getJSONBool(data, "CloudProtectionEnabled")

	if age, ok := getJSONInt(data, "SignatureAge"); ok {
		status.SignatureAge = formatAge(age)
		status.UpToDate = age <= 7
	} else {
		status.SignatureAge = "Unknown"
		status.UpToDate = false
	}

	if t, ok := getJSONString(data, "QuickScanEndTime"); ok && t != "" {
		status.LastScan = "Quick: " + formatTimeStr(t)
	} else if t, ok := getJSONString(data, "FullScanEndTime"); ok && t != "" {
		status.LastScan = "Full: " + formatTimeStr(t)
	} else {
		status.LastScan = "Unknown"
	}

	if age, ok := getJSONInt(data, "QuickScanAge"); ok {
		status.QuickScanAge = age
	}
	if age, ok := getJSONInt(data, "FullScanAge"); ok {
		status.FullScanAge = age
	}

	return status, nil
}

// getJSONBool safely extracts a boolean from a JSON map.
func getJSONBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		}
	}
	return false
}

// getJSONInt safely extracts an integer from a JSON map.
func getJSONInt(data map[string]interface{}, key string) (int, bool) {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val), true
		case int:
			return val, true
		case string:
			// Handle "-" or empty string (PowerShell edge case for unset values)
			if val == "-" || val == "" {
				return 0, true
			}
		case map[string]interface{}:
			// Handle {"Value": N} objects from PowerShell
			if inner, ok := val["Value"]; ok {
				return getJSONInt(map[string]interface{}{"tmp": inner}, "tmp")
			}
		}
	}
	return 0, false
}

// getJSONString safely extracts a string from a JSON map.
func getJSONString(data map[string]interface{}, key string) (string, bool) {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case string:
			return val, true
		case map[string]interface{}:
			// Handle {"Value": "something"} objects from PowerShell
			if inner, ok := val["Value"]; ok {
				return getJSONString(map[string]interface{}{"tmp": inner}, "tmp")
			}
		}
	}
	return "", false
}

// formatAge formats an age-in-days value to a human string.
func formatAge(days int) string {
	if days <= 0 {
		return "Today"
	}
	if days == 1 {
		return "1 day ago"
	}
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}
	months := days / 30
	if months == 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", months)
}

// formatTimeStr formats a DateTime string from PowerShell to a short form.
func formatTimeStr(t string) string {
	t = strings.ReplaceAll(t, "T", " ")
	if len(t) > 19 {
		t = t[:19]
	}
	return t
}

// parseDefenderPreferenceJSON parses Get-MpPreference JSON output
// as a fallback when Get-MpComputerStatus is unavailable.
func parseDefenderPreferenceJSON(jsonStr string) (*DefenderStatus, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse defender preference JSON: %w", err)
	}

	status := &DefenderStatus{
		SignatureAge: "Unknown",
		LastScan:     "Unknown",
	}

	// In MpPreference, DisableRealtimeMonitoring being false means it IS enabled
	disableRTM := getJSONBool(data, "DisableRealtimeMonitoring")
	status.RealTimeProtection = !disableRTM
	status.Enabled = true
	status.AMServiceEnabled = true
	status.AntispywareEnabled = true
	status.UpToDate = false

	return status, nil
}
