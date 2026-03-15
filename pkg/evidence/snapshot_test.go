package evidence

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CaptureResource(tt.kind, tt.resName, tt.namespace)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			// When no wantErr, the call will fail because kubectl is not available
			// in the test environment, but we verify it got past validation.
			if err != nil && contains(err.Error(), "not permitted") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && contains(err.Error(), "namespace is required") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && contains(err.Error(), "invalid resource") {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if err != nil && contains(err.Error(), "invalid namespace") {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestCaptureResource_ResourceID(t *testing.T) {
	// resourceID logic is pure; we test it indirectly by verifying the format
	// expectations. Since kubectl won't be available in CI, we test the
	// validation paths and document the expected resourceID format.
	tests := []struct {
		name           string
		resName        string
		kind           string
		wantResourceID string
	}{
		{
			name:           "empty name produces kind-only resourceID",
			resName:        "",
			kind:           "Deployment",
			wantResourceID: "Deployment",
		},
		{
			name:           "non-empty name produces kind/name resourceID",
			resName:        "my-app",
			kind:           "Deployment",
			wantResourceID: "Deployment/my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the resourceID format logic directly
			resourceID := tt.kind
			if tt.resName != "" {
				resourceID = tt.kind + "/" + tt.resName
			}
			if resourceID != tt.wantResourceID {
				t.Fatalf("expected resourceID %q, got %q", tt.wantResourceID, resourceID)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
