/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestContainerRunCommand_FlagParsing tests flag parsing for container run command
func TestContainerRunCommand_FlagParsing(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectedErr bool
		checkFlags  func(*testing.T, *cobra.Command)
	}{
		{
			name:        "TC-038-U001: Basic shorthand flags",
			args:        []string{"-d", "-i", "-t", "ubuntu:22.04"},
			expectedErr: false,
			checkFlags: func(t *testing.T, cmd *cobra.Command) {
				detach, _ := cmd.Flags().GetBool("detach")
				interactive, _ := cmd.Flags().GetBool("interactive")
				tty, _ := cmd.Flags().GetBool("tty")

				assert.True(t, detach, "Detach flag should be true")
				assert.True(t, interactive, "Interactive flag should be true")
				assert.True(t, tty, "TTY flag should be true")
			},
		},
		{
			name:        "TC-038-U002: Combined shorthand flags",
			args:        []string{"-dit", "ubuntu:22.04"},
			expectedErr: false,
			checkFlags: func(t *testing.T, cmd *cobra.Command) {
				detach, _ := cmd.Flags().GetBool("detach")
				interactive, _ := cmd.Flags().GetBool("interactive")
				tty, _ := cmd.Flags().GetBool("tty")

				assert.True(t, detach, "Detach flag should be true")
				assert.True(t, interactive, "Interactive flag should be true")
				assert.True(t, tty, "TTY flag should be true")
			},
		},
		{
			name:        "TC-038-U003: Flags with values",
			args:        []string{"-d", "--name", "test-container", "-p", "8080:80", "-v", "/host:/container", "-e", "NODE_ENV=production", "ubuntu:22.04"},
			expectedErr: false,
			checkFlags: func(t *testing.T, cmd *cobra.Command) {
				detach, _ := cmd.Flags().GetBool("detach")
				name, _ := cmd.Flags().GetString("name")
				ports, _ := cmd.Flags().GetStringSlice("port")
				volumes, _ := cmd.Flags().GetStringSlice("volume")
				env, _ := cmd.Flags().GetStringSlice("env")

				assert.True(t, detach, "Detach flag should be true")
				assert.Equal(t, "test-container", name, "Container name should be 'test-container'")
				assert.Contains(t, ports, "8080:80", "Port mapping should contain '8080:80'")
				assert.Contains(t, volumes, "/host:/container", "Volume mapping should contain '/host:/container'")
				assert.Contains(t, env, "NODE_ENV=production", "Environment should contain 'NODE_ENV=production'")
			},
		},
		{
			name:        "TC-038-U004: Multiple port and volume mappings",
			args:        []string{"-d", "-p", "8080:80", "-p", "9090:90", "-v", "/data:/app/data", "-v", "/logs:/app/logs", "ubuntu:22.04"},
			expectedErr: false,
			checkFlags: func(t *testing.T, cmd *cobra.Command) {
				ports, _ := cmd.Flags().GetStringSlice("port")
				volumes, _ := cmd.Flags().GetStringSlice("volume")

				assert.Len(t, ports, 2, "Should have 2 port mappings")
				assert.Contains(t, ports, "8080:80", "Should contain first port mapping")
				assert.Contains(t, ports, "9090:90", "Should contain second port mapping")

				assert.Len(t, volumes, 2, "Should have 2 volume mappings")
				assert.Contains(t, volumes, "/data:/app/data", "Should contain first volume mapping")
				assert.Contains(t, volumes, "/logs:/app/logs", "Should contain second volume mapping")
			},
		},
		{
			name:        "TC-038-U005: Long form flags",
			args:        []string{"--detach", "--interactive", "--tty", "--name", "test", "ubuntu:22.04"},
			expectedErr: false,
			checkFlags: func(t *testing.T, cmd *cobra.Command) {
				detach, _ := cmd.Flags().GetBool("detach")
				interactive, _ := cmd.Flags().GetBool("interactive")
				tty, _ := cmd.Flags().GetBool("tty")
				name, _ := cmd.Flags().GetString("name")

				assert.True(t, detach, "Detach flag should be true")
				assert.True(t, interactive, "Interactive flag should be true")
				assert.True(t, tty, "TTY flag should be true")
				assert.Equal(t, "test", name, "Container name should be 'test'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new command instance to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "run [flags] <image> [command...]",
				Args: cobra.MinimumNArgs(1),
				Run:  func(cmd *cobra.Command, args []string) {}, // No-op for testing
			}
			addContainerRunFlags(cmd) // flags of the real run command

			// Parse the flags
			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectedErr {
				assert.Error(t, err, "Expected error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tc.name)
				if tc.checkFlags != nil {
					tc.checkFlags(t, cmd)
				}
			}
		})
	}
}

