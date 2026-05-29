/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// NextVMID asks Proxmox for the next free VMID via /cluster/nextid. The cluster
// endpoint returns a JSON string like "100".
func (c *Client) NextVMID() (int, error) {
	var s string
	if err := c.get("/cluster/nextid", &s); err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse nextid %q: %w", s, err)
	}
	return id, nil
}

// CreateVM POSTs to /nodes/{node}/qemu with the supplied form. Returns the
// task UPID for the caller to wait on.
func (c *Client) CreateVM(node string, form url.Values) (string, error) {
	path := fmt.Sprintf("/nodes/%s/qemu", url.PathEscape(node))
	var upid string
	if err := c.postForm(path, form, &upid); err != nil {
		return "", err
	}
	return upid, nil
}

// CreateCT POSTs to /nodes/{node}/lxc. LXC creation is synchronous enough
// that Proxmox still returns a UPID — waiting is required for the container
// to become usable (template extraction takes time).
func (c *Client) CreateCT(node string, form url.Values) (string, error) {
	path := fmt.Sprintf("/nodes/%s/lxc", url.PathEscape(node))
	var upid string
	if err := c.postForm(path, form, &upid); err != nil {
		return "", err
	}
	return upid, nil
}

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VM or CT",
	}
	cmd.AddCommand(newCreateVMCmd())
	cmd.AddCommand(newCreateCTCmd())
	return cmd
}

// commonCreateFlags wraps flags shared between VM and CT creation. Consolidated
// here because Proxmox uses the same parameter names (cores, memory, storage, ...)
// for both endpoints.
type commonCreateFlags struct {
	profileName string
	name        string
	node        string
	vmid        int
	cores       int
	memory      int
	disk        int // GB
	storage     string
	template    string // volid or image name; for CT: /var/lib/vz/template/cache/xxx.tar.zst
	sshKey      string // path to public key for cloud-init (VM) or root SSH (CT)
	password    string // root/default-user password (CT required; VM optional via cloud-init)
	wait        bool
	timeoutSec  int
}

func bindCreateFlags(f *cobra.Command, c *commonCreateFlags) {
	ff := f.Flags()
	ff.StringVar(&c.profileName, "profile", "", "Proxmox profile to use")
	ff.StringVar(&c.name, "name", "", "Name for the new VM/CT (required)")
	ff.StringVar(&c.node, "node", "", "Target node name (required)")
	ff.IntVar(&c.vmid, "vmid", 0, "VMID to assign (0 = auto-allocate via /cluster/nextid)")
	ff.IntVar(&c.cores, "cores", 1, "Number of CPU cores")
	ff.IntVar(&c.memory, "memory", 1024, "Memory in MiB")
	ff.IntVar(&c.disk, "disk", 8, "Disk size in GiB")
	ff.StringVar(&c.storage, "storage", "local-lvm", "Storage pool for disks")
	ff.StringVar(&c.template, "template", "", "Template volid or image name (required)")
	ff.StringVar(&c.sshKey, "ssh-key", "", "Path to SSH public key (optional)")
	ff.StringVar(&c.password, "password", "", "Root/default password (CT required unless --ssh-key)")
	ff.BoolVar(&c.wait, "wait", true, "Wait for Proxmox task to finish")
	ff.IntVar(&c.timeoutSec, "timeout", 600, "Task wait timeout in seconds")
}

func newCreateVMCmd() *cobra.Command {
	cf := &commonCreateFlags{}
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Create a QEMU VM",
		Long: `Create a QEMU VM from a cloud-init-ready disk image. --template should be
a volid pointing at an image on Proxmox storage (e.g., "local:iso/ubuntu-22.04-cloud.img")
or a full path on the node (e.g., "/var/lib/vz/images/ubuntu-22.04-cloud.img").

The VM is created with a cloud-init drive and sane defaults (virtio-scsi,
vmbr0 bridge). Post-creation cloud-init user/ssh config is done via
'portunix proxmox cloud-init'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateVM(cf)
		},
	}
	bindCreateFlags(cmd, cf)
	return cmd
}

func newCreateCTCmd() *cobra.Command {
	cf := &commonCreateFlags{}
	cmd := &cobra.Command{
		Use:   "ct",
		Short: "Create an LXC container",
		Long: `Create an LXC container from a vztmpl on Proxmox storage. --template
typically looks like "local:vztmpl/ubuntu-22.04-standard_22.04-1_amd64.tar.zst".

One of --password or --ssh-key must be provided so root can log in.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateCT(cf)
		},
	}
	bindCreateFlags(cmd, cf)
	return cmd
}

