package bootstrap

import (
	"context"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/compiler"
	"github.com/thomasvincent/mantl/pkg/toolcommand"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// ExecutionDAG defines the sequence of platform bring-up.
type ExecutionDAG struct {
	SpecFile     string
	BuildDir     string
	Distribution string
	KubeContext  string
	SourceDir    string
	Applications []compiler.GitOpsApplication
}

var validKindClusterNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ProvisionInfra handles the creation of the underlying infrastructure.
func (d *ExecutionDAG) ProvisionInfra(ctx context.Context, provider, clusterName string) error {
	if provider == "local" {
		return d.CreateLocalCluster(ctx, clusterName)
	}

	slog.Info("provisioning infrastructure", "provider", provider)

	// 1. Derive the blueprint directory from provider and distribution.
	//    The convention matches the directory names under infra/terraform/blueprints/:
	//    {provider}-{distribution}  (e.g. aws-eks, gcp-gke, azure-aks)
	distribution := d.Distribution
	if distribution == "" {
		return fmt.Errorf("distribution is required to select a terraform blueprint")
	}
	blueprintName := fmt.Sprintf("%s-%s", provider, distribution)
	blueprintPath := filepath.Join(d.SourceDir, "infra/terraform/blueprints", blueprintName)
	if _, statErr := os.Stat(blueprintPath); os.IsNotExist(statErr) {
		return fmt.Errorf("terraform blueprint %q not found at %s", blueprintName, blueprintPath)
	}

	tfPath, err := terraformExecutable()
	if err != nil {
		return err
	}

	tf, err := tfexec.NewTerraform(blueprintPath, tfPath)
	if err != nil {
		return fmt.Errorf("failed to create tfexec instance: %w", err)
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	// 3. Initialize
	slog.Info("initializing Terraform")
	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// 4. Apply
	absBuildDir, err := filepath.Abs(d.BuildDir)
	if err != nil {
		return fmt.Errorf("resolve build directory: %w", err)
	}
	tfVarsPath := filepath.Join(absBuildDir, "terraform.tfvars.json")

	slog.Info("applying infrastructure configuration", "variablesFile", tfVarsPath)
	if err := tf.Apply(ctx, tfexec.VarFile(tfVarsPath)); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// CreateLocalCluster spins up a Kubernetes cluster using Kind.
func (d *ExecutionDAG) CreateLocalCluster(ctx context.Context, name string) error {
	if !validKindClusterNameRe.MatchString(name) || len(name) > 253 {
		return fmt.Errorf("invalid cluster name %q", name)
	}

	slog.Info("creating local cluster", "name", name)

	// Check if cluster already exists
	checkCmd := d.command(ctx, "kind", "get", "clusters")
	output, checkErr := checkCmd.Output()
	if checkErr != nil {
		// Non-fatal: kind may not have any clusters yet; log and proceed to create.
		slog.Warn("local cluster query failed", "error", checkErr)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == name {
			slog.Info("local cluster already exists", "name", name)
			return nil
		}
	}

	cmd := d.command(ctx, "kind", "create", "cluster", "--name", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create kind cluster: %w", err)
	}

	return nil
}

// InstallArgoCD installs ArgoCD into the current cluster.
func (d *ExecutionDAG) InstallArgoCD(ctx context.Context) error {
	slog.Info("installing ArgoCD")

	// Apply a declarative namespace so installation is repeatable.
	namespace := d.command(ctx, "kubectl", "apply", "-f", "-")
	namespace.Stdin = strings.NewReader("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: argocd\n")
	if err := namespace.Run(); err != nil {
		return fmt.Errorf("ensure argocd namespace: %w", err)
	}

	// 2. Apply install manifest
	installUrl := "https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.3/manifests/install.yaml"
	cmd := d.command(ctx, "kubectl", "apply", "-n", "argocd", "-f", installUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install argo-cd: %w", err)
	}

	// 3. Wait for ArgoCD to be ready
	slog.Info("waiting for ArgoCD deployments")
	waitCmd := d.command(ctx, "kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment", "-n", "argocd", "--all")
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("argocd deployments did not become ready: %w", err)
	}

	return nil
}

// BootstrapGitOps applies the root application to the cluster.
func (d *ExecutionDAG) BootstrapGitOps(ctx context.Context) error {
	slog.Info("bootstrapping GitOps applications")

	files, err := d.bootstrapFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		slog.Info("applying GitOps manifest", "file", filepath.Base(file))
		cmd := d.command(ctx, "kubectl", "apply", "-f", file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply %s: %w", file, err)
		}
	}

	return nil
}

func (d *ExecutionDAG) command(ctx context.Context, name string, args ...string) *exec.Cmd {
	if name == "kubectl" {
		if d.KubeContext != "" {
			args = append([]string{"--context", d.KubeContext}, args...)
		}
		command := toolcommand.Kubectl(ctx, args...)
		command.WaitDelay = time.Second
		return command
	}
	command := exec.CommandContext(ctx, name, args...)
	command.WaitDelay = time.Second
	return command
}

// ApplicationsHealthy requires a complete pair for every reported application.
func ApplicationsHealthy(raw string) bool {
	rows := 0
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		sync, health, ok := strings.Cut(line, ",")
		if !ok || sync != "Synced" || health != "Healthy" {
			return false
		}
		rows++
	}
	return rows > 0
}

func (d *ExecutionDAG) bootstrapFiles() ([]string, error) {
	if len(d.Applications) == 0 {
		return nil, fmt.Errorf("desired application topology is required for bootstrap")
	}
	files := []string{}
	for _, application := range d.Applications {
		if !validKindClusterNameRe.MatchString(application.Name) || len(application.Name) > 63 {
			return nil, fmt.Errorf("invalid desired application name")
		}
		file := filepath.Join(d.BuildDir, "gitops", application.Name+".yaml")
		info, err := os.Lstat(file)
		if err != nil {
			return nil, fmt.Errorf("inspect desired application manifest: %w", err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("desired application manifest must be a regular file")
		}
		files = append(files, file)
	}
	return files, nil
}
