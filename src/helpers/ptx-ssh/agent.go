/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// AgentStart spawns ssh-agent and echoes its environment in a shell-friendly
// form. The caller is expected to `eval $(portunix ssh agent start)`.
func AgentStart() error {
	if runtime.GOOS == "windows" {
		return errors.New("agent start is not supported on Windows (use the system ssh-agent service)")
	}
	out, err := exec.Command("ssh-agent", "-s").Output()
	if err != nil {
		return fmt.Errorf("start ssh-agent: %w", err)
	}
	os.Stdout.Write(out)
	return nil
}

// AgentAdd invokes ssh-add with the given identity path. Passphrase prompts
// are forwarded to the controlling terminal by ssh-add itself.
func AgentAdd(identity string) error {
	args := []string{}
	if identity != "" {
		expanded, err := expandPath(identity)
		if err != nil {
			return err
		}
		args = append(args, expanded)
	}
	cmd := exec.Command("ssh-add", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AgentStop sends the shutdown request to the current agent and prints the
// shell unset lines so the caller can `eval` them.
func AgentStop() error {
	sock := os.Getenv("SSH_AUTH_SOCK")
	pidEnv := os.Getenv("SSH_AGENT_PID")
	if sock == "" && pidEnv == "" {
		return errors.New("no SSH_AGENT_PID / SSH_AUTH_SOCK in environment; agent may not be running")
	}
	if pidEnv != "" {
		if pid, err := strconv.Atoi(pidEnv); err == nil {
			proc, err := os.FindProcess(pid)
			if err == nil {
				_ = proc.Kill()
			}
		}
	}
	fmt.Println("unset SSH_AUTH_SOCK;")
	fmt.Println("unset SSH_AGENT_PID;")
	return nil
}

// readLineFromReader is a tiny helper used by tests that fake the terminal.
func readLineFromReader(r io.Reader) (string, error) {
	rd := bufio.NewReader(r)
	line, err := rd.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
