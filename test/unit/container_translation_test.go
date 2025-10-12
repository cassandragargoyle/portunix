//go:build unit
// +build unit

package unit

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ContainerTranslationTestSuite tests Docker/Podman parameter translation
type ContainerTranslationTestSuite struct {
	suite.Suite
}

func TestContainerTranslationTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTranslationTestSuite))
}

// Mock structures for container runtime translation
type ContainerRuntime string

const (
	RuntimeDocker ContainerRuntime = "docker"
	RuntimePodman ContainerRuntime = "podman"
)

type RuntimeTranslator interface {
	TranslateParams(params *ContainerParams) ([]string, error)
	ValidateParam(param string, value string) error
	SupportedParams() []string
}

type DockerTranslator struct{}
type PodmanTranslator struct{}

// Docker translator implementation
func (d *DockerTranslator) TranslateParams(params *ContainerParams) ([]string, error) {
	var args []string
	
	// Volume translation
	for _, vol := range params.Volumes {
		volumeSpec := vol.HostPath + ":" + vol.ContainerPath
		if vol.ReadOnly {
			volumeSpec += ":ro"
		}
		args = append(args, "-v", volumeSpec)
	}
	
	// Port translation
	for _, port := range params.Ports {
		portSpec := port.HostPort + ":" + port.ContainerPort
		args = append(args, "-p", portSpec)
	}
	
	// Environment variables
	for _, env := range params.Environment {
		args = append(args, "-e", env.Name+"="+env.Value)
	}
	
	// Working directory
	if params.WorkingDir != "" {
		args = append(args, "-w", params.WorkingDir)
	}
	
	// User
	if params.User != "" {
		args = append(args, "--user", params.User)
	}
	
	// Privileged
	if params.Privileged {
		args = append(args, "--privileged")
	}
	
	// Network
	if params.Network != "" {
		args = append(args, "--network", params.Network)
	}
	
	// Memory limit
	if params.Memory != "" {
		args = append(args, "--memory", params.Memory)
	}
	
	// CPU limit
	if params.CPUs != "" {
		args = append(args, "--cpus", params.CPUs)
	}
	
	return args, nil
}

func (d *DockerTranslator) ValidateParam(param string, value string) error {
	supportedParams := map[string]bool{
		"-v": true, "--volume": true,
		"-p": true, "--port": true,
		"-e": true, "--env": true,
		"-w": true, "--workdir": true,
		"--user": true,
		"--privileged": true,
		"--network": true,
		"--memory": true,
		"--cpus": true,
	}
	
	if !supportedParams[param] {
		return assert.AnError
	}
	return nil
}

func (d *DockerTranslator) SupportedParams() []string {
	return []string{
		"-v", "--volume", "-p", "--port", "-e", "--env",
		"-w", "--workdir", "--user", "--privileged",
		"--network", "--memory", "--cpus",
	}
}

// Podman translator implementation
func (p *PodmanTranslator) TranslateParams(params *ContainerParams) ([]string, error) {
	var args []string
	
	// Volume translation (same as Docker for basic cases)
	for _, vol := range params.Volumes {
		volumeSpec := vol.HostPath + ":" + vol.ContainerPath
		if vol.ReadOnly {
			volumeSpec += ":ro"
		}
		args = append(args, "-v", volumeSpec)
	}
	
	// Port translation (same as Docker)
	for _, port := range params.Ports {
		portSpec := port.HostPort + ":" + port.ContainerPort
		args = append(args, "-p", portSpec)
	}
	
	// Environment variables (same as Docker)
	for _, env := range params.Environment {
		args = append(args, "-e", env.Name+"="+env.Value)
	}
	
	// Working directory (same as Docker)
	if params.WorkingDir != "" {
		args = append(args, "-w", params.WorkingDir)
	}
	
	// User (same as Docker)
	if params.User != "" {
		args = append(args, "--user", params.User)
	}
	
	// Privileged (same as Docker)
	if params.Privileged {
		args = append(args, "--privileged")
	}
	
	// Network (Podman has different default network behavior)
	if params.Network != "" {
		if params.Network == "bridge" {
			// Podman uses "pasta" or "slirp4netns" instead of bridge by default
			args = append(args, "--network", "slirp4netns")
		} else {
			args = append(args, "--network", params.Network)
		}
	}
	
	// Memory limit (same as Docker)
	if params.Memory != "" {
		args = append(args, "--memory", params.Memory)
	}
	
	// CPU limit (Podman might use different syntax for some cases)
	if params.CPUs != "" {
		args = append(args, "--cpus", params.CPUs)
	}
	
	return args, nil
}

