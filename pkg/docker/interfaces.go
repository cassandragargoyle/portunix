package docker

import (
	"context"

	"portunix.cz/app/system"
)

//go:generate mockgen -source=interfaces.go -destination=../../test/mocks/docker_mock.go

// DockerClient interface for mocking Docker operations.
type DockerClient interface {
	PullImage(ctx context.Context, image string) error
	CreateContainer(ctx context.Context, config ContainerConfig) (string, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string, force bool) error
	ListContainers(ctx context.Context) ([]ContainerInfo, error)
	ExecuteCommand(ctx context.Context, containerID string, command []string) error
	GetContainerLogs(ctx context.Context, containerID string, follow bool) error
}

// SystemDetector interface for mocking system detection.
type SystemDetector interface {
	GetSystemInfo() (*system.SystemInfo, error)
	DetectPackageManager(image string) (*PackageManagerInfo, error)
	IsDockerInstalled() bool
}

// NetworkManager interface for network operations.
type NetworkManager interface {
	TestConnectivity(host string, port string) error
	GetContainerIP(containerID string) (string, error)
	WaitForService(host, port string, timeout int) error
}

// FileSystem interface for file operations.
type FileSystem interface {
	CreateTempDir(prefix string) (string, error)
	WriteFile(path string, content []byte, perm int) error
	ReadFile(path string) ([]byte, error)
	Exists(path string) bool
	Remove(path string) error
	MkdirAll(path string, perm int) error
}

// ContainerConfig represents container configuration.
type ContainerConfig struct {
	Image       string
	Name        string
	Ports       []string
	Volumes     []string
	Environment []string
	Command     []string
	WorkingDir  string
	Privileged  bool
	Network     string
}

// ContainerInfo represents container information.
type ContainerInfo struct {
	ID     string
	Name   string
	Image  string
	Status string
	Ports  []string
	State  string
}

// PackageManagerInfo represents package manager information.
type PackageManagerInfo struct {
	Type      string   // apt-get, yum, dnf, apk, etc.
	Commands  []string // install commands
	UpdateCmd string   // update command
	SearchCmd string   // search command
	Available bool     // whether package manager is available
}