func validateCommonCreate(cf *commonCreateFlags) error {
	if cf.name == "" {
		return fmt.Errorf("--name is required")
	}
	if cf.node == "" {
		return fmt.Errorf("--node is required")
	}
	if cf.template == "" {
		return fmt.Errorf("--template is required")
	}
	return nil
}

func runCreateVM(cf *commonCreateFlags) error {
	if err := validateCommonCreate(cf); err != nil {
		return err
	}
	client, _, err := clientForProfile(cf.profileName)
	if err != nil {
		return err
	}
	vmid := cf.vmid
	if vmid == 0 {
		vmid, err = client.NextVMID()
		if err != nil {
			return fmt.Errorf("allocate next vmid: %w", err)
		}
	}

	form := url.Values{}
	form.Set("vmid", strconv.Itoa(vmid))
	form.Set("name", cf.name)
	form.Set("cores", strconv.Itoa(cf.cores))
	form.Set("memory", strconv.Itoa(cf.memory))
	form.Set("net0", "virtio,bridge=vmbr0")
	form.Set("scsihw", "virtio-scsi-pci")
	// scsi0 declares the main disk imported from --template.
	form.Set("scsi0", fmt.Sprintf("%s:%d,import-from=%s", cf.storage, cf.disk, cf.template))
	// Cloud-init drive on ide2.
	form.Set("ide2", fmt.Sprintf("%s:cloudinit", cf.storage))
	form.Set("boot", "order=scsi0")
	form.Set("serial0", "socket")
	form.Set("vga", "serial0")
	form.Set("agent", "enabled=1")

	if cf.sshKey != "" {
		keyData, err := readFile(cf.sshKey)
		if err != nil {
			return fmt.Errorf("read ssh key: %w", err)
		}
		form.Set("sshkeys", url.QueryEscape(keyData))
	}
	if cf.password != "" {
		form.Set("cipassword", cf.password)
	}

	upid, err := client.CreateVM(cf.node, form)
	if err != nil {
		return err
	}
	fmt.Printf("VM %d creation submitted: %s\n", vmid, upid)
	if !cf.wait {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cf.timeoutSec)*time.Second)
	defer cancel()
	if err := client.WaitTask(ctx, cf.node, upid); err != nil {
		return err
	}
	fmt.Printf("VM %d (%s) created on %s.\n", vmid, cf.name, cf.node)
	return nil
}

func runCreateCT(cf *commonCreateFlags) error {
	if err := validateCommonCreate(cf); err != nil {
		return err
	}
	if cf.password == "" && cf.sshKey == "" {
		return fmt.Errorf("--password or --ssh-key is required for CT creation")
	}
	client, _, err := clientForProfile(cf.profileName)
	if err != nil {
		return err
	}
	vmid := cf.vmid
	if vmid == 0 {
		vmid, err = client.NextVMID()
		if err != nil {
			return fmt.Errorf("allocate next vmid: %w", err)
		}
	}

	form := url.Values{}
	form.Set("vmid", strconv.Itoa(vmid))
	form.Set("hostname", cf.name)
	form.Set("ostemplate", cf.template)
	form.Set("cores", strconv.Itoa(cf.cores))
	form.Set("memory", strconv.Itoa(cf.memory))
	form.Set("rootfs", fmt.Sprintf("%s:%d", cf.storage, cf.disk))
	form.Set("net0", "name=eth0,bridge=vmbr0,ip=dhcp")
	form.Set("unprivileged", "1")

	if cf.password != "" {
		form.Set("password", cf.password)
	}
	if cf.sshKey != "" {
		keyData, err := readFile(cf.sshKey)
		if err != nil {
			return fmt.Errorf("read ssh key: %w", err)
		}
		form.Set("ssh-public-keys", keyData)
	}

	upid, err := client.CreateCT(cf.node, form)
	if err != nil {
		return err
	}
	fmt.Printf("CT %d creation submitted: %s\n", vmid, upid)
	if !cf.wait {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cf.timeoutSec)*time.Second)
	defer cancel()
	if err := client.WaitTask(ctx, cf.node, upid); err != nil {
		return err
	}
	fmt.Printf("CT %d (%s) created on %s.\n", vmid, cf.name, cf.node)
	return nil
}
