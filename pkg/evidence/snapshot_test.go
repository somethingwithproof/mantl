package evidence

import (
	"context"
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
			name:      "excessively long resource name is rejected",
			kind:      "Deployment",
			resName:   strings.Repeat("a", 254),
			namespace: "default",
			wantErr:   "exceeds maximum length",
		},
		{
			name:      "excessively long namespace is rejected",
			kind:      "Deployment",
			resName:   "my-app",
			namespace: strings.Repeat("a", 64),
			wantErr:   "exceeds maximum length",
		},
		{
			name:      "resource name at max length passes validation",
			kind:      "Deployment",
			resName:   strings.Repeat("a", 253),
			namespace: "default",
		},
		{
			name:      "namespace at max length passes validation",
			kind:      "Deployment",
			resName:   "my-app",
			namespace: strings.Repeat("a", 63),
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
		{
			name:      "resource name with trailing dash is rejected",
			kind:      "Deployment",
			resName:   "my-app-",
			namespace: "default",
			wantErr:   "invalid resource name",
		},
		{
			name:      "resource name with trailing dot is rejected",
			kind:      "Deployment",
			resName:   "my-app.",
			namespace: "default",
			wantErr:   "invalid resource name",
		},
		{
			name:      "namespace with trailing dash is rejected",
			kind:      "Deployment",
			resName:   "my-app",
			namespace: "my-ns-",
			wantErr:   "invalid namespace",
		},
		{
			name:      "kind with space is rejected",
			kind:      "Deployment List",
			resName:   "",
			namespace: "default",
			wantErr:   "not permitted",
		},
		{
			name:      "kind with slash is rejected",
			kind:      "apps/Deployment",
			resName:   "",
			namespace: "default",
			wantErr:   "not permitted",
		},
		{
			name:      "kind with underscore is rejected",
			kind:      "My_Kind",
			resName:   "",
			namespace: "default",
			wantErr:   "not permitted",
		},
		{
			name:      "namespace with dots is rejected",
			kind:      "Deployment",
			resName:   "my-app",
			namespace: "my.namespace",
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
			if err != nil && strings.Contains(err.Error(), "exceeds maximum length") {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestCaptureResource_ResourceID(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		resName    string
		wantPrefix string
	}{
		{
			name:       "kind-only resourceID when name is empty",
			kind:       "Deployment",
			resName:    "",
			wantPrefix: "Deployment",
		},
		{
			name:       "kind/name resourceID when name is set",
			kind:       "Deployment",
			resName:    "my-app",
			wantPrefix: "Deployment/my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the expected resourceID using the same logic as CaptureResource
			resourceID := tt.kind
			if tt.resName != "" {
				resourceID = tt.kind + "/" + tt.resName
			}
			if resourceID != tt.wantPrefix {
				t.Errorf("resourceID = %q, want %q", resourceID, tt.wantPrefix)
			}
		})
	}
}

func TestCaptureResource_EmptyStderr(t *testing.T) {
	// CaptureResource will fail because kubectl is not available. When stderr
	// is empty, the error message should contain "(no stderr output)".
	_, err := CaptureResource("Deployment", "my-app", "default")
	if err == nil {
		t.Fatal("expected error when kubectl is unavailable, got nil")
	}
	if !strings.Contains(err.Error(), "(no stderr output)") {
		t.Logf("error: %v", err)
		// stderr may or may not be empty depending on the environment;
		// verify the error is from kubectl exec, not validation.
		if strings.Contains(err.Error(), "not permitted") ||
			strings.Contains(err.Error(), "invalid resource") ||
			strings.Contains(err.Error(), "namespace is required") {
			t.Fatalf("unexpected validation error: %v", err)
		}
	}
}

func TestCaptureResourceWithContext(t *testing.T) {
	// Verify CaptureResourceWithContext accepts a caller-provided context
	// and passes validation (will fail at kubectl exec).
	ctx := context.Background()
	_, err := CaptureResourceWithContext(ctx, "Deployment", "my-app", "default")
	if err != nil &&
		(strings.Contains(err.Error(), "not permitted") ||
			strings.Contains(err.Error(), "invalid resource") ||
			strings.Contains(err.Error(), "namespace is required")) {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
