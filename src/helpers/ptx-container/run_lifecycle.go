/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/term"
)

// extractLifecycleFlags scans args for the issue-#027 flags and returns the
// remaining args (in their original order, suitable to splice into the
// runtime's `run` line) plus a parsed LifecycleConfig.
//
// Recognised flags:
//
//	--ttl DURATION
//	--auto-cleanup
//	--cleanup-policy on-exit|ttl|manual
//	--health-check CMD
//	--max-memory SIZE   (alias of --memory used by issue text; both pass through)
//	--max-cpu COUNT     (alias of --cpus)
//
// Unrecognised args are preserved verbatim so the existing runtime passthrough
// keeps working for native flags like --name, -p, -v, -d.
func extractLifecycleFlags(args []string) ([]string, LifecycleConfig, error) {
	var lc LifecycleConfig
	var rest []string
	i := 0
	for i < len(args) {
		switch args[i] {
		case "--ttl":
			if i+1 >= len(args) {
				return nil, lc, fmt.Errorf("--ttl requires a duration")
			}
			d, err := ParseDuration(args[i+1])
			if err != nil {
				return nil, lc, err
			}
			lc.TTL = d
			i += 2
			continue
		case "--auto-cleanup":
			lc.AutoCleanup = true
			i++
			continue
		case "--cleanup-policy":
			if i+1 >= len(args) {
				return nil, lc, fmt.Errorf("--cleanup-policy requires a value")
			}
			if err := ValidatePolicy(args[i+1]); err != nil {
				return nil, lc, err
			}
			lc.Policy = args[i+1]
			i += 2
			continue
		case "--health-check":
			if i+1 >= len(args) {
				return nil, lc, fmt.Errorf("--health-check requires a command")
			}
			lc.HealthCheck = args[i+1]
			i += 2
			continue
		case "--max-memory":
			if i+1 >= len(args) {
				return nil, lc, fmt.Errorf("--max-memory requires a value")
			}
			// Translate to runtime-native flag. Issue #027 introduced the alias
			// for readability; runtime knows --memory.
			rest = append(rest, "--memory", args[i+1])
			i += 2
			continue
		case "--max-cpu", "--max-cpus":
			if i+1 >= len(args) {
				return nil, lc, fmt.Errorf("%s requires a value", args[i])
			}
			rest = append(rest, "--cpus", args[i+1])
			i += 2
			continue
		default:
			rest = append(rest, args[i])
			i++
		}
	}
	return rest, lc, nil
}

// runManagedContainer runs `<runtime> run [base flags] <label/health flags>
// <user flags> <image> [cmd...]` so the lifecycle metadata is recorded at
// container creation time. It mirrors runPodmanContainer/runDockerContainer
// but injects the label/health arguments returned by LifecycleConfig.
//
// When AutoCleanup with policy=on-exit is requested for a non-detached run,
// the runtime --rm flag still applies (image is removed when the foreground
// process exits, which is the same effect as our policy). For detached runs
// we drop --rm so the lifecycle service can remove the container later.
func runManagedContainer(runtime string, lc LifecycleConfig, runArgs []string) {
	detached := containsAny(runArgs, "-d", "--detach")

	args := []string{"run"}
	if !detached {
		// Preserve the existing TTY behaviour from runPodman/runDocker: only
		// allocate -t when stdin is a terminal so non-interactive callers
		// (CI, tests) don't get "input device is not a TTY".
		if term.IsTerminal(int(os.Stdin.Fd())) {
			args = append(args, "-it")
		} else {
			args = append(args, "-i")
		}
		// Detached managed runs cannot use --rm — the lifecycle service
		// needs the container to exist past its own exit so it can poll
		// expires_at. For foreground runs with policy=on-exit, --rm matches
		// the policy semantics so we still want it.
		if lc.resolvePolicy() == PolicyOnExit {
			args = append(args, "--rm")
		}
	}

	// Inject lifecycle labels and health command.
	args = append(args, lc.ToLabels(time.Now())...)
	args = append(args, lc.RuntimeHealthArgs()...)

	// Then user flags + image + command. extractLifecycleFlags has stripped
	// the lifecycle flags, so what's left is the runtime's native vocabulary.
	args = append(args, runArgs...)

	if debugMode {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG %s args: %v\n", runtime, args)
	}

	cmd := exec.Command(runtime, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ %s run failed: %v\n", runtime, err)
		os.Exit(1)
	}
}

func containsAny(s []string, needles ...string) bool {
	for _, item := range s {
		for _, n := range needles {
			if item == n {
				return true
			}
		}
	}
	return false
}
