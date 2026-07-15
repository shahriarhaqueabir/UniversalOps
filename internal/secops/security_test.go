package secops

import (
	"testing"
)

func TestParseNetAccounts(t *testing.T) {
	output := `Force user logoff when required by user policy:       No
Minimum password age:                                0 days
Maximum password age:                                42 days
Minimum password length:                             7 characters
Lockout threshold:                                   5 invalid attempts
Lockout duration:                                    30 minutes
Lockout window:                                      30 minutes`

	policy := parseNetAccounts(output)

	if policy.MaxAge != 42 {
		t.Errorf("expected MaxAge 42, got %d", policy.MaxAge)
	}
	if policy.MinLength != 7 {
		t.Errorf("expected MinLength 7, got %d", policy.MinLength)
	}
	if policy.LockoutThreshold != 5 {
		t.Errorf("expected LockoutThreshold 5, got %d", policy.LockoutThreshold)
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Maximum password age:                                42 days", 42},
		{"Minimum password length:                             7 characters", 7},
		{"Lockout threshold:                                   5 invalid attempts", 5},
		{"No limit", 0},
	}
	for _, tt := range tests {
		result := extractNumber(tt.input)
		if result != tt.expected {
			t.Errorf("extractNumber(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
