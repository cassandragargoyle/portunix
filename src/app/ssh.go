/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// LnxExecutePyScriptSsh is a deprecated shim kept for backwards compatibility
// with internal callers. The previous implementation used a hardcoded private
// key path and ssh.InsecureIgnoreHostKey(), which is explicitly unsafe.
//
// The functional replacement is the ptx-ssh helper binary, invoked via the
// Portunix dispatcher. This shim shells out to ptx-ssh so existing callers
// keep working without re-introducing the legacy InsecureIgnoreHostKey path.
//
// Deprecated: call the ptx-ssh helper directly. This function will be removed
// in a future release — track Issue #174 follow-up.
func LnxExecutePyScriptSsh(host string, port string, user string, localScriptPath string) error {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "22"
	}
	if user == "" {
		return errors.New("LnxExecutePyScriptSsh: user is required (deprecated shim over ptx-ssh)")
	}
	if localScriptPath == "" {
		return errors.New("LnxExecutePyScriptSsh: script path is required")
	}

	bin, err := locatePtxSSH()
	if err != nil {
		return fmt.Errorf("LnxExecutePyScriptSsh: ptx-ssh helper not found: %w", err)
	}

	target := fmt.Sprintf("%s@%s:%s", user, host, port)
	remoteScriptPath := fmt.Sprintf("/tmp/%s", filepath.Base(localScriptPath))

	// Upload via ptx-ssh copy.
	upload := exec.Command(bin, "ssh", "copy", localScriptPath,
		fmt.Sprintf("%s@%s:%s", user, host, remoteScriptPath))
	upload.Stdin = os.Stdin
	upload.Stdout = os.Stdout
	upload.Stderr = os.Stderr
	if err := upload.Run(); err != nil {
		return fmt.Errorf("LnxExecutePyScriptSsh: upload failed: %w", err)
	}

	// Run via ptx-ssh exec.
	run := exec.Command(bin, "ssh", "exec", target,
		fmt.Sprintf("chmod +x %s && %s", remoteScriptPath, remoteScriptPath))
	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	return run.Run()
}

func locatePtxSSH() (string, error) {
	name := "ptx-ssh"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), name)
		if _, serr := os.Stat(candidate); serr == nil {
			return candidate, nil
		}
	}
	return exec.LookPath(name)
}