// TestContainerRunCommand_ArgumentParsing tests argument parsing (image and command)
func TestContainerRunCommand_ArgumentParsing(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedImage string
		expectedCmd   []string
		expectedErr   bool
	}{
		{
			name:          "TC-038-U006: Basic image only",
			args:          []string{"ubuntu:22.04"},
			expectedImage: "ubuntu:22.04",
			expectedCmd:   []string{},
			expectedErr:   false,
		},
		{
			name:          "TC-038-U007: Image with simple command",
			args:          []string{"ubuntu:22.04", "echo", "hello"},
			expectedImage: "ubuntu:22.04",
			expectedCmd:   []string{"echo", "hello"},
			expectedErr:   false,
		},
		{
			name:          "TC-038-U008: Image with complex command",
			args:          []string{"ubuntu:22.04", "bash", "-c", "apt-get update && sleep 10"},
			expectedImage: "ubuntu:22.04",
			expectedCmd:   []string{"bash", "-c", "apt-get update && sleep 10"},
			expectedErr:   false,
		},
		{
			name:          "TC-038-U008b: Flags, then separator after the image",
			args:          []string{"-d", "--name", "test", "ubuntu:22.04", "--", "bash", "-c", "echo test"},
			expectedImage: "ubuntu:22.04",
			expectedCmd:   []string{"bash", "-c", "echo test"},
			expectedErr:   false,
		},
		{
			name:          "TC-038-U008c: Flag-like arguments after the image belong to the command",
			args:          []string{"ubuntu:22.04", "ls", "-d", "/tmp"},
			expectedImage: "ubuntu:22.04",
			expectedCmd:   []string{"ls", "-d", "/tmp"},
			expectedErr:   false,
		},
		{
			name:        "TC-038-U009: No arguments - should fail",
			args:        []string{},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "run [flags] <image> [command...]",
				Args: cobra.MinimumNArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					if !tc.expectedErr && len(args) > 0 {
						image, command := splitContainerRunArgs(args)

						assert.Equal(t, tc.expectedImage, image, "Image should match expected")
						assert.Equal(t, tc.expectedCmd, command, "Command should match expected")
					}
				},
			}
			addContainerRunFlags(cmd) // flags of the real run command

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectedErr {
				assert.Error(t, err, "Expected error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tc.name)
			}
		})
	}
}

// TestContainerRunCommand_FlagValidation tests flag validation and error handling
func TestContainerRunCommand_FlagValidation(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectedErr bool
		errorMsg    string
	}{
		{
			name:        "TC-038-U010: Invalid shorthand flag",
			args:        []string{"-x", "ubuntu:22.04"},
			expectedErr: true,
			errorMsg:    "unknown shorthand flag",
		},
		{
			name:        "TC-038-U011: Invalid long flag",
			args:        []string{"--invalid", "ubuntu:22.04"},
			expectedErr: true,
			errorMsg:    "unknown flag",
		},
		{
			// --name takes "ubuntu:22.04" as its value, so no image is left
			name:        "TC-038-U012: Missing flag value",
			args:        []string{"--name", "ubuntu:22.04"},
			expectedErr: true,
			errorMsg:    "requires at least 1 arg(s)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "run [flags] <image> [command...]",
				Args: cobra.MinimumNArgs(1),
				Run:  func(cmd *cobra.Command, args []string) {},
			}
			addContainerRunFlags(cmd) // flags of the real run command

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectedErr {
				assert.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tc.name)
			}
		})
	}
}

// TestContainerRunCommand_OriginalIssueRepro tests the original issue reproduction
func TestContainerRunCommand_OriginalIssueRepro(t *testing.T) {
	t.Run("TC-038-U013: Original issue command should work", func(t *testing.T) {
		// This is the exact command that was failing in the original issue
		cmd := &cobra.Command{
			Use:  "run [flags] <image> [command...]",
			Args: cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify the arguments are parsed correctly
				detach, _ := cmd.Flags().GetBool("detach")
				name, _ := cmd.Flags().GetString("name")

				assert.True(t, detach, "Detach flag should be true")
				assert.Equal(t, "portunix-e2e-test", name, "Container name should be correct")
				assert.Equal(t, "ubuntu:22.04", args[0], "Image should be ubuntu:22.04")
				assert.Equal(t, []string{"bash", "-c", "apt-get update && apt-get install -y curl wget git python3 python3-pip nodejs npm && sleep 3600"}, args[1:], "Command should be parsed correctly")
			},
		}
		addContainerRunFlags(cmd) // flags of the real run command

		// The original failing command
		args := []string{"-d", "--name", "portunix-e2e-test", "ubuntu:22.04", "bash", "-c", "apt-get update && apt-get install -y curl wget git python3 python3-pip nodejs npm && sleep 3600"}

		cmd.SetArgs(args)
		err := cmd.Execute()

		// This should NOT fail with "unknown shorthand flag: 'd'" error
		assert.NoError(t, err, "Original issue command should execute without flag parsing errors")
	})
}
