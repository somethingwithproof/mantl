package runtimeevents

import (
	"strings"
	"testing"
)

func TestFalcoIdentityAndValidation(t *testing.T) {
	data := []byte(`{"rule":"Sensitive File Access","priority":"WARNING","time":"2026-01-01T12:00:00Z","output":"SENSITIVE_SENTINEL","output_fields":{"k8s.ns.name":"tenant","k8s.pod.name":"app"}}`)
	signal, err := Parse(data)
	if err != nil || signal.Severity != "medium" || strings.Contains(signal.Rule, "SENSITIVE_SENTINEL") {
		t.Fatalf("signal=%+v err=%v", signal, err)
	}
	again, _ := Parse(data)
	if again.ID != signal.ID {
		t.Fatal("event identity not deterministic")
	}
	if _, err := Parse([]byte(`{"rule":"test","priority":"WARNING","time":"bad"}`)); err == nil {
		t.Fatal("invalid event accepted")
	}
}
