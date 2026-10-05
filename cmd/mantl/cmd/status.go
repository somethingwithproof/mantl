package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thomasvincent/mantl/pkg/compiler"
)

// statusCheckTimeout bounds the cluster query so status stays fast and never
// blocks on the bootstrap-oriented convergence wait.
const statusCheckTimeout = 5 * time.Second

// healthJSONPath emits one "sync,health" line per ArgoCD application. Pairing the
// two fields per item (rather than two independent ranges) keeps sync and health
// correlated, so a missing health field cannot be silently dropped.
const healthJSONPath = `jsonpath={range .items[*]}{.status.sync.status},{.status.health.status}{"\n"}{end}`

// kubectlHealthQuery returns the raw ArgoCD application status (stdout only). It
// is a package var so tests can exercise platformHealth without a live cluster.
var kubectlHealthQuery = func(ctx context.Context) ([]byte, error) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return nil, fmt.Errorf("kubectl not found in PATH: %w", err)
	}
	// Output (not CombinedOutput) so kubectl warnings on stderr never reach the
	// parser; stderr is surfaced via ExitError on failure instead.
	out, err := exec.CommandContext(ctx, "kubectl", kubectlArgs("get", "applications", "-n", "argocd", "-o", healthJSONPath)...).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	return out, nil
}

var statusCmd = &cobra.Command{
	Use:   "status [spec-file]",
	Short: "Show the live status of the Mantl platform",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cluster, err := compiler.ParseSpec(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Fetching status for Mantl platform: %s\n", cluster.Name)
		fmt.Println("-------------------------------------------")
		fmt.Printf("Platform Status: %s\n", platformHealth())

		fmt.Println("\nEnabled Features:")
		fmt.Printf("  Observability: %v\n", cluster.Spec.Features.Observability)
		fmt.Printf("  Security:      %v\n", cluster.Spec.Features.Security)
		fmt.Printf("  Compliance:    %v\n", cluster.Spec.Features.Compliance)

		fmt.Printf("\nCompliance Score: %s\n", complianceScore())
	},
}

// platformHealth returns a one-shot view of ArgoCD application convergence.
// Unlike bootstrap's VerifyConvergence, it does not wait: a status command must
// return promptly whether or not the platform has converged.
func platformHealth() string {
	ctx, cancel := context.WithTimeout(context.Background(), statusCheckTimeout)
	defer cancel()

	out, err := kubectlHealthQuery(ctx)
	if err != nil {
		return fmt.Sprintf("UNKNOWN (cluster unreachable: %v)", err)
	}
	return evaluateHealth(string(out))
}

// evaluateHealth maps the per-application "sync,health" lines to a status string.
// It fails closed: any application whose sync or health is missing or not in the
// expected state yields UNKNOWN or DEGRADED, never HEALTHY. HEALTHY requires every
// application to report both Synced and Healthy.
func evaluateHealth(raw string) string {
	var apps [][2]string
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sync, health, ok := strings.Cut(line, ",")
		if !ok || strings.TrimSpace(sync) == "" || strings.TrimSpace(health) == "" {
			return "UNKNOWN (incomplete application status)"
		}
		apps = append(apps, [2]string{strings.TrimSpace(sync), strings.TrimSpace(health)})
	}

	if len(apps) == 0 {
		return "UNKNOWN (no ArgoCD applications found)"
	}
	for _, app := range apps {
		if app[0] != "Synced" {
			return "DEGRADED (applications not synced)"
		}
	}
	for _, app := range apps {
		if app[1] != "Healthy" {
			return "DEGRADED (applications not healthy)"
		}
	}
	return "HEALTHY"
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
