/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"portunix.ai/app/docker"
	"portunix.ai/app/podman"
)

// runCall records one call of a runtime's RunContainer
type runCall struct {
	image   string
	command []string
	docker  docker.ContainerRunOptions
	podman  podman.ContainerRunOptions
}

// runContainerRun runs the real run command with a fake runtime selection and
// fake runtime calls, so no container is started; it returns the calls made
func runContainerRun(t *testing.T, runtime string, runtimeErr error, argv []string) (dockerCalls, podmanCalls []runCall) {
	t.Helper()

	origSelect, origDocker, origPodman := selectContainerRuntime, dockerRunContainer, podmanRunContainer
	t.Cleanup(func() {
		selectContainerRuntime, dockerRunContainer, podmanRunContainer = origSelect, origDocker, origPodman
	})
	selectContainerRuntime = func() (string, error) { return runtime, runtimeErr }
	dockerRunContainer = func(image string, command []string, options docker.ContainerRunOptions) error {
		dockerCalls = append(dockerCalls, runCall{image: image, command: command, docker: options})
		return nil
	}
	podmanRunContainer = func(image string, command []string, options podman.ContainerRunOptions) error {
		podmanCalls = append(podmanCalls, runCall{image: image, command: command, podman: options})
		return nil
	}

	cmd := &cobra.Command{Use: "run"}
	addContainerRunFlags(cmd)
	require.NoError(t, cmd.ParseFlags(argv))
	containerRunCmd.Run(cmd, cmd.Flags().Args())
	return dockerCalls, podmanCalls
}

// TestContainerRunCommand_RuntimeDelegation checks that the run command calls
// the selected runtime and only that one
func TestContainerRunCommand_RuntimeDelegation(t *testing.T) {
	argv := []string{"ubuntu:22.04", "echo", "hello"}

	t.Run("TC-038-I001: Docker runtime delegation", func(t *testing.T) {
		dockerCalls, podmanCalls := runContainerRun(t, "docker", nil, argv)
		require.Len(t, dockerCalls, 1)
		assert.Empty(t, podmanCalls)
		assert.Equal(t, "ubuntu:22.04", dockerCalls[0].image)
		assert.Equal(t, []string{"echo", "hello"}, dockerCalls[0].command)
	})

	t.Run("TC-038-I002: Podman runtime delegation", func(t *testing.T) {
		dockerCalls, podmanCalls := runContainerRun(t, "podman", nil, argv)
		require.Len(t, podmanCalls, 1)
		assert.Empty(t, dockerCalls)
		assert.Equal(t, "ubuntu:22.04", podmanCalls[0].image)
		assert.Equal(t, []string{"echo", "hello"}, podmanCalls[0].command)
	})

	t.Run("Separator after the image is not passed to the runtime", func(t *testing.T) {
		_, podmanCalls := runContainerRun(t, "podman", nil, []string{"alpine:latest", "--", "sh", "-c", "echo test"})
		require.Len(t, podmanCalls, 1)
		assert.Equal(t, []string{"sh", "-c", "echo test"}, podmanCalls[0].command)
	})
}

// TestContainerRunCommand_OptionsTranslation checks that every run flag
// reaches the runtime options
func TestContainerRunCommand_OptionsTranslation(t *testing.T) {
	argv := []string{
		"-d", "-i", "-t", "--name", "web",
		"-p", "8080:80", "-p", "8443:443",
		"-v", "/data:/data",
		"-e", "A=1", "-e", "B=2",
		"--network", "backend",
		"nginx:latest",
	}

	t.Run("TC-038-I003: Docker options translation", func(t *testing.T) {
		dockerCalls, _ := runContainerRun(t, "docker", nil, argv)
		require.Len(t, dockerCalls, 1)
		assert.Equal(t, docker.ContainerRunOptions{
			Detach:      true,
			Interactive: true,
			TTY:         true,
			Name:        "web",
			Ports:       []string{"8080:80", "8443:443"},
			Volumes:     []string{"/data:/data"},
			Environment: []string{"A=1", "B=2"},
			Network:     "backend",
		}, dockerCalls[0].docker)
		assert.Empty(t, dockerCalls[0].command)
	})

	t.Run("TC-038-I004: Podman options translation", func(t *testing.T) {
		_, podmanCalls := runContainerRun(t, "podman", nil, argv)
		require.Len(t, podmanCalls, 1)
		assert.Equal(t, podman.ContainerRunOptions{
			Detach:      true,
			Interactive: true,
			TTY:         true,
			Name:        "web",
			Ports:       []string{"8080:80", "8443:443"},
			Volumes:     []string{"/data:/data"},
			Environment: []string{"A=1", "B=2"},
			Network:     "backend",
		}, podmanCalls[0].podman)
	})
}

// TestContainerRunCommand_ErrorHandling checks that no runtime is called when
// the runtime cannot be used
func TestContainerRunCommand_ErrorHandling(t *testing.T) {
	argv := []string{"ubuntu:22.04"}

	t.Run("TC-038-I005: Runtime unavailable", func(t *testing.T) {
		dockerCalls, podmanCalls := runContainerRun(t, "", errors.New("Podman not installed"), argv)
		assert.Empty(t, dockerCalls)
		assert.Empty(t, podmanCalls)
	})

	t.Run("TC-038-I006: Unsupported runtime", func(t *testing.T) {
		dockerCalls, podmanCalls := runContainerRun(t, "lxc", nil, argv)
		assert.Empty(t, dockerCalls)
		assert.Empty(t, podmanCalls)
	})
}
