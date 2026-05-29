/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// setupProxmoxEnvironment provisions (or validates) a Proxmox VE VM/CT for
// playbook execution (issue #167, Phase 4). It shells out to the ptx-proxmox
// helper so all Proxmox-specific behaviour lives in one place.
//
// The expected spec.environment shape is:
//
//	environment:
//	  type: proxmox
//	  host: pve.example.com      # (informational; the active proxmox profile is used)
//	  target: synapse-registry    # VM/CT name or vmid
//	  create_if_missing: true
//	  node: pve1
//	  template: local:iso/ubuntu-22.04-cloud.img   # VM image or CT vztmpl
//	  kind: vm                    # "vm" (default) or "ct"
//	  resources:                  # passed through to `proxmox create`
//	    cores: 2
//	    memory: 4096
//	    disk: 32
//	    storage: local-lvm
//	  ssh_user: ubuntu
//	  ssh_key: ~/.ssh/id_rsa.pub
//
// CLI flags override spec values: --target always wins over spec.environment.target.
func setupProxmoxEnvironment(options ExecutionOptions) (*EnvironmentContext, error) {
	env := options.PlaybookEnvironment // populated by executor before calling setupEnvironment
	if env == nil {
		env = map[string]interface{}{}
	}

	target := options.Target
	if target == "" {
		if t, ok := env["target"].(string); ok {
			target = t
		}
	}
	if target == "" {
		return nil, fmt.Errorf("--target (or spec.environment.target) is required for proxmox environment")
	}

	portunixPath, err := getPortunixBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("find portunix binary: %w", err)
	}

	// Auto-create if requested and the target does not already exist.
	if b, _ := env["create_if_missing"].(bool); b {
		if exists, err := proxmoxResourceExists(portunixPath, target); err != nil {
			return nil, err
		} else if !exists {
			if err := proxmoxCreateFromSpec(portunixPath, target, env, options.Verbose); err != nil {
				return nil, fmt.Errorf("create_if_missing: %w", err)
			}
		}
	}

	// Ensure the target is running before attempting SSH.
	if err := proxmoxEnsureRunning(portunixPath, target, options.Verbose); err != nil {
		return nil, err
	}

	// Resolve IP through ptx-proxmox (uses QEMU guest agent / LXC interfaces).
	ip, err := proxmoxResolveIP(portunixPath, target)
	if err != nil {
		return nil, err
	}

	sshUser := "root"
	if u, ok := env["ssh_user"].(string); ok && u != "" {
		sshUser = u
	}
	sshKey := ""
	if k, ok := env["ssh_key"].(string); ok {
		sshKey = k
	}

	envCtx := &EnvironmentContext{
		Type:       "proxmox",
		Target:     target,
		SSHHost:    ip,
		SSHPort:    "22",
		SSHUser:    sshUser,
		SSHKeyPath: sshKey,
		// Inventory matches the format used by virt environment so downstream
		// ansible-playbook invocations work without changes.
		Inventory: generateVMInventory(target, ip, "22", sshKey),
	}
	if options.Verbose {
		fmt.Printf("✅ Proxmox environment ready: %s (%s@%s)\n", target, sshUser, ip)
	}
	return envCtx, nil
}

// proxmoxResourceExists returns true if the target VM/CT is visible in the
// cluster. Uses `proxmox list --format json` and greps the result — cheap and
// avoids adding a new dedicated subcommand.
func proxmoxResourceExists(portunixPath, target string) (bool, error) {
	out, err := runProxmox(portunixPath, "list", "--format", "json")
	if err != nil {
		return false, err
	}
	var items []struct {
		VMID int    `json:"VMID"`
		Name string `json:"Name"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return false, fmt.Errorf("parse proxmox list output: %w", err)
	}
	id, isNumeric := atoiSafe(target)
	for _, it := range items {
		if (isNumeric && it.VMID == id) || it.Name == target {
			return true, nil
		}
	}
	return false, nil
}

// proxmoxCreateFromSpec assembles `proxmox create vm|ct` flags from the
// environment map and invokes the helper. Errors surface the helper's stderr
// verbatim to give users actionable diagnostics.
func proxmoxCreateFromSpec(portunixPath, target string, env map[string]interface{}, verbose bool) error {
	kind := "vm"
	if k, ok := env["kind"].(string); ok && k != "" {
		kind = k
	}
	args := []string{"create", kind, "--name", target}
	if n, ok := env["node"].(string); ok && n != "" {
		args = append(args, "--node", n)
	} else {
		return fmt.Errorf("spec.environment.node is required when create_if_missing is true")
	}
	if t, ok := env["template"].(string); ok && t != "" {
		args = append(args, "--template", t)
	} else {
		return fmt.Errorf("spec.environment.template is required when create_if_missing is true")
	}
	if res, ok := env["resources"].(map[string]interface{}); ok {
		if v := intOrString(res["cores"]); v != "" {
			args = append(args, "--cores", v)
		}
		if v := intOrString(res["memory"]); v != "" {
			args = append(args, "--memory", v)
		}
		if v := intOrString(res["disk"]); v != "" {
			args = append(args, "--disk", v)
		}
		if v, ok := res["storage"].(string); ok && v != "" {
			args = append(args, "--storage", v)
		}
	}
	if k, ok := env["ssh_key"].(string); ok && k != "" {
		args = append(args, "--ssh-key", k)
	}
	if p, ok := env["password"].(string); ok && p != "" {
		args = append(args, "--password", p)
	}
	if verbose {
		fmt.Printf("🛠️  Creating proxmox %s %q ...\n", kind, target)
	}
	if _, err := runProxmox(portunixPath, args...); err != nil {
		return err
	}
	return nil
}

// proxmoxEnsureRunning starts the VM/CT if status shows it's stopped. Running
// guests are left alone. Paused/suspended states trigger a start attempt so
// playbooks don't silently fail on SSH.
func proxmoxEnsureRunning(portunixPath, target string, verbose bool) error {
	out, err := runProxmox(portunixPath, "list", "--format", "json")
	if err != nil {
		return err
	}
	var items []struct {
		VMID   int    `json:"VMID"`
		Name   string `json:"Name"`
		Status string `json:"Status"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return err
	}
	id, isNumeric := atoiSafe(target)
	for _, it := range items {
		if (isNumeric && it.VMID == id) || it.Name == target {
			if it.Status == "running" {
				return nil
			}
			if verbose {
				fmt.Printf("▶️  Starting proxmox target %q (was %s)\n", target, it.Status)
			}
			_, err := runProxmox(portunixPath, "start", target)
			return err
		}
	}
	return fmt.Errorf("proxmox target %q not found after creation", target)
}

