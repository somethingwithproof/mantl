package framework

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

var safeName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func Load(dir, name, version string) (*Framework, error) {
	if !safeName.MatchString(name) {
		return nil, fmt.Errorf("invalid framework name %q", name)
	}
	data, err := os.ReadFile(filepath.Join(dir, name, "framework.yaml"))
	if err != nil {
		return nil, fmt.Errorf("load framework %s: %w", name, err)
	}
	var fw Framework
	if err := yaml.Unmarshal(data, &fw); err != nil {
		return nil, fmt.Errorf("decode framework %s: %w", name, err)
	}
	if fw.Version == "" || len(fw.Controls) == 0 {
		return nil, fmt.Errorf("framework %s has no supported control catalog", name)
	}
	if version != "" && version != fw.Version {
		return nil, fmt.Errorf("framework version %q does not match packaged version %q", version, fw.Version)
	}
	return &fw, nil
}
func Selected(control string, included []string) bool {
	if len(included) == 0 {
		return true
	}
	for _, id := range included {
		if id == control {
			return true
		}
	}
	return false
}
func Next(schedule string, last time.Time) (time.Time, error) {
	if len(strings.Fields(schedule)) != 5 {
		return time.Time{}, fmt.Errorf("expected five-field UTC cron schedule")
	}
	parsed, err := cron.ParseStandard(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid collector schedule: %w", err)
	}
	next := parsed.Next(last.UTC())
	if next.IsZero() {
		return time.Time{}, fmt.Errorf("collector schedule never runs")
	}
	return next, nil
}
