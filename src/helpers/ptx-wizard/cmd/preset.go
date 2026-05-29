/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"portunix.ai/app/wizard/engine"
)

// loadPreset locates a preset file for the given wizard and feeds its
// variables into the engine. Conventions tried in order:
//
//  1. <wizard-dir>/<wizard-name>.<preset>.yaml
//  2. <wizard-dir>/presets/<preset>.yaml
//
// A missing preset returns an error so the user notices the typo.
func loadPreset(eng *engine.WizardEngine, wizardPath, preset string) error {
	dir := filepath.Dir(wizardPath)
	base := strings.TrimSuffix(filepath.Base(wizardPath), filepath.Ext(wizardPath))

	candidates := []string{
		filepath.Join(dir, fmt.Sprintf("%s.%s.yaml", base, preset)),
		filepath.Join(dir, "presets", preset+".yaml"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return eng.LoadVariablesFromConfig(c)
		}
	}

	return fmt.Errorf("preset file not found (tried: %s)", strings.Join(candidates, ", "))
}
