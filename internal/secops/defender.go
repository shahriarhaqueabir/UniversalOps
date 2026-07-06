package secops

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
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
	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command", "Get-MpComputerStatus | ConvertTo-Json -Compress -Depth 2")
		output, err := cmd.Output()
		if err == nil {
			status, parseErr := parseDefenderJSON(string(output))
			if parseErr == nil {
				return status, nil
			}
		}

		// Try with a different approach if the first fails
		cmd2 := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command",
			"$s=Get-MpComputerStatus; $s | Select-Object @{N='Enabled';E={$_.AntivirusEnabled}}, @{N='SignatureAge';E={$_.SignatureAge}}, @{N='SignatureLastUpdated';E={$_.SignatureLastUpdated}}, @{N='QuickScanEndTime';E={$_.QuickScanEndTime}}, @{N='FullScanEndTime';E={$_.FullScanEndTime}}, @{N='RealTimeProtectionEnabled';E={$_.RealTimeProtectionEnabled}}, @{N='CloudProtectionEnabled';E={$_.CloudProtectionEnabled}}, AMServiceEnabled, AntispywareEnabled, NISEnabled | ConvertTo-Json -Compress")
		output2, err2 := cmd2.Output()
		if err2 == nil {
			status, parseErr := parseDefenderJSON(string(output2))
			if parseErr == nil {
				return status, nil
			}
		}

		// Fallback: try WMIC (available in locked-down environments where PowerShell is restricted)
		status, err := defenderWMICFallback()
		if err == nil {
			return status, nil
		}

		return nil, fmt.Errorf("failed to query Windows Defender: all approaches failed")
	}

	return nil, fmt.Errorf("Windows Defender is not available on this platform")
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
