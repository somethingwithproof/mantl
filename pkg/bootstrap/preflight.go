// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomasvincent/mantl/pkg/toolcommand"
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
	if _, err := toolcommand.OpenTofuOrTerraform(); err != nil {
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

func (d *ExecutionDAG) preflightLocal(ctx context.Context, clusterName string) error {
	if !validKindClusterNameRe.MatchString(clusterName) || len(clusterName) > 63 {
		return fmt.Errorf("invalid kind cluster name")
	}
	kind := d.command(ctx, "kind", "get", "clusters")
	if kind.Err != nil {
		return fmt.Errorf("preflight kind: %w", kind.Err)
	}
	if err := kind.Run(); err != nil {
		return fmt.Errorf("preflight kind container runtime: %w", err)
	}
	return nil
}
