package compliance

import "sort"

// resolveTemplates maps framework template names to policy file paths using a
// name index and a small alias table (for templates whose canonical policy ships
// under a different name). It returns resolved paths (sorted, deduped) and the
// template names with no policy. Pure: no I/O, so it tests without a cluster.
func resolveTemplates(templates []string, index, aliases map[string]string) (resolved, unresolved []string) {
	seen := map[string]bool{}
	for _, t := range templates {
		name := t
		if alias, ok := aliases[t]; ok {
			name = alias
		}
		if path, ok := index[name]; ok {
			if !seen[path] {
				seen[path] = true
				resolved = append(resolved, path)
			}
			continue
		}
		unresolved = append(unresolved, t)
	}
	sort.Strings(resolved)
	return resolved, unresolved
}
