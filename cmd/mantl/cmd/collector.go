// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

func init() {
	command := &cobra.Command{Use: "collector-rbac NAMESPACE", Short: "Render a reviewed namespaced collector identity; cloud permissions remain external", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		ns := args[0]
		if len(validation.IsDNS1123Label(ns)) != 0 {
			return fmt.Errorf("valid collection namespace required")
		}
		name := "mantl-evidence-collector"
		objects := []interface{}{
			&core.ServiceAccount{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"}, ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}},
			&rbac.Role{TypeMeta: metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"}, ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, Rules: []rbac.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods", "services", "serviceaccounts"}, Verbs: []string{"get", "list"}},
				{APIGroups: []string{"apps"}, Resources: []string{"deployments"}, Verbs: []string{"get", "list"}},
				{APIGroups: []string{"networking.k8s.io"}, Resources: []string{"networkpolicies"}, Verbs: []string{"get", "list"}},
				{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"roles", "rolebindings"}, Verbs: []string{"get", "list"}},
			}},
			&rbac.RoleBinding{TypeMeta: metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"}, ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, RoleRef: rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: name}, Subjects: []rbac.Subject{{Kind: "ServiceAccount", Name: name, Namespace: ns}}},
		}
		for _, object := range objects {
			b, err := yaml.Marshal(object)
			if err != nil {
				return err
			}
			if _, err = fmt.Fprintln(cmd.OutOrStdout(), "---\n"+string(b)); err != nil {
				return err
			}
		}
		return nil
	}}
	complianceCmd.AddCommand(command)
}
