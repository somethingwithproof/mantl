package bootstrap

import (
	"context"
	"fmt"
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
	Context      context.Context
	KubeContext  string
	SourceDir    string
}

var validKindClusterNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ProvisionInfra handles the creation of the underlying infrastructure.
func (d *ExecutionDAG) ProvisionInfra(provider string, clusterName string) error {
	if provider == "local" {
		return d.CreateLocalCluster(clusterName)
	}

	fmt.Printf("Executing: terraform apply (Provider: %s)\n", provider)

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

	// 2. Find terraform binary (prefer tofu, fall back to terraform)
	tfPath, err := exec.LookPath("tofu")
	if err != nil {
		tfPath, err = exec.LookPath("terraform")
		if err != nil {
			return fmt.Errorf("neither tofu nor terraform binary found in PATH: %w", err)
		}
	}

	tf, err := tfexec.NewTerraform(blueprintPath, tfPath)
	if err != nil {
		return fmt.Errorf("failed to create tfexec instance: %w", err)
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	// 3. Initialize
	fmt.Println("  Initializing Terraform...")
	if err := tf.Init(d.context(), tfexec.Upgrade(false)); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// 4. Apply
	absBuildDir, _ := filepath.Abs(d.BuildDir)
	tfVarsPath := filepath.Join(absBuildDir, "terraform.tfvars.json")

	fmt.Printf("  Applying configuration with vars: %s\n", tfVarsPath)
	if err := tf.Apply(d.context(), tfexec.VarFile(tfVarsPath)); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// CreateLocalCluster spins up a Kubernetes cluster using Kind.
func (d *ExecutionDAG) CreateLocalCluster(name string) error {
	if !validKindClusterNameRe.MatchString(name) || len(name) > 253 {
		return fmt.Errorf("invalid cluster name %q", name)
	}

	fmt.Printf("Executing: kind create cluster --name %s\n", name)

	// Check if cluster already exists
	checkCmd := d.command("kind", "get", "clusters")
	output, checkErr := checkCmd.Output()
	if checkErr != nil {
		// Non-fatal: kind may not have any clusters yet; log and proceed to create.
		fmt.Printf("  Warning: 'kind get clusters' failed (%s), proceeding with creation.\n", checkErr.Error())
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == name {
			fmt.Printf("  Cluster '%s' already exists, skipping creation.\n", name)
			return nil
		}
	}

	cmd := d.command("kind", "create", "cluster", "--name", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create kind cluster: %w", err)
	}

	return nil
}

// InstallArgoCD installs ArgoCD into the current cluster.
func (d *ExecutionDAG) InstallArgoCD() error {
	fmt.Println("Executing: Installing ArgoCD...")

	// Apply a declarative namespace so installation is repeatable.
	namespace := d.command("kubectl", "apply", "-f", "-")
	namespace.Stdin = strings.NewReader("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: argocd\n")
	if err := namespace.Run(); err != nil {
		return fmt.Errorf("ensure argocd namespace: %w", err)
	}

	// 2. Apply install manifest
	installUrl := "https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.3/manifests/install.yaml"
	cmd := d.command("kubectl", "apply", "-n", "argocd", "-f", installUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install argo-cd: %w", err)
	}

	// 3. Wait for ArgoCD to be ready
	fmt.Println("  Waiting for ArgoCD deployments to be ready...")
	waitCmd := d.command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment", "-n", "argocd", "--all")
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("argocd deployments did not become ready: %w", err)
	}

	return nil
}

// BootstrapGitOps applies the root application to the cluster.
func (d *ExecutionDAG) BootstrapGitOps() error {
	fmt.Println("Executing: kubectl apply -f .mantl/build/gitops/*.yaml")

	gitopsDir := filepath.Join(d.BuildDir, "gitops")
	files, err := filepath.Glob(filepath.Join(gitopsDir, "*.yaml"))
	if err != nil {
		return err
	}

	for _, file := range files {
		fmt.Printf("  Applying %s...\n", filepath.Base(file))
		cmd := d.command("kubectl", "apply", "-f", file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply %s: %w", file, err)
		}
	}

	return nil
}

// VerifyConvergence polls the cluster until the platform is healthy.
func (d *ExecutionDAG) VerifyConvergence() error {
	fmt.Println("Executing: Platform health check...")

	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.context().Done():
			return d.context().Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for platform convergence")
		case <-ticker.C:
			cmd := d.command("kubectl", "get", "applications", "-n", "argocd",
				"-o", `jsonpath={range .items[*]}{.status.sync.status},{.status.health.status}{"\n"}{end}`)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("  Waiting for ArgoCD API... (%s)\n", err.Error())
				continue
			}

			if ApplicationsHealthy(string(output)) {
				return nil
			}

			fmt.Println("  Waiting for applications to sync and become healthy...")
		}
	}
}

func (d *ExecutionDAG) context() context.Context {
	if d.Context != nil {
		return d.Context
	}
	return context.Background()
}
func (d *ExecutionDAG) command(name string, args ...string) *exec.Cmd {
	if name == "kubectl" && d.KubeContext != "" {
		args = append([]string{"--context", d.KubeContext}, args...)
	}
	return exec.CommandContext(d.context(), name, args...)
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
