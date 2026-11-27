/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newHelpFlagTestTree builds a root with the help level flags and a
// subcommand that records whether it ran
func newHelpFlagTestTree(ran *bool, passThrough bool) *cobra.Command {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().BoolVar(&helpExpert, "help-expert", false, "")
	root.PersistentFlags().BoolVar(&helpAI, "help-ai", false, "")
	sub := &cobra.Command{
		Use:                "sub",
		DisableFlagParsing: passThrough,
		Run:                func(cmd *cobra.Command, args []string) { *ran = true },
	}
	root.AddCommand(sub)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	return root
}

// TestGuardHelpLevelFlags_SubcommandShowsHelp verifies that --help-expert and
// --help-ai on a subcommand show help instead of running it (issue #203)
func TestGuardHelpLevelFlags_SubcommandShowsHelp(t *testing.T) {
	for _, flag := range []string{"--help-expert", "--help-ai"} {
		t.Run(flag, func(t *testing.T) {
			t.Cleanup(func() { helpExpert, helpAI = false, false })
			ran := false
			root := newHelpFlagTestTree(&ran, false)
			guardHelpLevelFlags(root)
			root.SetArgs([]string{"sub", flag})
			if err := root.Execute(); err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if ran {
				t.Errorf("subcommand ran with %s", flag)
			}
		})
	}
}

// TestGuardHelpLevelFlags_SubcommandRunsWithoutFlag verifies normal execution
func TestGuardHelpLevelFlags_SubcommandRunsWithoutFlag(t *testing.T) {
	t.Cleanup(func() { helpExpert, helpAI = false, false })
	ran := false
	root := newHelpFlagTestTree(&ran, false)
	guardHelpLevelFlags(root)
	root.SetArgs([]string{"sub"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !ran {
		t.Error("subcommand did not run")
	}
}

// TestGuardHelpLevelFlags_PassThroughCommand verifies that a command with
// DisableFlagParsing shows help for a raw help level argument (e.g. the
// `docker install` alias, which ran the installation before) and runs
// otherwise (issue #203)
func TestGuardHelpLevelFlags_PassThroughCommand(t *testing.T) {
	cases := []struct {
		args    []string
		wantRun bool
	}{
		{[]string{"sub", "--help-expert"}, false},
		{[]string{"sub", "--dry-run", "--help-ai"}, false},
		{[]string{"sub", "--help-ai=true"}, false},
		{[]string{"sub", "--help-expert=false"}, true},
		{[]string{"sub", "--dry-run"}, true},
		{[]string{"sub", "--", "--help-expert"}, true},
	}
	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			t.Cleanup(func() { helpExpert, helpAI = false, false })
			ran := false
			root := newHelpFlagTestTree(&ran, true)
			guardHelpLevelFlags(root)
			root.SetArgs(c.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if ran != c.wantRun {
				t.Errorf("ran = %v, want %v", ran, c.wantRun)
			}
		})
	}
}

func TestParseMapFolder(t *testing.T) {
	valid := []struct {
		in, host, sandbox string
		readOnly          bool
	}{
		{`C:\Host:C:\Sandbox:true`, `C:\Host`, `C:\Sandbox`, true},
		{`D:\a b\c:C:\Users\WDAGUtilityAccount\Desktop\x:false`, `D:\a b\c`, `C:\Users\WDAGUtilityAccount\Desktop\x`, false},
		{`C:/Host:C:/Sandbox:false`, `C:/Host`, `C:/Sandbox`, false},
		{`relative:C:\Sandbox:true`, `relative`, `C:\Sandbox`, true},
		{`/host:/sandbox:true`, `/host`, `/sandbox`, true},
	}
	for _, c := range valid {
		got, err := parseMapFolder(c.in)
		if err != nil {
			t.Errorf("parseMapFolder(%q) error: %v", c.in, err)
			continue
		}
		if got.HostPath != c.host || got.SandboxPath != c.sandbox || got.ReadOnly != c.readOnly {
			t.Errorf("parseMapFolder(%q) = %+v, want %s | %s | %v", c.in, got, c.host, c.sandbox, c.readOnly)
		}
	}

	for _, in := range []string{
		`C:\Host`,
		`C:\Host:true`,
		`C:\Host:C:\Sandbox:maybe`,
		`a:b:c:true`,
		`:C:\Sandbox:true`,
	} {
		if _, err := parseMapFolder(in); err == nil {
			t.Errorf("parseMapFolder(%q) should fail", in)
		}
	}
}
