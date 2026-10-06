// SPDX-License-Identifier: Apache-2.0

package runtimeevents

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/validation"
)

type Event struct {
	Rule     string                 `json:"rule"`
	Priority string                 `json:"priority"`
	Time     string                 `json:"time"`
	Fields   map[string]interface{} `json:"output_fields"`
}
type Signal struct {
	ID, Rule, Namespace, Pod, Severity string
	At                                 time.Time
}

func Parse(data []byte) (Signal, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return Signal{}, fmt.Errorf("decode Falco event: %w", err)
	}
	at, err := time.Parse(time.RFC3339Nano, event.Time)
	if err != nil || event.Rule == "" {
		return Signal{}, fmt.Errorf("falco event requires a rule and RFC3339 timestamp")
	}
	namespace, _ := event.Fields["k8s.ns.name"].(string)
	pod, _ := event.Fields["k8s.pod.name"].(string)
	if len(validation.IsDNS1123Label(namespace)) != 0 || len(validation.IsDNS1123Subdomain(pod)) != 0 {
		return Signal{}, fmt.Errorf("falco event requires Kubernetes namespace and pod identity")
	}
	severity := ""
	switch strings.ToUpper(event.Priority) {
	case "EMERGENCY", "ALERT", "CRITICAL":
		severity = "critical"
	case "ERROR":
		severity = "high"
	case "WARNING":
		severity = "medium"
	case "NOTICE", "INFORMATIONAL", "DEBUG":
		severity = "low"
	default:
		return Signal{}, fmt.Errorf("unsupported Falco priority")
	}
	sum := sha256.Sum256([]byte(event.Rule + "\x00" + at.UTC().Format(time.RFC3339Nano) + "\x00" + namespace + "\x00" + pod))
	return Signal{ID: "falco-" + hex.EncodeToString(sum[:16]), Rule: event.Rule, Namespace: namespace, Pod: pod, Severity: severity, At: at.UTC()}, nil
}
