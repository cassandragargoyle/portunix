/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"testing"
	"time"
)

func TestStripCommandSeparator(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"no separator", []string{"ubuntu:22.04", "bash"}, []string{"ubuntu:22.04", "bash"}},
		{"after image", []string{"ubuntu:22.04", "--", "bash", "-c", "echo test"}, []string{"ubuntu:22.04", "bash", "-c", "echo test"}},
		{"only first removed", []string{"ubuntu:22.04", "--", "sh", "-c", "echo", "--", "x"}, []string{"ubuntu:22.04", "sh", "-c", "echo", "--", "x"}},
		{"trailing", []string{"ubuntu:22.04", "--"}, []string{"ubuntu:22.04"}},
		{"empty", nil, nil},
	}
	for _, c := range cases {
		got := stripCommandSeparator(c.in)
		if !equalStringSlice(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStripCommandSeparator_DoesNotMutateInput(t *testing.T) {
	in := []string{"ubuntu:22.04", "--", "bash"}
	_ = stripCommandSeparator(in)
	if !equalStringSlice(in, []string{"ubuntu:22.04", "--", "bash"}) {
		t.Errorf("input mutated: %v", in)
	}
}

func TestHasDetachFlag(t *testing.T) {
	cases := []struct {
		in   []string
		want bool
	}{
		{[]string{"-d", "--name", "test", "ubuntu:22.04"}, true},
		{[]string{"--detach", "ubuntu:22.04"}, true},
		{[]string{"ubuntu:22.04", "bash"}, false},
		// -d after the separator is an argument of the container command
		{[]string{"ubuntu:22.04", "--", "ls", "-d", "/tmp"}, false},
		{[]string{"-d", "ubuntu:22.04", "--", "sleep", "infinity"}, true},
	}
	for _, c := range cases {
		if got := hasDetachFlag(c.in); got != c.want {
			t.Errorf("hasDetachFlag(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

// runArgsFor mirrors handleContainerRun for the direct (non-managed) path
func runArgsFor(args []string, tty bool) []string {
	detached := hasDetachFlag(args)
	cleaned := stripCommandSeparator(args)
	return buildRunArgs(cleaned[0], cleaned[1:], detached, tty)
}

func TestBuildRunArgs_Separator(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		tty  bool
		want []string
	}{
		{
			"detached with separator (help example)",
			[]string{"-d", "--name", "test", "ubuntu:22.04", "--", "bash", "-c", "echo test"},
			false,
			[]string{"run", "-d", "--name", "test", "ubuntu:22.04", "bash", "-c", "echo test"},
		},
		{
			"detached without separator",
			[]string{"-d", "--name", "test", "ubuntu:22.04", "bash", "-c", "echo test"},
			false,
			[]string{"run", "-d", "--name", "test", "ubuntu:22.04", "bash", "-c", "echo test"},
		},
		{
			"interactive tty with separator",
			[]string{"ubuntu:22.04", "--", "sh", "-c", "echo -- x"},
			true,
			[]string{"run", "-it", "--rm", "ubuntu:22.04", "sh", "-c", "echo -- x"},
		},
		{
			"interactive non-tty keeps second separator",
			[]string{"ubuntu:22.04", "--", "echo", "--", "x"},
			false,
			[]string{"run", "-i", "--rm", "ubuntu:22.04", "echo", "--", "x"},
		},
		{
			"interactive without separator",
			[]string{"ubuntu:22.04", "echo", "hello"},
			false,
			[]string{"run", "-i", "--rm", "ubuntu:22.04", "echo", "hello"},
		},
		{
			"-d in command does not detach",
			[]string{"ubuntu:22.04", "--", "ls", "-d", "/tmp"},
			false,
			[]string{"run", "-i", "--rm", "ubuntu:22.04", "ls", "-d", "/tmp"},
		},
	}
	for _, c := range cases {
		got := runArgsFor(c.in, c.tty)
		if !equalStringSlice(got, c.want) {
			t.Errorf("%s:\n got  %v\n want %v", c.name, got, c.want)
		}
	}
}

func TestExtractLifecycleFlags_StopsAtSeparator(t *testing.T) {
	args := []string{"--ttl", "1h", "ubuntu:22.04", "--", "echo", "--ttl", "--auto-cleanup"}
	rest, lc, err := extractLifecycleFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lc.TTL != time.Hour {
		t.Errorf("TTL=%s", lc.TTL)
	}
	if lc.AutoCleanup {
		t.Error("--auto-cleanup after separator must not be parsed")
	}
	want := []string{"ubuntu:22.04", "--", "echo", "--ttl", "--auto-cleanup"}
	if !equalStringSlice(rest, want) {
		t.Errorf("rest = %v\nwant %v", rest, want)
	}
}

// managedRunArgsFor mirrors handleContainerRun for the managed (lifecycle) path
func managedRunArgsFor(t *testing.T, args []string, tty bool) []string {
	t.Helper()
	cleaned, lc, err := extractLifecycleFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	detached := hasDetachFlag(cleaned)
	cleaned = stripCommandSeparator(cleaned)
	return buildManagedRunArgs(lc, cleaned, detached, tty, time.Unix(0, 0))
}

// tailAfterRun returns the last n args (user flags + image + command)
func tailAfterRun(args []string, n int) []string {
	return args[len(args)-n:]
}

func TestBuildManagedRunArgs_Separator(t *testing.T) {
	cases := []struct {
		name     string
		in       []string
		tty      bool
		wantHead []string
		wantTail []string
	}{
		{
			"detached with separator",
			[]string{"--ttl", "2h", "-d", "--name", "test", "ubuntu:22.04", "--", "bash", "-c", "echo test"},
			false,
			[]string{"run", "--label"},
			[]string{"-d", "--name", "test", "ubuntu:22.04", "bash", "-c", "echo test"},
		},
		{
			"detached without separator",
			[]string{"--ttl", "2h", "-d", "--name", "test", "ubuntu:22.04", "bash", "-c", "echo test"},
			false,
			[]string{"run", "--label"},
			[]string{"-d", "--name", "test", "ubuntu:22.04", "bash", "-c", "echo test"},
		},
		{
			"interactive with separator keeps second one",
			[]string{"--auto-cleanup", "ubuntu:22.04", "--", "sh", "-c", "echo", "--", "x"},
			true,
			[]string{"run", "-it", "--rm"},
			[]string{"ubuntu:22.04", "sh", "-c", "echo", "--", "x"},
		},
		{
			"interactive non-tty without separator",
			[]string{"--auto-cleanup", "ubuntu:22.04", "echo", "hello"},
			false,
			[]string{"run", "-i", "--rm"},
			[]string{"ubuntu:22.04", "echo", "hello"},
		},
	}
	for _, c := range cases {
		got := managedRunArgsFor(t, c.in, c.tty)
		if len(got) < len(c.wantHead)+len(c.wantTail) {
			t.Errorf("%s: too few args: %v", c.name, got)
			continue
		}
		if !equalStringSlice(got[:len(c.wantHead)], c.wantHead) {
			t.Errorf("%s: head\n got  %v\n want %v", c.name, got, c.wantHead)
		}
		if !equalStringSlice(tailAfterRun(got, len(c.wantTail)), c.wantTail) {
			t.Errorf("%s: tail\n got  %v\n want %v", c.name, got, c.wantTail)
		}
		for _, a := range got {
			if a == "--" && !containsSeparatorInCommand(c.wantTail) {
				t.Errorf("%s: separator leaked to runtime args: %v", c.name, got)
			}
		}
	}
}

func containsSeparatorInCommand(tail []string) bool {
	for _, a := range tail {
		if a == "--" {
			return true
		}
	}
	return false
}

func TestHasHelpFlag(t *testing.T) {
	cases := []struct {
		in   []string
		want bool
	}{
		{[]string{"--help"}, true},
		{[]string{"-h"}, true},
		{[]string{"-d", "ubuntu:22.04", "--help"}, true},
		{[]string{"ubuntu:22.04", "bash"}, false},
		// -h after the separator is an argument of the container command
		{[]string{"ubuntu:22.04", "--", "ls", "-h", "/"}, false},
		{[]string{"ubuntu:22.04", "--", "cmd", "--help"}, false},
		{nil, false},
	}
	for _, c := range cases {
		if got := hasHelpFlag(c.in); got != c.want {
			t.Errorf("hasHelpFlag(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseExecArgs(t *testing.T) {
	cases := []struct {
		name     string
		in       []string
		wantName string
		wantCmd  []string
		wantHelp bool
	}{
		{"without separator", []string{"t238", "echo", "x"}, "t238", []string{"echo", "x"}, false},
		{"separator after name", []string{"t238", "--", "echo", "x"}, "t238", []string{"echo", "x"}, false},
		{"separator before name", []string{"--", "t238", "echo", "x"}, "t238", []string{"echo", "x"}, false},
		{"only first separator removed", []string{"t238", "--", "sh", "-c", "echo -- x", "--"}, "t238", []string{"sh", "-c", "echo -- x", "--"}, false},
		{"help flag in command", []string{"t238", "ls", "-h", "/"}, "t238", []string{"ls", "-h", "/"}, false},
		{"help flag in command after separator", []string{"t238", "--", "ls", "--help"}, "t238", []string{"ls", "--help"}, false},
		{"long help", []string{"--help"}, "", nil, true},
		{"short help before name", []string{"-h", "t238", "ls"}, "", nil, true},
		{"no args", nil, "", nil, true},
		{"name only", []string{"t238"}, "", nil, true},
		{"name and separator only", []string{"t238", "--"}, "", nil, true},
	}
	for _, c := range cases {
		name, cmd, help := parseExecArgs(c.in)
		if name != c.wantName || help != c.wantHelp || !equalStringSlice(cmd, c.wantCmd) {
			t.Errorf("%s: got (%q, %v, %v), want (%q, %v, %v)",
				c.name, name, cmd, help, c.wantName, c.wantCmd, c.wantHelp)
		}
	}
}

func TestBuildExecArgs(t *testing.T) {
	got := buildExecArgs("t238", []string{"echo", "x"}, false)
	if want := []string{"exec", "-i", "t238", "echo", "x"}; !equalStringSlice(got, want) {
		t.Errorf("non-tty: got %v, want %v", got, want)
	}
	got = buildExecArgs("t238", []string{"bash"}, true)
	if want := []string{"exec", "-it", "t238", "bash"}; !equalStringSlice(got, want) {
		t.Errorf("tty: got %v, want %v", got, want)
	}
}
