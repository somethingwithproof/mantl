// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Preflight checks local prerequisites without provisioning or contacting a
// cluster. Cloud context names are configuration selectors, not identity proof.
func (d *ExecutionDAG) Preflight(ctx context.Context, provider, clusterName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	kubectl := d.command(ctx, "kubectl", "version", "--client")
	if kubectl.Err != nil {
		return fmt.Errorf("preflight kubectl: %w", kubectl.Err)
	}
	if provider == "local" {
		return d.preflightLocal(ctx, clusterName)
	}
	if d.KubeContext == "" {
		return fmt.Errorf("explicit context is required for cloud bootstrap")
	}
	if d.Distribution == "" {
		return fmt.Errorf("distribution is required to select a terraform blueprint")
	}
	blueprint := filepath.Join(d.SourceDir, "infra/terraform/blueprints", provider+"-"+d.Distribution)
	info, err := os.Stat(blueprint)
	if err != nil {
		return fmt.Errorf("preflight blueprint: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("terraform blueprint must be a directory")
	}
	if _, err := terraformExecutable(); err != nil {
		return err
	}
	command := d.command(ctx, "kubectl", "config", "get-contexts", d.KubeContext, "-o", "name")
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("preflight kube-context: %w", err)
	}
	if strings.TrimSpace(string(output)) != d.KubeContext {
		return fmt.Errorf("requested cloud kube-context is absent from local configuration")
	}
	return nil
}

func terraformExecutable() (string, error) {
	if path, err := exec.LookPath("tofu"); err == nil {
		return path, nil
	}
	path, err := exec.LookPath("terraform")
	if err != nil {
		return "", fmt.Errorf("neither tofu nor terraform binary found in PATH: %w", err)
	}
	return path, nil
}

func (d *ExecutionDAG) preflightLocal(ctx context.Context, clusterName string) error {
	if !validKindClusterNameRe.MatchString(clusterName) || len(clusterName) > 63 {
		return fmt.Errorf("invalid kind cluster name")
	}
	if _, err := exec.LookPath("kind"); err != nil {
		return fmt.Errorf("preflight kind: %w", err)
	}
	if err := d.command(ctx, "kind", "get", "clusters").Run(); err != nil {
		return fmt.Errorf("preflight kind container runtime: %w", err)
	}
	return nil
}
