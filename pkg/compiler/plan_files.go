// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
)

// ReadPlan loads bounded metadata. Inventories must be kept in a trusted location;
// hashes establish integrity, not reviewer identity or authenticity.
func ReadPlan(filename string) (Plan, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return Plan{}, fmt.Errorf("open saved plan: %w", err)
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return Plan{}, fmt.Errorf("inspect saved plan: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Plan{}, fmt.Errorf("saved plan must be a regular file")
	}

	return DecodePlan(file)
}

// ReadPlanArtifacts returns verified bytes held independently of the source files.
// Root-bound access prevents paths or directory symlinks from escaping the build.
func ReadPlanArtifacts(plan Plan, directory string) (Plan, error) {
	if err := ValidatePlan(plan); err != nil {
		return Plan{}, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return Plan{}, fmt.Errorf("open reviewed artifact directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	verified := plan
	verified.Artifacts = append([]Artifact(nil), plan.Artifacts...)
	for i, artifact := range verified.Artifacts {
		data, err := readPlanArtifact(root, artifact)
		if err != nil {
			return Plan{}, fmt.Errorf("verify %s: %w", artifact.Path, err)
		}
		verified.Artifacts[i].Content = data
	}
	return verified, nil
}

func readPlanArtifact(root *os.Root, artifact Artifact) ([]byte, error) {
	info, err := root.Lstat(artifact.Path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("artifact must be a regular file")
	}
	file, err := root.OpenFile(artifact.Path, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() != int64(artifact.Size) {
		return nil, fmt.Errorf("artifact type or size differs from reviewed inventory")
	}
	data, err := io.ReadAll(io.LimitReader(file, int64(artifact.Size)+1))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	if len(data) != artifact.Size || hex.EncodeToString(hash[:]) != artifact.SHA256 {
		return nil, fmt.Errorf("artifact hash or size differs from reviewed inventory")
	}
	return data, nil
}
