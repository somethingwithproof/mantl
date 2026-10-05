package framework

// PolicyControls preserves every selected mapping for a policy, including aliases.
// Callers decide namespace scope and sorting; duplicate mappings remain visible.
func PolicyControls(fw *Framework, included []string, frameworkName, policy string, aliases map[string]string) []string {
	var controls []string
	for _, control := range fw.Controls {
		if !Selected(control.ID, included) {
			continue
		}
		for _, mapping := range control.Mappings {
			if policyMatches(mapping.PolicyRef, frameworkName, policy, aliases) {
				controls = append(controls, control.ID)
			}
		}
	}
	return controls
}

func policyMatches(ref *PolicyRef, frameworkName, policy string, aliases map[string]string) bool {
	if ref == nil {
		return false
	}
	template := ref.Template
	return policy == template || policy == aliases[template] || policy == frameworkName+"-"+template
}

// NamespaceSelected includes cluster reports, otherwise respects the explicit
// namespace list or the profile namespace fallback.
func NamespaceSelected(namespaces []string, fallback, namespace string) bool {
	if namespace == "" {
		return true
	}
	if len(namespaces) == 0 {
		return namespace == fallback
	}
	for _, candidate := range namespaces {
		if candidate == namespace {
			return true
		}
	}
	return false
}
