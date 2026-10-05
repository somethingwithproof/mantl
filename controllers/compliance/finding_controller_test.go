package compliance

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuildFindingName_Basic(t *testing.T) {
	tests := []struct {
		name         string
		policy       string
		rule         string
		namespace    string
		resourceName string
		wantContains string
	}{
		{
			name:         "simple names",
			policy:       "require-labels",
			rule:         "check-labels",
			namespace:    "default",
			resourceName: "my-app",
			wantContains: "require-labels",
		},
		{
			name:         "uppercase converted to lowercase",
			policy:       "RequireLabels",
			rule:         "CheckLabels",
			namespace:    "default",
			resourceName: "my-app",
			wantContains: "requirelabels",
		},
		{
			name:         "special characters replaced with dashes",
			policy:       "policy.name/v1",
			rule:         "rule_check",
			namespace:    "ns",
			resourceName: "res",
			wantContains: "policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFindingName(tt.policy, tt.rule, tt.namespace, tt.resourceName)
			if got == "" {
				t.Fatal("buildFindingName returned empty string")
			}
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("buildFindingName() = %q, want it to contain %q", got, tt.wantContains)
			}
			if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(got) {
				t.Errorf("buildFindingName() = %q, want a lowercase DNS label", got)
			}
		})
	}
}

func TestBuildFindingName_LengthLimit(t *testing.T) {
	// Very long inputs should be truncated to 253 chars
	longPolicy := strings.Repeat("a", 100)
	longRule := strings.Repeat("b", 100)
	longNS := strings.Repeat("c", 100)
	longName := strings.Repeat("d", 100)

	got := buildFindingName(longPolicy, longRule, longNS, longName)
	if len(got) > 253 {
		t.Errorf("buildFindingName() length = %d, want <= 253", len(got))
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"", "medium"},
		{"unknown", "medium"},
		{"CRITICAL", "medium"}, // case sensitive - unknown falls back to medium
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSeverity(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSeverity(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
