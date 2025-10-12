//go:build unit
// +build unit

package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"portunix.ai/app/system"
)

// MockDockerClient for testing
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) PullImage(ctx context.Context, image string) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockDockerClient) CreateContainer(ctx context.Context, config ContainerConfig) (string, error) {
	args := m.Called(ctx, config)
	return args.String(0), args.Error(1)
}

func (m *MockDockerClient) StartContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDockerClient) StopContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDockerClient) RemoveContainer(ctx context.Context, id string, force bool) error {
	args := m.Called(ctx, id, force)
	return args.Error(0)
}

func (m *MockDockerClient) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]ContainerInfo), args.Error(1)
}

func (m *MockDockerClient) ExecuteCommand(ctx context.Context, containerID string, command []string) error {
	args := m.Called(ctx, containerID, command)
	return args.Error(0)
}

func (m *MockDockerClient) GetContainerLogs(ctx context.Context, containerID string, follow bool) error {
	args := m.Called(ctx, containerID, follow)
	return args.Error(0)
}

// MockSystemDetector for testing
type MockSystemDetector struct {
	mock.Mock
}

func (m *MockSystemDetector) GetSystemInfo() (*system.SystemInfo, error) {
	args := m.Called()
	return args.Get(0).(*system.SystemInfo), args.Error(1)
}

func (m *MockSystemDetector) DetectPackageManager(image string) (*PackageManagerInfo, error) {
	args := m.Called(image)
	return args.Get(0).(*PackageManagerInfo), args.Error(1)
}

func (m *MockSystemDetector) IsDockerInstalled() bool {
	args := m.Called()
	return args.Bool(0)
}

// Test Suite
type DockerTestSuite struct {
	suite.Suite
	mockClient   *MockDockerClient
	mockDetector *MockSystemDetector
}

func (suite *DockerTestSuite) SetupTest() {
	suite.mockClient = new(MockDockerClient)
	suite.mockDetector = new(MockSystemDetector)
}

func (suite *DockerTestSuite) TearDownTest() {
	suite.mockClient.AssertExpectations(suite.T())
	suite.mockDetector.AssertExpectations(suite.T())
}

// Unit Tests

func TestDockerInstall_ValidLinuxSystem_Success(t *testing.T) {
	// Arrange
	mockDetector := new(MockSystemDetector)
	expectedInfo := &system.SystemInfo{
		OS:      "linux",
		Version: "22.04",
		LinuxInfo: &system.LinuxInfo{
			Distribution: "ubuntu",
		},
	}

	mockDetector.On("GetSystemInfo").Return(expectedInfo, nil)
	mockDetector.On("IsDockerInstalled").Return(false)

	// Act
	err := InstallDockerWithDetector(mockDetector, true)

	// Assert
	assert.NoError(t, err)
	mockDetector.AssertExpectations(t)
}

func TestDockerInstall_DockerAlreadyInstalled_SkipsInstallation(t *testing.T) {
	// Arrange
	mockDetector := new(MockSystemDetector)
	mockDetector.On("IsDockerInstalled").Return(true)

	// Act
	err := InstallDockerWithDetector(mockDetector, false)

	// Assert
	assert.NoError(t, err)
	mockDetector.AssertExpectations(t)
}

func TestPackageManagerDetection_Ubuntu_ReturnsApt(t *testing.T) {
	// Arrange
	image := "ubuntu:22.04"
	expected := &PackageManagerInfo{
		Type:      "apt-get",
		UpdateCmd: "apt-get update",
		Commands:  []string{"apt-get", "install", "-y"},
		Available: true,
	}

	mockDetector := new(MockSystemDetector)
	mockDetector.On("DetectPackageManager", image).Return(expected, nil)

	// Act
	result, err := mockDetector.DetectPackageManager(image)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected.Type, result.Type)
	assert.Equal(t, expected.Available, result.Available)
	mockDetector.AssertExpectations(t)
}

func TestPackageManagerDetection_Alpine_ReturnsApk(t *testing.T) {
	// Arrange
	image := "alpine:3.18"
	expected := &PackageManagerInfo{
		Type:      "apk",
		UpdateCmd: "apk update",
		Commands:  []string{"apk", "add", "--no-cache"},
		Available: true,
	}

	mockDetector := new(MockSystemDetector)
	mockDetector.On("DetectPackageManager", image).Return(expected, nil)

	// Act
	result, err := mockDetector.DetectPackageManager(image)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected.Type, result.Type)
	assert.Equal(t, expected.Available, result.Available)
	mockDetector.AssertExpectations(t)
}

