/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"strings"
	"testing"
)

// TestContainerExecCommand checks the argument validation of the real exec
// command; it does not execute the command, so no container runtime is called
func TestContainerExecCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
			errMsg:  "requires at least 2 arg(s), only received 0",
		},
		{
			name:    "only container name",
			args:    []string{"test-container"},
			wantErr: true,
			errMsg:  "requires at least 2 arg(s), only received 1",
		},
		{
			name: "valid basic command",
			args: []string{"test-container", "ls", "-la"},
		},
		{
			name: "quoted command",
			args: []string{"test-container", "ls -la /app/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := containerExecCmd.Args(containerExecCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Args(%v) error = %q, want %q", tt.args, err, tt.errMsg)
			}
		})
	}
}

func TestContainerExecInteractiveFlag(t *testing.T) {
	// Test that the interactive flag is properly defined
	cmd := containerExecCmd

	// Check if interactive flag exists
	flag := cmd.Flags().Lookup("interactive")
	if flag == nil {
		t.Error("Interactive flag not found")
		return
	}

	if flag.Shorthand != "i" {
		t.Errorf("Interactive flag shorthand = %v, want 'i'", flag.Shorthand)
	}

	if flag.Usage != "Keep STDIN open and allocate pseudo-TTY" {
		t.Errorf("Interactive flag usage = %v, want 'Keep STDIN open and allocate pseudo-TTY'", flag.Usage)
	}
}

func TestContainerExecArgumentParsing(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantInteractive bool
		wantName        string
		wantCmd         []string
		wantErr         bool
	}{
		{name: "simple command", args: []string{"test-container", "ls"}, wantName: "test-container", wantCmd: []string{"ls"}},
		{name: "command with arguments", args: []string{"web-server", "cat", "/etc/hosts"}, wantName: "web-server", wantCmd: []string{"cat", "/etc/hosts"}},
		{name: "complex command", args: []string{"db-container", "mysql", "-u", "root", "-p"}, wantName: "db-container", wantCmd: []string{"mysql", "-u", "root", "-p"}},
		{name: "interactive -i", args: []string{"-i", "test-container", "sh"}, wantInteractive: true, wantName: "test-container", wantCmd: []string{"sh"}},
		{name: "interactive --interactive", args: []string{"--interactive", "test-container", "sh"}, wantInteractive: true, wantName: "test-container", wantCmd: []string{"sh"}},
		// help example "portunix container exec -it web-server bash"
		{name: "interactive -it", args: []string{"-it", "web-server", "bash"}, wantInteractive: true, wantName: "web-server", wantCmd: []string{"bash"}},
		// flags after the container name belong to the command
		{name: "command flag -i kept", args: []string{"web-server", "grep", "-i", "error", "/var/log/app.log"}, wantName: "web-server", wantCmd: []string{"grep", "-i", "error", "/var/log/app.log"}},
		{name: "only flags", args: []string{"-it", "web-server"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interactive, name, command, err := parseContainerExecArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseContainerExecArgs(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if interactive != tt.wantInteractive {
				t.Errorf("interactive = %v, want %v", interactive, tt.wantInteractive)
			}
			if name != tt.wantName {
				t.Errorf("container name = %q, want %q", name, tt.wantName)
			}
			if strings.Join(command, "\x00") != strings.Join(tt.wantCmd, "\x00") {
				t.Errorf("command = %q, want %q", command, tt.wantCmd)
			}
		})
	}
}
