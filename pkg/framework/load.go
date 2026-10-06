// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/thomasvincent/mantl/pkg/controlbundle"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

var safeName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func Load(dir, name, version string) (*Framework, error) {
	return LoadPinned(dir, name, version, "")
}

// LoadPinned validates a content bundle before reading a selected framework.
func LoadPinned(dir, name, version, digest string) (*Framework, error) {
	if !safeName.MatchString(name) {
		return nil, fmt.Errorf("invalid framework name %q", name)
	}
	bundleDigest := ""
	parent := filepath.Dir(dir)
	if _, statErr := os.Stat(filepath.Join(parent, "bundle.json")); statErr == nil || digest != "" {
		manifest, verifyErr := controlbundle.Verify(parent, digest)
		if verifyErr != nil {
			return nil, fmt.Errorf("verify framework bundle: %w", verifyErr)
		}
		bundleDigest = manifest.Digest()
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
	fw.ContentDigest = "sha256:" + evidence.Hash(data)
	if bundleDigest != "" {
		fw.ContentDigest = bundleDigest
	}
	seen := map[string]bool{}
	for _, control := range fw.Controls {
		if control.ID == "" || seen[control.ID] {
			return nil, fmt.Errorf("framework control IDs must be nonempty and unique")
		}
		seen[control.ID] = true
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

func LoadSelected(dir, name, version, digest string, included []string) (*Framework, error) {
	fw, err := LoadPinned(dir, name, version, digest)
	if err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, control := range fw.Controls {
		known[control.ID] = true
	}
	for _, id := range included {
		if !known[id] {
			return nil, fmt.Errorf("selected control %q is not in the framework", id)
		}
	}
	return fw, nil
}
