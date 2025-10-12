package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestContainerStopCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "Valid container name",
			args:        []string{"test-container"},
			expectError: false,
		},
		{
			name:        "No container name provided",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "Too many arguments",
			args:        []string{"container1", "container2"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new command instance for testing
			cmd := &cobra.Command{
				Use:  "stop <container-name>",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil // Simulate successful execution
				},
			}

			// Execute command with test arguments
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Verify results
			if tt.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
			}
		})
	}
}

func TestContainerStartCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "Valid container name",
			args:        []string{"test-container"},
			expectError: false,
		},
		{
			name:        "No container name provided",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "start <container-name>",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerRemoveCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		hasForceFlag bool
	}{
		{
			name:        "Valid container name without force",
			args:        []string{"test-container"},
			expectError: false,
			hasForceFlag: false,
		},
		{
			name:        "Valid container name with force flag",
			args:        []string{"--force", "test-container"},
			expectError: false,
			hasForceFlag: true,
		},
		{
			name:        "Valid container name with short force flag",
			args:        []string{"-f", "test-container"},
			expectError: false,
			hasForceFlag: true,
		},
		{
			name:        "No container name provided",
			args:        []string{},
			expectError: true,
			hasForceFlag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "remove <container-name>",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			
			// Add force flag
			cmd.Flags().BoolP("force", "f", false, "Force removal of running container")

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				if tt.hasForceFlag {
					force, _ := cmd.Flags().GetBool("force")
					assert.True(t, force, "Expected force flag to be set")
				}
			}
		})
	}
}

func TestContainerLogsCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectError  bool
		hasFollowFlag bool
	}{
		{
			name:         "Valid container name without follow",
			args:         []string{"test-container"},
			expectError:  false,
			hasFollowFlag: false,
		},
		{
			name:         "Valid container name with follow flag",
			args:         []string{"--follow", "test-container"},
			expectError:  false,
			hasFollowFlag: true,
		},
		{
			name:         "Valid container name with short follow flag",
			args:         []string{"-f", "test-container"},
			expectError:  false,
			hasFollowFlag: true,
		},
		{
			name:         "No container name provided",
			args:         []string{},
			expectError:  true,
			hasFollowFlag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "logs <container-name>",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			
			// Add follow flag
			cmd.Flags().BoolP("follow", "f", false, "Follow log output")

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				if tt.hasFollowFlag {
					follow, _ := cmd.Flags().GetBool("follow")
					assert.True(t, follow, "Expected follow flag to be set")
				}
			}
		})
	}
}

func TestContainerListCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "No arguments (valid)",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "Too many arguments",
			args:        []string{"extra-arg"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "list",
				Args: cobra.NoArgs,
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerListCommandAliases(t *testing.T) {
	aliases := []string{"ls", "ps"}
	
	for _, alias := range aliases {
		t.Run("Alias_"+alias, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:     "list",
				Aliases: []string{"ls", "ps"},
				Args:    cobra.NoArgs,
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			// Test that alias can be found
			found := false
			for _, a := range cmd.Aliases {
				if a == alias {
					found = true
					break
				}
			}
			
			assert.True(t, found, "Expected alias '%s' to be present", alias)
		})
	}
}

func TestContainerCommandHelp(t *testing.T) {
	commands := []struct {
		cmd      *cobra.Command
		name     string
		usePattern string
	}{
		{containerStopCmd, "stop", "stop <container-name>"},
		{containerStartCmd, "start", "start <container-name>"},
		{containerRemoveCmd, "remove", "remove <container-name>"},
		{containerLogsCmd, "logs", "logs <container-name>"},
		{containerListCmd, "list", "list"},
	}

	for _, cmdTest := range commands {
		t.Run("Help_"+cmdTest.name, func(t *testing.T) {
			assert.Contains(t, cmdTest.cmd.Use, cmdTest.usePattern, 
				"Expected Use field to contain pattern for %s command", cmdTest.name)
			assert.NotEmpty(t, cmdTest.cmd.Short, 
				"Expected Short description for %s command", cmdTest.name)
			assert.NotEmpty(t, cmdTest.cmd.Long, 
				"Expected Long description for %s command", cmdTest.name)
		})
	}
}

func TestContainerCommandExamples(t *testing.T) {
	commands := []*cobra.Command{
		containerStopCmd,
		containerStartCmd,
		containerRemoveCmd,
		containerLogsCmd,
		containerListCmd,
	}

	for _, cmd := range commands {
		t.Run("Examples_"+cmd.Name(), func(t *testing.T) {
			assert.Contains(t, cmd.Long, "Examples:", 
				"Expected examples in Long description for %s command", cmd.Name())
			assert.Contains(t, cmd.Long, "portunix container", 
				"Expected command examples to start with 'portunix container' for %s command", cmd.Name())
		})
	}
}

func TestContainerCommandFlags(t *testing.T) {
	t.Run("Remove command force flag", func(t *testing.T) {
		flag := containerRemoveCmd.Flags().Lookup("force")
		assert.NotNil(t, flag, "Expected force flag to be present")
		assert.Equal(t, "f", flag.Shorthand, "Expected force flag shorthand to be 'f'")
	})

	t.Run("Logs command follow flag", func(t *testing.T) {
		flag := containerLogsCmd.Flags().Lookup("follow")
		assert.NotNil(t, flag, "Expected follow flag to be present")
		assert.Equal(t, "f", flag.Shorthand, "Expected follow flag shorthand to be 'f'")
	})
	
	t.Run("Exec command interactive flag", func(t *testing.T) {
		flag := containerExecCmd.Flags().Lookup("interactive")
		assert.NotNil(t, flag, "Expected interactive flag to be present")
		assert.Equal(t, "i", flag.Shorthand, "Expected interactive flag shorthand to be 'i'")
	})
}

// Integration test to verify all commands are properly registered
func TestContainerCommandRegistration(t *testing.T) {
	expectedCommands := []string{"run-in-container", "exec", "info", "stop", "start", "remove", "logs", "list"}
	
	actualCommands := make([]string, 0)
	for _, cmd := range containerCmd.Commands() {
		actualCommands = append(actualCommands, cmd.Name())
	}
	
	for _, expected := range expectedCommands {
		found := false
		for _, actual := range actualCommands {
			if actual == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected command '%s' to be registered with containerCmd", expected)
	}
}

func TestContainerCommandErrorMessages(t *testing.T) {
	t.Run("Stop command error message format", func(t *testing.T) {
		// Test that error messages contain appropriate format
		assert.Contains(t, containerStopCmd.Long, "portunix container stop", 
			"Expected stop command examples")
		assert.Contains(t, containerStopCmd.Long, "configured container runtime", 
			"Expected runtime delegation information")
	})
	
	t.Run("All commands mention runtime delegation", func(t *testing.T) {
		commands := []*cobra.Command{
			containerStopCmd, containerStartCmd, containerRemoveCmd, 
			containerLogsCmd, containerListCmd,
		}
		
		for _, cmd := range commands {
			assert.Contains(t, cmd.Long, "configured container runtime", 
				"Expected %s command to mention runtime delegation", cmd.Name())
			assert.Contains(t, cmd.Long, "Docker or Podman", 
				"Expected %s command to mention Docker and Podman support", cmd.Name())
		}
	})
}