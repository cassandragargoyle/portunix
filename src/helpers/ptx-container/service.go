/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Service paths follow the existing ptx-mcp convention (ptx-mcp/common.go)
// of stashing daemon state under ~/.portunix/<helper>/.
const (
	serviceSubDir       = "container"
	servicePIDName      = "lifecycle.pid"
	serviceLogName      = "lifecycle.log"
	defaultPollInterval = 30 * time.Second
)

// LifecycleServiceConfig is the minimal config the daemon needs at runtime.
// Wider configurability (e.g. pluggable backends, structured logs) was
// considered for issue #027 but rejected to keep the surface small — there is
// only one polling loop with one knob.
type LifecycleServiceConfig struct {
	Interval time.Duration
}

// handleContainerService is the dispatch entry point for `container service`.
// It accepts: start [--interval D] [--detach] | stop | status | run.
//   - run runs the polling loop in the foreground (used internally by --detach
//     and exposed for systemd-style supervision).
//   - start handles the user-facing daemonisation.
func handleContainerService(args []string) {
	if len(args) == 0 {
		showServiceHelp()
		return
	}
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showServiceHelp()
			return
		}
	}

	switch args[0] {
	case "start":
		serviceStart(args[1:])
	case "stop":
		serviceStop(args[1:])
	case "status":
		serviceStatus(args[1:])
	case "run":
		serviceRun(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown service subcommand: %s\n", args[0])
		showServiceHelp()
		os.Exit(2)
	}
}

func serviceStart(args []string) {
	cfg := LifecycleServiceConfig{Interval: defaultPollInterval}
	detach := true // default to background; --foreground keeps it attached

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--interval":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "❌ --interval requires a duration")
				os.Exit(2)
			}
			d, err := ParseDuration(args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				os.Exit(2)
			}
			cfg.Interval = d
			i++
		case "--detach":
			detach = true
		case "--foreground":
			detach = false
		default:
			fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", args[i])
			os.Exit(2)
		}
	}

	if pid, ok := readServicePID(); ok && processAlive(pid) {
		fmt.Fprintf(os.Stderr, "❌ Lifecycle service already running (pid %d)\n", pid)
		os.Exit(1)
	}

	if !detach {
		runLifecycleLoop(cfg)
		return
	}

	// Re-exec ourselves with `service run` so we end up with a stand-alone
	// long-lived process. We rely on os.StartProcess + Setsid (Linux) so the
	// child detaches from the parent. On Windows this still spawns a child
	// process; the caller can stop it via `service stop`.
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot resolve executable: %v\n", err)
		os.Exit(1)
	}

	logPath, err := serviceLogPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot open log file %s: %v\n", logPath, err)
		os.Exit(1)
	}

	// `portunix container service run --interval D` is the canonical re-entry
	// point. The dispatcher routes top-level "container" to ptx-container, so
	// this works whether ptx-container is invoked standalone or via the main
	// portunix binary.
	cmd := exec.Command(exe, "container", "service", "run",
		"--interval", cfg.Interval.String())
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.SysProcAttr = newDetachedSysProcAttr()

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		fmt.Fprintf(os.Stderr, "❌ Failed to start daemon: %v\n", err)
		os.Exit(1)
	}
	// Parent does not Wait — daemon owns the log file fd from now on. We
	// release ours after the child has dup'd it.
	go func() { _ = logFile.Close() }()

	if err := writeServicePID(cmd.Process.Pid); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Started daemon (pid %d) but failed to write pid file: %v\n",
			cmd.Process.Pid, err)
		os.Exit(1)
	}
	fmt.Printf("✅ Lifecycle service started (pid %d, interval %s)\n",
		cmd.Process.Pid, cfg.Interval)
	fmt.Printf("   logs: %s\n", logPath)
}

func serviceRun(args []string) {
	cfg := LifecycleServiceConfig{Interval: defaultPollInterval}
	for i := 0; i < len(args); i++ {
		if args[i] == "--interval" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--interval requires a duration")
				os.Exit(2)
			}
			d, err := ParseDuration(args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}
			cfg.Interval = d
			i++
		}
	}
	// When invoked as the daemon child (via `service start --detach`), make
	// sure we own a fresh PID file. Foreground starts already wrote it.
	_ = writeServicePID(os.Getpid())
	runLifecycleLoop(cfg)
}

