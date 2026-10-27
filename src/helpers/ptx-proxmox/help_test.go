/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout returns everything fn writes to os.Stdout
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(out)
}

func TestHandleMetaArg_HelpAI(t *testing.T) {
	var handled bool
	out := captureStdout(t, func() { handled = handleMetaArg("--help-ai") })
	if !handled {
		t.Fatal("--help-ai was not handled")
	}

	var help struct {
		Tool     string `json:"tool"`
		Commands []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(out), &help); err != nil {
		t.Fatalf("--help-ai output is not valid JSON: %v\n%s", err, out)
	}
	if help.Tool != "ptx-proxmox" {
		t.Errorf("tool = %q, want %q", help.Tool, "ptx-proxmox")
	}
	if len(help.Commands) == 0 {
		t.Fatal("--help-ai lists no commands")
	}

	// Every listed command must exist in the Cobra tree
	for _, c := range help.Commands {
		found, _, err := rootCmd.Find(strings.Fields(c.Name))
		if err != nil || found == nil || !strings.HasSuffix(found.CommandPath(), c.Name) {
			t.Errorf("command %q from --help-ai not found in command tree", c.Name)
		}
		if c.Description == "" {
			t.Errorf("command %q has empty description", c.Name)
		}
	}
}

func TestHandleMetaArg_HelpExpert(t *testing.T) {
	var handled bool
	out := captureStdout(t, func() { handled = handleMetaArg("--help-expert") })
	if !handled {
		t.Fatal("--help-expert was not handled")
	}
	for _, section := range []string{"DESCRIPTION:", "COMMANDS:", "EXAMPLES:"} {
		if !strings.Contains(out, section) {
			t.Errorf("--help-expert output missing section %q", section)
		}
	}
}

func TestHandleMetaArg_OtherArgsPassThrough(t *testing.T) {
	for _, arg := range []string{"--help", "--version", "unknown"} {
		if handleMetaArg(arg) {
			t.Errorf("handleMetaArg(%q) = true, want false", arg)
		}
	}
}

func TestHandleMetaArgs_DirectAndDispatcherForm(t *testing.T) {
	cases := [][]string{{"--help-ai"}, {"--help-expert"}}
	for _, name := range []string{"proxmox"} {
		cases = append(cases, []string{name, "--help-ai"}, []string{name, "--help-expert"})
	}
	for _, args := range cases {
		var handled bool
		out := captureStdout(t, func() { handled = handleMetaArgs(args) })
		if !handled {
			t.Errorf("handleMetaArgs(%q) = false, want true", args)
		}
		if strings.TrimSpace(out) == "" {
			t.Errorf("handleMetaArgs(%q) printed nothing", args)
		}
	}
}

func TestHandleMetaArgs_NotHandled(t *testing.T) {
	cases := [][]string{
		nil,
		{"--help"},
		{"unknown", "--help-ai"},
		{"proxmox", "sub", "--help-ai"},
		{"proxmox", "--help"},
	}
	for _, args := range cases {
		if handleMetaArgs(args) {
			t.Errorf("handleMetaArgs(%q) = true, want false", args)
		}
	}
}