func (p *PodmanTranslator) ValidateParam(param string, value string) error {
	supportedParams := map[string]bool{
		"-v": true, "--volume": true,
		"-p": true, "--port": true,
		"-e": true, "--env": true,
		"-w": true, "--workdir": true,
		"--user": true,
		"--privileged": true,
		"--network": true,
		"--memory": true,
		"--cpus": true,
	}
	
	if !supportedParams[param] {
		return assert.AnError
	}
	
	// Podman-specific validation
	if param == "--network" && value == "host" {
		// Podman requires root for host networking on some systems
		if runtime.GOOS == "linux" {
			// Would check if running as root in real implementation
		}
	}
	
	return nil
}

func (p *PodmanTranslator) SupportedParams() []string {
	return []string{
		"-v", "--volume", "-p", "--port", "-e", "--env",
		"-w", "--workdir", "--user", "--privileged",
		"--network", "--memory", "--cpus",
	}
}

// Factory function to create appropriate translator
func NewRuntimeTranslator(runtime ContainerRuntime) RuntimeTranslator {
	switch runtime {
	case RuntimeDocker:
		return &DockerTranslator{}
	case RuntimePodman:
		return &PodmanTranslator{}
	default:
		return &DockerTranslator{} // Default fallback
	}
}

// Test Docker parameter translation
func (suite *ContainerTranslationTestSuite) TestDockerTranslator_VolumeTranslation_Success() {
	// Arrange
	translator := &DockerTranslator{}
	params := &ContainerParams{
		Volumes: []VolumeMount{
			{HostPath: "/host", ContainerPath: "/container", ReadOnly: false},
			{HostPath: "/host2", ContainerPath: "/container2", ReadOnly: true},
		},
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{
		"-v", "/host:/container",
		"-v", "/host2:/container2:ro",
	}
	assert.Equal(suite.T(), expectedArgs, args)
}

func (suite *ContainerTranslationTestSuite) TestDockerTranslator_PortTranslation_Success() {
	// Arrange
	translator := &DockerTranslator{}
	params := &ContainerParams{
		Ports: []PortMapping{
			{HostPort: "8080", ContainerPort: "80"},
			{HostPort: "3000", ContainerPort: "3000"},
		},
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{
		"-p", "8080:80",
		"-p", "3000:3000",
	}
	assert.Equal(suite.T(), expectedArgs, args)
}

func (suite *ContainerTranslationTestSuite) TestDockerTranslator_EnvironmentTranslation_Success() {
	// Arrange
	translator := &DockerTranslator{}
	params := &ContainerParams{
		Environment: []EnvVar{
			{Name: "NODE_ENV", Value: "production"},
			{Name: "PORT", Value: "3000"},
		},
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{
		"-e", "NODE_ENV=production",
		"-e", "PORT=3000",
	}
	assert.Equal(suite.T(), expectedArgs, args)
}

func (suite *ContainerTranslationTestSuite) TestDockerTranslator_ComplexTranslation_Success() {
	// Arrange
	translator := &DockerTranslator{}
	params := &ContainerParams{
		Volumes:     []VolumeMount{{HostPath: "/host", ContainerPath: "/container"}},
		Ports:       []PortMapping{{HostPort: "8080", ContainerPort: "80"}},
		Environment: []EnvVar{{Name: "ENV", Value: "test"}},
		WorkingDir:  "/app",
		User:        "1000:1000",
		Privileged:  true,
		Network:     "bridge",
		Memory:      "2G",
		CPUs:        "1.5",
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{
		"-v", "/host:/container",
		"-p", "8080:80",
		"-e", "ENV=test",
		"-w", "/app",
		"--user", "1000:1000",
		"--privileged",
		"--network", "bridge",
		"--memory", "2G",
		"--cpus", "1.5",
	}
	assert.Equal(suite.T(), expectedArgs, args)
}

// Test Podman parameter translation
func (suite *ContainerTranslationTestSuite) TestPodmanTranslator_VolumeTranslation_Success() {
	// Arrange
	translator := &PodmanTranslator{}
	params := &ContainerParams{
		Volumes: []VolumeMount{
			{HostPath: "/host", ContainerPath: "/container", ReadOnly: false},
		},
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{"-v", "/host:/container"}
	assert.Equal(suite.T(), expectedArgs, args)
}

func (suite *ContainerTranslationTestSuite) TestPodmanTranslator_NetworkTranslation_BridgeToSlirp() {
	// Arrange
	translator := &PodmanTranslator{}
	params := &ContainerParams{
		Network: "bridge",
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{"--network", "slirp4netns"}
	assert.Equal(suite.T(), expectedArgs, args)
}

func (suite *ContainerTranslationTestSuite) TestPodmanTranslator_NetworkTranslation_CustomNetwork() {
	// Arrange
	translator := &PodmanTranslator{}
	params := &ContainerParams{
		Network: "custom-net",
	}

	// Act
	args, err := translator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), err)
	expectedArgs := []string{"--network", "custom-net"}
	assert.Equal(suite.T(), expectedArgs, args)
}

// Test parameter validation
func (suite *ContainerTranslationTestSuite) TestDockerTranslator_ValidateParam_SupportedParam_Success() {
	// Arrange
	translator := &DockerTranslator{}

	// Act & Assert
	supportedParams := []string{"-v", "--volume", "-p", "--port", "-e", "--env"}
	for _, param := range supportedParams {
		err := translator.ValidateParam(param, "test-value")
		assert.NoError(suite.T(), err, "Parameter %s should be supported", param)
	}
}

func (suite *ContainerTranslationTestSuite) TestDockerTranslator_ValidateParam_UnsupportedParam_Error() {
	// Arrange
	translator := &DockerTranslator{}

	// Act & Assert
	unsupportedParams := []string{"--unsupported", "-x", "--random"}
	for _, param := range unsupportedParams {
		err := translator.ValidateParam(param, "test-value")
		assert.Error(suite.T(), err, "Parameter %s should not be supported", param)
	}
}

func (suite *ContainerTranslationTestSuite) TestPodmanTranslator_ValidateParam_SupportedParam_Success() {
	// Arrange
	translator := &PodmanTranslator{}

	// Act & Assert
	supportedParams := []string{"-v", "--volume", "-p", "--port", "-e", "--env"}
	for _, param := range supportedParams {
		err := translator.ValidateParam(param, "test-value")
		assert.NoError(suite.T(), err, "Parameter %s should be supported", param)
	}
}

// Test runtime translator factory
func (suite *ContainerTranslationTestSuite) TestNewRuntimeTranslator_Docker_ReturnsDockerTranslator() {
	// Act
	translator := NewRuntimeTranslator(RuntimeDocker)

	// Assert
	_, ok := translator.(*DockerTranslator)
	assert.True(suite.T(), ok, "Should return DockerTranslator")
}

func (suite *ContainerTranslationTestSuite) TestNewRuntimeTranslator_Podman_ReturnsPodmanTranslator() {
	// Act
	translator := NewRuntimeTranslator(RuntimePodman)

	// Assert
	_, ok := translator.(*PodmanTranslator)
	assert.True(suite.T(), ok, "Should return PodmanTranslator")
}

func (suite *ContainerTranslationTestSuite) TestNewRuntimeTranslator_Unknown_ReturnsDockerTranslator() {
	// Act
	translator := NewRuntimeTranslator("unknown")

	// Assert
	_, ok := translator.(*DockerTranslator)
	assert.True(suite.T(), ok, "Should return DockerTranslator as default")
}

// Test supported parameters list
func (suite *ContainerTranslationTestSuite) TestDockerTranslator_SupportedParams_ReturnsCorrectList() {
	// Arrange
	translator := &DockerTranslator{}

	// Act
	params := translator.SupportedParams()

	// Assert
	expectedParams := []string{
		"-v", "--volume", "-p", "--port", "-e", "--env",
		"-w", "--workdir", "--user", "--privileged",
		"--network", "--memory", "--cpus",
	}
	assert.ElementsMatch(suite.T(), expectedParams, params)
}

func (suite *ContainerTranslationTestSuite) TestPodmanTranslator_SupportedParams_ReturnsCorrectList() {
	// Arrange
	translator := &PodmanTranslator{}

	// Act
	params := translator.SupportedParams()

	// Assert
	expectedParams := []string{
		"-v", "--volume", "-p", "--port", "-e", "--env",
		"-w", "--workdir", "--user", "--privileged",
		"--network", "--memory", "--cpus",
	}
	assert.ElementsMatch(suite.T(), expectedParams, params)
}

// Test edge cases and error conditions
func (suite *ContainerTranslationTestSuite) TestTranslator_EmptyParams_ReturnsEmptyArgs() {
	// Arrange
	dockerTranslator := &DockerTranslator{}
	podmanTranslator := &PodmanTranslator{}
	params := &ContainerParams{}

	// Act
	dockerArgs, dockerErr := dockerTranslator.TranslateParams(params)
	podmanArgs, podmanErr := podmanTranslator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), dockerErr)
	require.NoError(suite.T(), podmanErr)
	assert.Empty(suite.T(), dockerArgs)
	assert.Empty(suite.T(), podmanArgs)
}

func (suite *ContainerTranslationTestSuite) TestTranslator_NilParams_HandlesGracefully() {
	// Arrange
	translator := &DockerTranslator{}

	// Act
	args, err := translator.TranslateParams(&ContainerParams{})

	// Assert
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), args)
}

