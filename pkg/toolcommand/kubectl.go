// SPDX-License-Identifier: Apache-2.0

// Package toolcommand constructs bounded external-tool invocations without PATH lookup.
package toolcommand

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/unix"
)

var toolDirectories = []string{"/usr/local/bin", "/usr/bin", "/opt/homebrew/bin"}

// Kubectl uses a fixed installation directory or an explicitly configured absolute
// MANTL_KUBECTL_PATH. Developer tool managers should supply their pinned binary path.
// A resolution failure is retained in Cmd.Err, so no process can start.
func Kubectl(ctx context.Context, args ...string) *exec.Cmd {
	return tool(ctx, "kubectl", os.Getenv("MANTL_KUBECTL_PATH"), args...)
}

// Kind uses the same directory rule as Kubectl. MANTL_KIND_PATH overrides it.
func Kind(ctx context.Context, args ...string) *exec.Cmd {
	return tool(ctx, "kind", os.Getenv("MANTL_KIND_PATH"), args...)
}

// OpenTofuOrTerraform prefers tofu, then terraform. An explicit MANTL_TOFU_PATH
// or MANTL_TERRAFORM_PATH does not fall back to another binary.
func OpenTofuOrTerraform() (string, error) {
	if configured := os.Getenv("MANTL_TOFU_PATH"); configured != "" {
		return resolveTool("tofu", configured, nil)
	}
	if path, err := resolveTool("tofu", "", toolDirectories); err == nil {
		return path, nil
	}
	if configured := os.Getenv("MANTL_TERRAFORM_PATH"); configured != "" {
		return resolveTool("terraform", configured, nil)
	}
	path, err := resolveTool("terraform", "", toolDirectories)
	if err != nil {
		return "", fmt.Errorf("neither tofu nor terraform is available in fixed installation directories; set MANTL_TOFU_PATH or MANTL_TERRAFORM_PATH")
	}
	return path, nil
}

func tool(ctx context.Context, name, configured string, args ...string) *exec.Cmd {
	path, err := resolveTool(name, configured, toolDirectories)
	command := exec.CommandContext(ctx, path, args...)
	if err != nil {
		command.Err = err
	}
	return command
}

func resolveKubectl(configured string, directories []string) (string, error) {
	return resolveTool("kubectl", configured, directories)
}

func resolveTool(name, configured string, directories []string) (string, error) {
	if configured != "" {
		return executable(name, configured)
	}
	for _, directory := range directories {
		if path, err := executable(name, filepath.Join(directory, name)); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s is unavailable in fixed installation directories; set its absolute pinned binary path", name)
}

func executable(name, path string) (string, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", fmt.Errorf("%s binary path must be absolute and canonical", name)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s binary: %w", name, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("inspect %s binary: %w", name, err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 || info.Mode().Perm()&0022 != 0 {
		return "", fmt.Errorf("%s binary must be a regular executable without group or world write permission", name)
	}
	if err := unix.Faccessat(unix.AT_FDCWD, resolved, unix.X_OK, unix.AT_EACCESS); err != nil {
		return "", fmt.Errorf("%s binary is not executable by the current process: %w", name, err)
	}
	return resolved, nil
}
