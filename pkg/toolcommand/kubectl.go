// Package toolcommand constructs bounded external-tool invocations without PATH lookup.
package toolcommand

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var kubectlDirectories = []string{"/usr/local/bin", "/usr/bin", "/opt/homebrew/bin"}

// Kubectl uses a fixed installation directory or an explicitly configured absolute
// MANTL_KUBECTL_PATH. Developer tool managers should supply their pinned binary path.
// A resolution failure is retained in Cmd.Err, so no process can start.
func Kubectl(ctx context.Context, args ...string) *exec.Cmd {
	path, err := resolveKubectl(os.Getenv("MANTL_KUBECTL_PATH"), kubectlDirectories)
	command := exec.CommandContext(ctx, path, args...)
	if err != nil {
		command.Err = err
	}
	return command
}

func resolveKubectl(configured string, directories []string) (string, error) {
	if configured != "" {
		return executable(configured)
	}
	for _, directory := range directories {
		if path, err := executable(filepath.Join(directory, "kubectl")); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("kubectl is unavailable in fixed installation directories; set MANTL_KUBECTL_PATH to its absolute pinned binary path")
}

func executable(path string) (string, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", fmt.Errorf("kubectl binary path must be absolute and canonical")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve kubectl binary: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("inspect kubectl binary: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 || info.Mode().Perm()&0022 != 0 {
		return "", fmt.Errorf("kubectl binary must be a regular executable without group or world write permission")
	}
	return resolved, nil
}
