/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [wizard-name]",
	Short: "Create a new wizard from template",
	Long:  "Create a new wizard YAML file from a minimal template.",
	Args:  cobra.ExactArgs(1),
	Run:   createWizard,
}

func createWizard(_ *cobra.Command, args []string) {
	wizardName := args[0]
	fileName := wizardName + ".yaml"

	if _, err := os.Stat(fileName); err == nil {
		fmt.Printf("Error: File '%s' already exists\n", fileName)
		os.Exit(1)
	}

	title := titleCase(wizardName)
	template := `wizard:
  id: "` + wizardName + `"
  name: "` + title + ` Setup Wizard"
  version: "1.0"
  description: "Description of your wizard"

  variables:
    # Define variables here

  pages:
    - id: "welcome"
      type: "info"
      title: "Welcome"
      content: |
        Welcome to the ` + wizardName + ` setup wizard!

        This wizard will guide you through the setup process.
      next:
        page: "complete"

    - id: "complete"
      type: "success"
      title: "Setup Complete!"
      content: |
        Setup completed successfully!
`

	if err := os.WriteFile(fileName, []byte(template), 0o644); err != nil {
		fmt.Printf("Error creating wizard file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created wizard template: %s\n", fileName)
	fmt.Printf("Edit the file and run: portunix wizard validate %s\n", fileName)
}

// titleCase capitalizes each dash- or underscore-separated word.
// Replaces deprecated strings.Title.
func titleCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}