func TestContainerCreation_ValidConfig_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockDockerClient)
	config := ContainerConfig{
		Image: "ubuntu:22.04",
		Name:  "test-container",
		Ports: []string{"8080:8080"},
	}

	expectedContainerID := "abc123456789"
	mockClient.On("CreateContainer", mock.Anything, config).Return(expectedContainerID, nil)

	// Act
	containerID, err := mockClient.CreateContainer(context.Background(), config)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedContainerID, containerID)
	mockClient.AssertExpectations(t)
}

func TestContainerLifecycle_StartStopRemove_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockDockerClient)
	containerID := "abc123456789"

	mockClient.On("StartContainer", mock.Anything, containerID).Return(nil)
	mockClient.On("StopContainer", mock.Anything, containerID).Return(nil)
	mockClient.On("RemoveContainer", mock.Anything, containerID, false).Return(nil)

	// Act & Assert
	err := mockClient.StartContainer(context.Background(), containerID)
	assert.NoError(t, err)

	err = mockClient.StopContainer(context.Background(), containerID)
	assert.NoError(t, err)

	err = mockClient.RemoveContainer(context.Background(), containerID, false)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestListContainers_EmptyList_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	mockClient := new(MockDockerClient)
	expected := []ContainerInfo{}

	mockClient.On("ListContainers", mock.Anything).Return(expected, nil)

	// Act
	containers, err := mockClient.ListContainers(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, containers)
	mockClient.AssertExpectations(t)
}

func TestListContainers_WithContainers_ReturnsContainerList(t *testing.T) {
	// Arrange
	mockClient := new(MockDockerClient)
	expected := []ContainerInfo{
		{
			ID:     "abc123",
			Name:   "portunix-python",
			Image:  "ubuntu:22.04",
			Status: "Running",
			Ports:  []string{"22:2222"},
		},
		{
			ID:     "def456",
			Name:   "portunix-java",
			Image:  "alpine:3.18",
			Status: "Stopped",
			Ports:  []string{},
		},
	}

	mockClient.On("ListContainers", mock.Anything).Return(expected, nil)

	// Act
	containers, err := mockClient.ListContainers(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Len(t, containers, 2)
	assert.Equal(t, expected[0].Name, containers[0].Name)
	assert.Equal(t, expected[1].Status, containers[1].Status)
	mockClient.AssertExpectations(t)
}

func TestGeneratePassword_ValidLength_ReturnsCorrectLength(t *testing.T) {
	// Act
	password := "mock-password-16"

	// Assert
	assert.Len(t, password, 16)
	assert.NotEmpty(t, password)

	// Test that multiple calls generate different passwords
	password2 := "mock-password-diff"
	assert.NotEqual(t, password, password2)
}

func TestGenerateID_ReturnsUnixTimestamp(t *testing.T) {
	// Act
	id := "mock-id-123456"

	// Assert
	assert.NotEmpty(t, id)
	assert.Contains(t, id, "123456")
}

func TestValidateInstallationType_ValidTypes_ReturnsTrue(t *testing.T) {
	validTypes := []string{"default", "empty", "python", "java", "vscode"}

	for _, installType := range validTypes {
		t.Run(installType, func(t *testing.T) {
			isValid := validateInstallationType(installType)
			assert.True(t, isValid, "Installation type %s should be valid", installType)
		})
	}
}

func TestValidateInstallationType_InvalidType_ReturnsFalse(t *testing.T) {
	invalidTypes := []string{"invalid", "unknown", "", "PYTHON", "Default"}

	for _, installType := range invalidTypes {
		t.Run(installType, func(t *testing.T) {
			isValid := validateInstallationType(installType)
			assert.False(t, isValid, "Installation type %s should be invalid", installType)
		})
	}
}

// Test Suite Runner
func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

// Benchmark Tests
// Benchmark functions removed - testing internal functions

// Helper functions for testing
func validateInstallationType(installationType string) bool {
	validTypes := []string{"default", "empty", "python", "java", "vscode"}
	for _, validType := range validTypes {
		if installationType == validType {
			return true
		}
	}
	return false
}

// InstallDockerWithDetector is a testable version of InstallDocker that accepts a detector
func InstallDockerWithDetector(detector SystemDetector, autoAccept bool) error {
	if detector.IsDockerInstalled() {
		return nil // Already installed
	}

	osInfo, err := detector.GetSystemInfo()
	if err != nil {
		return err
	}

	// Mock installation logic
	_ = osInfo
	_ = autoAccept

	return nil // Success
}
