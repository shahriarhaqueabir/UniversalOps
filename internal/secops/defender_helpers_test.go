package secops

import (
	"testing"
)

func TestWmicBool(t *testing.T) {
	data := map[string]string{
		"TRUE":  "TRUE",
		"true":  "true",
		"FALSE": "FALSE",
		"false": "false",
		"1":     "1",
		"0":     "0",
	}
	if !wmicBool(data, "TRUE") {
		t.Error("wmicBool should return true for 'TRUE'")
	}
	if !wmicBool(data, "true") {
		t.Error("wmicBool should return true for 'true'")
	}
	if wmicBool(data, "FALSE") {
		t.Error("wmicBool should return false for 'FALSE'")
	}
	if wmicBool(data, "false") {
		t.Error("wmicBool should return false for 'false'")
	}
	if !wmicBool(data, "1") {
		t.Error("wmicBool should return true for '1'")
	}
	if wmicBool(data, "0") {
		t.Error("wmicBool should return false for '0'")
	}
	if wmicBool(data, "MISSING") {
		t.Error("wmicBool should return false for missing key")
	}
}

func TestWmicInt(t *testing.T) {
	data := map[string]string{
		"age":   "42",
		"count": "0",
		"neg":   "-1",
		"txt":   "notanumber",
	}
	v, ok := wmicInt(data, "age")
	if !ok || v != 42 {
		t.Errorf("wmicInt(age) = %d, %v, want 42, true", v, ok)
	}
	v, ok = wmicInt(data, "count")
	if !ok || v != 0 {
		t.Errorf("wmicInt(count) = %d, %v, want 0, true", v, ok)
	}
	v, ok = wmicInt(data, "neg")
	if !ok || v != -1 {
		t.Errorf("wmicInt(neg) = %d, %v, want -1, true", v, ok)
	}
	v, ok = wmicInt(data, "txt")
	if ok {
		t.Errorf("wmicInt(txt) should return false")
	}
	v, ok = wmicInt(data, "MISSING")
	if ok {
		t.Errorf("wmicInt(MISSING) should return false")
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		days int
		want string
	}{
		{0, "Today"},
		{1, "1 day ago"},
		{2, "2 days ago"},
		{7, "7 days ago"},
		{30, "1 month ago"},
		{90, "3 months ago"},
		{365, "12 months ago"},
	}
	for _, tt := range tests {
		got := formatAge(tt.days)
		if got != tt.want {
			t.Errorf("formatAge(%d) = %q, want %q", tt.days, got, tt.want)
		}
	}
}

func TestFormatTimeStr(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15 10:30:00"},
		{"2024-01-15T10:30:00+00:00", "2024-01-15 10:30:00"},
		{"20240115103000.123456-420", "20240115103000.1234"},
		{"invalid", "invalid"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatTimeStr(tt.input)
		if got != tt.want {
			t.Errorf("formatTimeStr(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetJSONBool(t *testing.T) {
	data := map[string]interface{}{
		"trueVal":  true,
		"falseVal": false,
		"nilVal":   nil,
	}
	if !getJSONBool(data, "trueVal") {
		t.Error("getJSONBool(trueVal) should return true")
	}
	if getJSONBool(data, "falseVal") {
		t.Error("getJSONBool(falseVal) should return false")
	}
	if getJSONBool(data, "nilVal") {
		t.Error("getJSONBool(nilVal) should return false")
	}
	if getJSONBool(data, "missing") {
		t.Error("getJSONBool(missing) should return false")
	}
}

func TestGetJSONInt(t *testing.T) {
	data := map[string]interface{}{
		"age":   float64(42),
		"count": float64(0),
	}
	v, ok := getJSONInt(data, "age")
	if !ok || v != 42 {
		t.Errorf("getJSONInt(age) = %d, %v, want 42, true", v, ok)
	}
	v, ok = getJSONInt(data, "count")
	if !ok || v != 0 {
		t.Errorf("getJSONInt(count) = %d, %v, want 0, true", v, ok)
	}
	v, ok = getJSONInt(data, "missing")
	if ok {
		t.Errorf("getJSONInt(missing) should return false")
	}
}

func TestGetJSONString(t *testing.T) {
	data := map[string]interface{}{
		"name":  "TestProduct",
		"empty": "",
		"nil":   nil,
	}
	v, ok := getJSONString(data, "name")
	if !ok || v != "TestProduct" {
		t.Errorf("getJSONString(name) = %q, %v, want TestProduct, true", v, ok)
	}
	v, ok = getJSONString(data, "empty")
	if !ok || v != "" {
		t.Errorf("getJSONString(empty) = %q, %v, want empty string, true", v, ok)
	}
	v, ok = getJSONString(data, "nil")
	if ok {
		t.Errorf("getJSONString(nil) should return false")
	}
	v, ok = getJSONString(data, "missing")
	if ok {
		t.Errorf("getJSONString(missing) should return false")
	}
}
