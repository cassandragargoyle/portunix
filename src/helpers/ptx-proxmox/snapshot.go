/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

// Snapshots lists snapshots for a VM or CT.
func (c *Client) Snapshots(r *ResourceRef) ([]Snapshot, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/snapshot",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID)
	var snaps []Snapshot
	if err := c.get(path, &snaps); err != nil {
		return nil, err
	}
	return snaps, nil
}

// CreateSnapshot POSTs a new snapshot and returns the task UPID.
func (c *Client) CreateSnapshot(r *ResourceRef, name, description string, withRAM bool) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/snapshot",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID)
	form := url.Values{}
	form.Set("snapname", name)
	if description != "" {
		form.Set("description", description)
	}
	if withRAM {
		form.Set("vmstate", "1")
	}
	var upid string
	if err := c.postForm(path, form, &upid); err != nil {
		return "", err
	}
	return upid, nil
}

// RevertSnapshot issues a rollback. The VM/CT is reverted to the snapshot
// state; any changes since the snapshot are lost.
func (c *Client) RevertSnapshot(r *ResourceRef, name string) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/%s/rollback",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID, url.PathEscape(name))
	var upid string
	if err := c.postForm(path, nil, &upid); err != nil {
		return "", err
	}
	return upid, nil
}

// DeleteSnapshot removes a snapshot. force=1 deletes snapshot metadata even
// if the underlying disk state is missing (recovery scenario).
func (c *Client) DeleteSnapshot(r *ResourceRef, name string, force bool) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/%s",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID, url.PathEscape(name))
	if force {
		path += "?force=1"
	}
	var upid string
	if err := c.delete(path, &upid); err != nil {
		return "", err
	}
	return upid, nil
}

func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage VM/CT snapshots",
	}
	cmd.AddCommand(newSnapshotListCmd())
	cmd.AddCommand(newSnapshotCreateCmd())
	cmd.AddCommand(newSnapshotRevertCmd())
	cmd.AddCommand(newSnapshotDeleteCmd())
	return cmd
}

func newSnapshotListCmd() *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:   "list <name|vmid>",
		Short: "List snapshots of a VM or CT",
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
			snaps, err := client.Snapshots(r)
			if err != nil {
				return err
			}
			if len(snaps) == 0 {
				fmt.Println("No snapshots.")
				return nil
			}
			sort.Slice(snaps, func(i, j int) bool { return snaps[i].SnapTime < snaps[j].SnapTime })
			fmt.Printf("%-25s %-15s %-22s %s\n", "NAME", "PARENT", "CREATED", "DESCRIPTION")
			fmt.Printf("%-25s %-15s %-22s %s\n", "----", "------", "-------", "-----------")
			for _, s := range snaps {
				ts := "-"
				if s.SnapTime > 0 {
					ts = time.Unix(s.SnapTime, 0).Format("2006-01-02 15:04:05")
				}
				fmt.Printf("%-25s %-15s %-22s %s\n", s.Name, s.Parent, ts, s.Description)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	return cmd
}

func newSnapshotCreateCmd() *cobra.Command {
	var (
		profileName string
		desc        string
		includeRAM  bool
		wait        bool
		timeoutSec  int
	)
	cmd := &cobra.Command{
		Use:   "create <name|vmid> <snapshot-name>",
		Short: "Create a snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshotOp(profileName, args[0], wait, timeoutSec,
				func(c *Client, r *ResourceRef) (string, error) {
					return c.CreateSnapshot(r, args[1], desc, includeRAM)
				}, fmt.Sprintf("Snapshot %q created.", args[1]))
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	cmd.Flags().StringVar(&desc, "description", "", "Snapshot description")
	cmd.Flags().BoolVar(&includeRAM, "include-ram", false, "Include RAM state (QEMU only)")
	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for task completion")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 600, "Task wait timeout in seconds")
	return cmd
}

func newSnapshotRevertCmd() *cobra.Command {
	var (
		profileName string
		wait        bool
		timeoutSec  int
	)
	cmd := &cobra.Command{
		Use:   "revert <name|vmid> <snapshot-name>",
		Short: "Revert a VM/CT to a snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshotOp(profileName, args[0], wait, timeoutSec,
				func(c *Client, r *ResourceRef) (string, error) {
					return c.RevertSnapshot(r, args[1])
				}, fmt.Sprintf("Reverted to snapshot %q.", args[1]))
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for task completion")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 600, "Task wait timeout in seconds")
	return cmd
}

func newSnapshotDeleteCmd() *cobra.Command {
	var (
		profileName string
		force       bool
		wait        bool
		timeoutSec  int
	)
	cmd := &cobra.Command{
		Use:   "delete <name|vmid> <snapshot-name>",
		Short: "Delete a snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshotOp(profileName, args[0], wait, timeoutSec,
				func(c *Client, r *ResourceRef) (string, error) {
					return c.DeleteSnapshot(r, args[1], force)
				}, fmt.Sprintf("Snapshot %q deleted.", args[1]))
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Proxmox profile to use")
	cmd.Flags().BoolVar(&force, "force", false, "Force delete even if disk state is missing")
	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for task completion")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 600, "Task wait timeout in seconds")
	return cmd
}

// runSnapshotOp is a small adapter that runs any per-snapshot action and
// optionally waits for the Proxmox task to finish. Centralised here so the
// four snapshot subcommands share consistent progress output.
func runSnapshotOp(profileName, ident string, wait bool, timeoutSec int,
	op func(*Client, *ResourceRef) (string, error), successMsg string) error {

	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	r, err := client.ResolveResource(ident)
	if err != nil {
		return err
	}
	upid, err := op(client, r)
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
	fmt.Println(successMsg)
	return nil
}
