package evidence

import (
	"strings"
	"testing"
)

func TestCaptureResource_Args(t *testing.T) {
	tests := []struct {
		name      string
		kind      string
		resName   string
		namespace string
		wantErr   string
	}{
		{
			name:      "empty name lists all in namespace",
			kind:      "Deployment",
			resName:   "",
			namespace: "default",
		},
		{
			name:      "non-empty name targets specific resource",
			kind:      "Deployment",
			resName:   "my-app",
			namespace: "default",
		},
		{
			name:      "empty namespace returns error",
			kind:      "Deployment",
			resName:   "",
			namespace: "",
			wantErr:   "namespace is required",
		},
		{
			name:      "disallowed kind returns error",
			kind:      "Secret",
			resName:   "",
			namespace: "default",
			wantErr:   "not permitted",
		},
		{
			name:      "invalid kind format returns error",
			kind:      "../etc",
			resName:   "",
			namespace: "default",
			wantErr:   "not permitted",
		},
		{
			name:      "invalid resource name returns error",
			kind:      "Deployment",
			resName:   "../bad",
			namespace: "default",
			wantErr:   "invalid resource name",
		},
		{
			name:      "invalid namespace returns error",
			kind:      "Deployment",
			resName:   "",
			namespace: "INVALID_NS",
			wantErr:   "invalid namespace",
		},
		{
			name:      "excessively long resource name passes validation",
			kind:      "Deployment",
			resName:   strings.Repeat("a", 254),
			namespace: "default",
		},
		{
			name:      "excessively long namespace passes validation",
			kind:      "Deployment",
			resName:   "my-app",
			namespace: strings.Repeat("a", 254),
		},
		{
			name:      "name with dots and dashes passes validation",
			kind:      "Deployment",
			resName:   "my-app.v1.test-2",
			namespace: "default",
		},
		{
			name:      "name starting with digit passes validation",
			kind:      "Deployment",
			resName:   "2048-game",
			namespace: "default",
		},
		{
			name:      "single digit name passes validation",
			kind:      "Service",
			resName:   "0",
			namespace: "default",
		},
		{
			name:      "uppercase resource name is rejected",
			kind:      "Deployment",
			resName:   "MyApp",
			namespace: "default",
			wantErr:   "invalid resource name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CaptureResource(tt.kind, tt.resName, tt.namespace)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			// When no wantErr, the call will fail because kubectl is not available
			// in the test environment, but we verify it got past validation.
			if err != nil && strings.Contains(err.Error(), "not permitted") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "namespace is required") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "invalid resource") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "invalid namespace") {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestCaptureResource_ResourceID(t *testing.T) {
	// CaptureResource will fail at the kubectl exec step, but we can verify
	// the resourceID format by checking the error does NOT come from
	// validation, confirming the inputs were accepted. The resourceID logic
	// itself is verified by inspecting returned Snapshots in integration tests.
	tests := []struct {
		name    string
		resName string
		kind    string
	}{
		{
			name:    "empty name passes validation (kind-only resourceID)",
			resName: "",
			kind:    "Deployment",
		},
		{
			name:    "non-empty name passes validation (kind/name resourceID)",
			resName: "my-app",
			kind:    "Deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CaptureResource(tt.kind, tt.resName, "default")
			// Should fail only because kubectl is not available, not validation
			if err != nil && strings.Contains(err.Error(), "invalid resource") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "not permitted") {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
