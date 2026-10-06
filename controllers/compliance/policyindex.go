// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"os"
	"path/filepath"
	"strings"
)

// buildPolicyIndex maps a policy's base file name (without .yaml) to its path,
// scanning the given directories. Missing directories are skipped so the index
// works whether policies live under platform/ or compliance/ trees.
func buildPolicyIndex(dirs []string) (map[string]string, error) {
	index := map[string]string{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".yaml")
			if _, exists := index[name]; !exists {
				index[name] = filepath.Join(dir, e.Name())
			}
		}
	}
	return index, nil
}

// templateAliases maps framework template names to the policy base name that
// actually ships, for cases where the names differ.
var templateAliases = map[string]string{
	"restrict-registries":          "restrict-image-registries",
	"require-pod-security-context": "pss-restricted",
}
