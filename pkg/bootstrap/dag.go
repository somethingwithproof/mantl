package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// ExecutionDAG defines the sequence of platform bring-up.
type ExecutionDAG struct {
	SpecFile string
	BuildDir string
}

// ProvisionInfra handles the creation of the underlying infrastructure.
func (d *ExecutionDAG) ProvisionInfra(provider string, clusterName string) error {
	if provider == "local" {
		return d.CreateLocalCluster(clusterName)
	}

	fmt.Printf("Executing: terraform apply (Provider: %s)\n", provider)

	// 1. Find terraform binary
	tfPath, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("terraform binary not found in PATH: %w", err)
	}

	// 2. Prepare workspace
	blueprintPath := filepath.Join("infra/terraform/blueprints", fmt.Sprintf("%s-eks", provider))
	tf, err := tfexec.NewTerraform(blueprintPath, tfPath)
	if err != nil {
		return fmt.Errorf("failed to create tfexec instance: %w", err)
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	// 3. Initialize
	fmt.Println("  Initializing Terraform...")
	if err := tf.Init(context.Background(), tfexec.Upgrade(true)); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// 4. Apply
	absBuildDir, _ := filepath.Abs(d.BuildDir)
	tfVarsPath := filepath.Join(absBuildDir, "terraform.tfvars.json")
	
	fmt.Printf("  Applying configuration with vars: %s\n", tfVarsPath)
	if err := tf.Apply(context.Background(), tfexec.VarFile(tfVarsPath)); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// CreateLocalCluster spins up a Kubernetes cluster using Kind.
func (d *ExecutionDAG) CreateLocalCluster(name string) error {
	fmt.Printf("Executing: kind create cluster --name %s\n", name)
	
	// Check if cluster already exists
	checkCmd := exec.Command("kind", "get", "clusters")
	output, _ := checkCmd.Output()
	if strings.Contains(string(output), name) {
		fmt.Printf("  Cluster '%s' already exists, skipping creation.\n", name)
		return nil
	}

	cmd := exec.Command("kind", "create", "cluster", "--name", name)
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

	// 1. Create namespace
	exec.Command("kubectl", "create", "namespace", "argocd").Run()

	// 2. Apply install manifest
	installUrl := "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"
	cmd := exec.Command("kubectl", "apply", "-n", "argocd", "-f", installUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install argo-cd: %w", err)
	}

	// 3. Wait for ArgoCD to be ready
	fmt.Println("  Waiting for ArgoCD deployments to be ready...")
	waitCmd := exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s", "deployment", "-n", "argocd", "--all")
	waitCmd.Run()

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
		cmd := exec.Command("kubectl", "apply", "-f", file)
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
		case <-timeout:
			return fmt.Errorf("timeout waiting for platform convergence")
		case <-ticker.C:
			cmd := exec.Command("kubectl", "get", "applications", "-n", "argocd",
				"-o", "jsonpath={.items[*].status.sync.status}={.items[*].status.health.status}")
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("  Waiting for ArgoCD API... (%s)\n", err.Error())
				continue
			}

			strOutput := string(output)
			if strOutput == "=" || strOutput == "" {
				fmt.Println("  Waiting for applications to be created...")
				continue
			}

			parts := strings.Split(strOutput, "=")
			if len(parts) != 2 {
				fmt.Println("  Waiting for application status...")
				continue
			}

			syncStatuses := strings.Fields(parts[0])
			healthStatuses := strings.Fields(parts[1])

			allSynced := true
			allHealthy := true

			for _, s := range syncStatuses {
				if s != "Synced" {
					allSynced = false
				}
			}
			for _, h := range healthStatuses {
				if h != "Healthy" {
					allHealthy = false
				}
			}

			if allSynced && allHealthy {
				fmt.Println("  All platform applications are Synced and Healthy!")
				return nil
			}

			fmt.Println("  Waiting for applications to sync and become healthy...")
		}
	}
}