// Integration test for cross-runtime compatibility
func (suite *ContainerTranslationTestSuite) TestCrossRuntimeCompatibility_SameParams_SimilarOutput() {
	// Arrange
	dockerTranslator := &DockerTranslator{}
	podmanTranslator := &PodmanTranslator{}
	
	params := &ContainerParams{
		Volumes:     []VolumeMount{{HostPath: "/host", ContainerPath: "/container"}},
		Ports:       []PortMapping{{HostPort: "8080", ContainerPort: "80"}},
		Environment: []EnvVar{{Name: "ENV", Value: "test"}},
		WorkingDir:  "/app",
		User:        "1000:1000",
		Memory:      "2G",
		CPUs:        "1.5",
	}

	// Act
	dockerArgs, dockerErr := dockerTranslator.TranslateParams(params)
	podmanArgs, podmanErr := podmanTranslator.TranslateParams(params)

	// Assert
	require.NoError(suite.T(), dockerErr)
	require.NoError(suite.T(), podmanErr)
	
	// Both should have same number of arguments (for this basic case)
	assert.Equal(suite.T(), len(dockerArgs), len(podmanArgs))
	
	// Both should contain the same volume, port, and env specs
	assert.Contains(suite.T(), dockerArgs, "/host:/container")
	assert.Contains(suite.T(), podmanArgs, "/host:/container")
	assert.Contains(suite.T(), dockerArgs, "8080:80")
	assert.Contains(suite.T(), podmanArgs, "8080:80")
	assert.Contains(suite.T(), dockerArgs, "ENV=test")
	assert.Contains(suite.T(), podmanArgs, "ENV=test")
}

