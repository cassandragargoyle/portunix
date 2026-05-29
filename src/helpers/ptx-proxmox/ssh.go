/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// ResolveIP returns the best-effort IPv4/IPv6 for the given VM or CT.
// For QEMU VMs, the guest agent endpoint is consulted first; LXC containers
// use the status interfaces endpoint.
func (c *Client) ResolveIP(r *ResourceRef, preferIPv4 bool) (string, error) {
	ips, err := c.GuestIPs(r)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for %s %d on %s "+
			"(ensure the guest is running and qemu-guest-agent is installed for VMs)",
			r.Type, r.VMID, r.Node)
	}
	// Prefer a non-loopback, non-link-local IPv4 unless the caller wants IPv6.
	pick := func(family string) string {
		for _, ip := range ips {
			if ip.Family != family {
				continue
			}
			parsed := net.ParseIP(ip.Address)
			if parsed == nil || parsed.IsLoopback() || parsed.IsLinkLocalUnicast() {
				continue
			}
			return ip.Address
		}
		return ""
	}
	if preferIPv4 {
		if v := pick("ipv4"); v != "" {
			return v, nil
		}
		if v := pick("ipv6"); v != "" {
			return v, nil
		}
	} else {
		if v := pick("ipv6"); v != "" {
			return v, nil
		}
		if v := pick("ipv4"); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("only loopback/link-local addresses found")
}

// GuestIPs returns all IPs reported by the agent/interfaces endpoint. Useful
// for diagnostics and exposed to callers who need something other than the
// single "best" IP.
func (c *Client) GuestIPs(r *ResourceRef) ([]GuestIP, error) {
	switch r.Type {
	case ResourceTypeVM:
		return c.vmGuestIPs(r)
	case ResourceTypeCT:
		return c.ctGuestIPs(r)
	default:
		return nil, fmt.Errorf("unknown resource type %q", r.Type)
	}
}

// vmGuestIPs calls the QEMU guest agent endpoint. Requires qemu-guest-agent
// running inside the VM.
func (c *Client) vmGuestIPs(r *ResourceRef) ([]GuestIP, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces",
		url.PathEscape(r.Node), r.VMID)
	var resp struct {
		Result []struct {
			Name        string `json:"name"`
			IPAddresses []struct {
				IPAddress     string `json:"ip-address"`
				IPAddressType string `json:"ip-address-type"`
			} `json:"ip-addresses"`
		} `json:"result"`
	}
	if err := c.get(path, &resp); err != nil {
		// Agent not running or endpoint not available — translate to a
		// clearer error so users know to install qemu-guest-agent.
		if strings.Contains(err.Error(), "HTTP 500") || strings.Contains(err.Error(), "HTTP 404") {
			return nil, fmt.Errorf("QEMU guest agent not responding on VM %d (install qemu-guest-agent inside the VM and enable it on the VM config)", r.VMID)
		}
		return nil, err
	}
	var out []GuestIP
	for _, iface := range resp.Result {
		for _, ip := range iface.IPAddresses {
			fam := "ipv4"
			if ip.IPAddressType == "ipv6" {
				fam = "ipv6"
			}
			out = append(out, GuestIP{Interface: iface.Name, Address: ip.IPAddress, Family: fam})
		}
	}
	return out, nil
}

// ctGuestIPs uses the LXC interfaces endpoint (no agent needed).
func (c *Client) ctGuestIPs(r *ResourceRef) ([]GuestIP, error) {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/interfaces",
		url.PathEscape(r.Node), r.VMID)
	var raw []struct {
		Name  string `json:"name"`
		Inet  string `json:"inet,omitempty"` // "10.0.0.5/24"
		Inet6 string `json:"inet6,omitempty"`
	}
	if err := c.get(path, &raw); err != nil {
		return nil, err
	}
	var out []GuestIP
	for _, iface := range raw {
		if iface.Inet != "" {
			addr := iface.Inet
			if idx := strings.Index(addr, "/"); idx > 0 {
				addr = addr[:idx]
			}
			out = append(out, GuestIP{Interface: iface.Name, Address: addr, Family: "ipv4"})
		}
		if iface.Inet6 != "" {
			addr := iface.Inet6
			if idx := strings.Index(addr, "/"); idx > 0 {
				addr = addr[:idx]
			}
			out = append(out, GuestIP{Interface: iface.Name, Address: addr, Family: "ipv6"})
		}
	}
	return out, nil
}

