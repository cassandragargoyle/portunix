/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package kit

import (
	"fmt"
	"os"
)

// LocalPathProvider exposes a pre-checked-out kit tree from the local
// filesystem. Used for forks, developer iteration, and the supported
// air-gapped workflow.
type LocalPathProvider struct {
	Path string
}

// NewLocalPathProvider constructs a Provider rooted at an absolute path.
func NewLocalPathProvider(path string) *LocalPathProvider {
	return &LocalPathProvider{Path: path}
}

// Fetch validates that the path looks like a spec-kit-pm checkout and
// returns it as the kit directory. fromCache is always true: nothing is
// fetched, no network was contacted.
func (p *LocalPathProvider) Fetch() (string, bool, error) {
	info, err := os.Stat(p.Path)
	if err != nil {
		return "", false, fmt.Errorf("local kit path: %w", err)
	}
	if !info.IsDir() {
		return "", false, fmt.Errorf("local kit path %q is not a directory", p.Path)
	}
	if !isPopulatedKit(p.Path) {
		return "", false, fmt.Errorf("local kit path %q does not look like a spec-kit-pm checkout (expected templates/, agents/, workflows/)", p.Path)
	}
	return p.Path, true, nil
}