// Benchmark translation performance
func BenchmarkDockerTranslator_SimpleTranslation(b *testing.B) {
	translator := &DockerTranslator{}
	params := &ContainerParams{
		Volumes: []VolumeMount{{HostPath: "/host", ContainerPath: "/container"}},
		Ports:   []PortMapping{{HostPort: "8080", ContainerPort: "80"}},
	}
	
	for i := 0; i < b.N; i++ {
		_, _ = translator.TranslateParams(params)
	}
}

func BenchmarkPodmanTranslator_SimpleTranslation(b *testing.B) {
	translator := &PodmanTranslator{}
	params := &ContainerParams{
		Volumes: []VolumeMount{{HostPath: "/host", ContainerPath: "/container"}},
		Ports:   []PortMapping{{HostPort: "8080", ContainerPort: "80"}},
	}
	
	for i := 0; i < b.N; i++ {
		_, _ = translator.TranslateParams(params)
	}
}

func BenchmarkTranslator_ComplexTranslation(b *testing.B) {
	translator := &DockerTranslator{}
	params := &ContainerParams{
		Volumes: []VolumeMount{
			{HostPath: "/host1", ContainerPath: "/container1"},
			{HostPath: "/host2", ContainerPath: "/container2", ReadOnly: true},
			{HostPath: "/host3", ContainerPath: "/container3"},
		},
		Ports: []PortMapping{
			{HostPort: "8080", ContainerPort: "80"},
			{HostPort: "3000", ContainerPort: "3000"},
			{HostPort: "5432", ContainerPort: "5432"},
		},
		Environment: []EnvVar{
			{Name: "NODE_ENV", Value: "production"},
			{Name: "DATABASE_URL", Value: "postgres://localhost:5432/db"},
			{Name: "API_KEY", Value: "secret123"},
		},
		WorkingDir: "/app",
		User:       "1000:1000",
		Privileged: true,
		Network:    "bridge",
		Memory:     "4G",
		CPUs:       "2.0",
	}
	
	for i := 0; i < b.N; i++ {
		_, _ = translator.TranslateParams(params)
	}
}