/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// readFile is a small helper used for --ssh-key loading across multiple
// commands. Returns the file content as a string or an error.
func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// UpdateVMConfig pushes a set of config keys to /nodes/{node}/qemu/{vmid}/config
// (PUT). Proxmox treats each form field as a single config directive.
func (c *Client) UpdateVMConfig(r *ResourceRef, form url.Values) error {
	if r.Type != ResourceTypeVM {
		return fmt.Errorf("cloud-init configuration only applies to QEMU VMs")
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(r.Node), r.VMID)
	return c.putForm(path, form, nil)
}

func newCloudInitCmd() *cobra.Command {
	var (
		profileName string
		user        string
		sshKeyPath  string
		password    string
		ip          string
		gateway     string
		nameserver  string
	)
	cmd := &cobra.Command{
		Use:   "cloud-init <name|vmid>",
		Short: "Configure cloud-init on an existing VM",
		Long: `Update the cloud-init parameters of a QEMU VM. The VM must have been
created with a cloud-init drive (use 'portunix proxmox create vm'); this
command does not add one.

Examples:
  portunix proxmox cloud-init my-vm --user deploy --ssh-key ~/.ssh/id_rsa.pub
  portunix proxmox cloud-init 101 --ip dhcp
  portunix proxmox cloud-init 101 --ip 10.0.0.5/24 --gateway 10.0.0.1 --nameserver 1.1.1.1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCloudInit(profileName, args[0], user, sshKeyPath, password, ip, gateway, nameserver)
		},
	}
	f := cmd.Flags()
	f.StringVar(&profileName, "profile", "", "Proxmox profile to use")
	f.StringVar(&user, "user", "", "cloud-init default user (ciuser)")
	f.StringVar(&sshKeyPath, "ssh-key", "", "Path to SSH public key to install")
	f.StringVar(&password, "password", "", "Default user password (cipassword)")
	f.StringVar(&ip, "ip", "", "IPv4 config: 'dhcp' or CIDR (e.g., 10.0.0.5/24)")
	f.StringVar(&gateway, "gateway", "", "IPv4 gateway (required with static IP)")
	f.StringVar(&nameserver, "nameserver", "", "DNS nameserver IP")
	return cmd
}

func runCloudInit(profileName, ident, user, sshKeyPath, password, ip, gateway, nameserver string) error {
	client, _, err := clientForProfile(profileName)
	if err != nil {
		return err
	}
	r, err := client.ResolveResource(ident)
	if err != nil {
		return err
	}
	if r.Type != ResourceTypeVM {
		return fmt.Errorf("cloud-init applies to QEMU VMs only (%s is %s)", ident, r.Type)
	}

	form := url.Values{}
	if user != "" {
		form.Set("ciuser", user)
	}
	if password != "" {
		form.Set("cipassword", password)
	}
	if sshKeyPath != "" {
		keyData, err := readFile(sshKeyPath)
		if err != nil {
			return fmt.Errorf("read ssh key: %w", err)
		}
		form.Set("sshkeys", url.QueryEscape(keyData))
	}
	if ip != "" {
		spec := "ip=" + ip
		if gateway != "" {
			spec += ",gw=" + gateway
		}
		form.Set("ipconfig0", spec)
	}
	if nameserver != "" {
		form.Set("nameserver", nameserver)
	}
	if len(form) == 0 {
		return fmt.Errorf("no cloud-init flags specified — pass at least one of --user/--ssh-key/--password/--ip/--nameserver")
	}

	if err := client.UpdateVMConfig(r, form); err != nil {
		return err
	}
	fmt.Printf("cloud-init configured on VM %d (%s).\n", r.VMID, r.Name)
	fmt.Println("Note: reboot the VM for cloud-init changes to take effect on next boot.")
	return nil
}
