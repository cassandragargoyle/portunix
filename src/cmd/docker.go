package cmd

import (
	"github.com/spf13/cobra"
)

// dockerCmd represents the docker command
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Manages Docker containers and Docker installation.",
	Long: `The docker command provides comprehensive Docker container management capabilities
for Portunix, allowing users to run Portunix installations and commands inside Docker
containers across Windows and Linux.

Key features:
- Intelligent OS-based Docker installation
- Multi-platform container support (Ubuntu, Alpine, CentOS, etc.)
- SSH server setup in containers
- Cache directory mounting for persistent downloads
- Flexible base image selection

Available commands:
  install              Install Docker with intelligent OS detection
  run-in-container     Run Portunix installation inside a Docker container
  build               Build Portunix Docker images
  start               Start existing container
  stop                Stop running container
  list                List Portunix containers
  remove              Remove containers
  logs                View container logs
  exec                Execute commands in container

Examples:
  portunix docker install
  portunix docker run-in-container default
  portunix docker run-in-container python --image alpine:3.18
  portunix docker list
  portunix docker logs <container-id>`,
}

func init() {
	rootCmd.AddCommand(dockerCmd)
}
