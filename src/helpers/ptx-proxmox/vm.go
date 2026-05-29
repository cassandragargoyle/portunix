/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

// clientForProfile loads the saved config and returns a live Client for the
// selected profile. For password profiles this triggers an interactive
// re-login — tickets are not persisted.
func clientForProfile(profileName string) (*Client, *Profile, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}
	_, profile, err := resolveProfile(cfg, profileName)
	if err != nil {
		return nil, nil, err
	}
	client, err := NewClient(profile)
	if err != nil {
		return nil, nil, err
	}
	if profile.AuthType == AuthTypePassword {
		password, err := promptPassword(fmt.Sprintf("Password for %s: ", profile.User))
		if err != nil {
			return nil, nil, fmt.Errorf("read password: %w", err)
		}
		if err := client.Login(password); err != nil {
			return nil, nil, fmt.Errorf("re-authenticate: %w", err)
		}
	}
	return client, profile, nil
}

// humanBytes formats a byte count as a short string (e.g., "4.0G"). Proxmox
// reports memory/disk in bytes; this keeps list output readable.
func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := uint64(unit), 0
	for n/div >= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(n)/float64(div), "KMGTPE"[exp])
}

func newListCmd() *cobra.Command {
	var (
		profileName string
		typeFilter  string
		nodeFilter  string
		format      string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List VMs and LXC containers",
		Long: `List all VMs and LXC containers visible to the authenticated user.

Examples:
  portunix proxmox list                    # all VMs and CTs on all nodes
  portunix proxmox list --type vm           # QEMU VMs only
  portunix proxmox list --type ct           # LXC containers only
  portunix proxmox list --node pve1         # resources on a specific node
  portunix proxmox list --format json       # machine-readable output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(profileName, typeFilter, nodeFilter, format)
		},
	}
	f := cmd.Flags()
	f.StringVar(&profileName, "profile", "", "Proxmox profile to use (defaults to current)")
	f.StringVar(&typeFilter, "type", "", "Filter by resource type: vm or ct")
	f.StringVar(&nodeFilter, "node", "", "Filter by Proxmox node name")
	f.StringVar(&format, "format", "table", "Output format: table or json")
	return cmd
}

func runList(profileName, typeFilter, nodeFilter, format string) error {
	switch typeFilter {
	case "", "vm", "ct":
	default:
		return fmt.Errorf("--type must be vm, ct, or empty (got %q)", typeFilter)
	}
	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	items, err := client.ListResources(typeFilter, nodeFilter)
	if err != nil {
		return err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Node != items[j].Node {
			return items[i].Node < items[j].Node
		}
		return items[i].VMID < items[j].VMID
	})

	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	}
	if len(items) == 0 {
		fmt.Println("No VMs or CTs found.")
		return nil
	}
	fmt.Printf("%-6s %-4s %-25s %-10s %-10s %-6s %-10s %s\n",
		"VMID", "TYPE", "NAME", "STATUS", "NODE", "CPUS", "MEMORY", "DISK")
	fmt.Printf("%-6s %-4s %-25s %-10s %-10s %-6s %-10s %s\n",
		"----", "----", "----", "------", "----", "----", "------", "----")
	for _, r := range items {
		t := "vm"
		if r.Type == ResourceTypeCT {
			t = "ct"
		}
		fmt.Printf("%-6d %-4s %-25s %-10s %-10s %-6d %-10s %s\n",
			r.VMID, t, r.Name, r.Status, r.Node, r.CPUs, humanBytes(r.MaxMem), humanBytes(r.MaxDisk))
	}
	return nil
}

func newInfoCmd() *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:   "info <name|vmid>",
		Short: "Show VM/CT configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(profileName, args[0])
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	return cmd
}

func runInfo(profileName, ident string) error {
	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	r, err := client.ResolveResource(ident)
	if err != nil {
		return err
	}
	cfg, err := client.Config(r)
	if err != nil {
		return err
	}
	fmt.Printf("Node:      %s\n", r.Node)
	fmt.Printf("VMID:      %d\n", r.VMID)
	fmt.Printf("Type:      %s\n", r.Type)
	fmt.Printf("Name:      %s\n", r.Name)
	fmt.Println("Config:")
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-20s %v\n", k+":", cfg[k])
	}
	return nil
}

func newStatusCmd() *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:   "status <name|vmid>",
		Short: "Show VM/CT runtime status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(profileName, args[0])
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	return cmd
}

func runStatus(profileName, ident string) error {
	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	r, err := client.ResolveResource(ident)
	if err != nil {
		return err
	}
	st, err := client.StatusCurrent(r)
	if err != nil {
		return err
	}
	fmt.Printf("Node:      %s\n", r.Node)
	fmt.Printf("VMID:      %d\n", r.VMID)
	fmt.Printf("Type:      %s\n", r.Type)
	fmt.Printf("Name:      %s\n", r.Name)
	// A handful of common keys first; dump the rest after.
	highlight := []string{"status", "uptime", "cpu", "mem", "maxmem", "netin", "netout", "qmpstatus"}
	for _, k := range highlight {
		if v, ok := st[k]; ok {
			fmt.Printf("  %-12s %v\n", k+":", v)
		}
	}
	return nil
}

// newLifecycleCmd builds start/stop/shutdown/restart/delete. All share the
// same flag set (profile + optional --wait) and only differ in the Client
// method they invoke.
func newLifecycleCmd(use, short string, action func(*Client, *ResourceRef) (string, error)) *cobra.Command {
	var (
		profileName string
		wait        bool
		timeoutSec  int
	)
	cmd := &cobra.Command{
		Use:   use + " <name|vmid>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLifecycle(profileName, args[0], wait, timeoutSec, action)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for Proxmox task to finish")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 300, "Task wait timeout in seconds")
	return cmd
}

func runLifecycle(profileName, ident string, wait bool, timeoutSec int,
	action func(*Client, *ResourceRef) (string, error)) error {

	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	r, err := client.ResolveResource(ident)
	if err != nil {
		return err
	}
	upid, err := action(client, r)
	if err != nil {
		return err
	}
	fmt.Printf("Task submitted: %s\n", upid)
	if !wait {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	if err := client.WaitTask(ctx, r.Node, upid); err != nil {
		return err
	}
	fmt.Println("Task completed successfully.")
	return nil
}

func newDeleteCmd() *cobra.Command {
	var (
		profileName string
		wait        bool
		timeoutSec  int
		purge       bool
		force       bool
	)
	cmd := &cobra.Command{
		Use:   "delete <name|vmid>",
		Short: "Destroy a VM or CT",
		Long: `Destroy a VM or CT. The instance must be stopped first unless --force is
used (which stops the instance before destroying it). --purge removes the
volume from backup jobs and HA config too.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := clientForProfile(profileName)
			if err != nil {
				return err
			}
			r, err := client.ResolveResource(args[0])
			if err != nil {
				return err
			}
			if force && r.Status == "running" {
				fmt.Println("Stopping instance before destroy...")
				upid, err := client.Stop(r)
				if err != nil {
					return fmt.Errorf("force stop: %w", err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				if err := client.WaitTask(ctx, r.Node, upid); err != nil {
					cancel()
					return err
				}
				cancel()
			}
			upid, err := client.Destroy(r, purge)
			if err != nil {
				return err
			}
			fmt.Printf("Task submitted: %s\n", upid)
			if !wait {
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
			defer cancel()
			if err := client.WaitTask(ctx, r.Node, upid); err != nil {
				return err
			}
			fmt.Println("Instance destroyed.")
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&profileName, "profile", "", "Proxmox profile to use")
	f.BoolVar(&wait, "wait", true, "Wait for Proxmox task to finish")
	f.IntVar(&timeoutSec, "timeout", 300, "Task wait timeout in seconds")
	f.BoolVar(&purge, "purge", false, "Also remove from backup jobs and HA config")
	f.BoolVar(&force, "force", false, "Stop the instance if running before destroying")
	return cmd
}
