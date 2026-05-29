/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// ListStorageContent returns entries from /nodes/{node}/storage/{storage}/content.
// contentFilter is the Proxmox content type ("iso", "vztmpl", "images", …).
func (c *Client) ListStorageContent(node, storage, contentFilter string) ([]StorageContent, error) {
	path := fmt.Sprintf("/nodes/%s/storage/%s/content",
		url.PathEscape(node), url.PathEscape(storage))
	if contentFilter != "" {
		path += "?content=" + url.QueryEscape(contentFilter)
	}
	var items []StorageContent
	if err := c.get(path, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ListNodes returns the cluster's node names (via /cluster/resources?type=node).
func (c *Client) ListNodes() ([]string, error) {
	var raw []clusterResource
	if err := c.get("/cluster/resources?type=node", &raw); err != nil {
		return nil, err
	}
	nodes := make([]string, 0, len(raw))
	for _, r := range raw {
		if r.Type == "node" && r.Node != "" {
			nodes = append(nodes, r.Node)
		}
	}
	sort.Strings(nodes)
	return nodes, nil
}

// ListStorages returns storage pools visible on the given node. An empty node
// falls back to all storages reported by /cluster/resources?type=storage.
func (c *Client) ListStorages(node string) ([]string, error) {
	if node == "" {
		var raw []clusterResource
		if err := c.get("/cluster/resources?type=storage", &raw); err != nil {
			return nil, err
		}
		seen := map[string]struct{}{}
		out := []string{}
		for _, r := range raw {
			if r.Type != "storage" {
				continue
			}
			// resource ID looks like "storage/pve1/local" — last segment is
			// the storage name.
			parts := strings.Split(r.ID, "/")
			name := parts[len(parts)-1]
			if _, dup := seen[name]; !dup {
				seen[name] = struct{}{}
				out = append(out, name)
			}
		}
		sort.Strings(out)
		return out, nil
	}
	var stors []struct {
		Storage string `json:"storage"`
		Content string `json:"content"`
	}
	path := fmt.Sprintf("/nodes/%s/storage", url.PathEscape(node))
	if err := c.get(path, &stors); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(stors))
	for _, s := range stors {
		out = append(out, s.Storage)
	}
	sort.Strings(out)
	return out, nil
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Browse available templates and ISOs",
	}
	cmd.AddCommand(newTemplateListCmd())
	return cmd
}

func newTemplateListCmd() *cobra.Command {
	var (
		profileName string
		node        string
		storage     string
		contentType string
		format      string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List templates and ISOs on Proxmox storage",
		Long: `List storage content across the cluster. Defaults to content type
"vztmpl,iso" which covers LXC templates and QEMU installer images. Use
--content to restrict further.

Examples:
  portunix proxmox template list
  portunix proxmox template list --node pve1 --storage local
  portunix proxmox template list --content vztmpl --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateList(profileName, node, storage, contentType, format)
		},
	}
	f := cmd.Flags()
	f.StringVar(&profileName, "profile", "", "Proxmox profile to use")
	f.StringVar(&node, "node", "", "Node to query (default: iterate all nodes)")
	f.StringVar(&storage, "storage", "", "Restrict to a single storage pool")
	f.StringVar(&contentType, "content", "vztmpl,iso", "Proxmox content type filter (vztmpl, iso, images, backup, ...)")
	f.StringVar(&format, "format", "table", "Output format: table or json")
	return cmd
}

type templateRow struct {
	Node    string `json:"node"`
	Storage string `json:"storage"`
	Content string `json:"content"`
	VolID   string `json:"volid"`
	Size    uint64 `json:"size"`
}

func runTemplateList(profileName, node, storage, contentType, format string) error {
	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	var nodes []string
	if node != "" {
		nodes = []string{node}
	} else {
		nodes, err = client.ListNodes()
		if err != nil {
			return err
		}
	}

	var rows []templateRow
	// contentType is the user-facing list; Proxmox accepts a single value per
	// call, so iterate.
	filters := strings.Split(contentType, ",")
	for _, n := range nodes {
		storages := []string{storage}
		if storage == "" {
			s, err := client.ListStorages(n)
			if err != nil {
				return err
			}
			storages = s
		}
		for _, s := range storages {
			for _, f := range filters {
				f = strings.TrimSpace(f)
				items, err := client.ListStorageContent(n, s, f)
				if err != nil {
					// A storage may not support a given content type
					// (returns 400). Ignore non-fatal errors so the
					// iteration keeps going.
					continue
				}
				for _, it := range items {
					rows = append(rows, templateRow{
						Node: n, Storage: s, Content: it.Content,
						VolID: it.VolID, Size: it.Size,
					})
				}
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Node != rows[j].Node {
			return rows[i].Node < rows[j].Node
		}
		return rows[i].VolID < rows[j].VolID
	})

	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}
	if len(rows) == 0 {
		fmt.Println("No matching content found.")
		return nil
	}
	fmt.Printf("%-10s %-15s %-8s %-10s %s\n", "NODE", "STORAGE", "TYPE", "SIZE", "VOLID")
	fmt.Printf("%-10s %-15s %-8s %-10s %s\n", "----", "-------", "----", "----", "-----")
	for _, r := range rows {
		fmt.Printf("%-10s %-15s %-8s %-10s %s\n",
			r.Node, r.Storage, r.Content, humanBytes(r.Size), r.VolID)
	}
	return nil
}