func serviceStop(_ []string) {
	pid, ok := readServicePID()
	if !ok {
		fmt.Println("ℹ️  No lifecycle service running")
		return
	}
	if !processAlive(pid) {
		fmt.Printf("⚠️  Stale pid file (pid %d not running) — clearing\n", pid)
		_ = removeServicePID()
		return
	}
	if err := signalProcess(pid, syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to signal pid %d: %v\n", pid, err)
		os.Exit(1)
	}
	// Daemon may be mid-tick inside a synchronous `podman rm -f`; that call
	// can block for several seconds when the runtime is busy. 15s is a
	// pragmatic upper bound that lets a typical cleanup finish without
	// resorting to SIGKILL.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if processAlive(pid) {
		fmt.Fprintf(os.Stderr, "⚠️  Service still running after SIGTERM, sending SIGKILL\n")
		_ = signalProcess(pid, syscall.SIGKILL)
	}
	_ = removeServicePID()
	fmt.Printf("✅ Lifecycle service stopped (pid %d)\n", pid)
}

func serviceStatus(_ []string) {
	pid, ok := readServicePID()
	if !ok {
		fmt.Println("Status: stopped")
		return
	}
	if !processAlive(pid) {
		fmt.Printf("Status: stopped (stale pid file %d)\n", pid)
		return
	}
	fmt.Printf("Status: running (pid %d)\n", pid)
	if path, err := serviceLogPath(); err == nil {
		fmt.Printf("Logs:   %s\n", path)
	}
}

// runLifecycleLoop polls FindExpired at the configured interval and removes
// expired containers. Errors are logged and do not stop the loop — a transient
// runtime failure (e.g. dockerd restart) should not take the daemon down.
func runLifecycleLoop(cfg LifecycleServiceConfig) {
	logf("lifecycle service starting (pid=%d interval=%s)", os.Getpid(), cfg.Interval)
	defer logf("lifecycle service exiting (pid=%d)", os.Getpid())

	stop := make(chan struct{})
	registerSignalShutdown(stop)
	// Make sure the pid file is removed on normal/graceful exit too.
	defer func() { _ = removeServicePID() }()

	tick := time.NewTicker(cfg.Interval)
	defer tick.Stop()

	// Run once immediately so that --interval doesn't gate the first sweep.
	tickOnce()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
			tickOnce()
		}
	}
}

func tickOnce() {
	expired, err := FindExpired(time.Now())
	if err != nil {
		logf("FindExpired error: %v", err)
		return
	}
	if len(expired) == 0 {
		return
	}
	for _, mc := range expired {
		if err := removeManaged(mc, false); err != nil {
			logf("cleanup %s:%s failed: %v", mc.Runtime, mc.Name, err)
			continue
		}
		logf("cleanup %s:%s removed (TTL expired)", mc.Runtime, mc.Name)
	}
}

func serviceDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	dir := filepath.Join(home, ".portunix", serviceSubDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}

func servicePIDPath() (string, error) {
	dir, err := serviceDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, servicePIDName), nil
}

func serviceLogPath() (string, error) {
	dir, err := serviceDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, serviceLogName), nil
}

func readServicePID() (int, bool) {
	path, err := servicePIDPath()
	if err != nil {
		return 0, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

func writeServicePID(pid int) error {
	path, err := servicePIDPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o600)
}

func removeServicePID() error {
	path, err := servicePIDPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// logf writes a single timestamped line to the service log. We open and close
// per-line so concurrent writers (e.g. signal handler running in parallel with
// the loop) don't interleave half-formed bytes — the cost is negligible at the
// 30-second cadence.
func logf(format string, args ...any) {
	line := fmt.Sprintf("%s "+format+"\n",
		append([]any{time.Now().UTC().Format(time.RFC3339)}, args...)...)
	path, err := serviceLogPath()
	if err != nil {
		// Last-resort: stderr still works when home dir lookup fails.
		_, _ = io.WriteString(os.Stderr, line)
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		_, _ = io.WriteString(os.Stderr, line)
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

func showServiceHelp() {
	fmt.Println("Usage: portunix container service <subcommand>")
	fmt.Println()
	fmt.Println("⚙️  CONTAINER LIFECYCLE BACKGROUND SERVICE")
	fmt.Println()
	fmt.Println("Polls all portunix-managed containers and removes those whose TTL has")
	fmt.Println("expired. Daemon state lives under ~/.portunix/container/.")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  start [--interval D] [--detach|--foreground]")
	fmt.Println("                  Start the background lifecycle service (default: detach)")
	fmt.Println("  stop            Stop the background lifecycle service")
	fmt.Println("  status          Show whether the service is running")
	fmt.Println("  run [--interval D]")
	fmt.Println("                  Run the polling loop in the foreground (used by --detach)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container service start --interval 1m")
	fmt.Println("  portunix container service status")
	fmt.Println("  portunix container service stop")
}