// sshDefaults holds per-invocation SSH settings. Constructed from CLI flags.
type sshDefaults struct {
	profileName  string
	user         string
	identityFile string
	port         int
	preferIPv6   bool
	extraOptions []string // passed as "-o key=value"
}

func bindSSHFlags(cmd *cobra.Command, d *sshDefaults) {
	f := cmd.Flags()
	f.StringVar(&d.profileName, "profile", "", "Proxmox profile to use")
	f.StringVarP(&d.user, "user", "u", "root", "SSH user on the VM/CT")
	f.StringVarP(&d.identityFile, "identity", "i", "", "SSH identity file (passed as -i)")
	f.IntVarP(&d.port, "port", "p", 22, "SSH port")
	f.BoolVar(&d.preferIPv6, "ipv6", false, "Prefer an IPv6 address when resolving the guest")
	f.StringArrayVarP(&d.extraOptions, "option", "o", nil, "Extra ssh -o KEY=VALUE options (repeatable)")
}

// buildSSHArgs constructs the argument list shared by ssh/scp invocations.
// target is the user@host portion for ssh, or host for scp (caller formats
// the source/destination paths).
func buildSSHArgs(d *sshDefaults, extra []string) []string {
	args := []string{}
	if d.port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", d.port))
	}
	if d.identityFile != "" {
		args = append(args, "-i", d.identityFile)
	}
	for _, o := range d.extraOptions {
		args = append(args, "-o", o)
	}
	return append(args, extra...)
}

// scpArgs mirrors buildSSHArgs but substitutes -p for -P (scp's port flag).
func buildSCPArgs(d *sshDefaults, extra []string) []string {
	args := []string{}
	if d.port != 22 {
		args = append(args, "-P", fmt.Sprintf("%d", d.port))
	}
	if d.identityFile != "" {
		args = append(args, "-i", d.identityFile)
	}
	for _, o := range d.extraOptions {
		args = append(args, "-o", o)
	}
	return append(args, extra...)
}

// resolveSSHTarget loads the profile, resolves the VM/CT, and returns its IP.
// Returned *Client is handy for callers who want to print extra diagnostics.
func resolveSSHTarget(d *sshDefaults, ident string) (*Client, *ResourceRef, string, error) {
	client, _, err := clientForProfile(d.profileName)
	if err != nil {
		return nil, nil, "", err
	}
	r, err := client.ResolveResource(ident)
	if err != nil {
		return nil, nil, "", err
	}
	ip, err := client.ResolveIP(r, !d.preferIPv6)
	if err != nil {
		return nil, nil, "", err
	}
	return client, r, ip, nil
}

// newResolveIPCmd exposes IP resolution as a stand-alone subcommand so the
// ptx-ansible Phase 4 integration (and scripts) can obtain a clean single-
// line IP without parsing `status` output.
func newResolveIPCmd() *cobra.Command {
	var (
		profileName string
		preferIPv6  bool
	)
	cmd := &cobra.Command{
		Use:   "resolve-ip <name|vmid>",
		Short: "Print the best-effort IP address of a VM or CT",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := clientForProfile(profileName)
			if err != nil {
				return err
			}
			r, err := client.ResolveResource(args[0])
			if err != nil {
				return err
			}
			ip, err := client.ResolveIP(r, !preferIPv6)
			if err != nil {
				return err
			}
			fmt.Println(ip)
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	cmd.Flags().BoolVar(&preferIPv6, "ipv6", false, "Prefer an IPv6 address")
	return cmd
}

func newSSHCmd() *cobra.Command {
	d := &sshDefaults{}
	cmd := &cobra.Command{
		Use:   "ssh <name|vmid> [-- ssh-args...]",
		Short: "Open an SSH session to a VM or CT",
		Long: `Open an interactive SSH session. The guest IP is resolved via the
QEMU guest agent (VMs) or the LXC interfaces endpoint (CTs), so the guest
must be running and network-configured.

Additional SSH arguments can be passed after --, e.g.:
  portunix proxmox ssh my-vm -- -L 8080:localhost:80`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, ip, err := resolveSSHTarget(d, args[0])
			if err != nil {
				return err
			}
			target := d.user + "@" + ip
			sshArgs := buildSSHArgs(d, append([]string{target}, args[1:]...))
			return runInteractive("ssh", sshArgs)
		},
	}
	bindSSHFlags(cmd, d)
	return cmd
}

