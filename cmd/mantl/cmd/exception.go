// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	api "github.com/thomasvincent/mantl/apis/compliance/v1alpha1"
	"github.com/thomasvincent/mantl/pkg/toolcommand"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func init() {
	command := &cobra.Command{Use: "exception", Short: "Approve reviewed exception revisions using Kubernetes authorization"}
	approve := &cobra.Command{Use: "approve NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		data, err := queryCompliance(ctx, "complianceexception", args[0], complianceNamespace)
		if err != nil {
			return err
		}
		var exception api.ComplianceException
		if err = json.Unmarshal(data, &exception); err != nil {
			return err
		}
		if exception.Spec.Owner == "" || exception.Spec.Justification == "" || !exception.Spec.ExpiresAt.After(time.Now().UTC()) {
			return fmt.Errorf("exception requires owner, justification and future expiry")
		}
		identity, err := toolcommand.Kubectl(ctx, kubectlArgs("auth", "whoami", "-o", "json")...).Output()
		if err != nil {
			return fmt.Errorf("resolve authenticated approver: %w", err)
		}
		var subject authenticationv1.SelfSubjectReview
		if err = json.Unmarshal(identity, &subject); err != nil {
			return err
		}
		if subject.Status.UserInfo.Username == "" {
			return fmt.Errorf("authenticated approver identity unavailable")
		}
		patch, err := json.Marshal([]map[string]interface{}{
			{"op": "test", "path": "/metadata/resourceVersion", "value": exception.ResourceVersion},
			{"op": "add", "path": "/status", "value": api.ComplianceExceptionStatus{Approved: true, ApprovedBy: subject.Status.UserInfo.Username, ApprovedAt: func() *metav1.Time { at := metav1.Now(); return &at }(), ApprovedGeneration: exception.Generation}},
		})
		if err != nil {
			return err
		}
		output, err := toolcommand.Kubectl(ctx, kubectlArgs("patch", "complianceexception", exception.Name, "-n", complianceNamespace, "--subresource=status", "--type=json", "-p", string(patch), "-o", "json")...).Output()
		if err != nil {
			return fmt.Errorf("approve authorized exception revision: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(output)
		return err
	}}
	command.AddCommand(approve)
	complianceCmd.AddCommand(command)
}
