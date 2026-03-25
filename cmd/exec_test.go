package cmd

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  []int
	}{
		{"3.11.5", []int{3, 11, 5}},
		{"3.8", []int{3, 8}},
		{"20", []int{20}},
		{"v20.10.0", []int{20, 10, 0}},
		{"3.12.0rc1", []int{3, 12, 0}},
		{"", nil},
		{"abc", nil},
	}

	for _, tt := range tests {
		got := parseVersion(tt.input)
		if tt.want == nil {
			if got != nil {
				t.Errorf("parseVersion(%q) = %v, want nil", tt.input, got)
			}
			continue
		}
		if len(got) != len(tt.want) {
			t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
				break
			}
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b []int
		want int
	}{
		{[]int{3, 11, 5}, []int{3, 11, 5}, 0},
		{[]int{3, 11, 5}, []int{3, 11, 4}, 1},
		{[]int{3, 11, 5}, []int{3, 11, 6}, -1},
		{[]int{3, 12}, []int{3, 11, 99}, 1},
		{[]int{4}, []int{3, 99, 99}, 1},
		{[]int{3, 8}, []int{3, 8, 0}, 0},
		{[]int{3}, []int{3, 0, 0}, 0},
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareVersions(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCheckVersionConstraint(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		want       bool
	}{
		// Python-style constraints
		{"3.11.5", ">=3.8", true},
		{"3.11.5", ">=3.8,<4.0", true},
		{"3.7.2", ">=3.8,<4.0", false},
		{"4.0.0", ">=3.8,<4.0", false},
		{"3.8.0", ">=3.8", true},

		// Node-style constraints
		{"20.10.0", ">=18", true},
		{"16.0.0", ">=18", false},

		// Edge cases
		{"3.11.5", "", true},
		{"", ">=3.8", true}, // Can't parse, allow
	}

	for _, tt := range tests {
		got := checkVersionConstraint(tt.version, tt.constraint)
		if got != tt.want {
			t.Errorf("checkVersionConstraint(%q, %q) = %v, want %v",
				tt.version, tt.constraint, got, tt.want)
		}
	}
}

func TestEvaluateConstraint(t *testing.T) {
	tests := []struct {
		version    []int
		constraint string
		want       bool
	}{
		{[]int{3, 11}, ">=3.8", true},
		{[]int{3, 7}, ">=3.8", false},
		{[]int{3, 8}, ">=3.8", true},
		{[]int{4, 0}, "<4.0", false},
		{[]int{3, 99}, "<4.0", true},
		{[]int{3, 8}, ">3.8", false},
		{[]int{3, 9}, ">3.8", true},
		{[]int{3, 8}, "<=3.8", true},
		{[]int{3, 9}, "<=3.8", false},
		{[]int{3, 8}, "==3.8", true},
		{[]int{3, 9}, "==3.8", false},
		{[]int{3, 8}, "!=3.8", false},
		{[]int{3, 9}, "!=3.8", true},
	}

	for _, tt := range tests {
		got := evaluateConstraint(tt.version, tt.constraint)
		if got != tt.want {
			t.Errorf("evaluateConstraint(%v, %q) = %v, want %v",
				tt.version, tt.constraint, got, tt.want)
		}
	}
}

func TestSetEnv(t *testing.T) {
	env := []string{"HOME=/home/user", "PATH=/usr/bin"}

	// Set existing
	result := setEnv(env, "PATH", "/usr/local/bin")
	found := false
	for _, e := range result {
		if e == "PATH=/usr/local/bin" {
			found = true
		}
	}
	if !found {
		t.Error("setEnv did not update existing PATH")
	}

	// Set new
	result = setEnv(result, "PYTHONPATH", "/tmp")
	found = false
	for _, e := range result {
		if e == "PYTHONPATH=/tmp" {
			found = true
		}
	}
	if !found {
		t.Error("setEnv did not add new PYTHONPATH")
	}
}

func TestEnvPath(t *testing.T) {
	path, err := envPath("abc123def456")
	if err != nil {
		t.Fatalf("envPath failed: %v", err)
	}
	if path == "" {
		t.Error("envPath returned empty string")
	}
	// Should end with the hash
	if !containsSubstring(path, "abc123def456") {
		t.Errorf("envPath(%q) = %q, expected to contain the hash", "abc123def456", path)
	}
	if !containsSubstring(path, ".aigogo/envs") {
		t.Errorf("envPath(%q) = %q, expected to contain .aigogo/envs", "abc123def456", path)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
