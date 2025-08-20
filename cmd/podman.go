package cmd

import (
	"github.com/spf13/cobra"
)

// podmanCmd represents the podman command
var podmanCmd = &cobra.Command{
	Use:   "podman",
	Short: "Manages Podman containers and Podman installation.",
	Long: `The podman command provides comprehensive Podman container management capabilities
for Portunix, allowing users to run Portunix installations and commands inside Podman
containers across Windows and Linux.

Key features:
- Intelligent OS-based Podman installation
- Multi-platform container support (Ubuntu, Alpine, CentOS, etc.)
- SSH server setup in containers
- Cache directory mounting for persistent downloads
- Flexible base image selection
- Rootless containers by default (enhanced security)
- Daemonless operation
- OCI-compatible with Docker images
- Pod support for Kubernetes-style container grouping

Available commands:
  install              Install Podman with intelligent OS detection
  run-in-container     Run Portunix installation inside a Podman container
  build               Build Portunix Podman images
  start               Start existing container
  stop                Stop running container
  list                List Portunix containers
  remove              Remove containers
  logs                View container logs
  exec                Execute commands in container

Examples:
  portunix podman install
  portunix podman run-in-container default
  portunix podman run-in-container python --image alpine:3.18
  portunix podman run-in-container java --rootless
  portunix podman list
  portunix podman logs <container-id>`,
}

func init() {
	rootCmd.AddCommand(podmanCmd)
}
