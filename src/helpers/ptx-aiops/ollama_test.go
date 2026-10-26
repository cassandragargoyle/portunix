/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import "testing"

// Sample inspect output trimmed to the fields relevant for GPU and port detection

const dockerInspectGPUs = `[{
  "Config": {"Image": "ollama/ollama:latest"},
  "HostConfig": {
    "Devices": [],
    "DeviceRequests": [{"Driver": "", "Count": -1, "DeviceIDs": null, "Capabilities": [["gpu"]], "Options": {}}],
    "PortBindings": {"11434/tcp": [{"HostIp": "", "HostPort": "11434"}]}
  },
  "NetworkSettings": {"Ports": {"11434/tcp": [{"HostIp": "0.0.0.0", "HostPort": "11434"}, {"HostIp": "::", "HostPort": "11434"}]}}
}]`

const dockerInspectCDI = `[{
  "HostConfig": {
    "DeviceRequests": [{"Driver": "cdi", "Count": 0, "DeviceIDs": ["nvidia.com/gpu=all"], "Capabilities": null}]
  }
}]`

const dockerInspectNoGPU = `[{
  "HostConfig": {
    "Devices": [],
    "DeviceRequests": null,
    "PortBindings": {"11434/tcp": [{"HostIp": "", "HostPort": "21434"}]}
  },
  "NetworkSettings": {"Ports": {"11434/tcp": [{"HostIp": "0.0.0.0", "HostPort": "21434"}]}}
}]`

const dockerInspectFuse = `[{
  "HostConfig": {
    "Devices": [{"PathOnHost": "/dev/fuse", "PathInContainer": "/dev/fuse", "CgroupPermissions": "rwm"}],
    "DeviceRequests": null
  }
}]`

const podmanInspectCDI = `[{
  "Config": {
    "CreateCommand": ["podman", "run", "-d", "--name", "ollama", "--device", "nvidia.com/gpu=all", "-p", "11434:11434", "docker.io/ollama/ollama"]
  },
  "HostConfig": {
    "Devices": [],
    "DeviceRequests": [{"PathOnHost": "/dev/dri/renderD129", "PathInContainer": "/dev/dri/renderD129"}],
    "PortBindings": {"11434/tcp": [{"HostIp": "", "HostPort": "11434"}]}
  },
  "NetworkSettings": {"Ports": {"11434/tcp": [{"HostIp": "", "HostPort": "11434"}]}}
}]`

const podmanInspectNvidiaDevices = `[{
  "Config": {"CreateCommand": ["podman", "run", "-d", "--name", "ollama", "docker.io/ollama/ollama"]},
  "HostConfig": {
    "Devices": [{"PathOnHost": "/dev/nvidia0", "PathInContainer": "/dev/nvidia0"}, {"PathOnHost": "/dev/nvidiactl", "PathInContainer": "/dev/nvidiactl"}]
  }
}]`

const podmanInspectDRIOnly = `[{
  "Config": {"CreateCommand": ["podman", "run", "-d", "--device", "/dev/dri", "docker.io/ollama/ollama"]},
  "HostConfig": {
    "Devices": [{"PathOnHost": "/dev/dri/renderD128", "PathInContainer": "/dev/dri/renderD128"}]
  }
}]`

const podmanInspectNoPort = `[{
  "Config": {"CreateCommand": ["podman", "run", "-d", "--name", "ollama", "alpine", "sleep", "3600"]},
  "HostConfig": {"Devices": [], "PortBindings": {}},
  "NetworkSettings": {"Ports": {}}
}]`

const dockerInspectStoppedRandomPort = `[{
  "HostConfig": {"PortBindings": {"11434/tcp": [{"HostIp": "", "HostPort": ""}]}},
  "NetworkSettings": {"Ports": {}}
}]`

const dockerInspectStoppedFixedPort = `[{
  "HostConfig": {"PortBindings": {"11434/tcp": [{"HostIp": "127.0.0.1", "HostPort": "21434"}]}},
  "NetworkSettings": {"Ports": {}}
}]`

func TestParseGPUEnabled(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{"docker --gpus all", dockerInspectGPUs, true},
		{"docker CDI device", dockerInspectCDI, true},
		{"docker no GPU", dockerInspectNoGPU, false},
		{"docker /dev/fuse only", dockerInspectFuse, false},
		{"podman CDI nvidia.com/gpu=all", podmanInspectCDI, true},
		{"podman CDI --device= form", `[{"Config": {"CreateCommand": ["podman", "run", "--device=nvidia.com/gpu=0", "img"]}}]`, true},
		{"podman --gpus", `[{"Config": {"CreateCommand": ["podman", "run", "--gpus", "all", "img"]}}]`, true},
		{"podman NVIDIA device nodes", podmanInspectNvidiaDevices, true},
		{"podman /dev/dri only", podmanInspectDRIOnly, false},
		{"podman no GPU", podmanInspectNoPort, false},
		{"invalid JSON", `not json`, false},
		{"empty array", `[]`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseGPUEnabled([]byte(tt.data)); got != tt.want {
				t.Errorf("parseGPUEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		wantHost string
		wantPort string
	}{
		{"docker 11434:11434", dockerInspectGPUs, "localhost", "11434"},
		{"docker 21434:11434", dockerInspectNoGPU, "localhost", "21434"},
		{"podman 11434:11434", podmanInspectCDI, "localhost", "11434"},
		{"podman no port published", podmanInspectNoPort, "", ""},
		{"no port data", dockerInspectCDI, "", ""},
		{"stopped, random port", dockerInspectStoppedRandomPort, "", ""},
		{"stopped, bound to 127.0.0.1", dockerInspectStoppedFixedPort, "127.0.0.1", "21434"},
		{"IPv6 host IP", `[{"NetworkSettings": {"Ports": {"11434/tcp": [{"HostIp": "::1", "HostPort": "11434"}]}}}]`, "[::1]", "11434"},
		{"other port only", `[{"NetworkSettings": {"Ports": {"8080/tcp": [{"HostIp": "", "HostPort": "8080"}]}}}]`, "", ""},
		{"invalid JSON", `not json`, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := parseHostPort([]byte(tt.data))
			if host != tt.wantHost || port != tt.wantPort {
				t.Errorf("parseHostPort() = (%q, %q), want (%q, %q)", host, port, tt.wantHost, tt.wantPort)
			}
		})
	}
}
