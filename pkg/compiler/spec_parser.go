// SPDX-License-Identifier: Apache-2.0

package compiler

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thomasvincent/mantl/apis/platform/v1alpha1"
	"sigs.k8s.io/yaml"
)

// ParseSpec reads the MantlCluster YAML spec from a file and unmarshals it.
func ParseSpec(filePath string) (*v1alpha1.MantlCluster, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var cluster v1alpha1.MantlCluster
	if err := yaml.UnmarshalStrict(data, &cluster); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	if err := Validate(&cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

var dnsName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// reservedNamespaces are control-plane or platform namespaces. A tenant admin
// binding must not be rendered into them.
var reservedNamespaces = map[string]struct{}{
	"kube-system": {}, "kube-public": {}, "kube-node-lease": {},
	"default": {}, "argocd": {}, "mantl-compliance-system": {},
}

func reservedNamespace(name string) bool {
	_, ok := reservedNamespaces[name]
	return ok
}

func Validate(cluster *v1alpha1.MantlCluster) error {
	if cluster == nil {
		return fmt.Errorf("cluster must not be nil")
	}

	if err := validateRequiredFields(cluster); err != nil {
		return err
	}

	if !dnsName.MatchString(cluster.Name) || len(cluster.Name) > 63 {
		return fmt.Errorf("metadata.name must be a DNS label")
	}
	if cluster.Spec.Environment != "" && cluster.Spec.Environment != "dev" && cluster.Spec.Environment != "staging" && cluster.Spec.Environment != "production" {
		return fmt.Errorf("unsupported environment")
	}
	pairs := map[string]string{"local": "kind", "aws": "eks", "gcp": "gke", "azure": "aks", "do": "doks", "linode": "lke", "oci": "oke", "ibm": "iks", "openstack": "kubernetes"}
	if distribution, ok := pairs[cluster.Spec.Provider.Kind]; !ok || distribution != cluster.Spec.Kubernetes.Distribution {
		return fmt.Errorf("unsupported provider/distribution combination")
	}
	if cluster.Spec.Profile.Size != "small" && cluster.Spec.Profile.Size != "medium" && cluster.Spec.Profile.Size != "full" {
		return fmt.Errorf("unsupported profile size")
	}
	if (cluster.Spec.Provider.Kind == "gcp" || cluster.Spec.Provider.Kind == "azure") && cluster.Spec.Provider.AccountID == "" {
		return fmt.Errorf("provider.accountId is required for the GCP project or Azure resource group")
	}
	if err := validateGitOps(cluster); err != nil {
		return err
	}
	if ns := cluster.Spec.ControlNamespace; ns != "" && (!dnsName.MatchString(ns) || len(ns) > 63) {
		return fmt.Errorf("controlNamespace must be a DNS label")
	}
	return validateTenants(cluster.Spec.Tenants, cluster.Spec.ControlNamespace)
}

func validateGitOps(cluster *v1alpha1.MantlCluster) error {
	git := cluster.Spec.GitOps
	if git.Repository != "" {
		u, err := url.Parse(git.Repository)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return fmt.Errorf("gitops.repository must be an HTTPS URL without credentials")
		}
	}
	for _, path := range []string{git.TenantPath, git.OperatorPath} {
		if path != "" && (filepath.IsAbs(path) || strings.Contains(path, "\\") || strings.HasPrefix(filepath.Clean(path), "..") || path == ".") {
			return fmt.Errorf("gitops paths must be relative repository paths")
		}
	}
	if len(cluster.Spec.Tenants) > 0 && git.TenantPath == "" {
		return fmt.Errorf("gitops.tenantPath is required; commit generated tenants at this path")
	}
	return nil
}

func validateTenants(tenants []v1alpha1.TenantSpec, controlNamespace string) error {
	seen := map[string]bool{}
	seenNames := map[string]bool{}
	for _, t := range tenants {
		ns := t.Namespace
		if ns == "" {
			ns = t.Name
		}
		if !dnsName.MatchString(t.Name) || !dnsName.MatchString(ns) || len(ns) > 63 || len(t.Name) > 63 || seen[ns] || seenNames[t.Name] || t.Name == "kustomization" {
			return fmt.Errorf("tenant names and namespaces must be unique DNS labels")
		}
		if reservedNamespace(ns) || ns == controlNamespace {
			return fmt.Errorf("tenant namespace %q is reserved", ns)
		}
		seen[ns] = true
		seenNames[t.Name] = true
	}
	return nil
}

func validateRequiredFields(cluster *v1alpha1.MantlCluster) error {
	type requiredField struct{ name, value string }
	fields := []requiredField{
		{"spec.provider.kind", cluster.Spec.Provider.Kind},
		{"spec.provider.region", cluster.Spec.Provider.Region},
		{"spec.kubernetes.distribution", cluster.Spec.Kubernetes.Distribution},
		{"spec.kubernetes.version", cluster.Spec.Kubernetes.Version},
		{"spec.profile.size", cluster.Spec.Profile.Size},
		{"spec.networking.domain", cluster.Spec.Networking.Domain},
	}
	for _, field := range fields {
		if field.value == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}
	return nil
}