func newExecCmd() *cobra.Command {
	d := &sshDefaults{}
	cmd := &cobra.Command{
		Use:   "exec <name|vmid> -- <command> [args...]",
		Short: "Run a command on a VM or CT over SSH",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, ip, err := resolveSSHTarget(d, args[0])
			if err != nil {
				return err
			}
			target := d.user + "@" + ip
			sshArgs := buildSSHArgs(d, append([]string{target}, args[1:]...))
			return runInteractive("ssh", sshArgs)
		},
	}
	bindSSHFlags(cmd, d)
	return cmd
}

func newCopyCmd() *cobra.Command {
	d := &sshDefaults{}
	var recursive bool
	cmd := &cobra.Command{
		Use:   "copy <src> <dst>",
		Short: "Copy a file to or from a VM/CT via SCP",
		Long: `Source or destination uses the form '<name|vmid>:/path'. The guest IP
is resolved via agent/interfaces; SCP is executed in the foreground.

Examples:
  portunix proxmox copy ./local.txt my-vm:/tmp/
  portunix proxmox copy my-vm:/var/log/app.log ./
  portunix proxmox copy -r ./dist my-vm:/opt/app/`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]
			srcIdent, srcPath, srcRemote := splitSCPTarget(src)
			dstIdent, dstPath, dstRemote := splitSCPTarget(dst)
			if srcRemote == dstRemote {
				return errors.New("exactly one of <src> or <dst> must be a remote target '<name|vmid>:/path'")
			}
			var ident string
			if srcRemote {
				ident = srcIdent
			} else {
				ident = dstIdent
			}
			_, _, ip, err := resolveSSHTarget(d, ident)
			if err != nil {
				return err
			}
			var srcArg, dstArg string
			if srcRemote {
				srcArg = d.user + "@" + ip + ":" + srcPath
				dstArg = dstPath
			} else {
				srcArg = srcPath
				dstArg = d.user + "@" + ip + ":" + dstPath
			}
			extra := []string{}
			if recursive {
				extra = append(extra, "-r")
			}
			extra = append(extra, srcArg, dstArg)
			scpArgs := buildSCPArgs(d, extra)
			return runInteractive("scp", scpArgs)
		},
	}
	bindSSHFlags(cmd, d)
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively copy directories")
	return cmd
}

// splitSCPTarget detects whether an argument is a remote reference
// (`<ident>:/path`). Returns (ident, path, isRemote). Only the first colon is
// treated as the separator so filenames containing colons in the local path
// are preserved.
func splitSCPTarget(s string) (string, string, bool) {
	// Windows-style absolute paths start with "C:" — guard against treating
	// them as remote targets.
	if len(s) >= 2 && s[1] == ':' {
		// drive letter followed by a path separator is definitely local
		return "", s, false
	}
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", s, false
	}
	return s[:idx], s[idx+1:], true
}

// runInteractive runs cmdName with args, wiring stdio so the user can interact
// with ssh / scp as they would on a terminal. Returns the underlying command
// error verbatim; exit codes propagate via cobra -> main.
func runInteractive(cmdName string, args []string) error {
	bin, err := exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("%s not found on PATH: %w", cmdName, err)
	}
	c := exec.Command(bin, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
