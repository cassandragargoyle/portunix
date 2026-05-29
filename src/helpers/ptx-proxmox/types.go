/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

// AuthType identifies how a profile authenticates against the Proxmox VE API.
type AuthType string

const (
	AuthTypeToken    AuthType = "token"    // PVE API Token (recommended for automation)
	AuthTypePassword AuthType = "password" // Username + password (ticket acquired on demand)
)

// Profile holds connection settings for a single Proxmox VE endpoint.
// One config file can hold multiple profiles; the active one is selected by
// Config.CurrentProfile.
type Profile struct {
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	User        string   `json:"user,omitempty"`
	AuthType    AuthType `json:"auth_type"`
	TokenID     string   `json:"token_id,omitempty"`
	TokenSecret string   `json:"token_secret,omitempty"`
	VerifyTLS   bool     `json:"verify_tls"`
}

// Config is the on-disk structure stored in the user config directory.
type Config struct {
	Profiles       map[string]*Profile `json:"profiles"`
	CurrentProfile string              `json:"current_profile"`
}

// VersionInfo is the subset of /version response we rely on to validate
// credentials. The full response carries additional fields we ignore.
type VersionInfo struct {
	Version string `json:"version"`
	Release string `json:"release"`
	RepoID  string `json:"repoid"`
}

// apiResponse is the envelope used by Proxmox REST endpoints.
// Successful calls return the payload in `data`; errors are non-2xx HTTP
// responses, so there is no dedicated error field here.
type apiResponse struct {
	Data interface{} `json:"data"`
}

// ResourceType is the VM/CT discriminator used by Proxmox /cluster/resources.
type ResourceType string

const (
	ResourceTypeVM ResourceType = "qemu" // QEMU VM
	ResourceTypeCT ResourceType = "lxc"  // LXC container
)

// PathSegment returns the URL fragment used under /nodes/{node}/... for this
// resource type. Example: "qemu" -> /nodes/n1/qemu/100/...
func (r ResourceType) PathSegment() string { return string(r) }

// ResourceRef identifies a VM or CT after name→id lookup. Every mutating
// operation needs the tuple (node, type, vmid).
type ResourceRef struct {
	Node    string
	Type    ResourceType
	VMID    int
	Name    string
	Status  string // running / stopped / …
	MaxMem  uint64 // bytes
	MaxDisk uint64 // bytes
	CPUs    int
}

// clusterResource is the raw shape returned by /cluster/resources?type=vm.
// Not exported — callers go through resolveResource() which returns
// ResourceRef.
type clusterResource struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // "qemu" | "lxc" | "node" | "storage" | ...
	Node    string `json:"node"`
	VMID    int    `json:"vmid,omitempty"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	MaxMem  uint64 `json:"maxmem,omitempty"`
	MaxDisk uint64 `json:"maxdisk,omitempty"`
	CPUs    int    `json:"cpus,omitempty"`
}

// Snapshot represents a single snapshot entry.
type Snapshot struct {
	Name        string `json:"name"`
	Parent      string `json:"parent,omitempty"`
	Description string `json:"description,omitempty"`
	SnapTime    int64  `json:"snaptime,omitempty"`
}

// StorageContent represents one entry in /nodes/{node}/storage/{storage}/content.
type StorageContent struct {
	VolID   string `json:"volid"`
	Content string `json:"content"`
	Format  string `json:"format,omitempty"`
	Size    uint64 `json:"size,omitempty"`
}

// GuestIP is a single IP address reported by the QEMU guest agent or LXC
// config. Phase 3 SSH resolution iterates these.
type GuestIP struct {
	Interface string
	Address   string
	Family    string // "ipv4" / "ipv6"
}