// proxmoxResolveIP shells out to `proxmox info` just to confirm reachability,
// then uses the agent endpoint via `proxmox status` is not sufficient —
// instead we use ssh resolution which the helper does internally during
// `proxmox exec/ssh`. Here we duplicate the logic by calling a dry-run ssh
// with a no-op command — cheap and reuses the helper's IP resolution code.
// If a future `proxmox resolve-ip` subcommand lands, this can be simplified.
func proxmoxResolveIP(portunixPath, target string) (string, error) {
	// Use `proxmox status <target>` to confirm the VM exists; the actual IP
	// is emitted via ssh later. For the inventory we need an IP string, so
	// invoke a lightweight helper subcommand: `proxmox resolve-ip` (added
	// below) which returns a single-line IP. If that subcommand is not yet
	// available (older helper), we fall back to scanning `proxmox info`
	// output for known IP-carrying config lines.
	out, err := runProxmox(portunixPath, "resolve-ip", target)
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	// Fallback: parse `proxmox info` output for `net0: ...,ip=X.X.X.X` etc.
	// (Best-effort — returns the first IPv4 we can find.)
	info, err2 := runProxmox(portunixPath, "info", target)
	if err2 != nil {
		return "", fmt.Errorf("resolve IP: %w (fallback: %v)", err, err2)
	}
	if ip := findIPv4(string(info)); ip != "" {
		return ip, nil
	}
	return "", fmt.Errorf("could not resolve IP for %q — ensure QEMU guest agent is running or add a `proxmox resolve-ip` subcommand", target)
}

// runProxmox runs `portunix proxmox <args...>` and returns stdout. Errors
// include the helper's stderr so users see the real problem.
func runProxmox(portunixPath string, args ...string) ([]byte, error) {
	cmd := exec.Command(portunixPath, append([]string{"proxmox"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("proxmox %s: %v\n%s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// atoiSafe is strconv.Atoi that returns (0, false) on error — useful in
// switches that accept either numeric IDs or names.
func atoiSafe(s string) (int, bool) {
	id, err := strconv.Atoi(s)
	return id, err == nil
}

// intOrString normalises a resource value coming from YAML (int, float64, or
// string) to a decimal string suitable for CLI flags. Empty string means the
// field was absent or of an unexpected type; the caller should skip the flag.
func intOrString(v interface{}) string {
	switch vv := v.(type) {
	case int:
		return strconv.Itoa(vv)
	case int64:
		return strconv.FormatInt(vv, 10)
	case float64:
		return strconv.FormatInt(int64(vv), 10)
	case string:
		return vv
	}
	return ""
}

// findIPv4 scans text for the first plausible dotted-quad and returns it.
// This is a fallback for older ptx-proxmox builds that lack resolve-ip.
func findIPv4(text string) string {
	const digits = "0123456789"
	isDigit := func(c byte) bool {
		for i := 0; i < len(digits); i++ {
			if c == digits[i] {
				return true
			}
		}
		return false
	}
	for i := 0; i < len(text); i++ {
		if !isDigit(text[i]) {
			continue
		}
		start := i
		dots := 0
		for i < len(text) && (isDigit(text[i]) || text[i] == '.') {
			if text[i] == '.' {
				dots++
			}
			i++
		}
		if dots == 3 {
			candidate := text[start:i]
			parts := strings.Split(candidate, ".")
			if len(parts) == 4 {
				ok := true
				for _, p := range parts {
					n, err := strconv.Atoi(p)
					if err != nil || n < 0 || n > 255 {
						ok = false
						break
					}
				}
				if ok && candidate != "0.0.0.0" && candidate != "127.0.0.1" {
					return candidate
				}
			}
		}
	}
	return ""
}
