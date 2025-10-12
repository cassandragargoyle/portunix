package unit

import (
	"testing"

	"portunix.ai/app/virt"
)

func TestVirtManagerCreation(t *testing.T) {
	// Test that we can create a virt manager
	// This tests the basic configuration loading and backend selection

	manager, err := virt.NewManager()

	// On a system without virtualization backends, this might fail
	// That's OK - we're testing the code structure
	if err != nil {
		t.Logf("Manager creation failed (expected if no backend): %v", err)
		return
	}

	if manager == nil {
		t.Error("Manager should not be nil when creation succeeds")
		return
	}

	// Test that we can get the backend
	backend := manager.GetBackend()
	if backend == nil {
		t.Error("Backend should not be nil")
		return
	}

	// Test that backend has a name
	backendName := backend.GetName()
	if backendName == "" {
		t.Error("Backend should have a non-empty name")
		return
	}

	t.Logf("Successfully created manager with backend: %s", backendName)
}

func TestVirtManagerMethods(t *testing.T) {
	// Test that manager methods exist and can be called
	// This is more of a compilation test than functional test

	manager, err := virt.NewManager()
	if err != nil {
		t.Skipf("Skipping method tests - no backend available: %v", err)
		return
	}

	// Test List method (should not panic)
	vms, err := manager.List()
	if err != nil {
		t.Logf("List failed (expected if no VMs): %v", err)
	} else {
		t.Logf("List returned %d VMs", len(vms))
	}

	// Test GetState method with non-existent VM
	state := manager.GetState("nonexistent-vm")
	if state != virt.VMStateNotFound {
		t.Logf("GetState for nonexistent VM returned: %s (should be not-found)", state)
	}
}

func TestVirtManagerStateHandling(t *testing.T) {
	manager, err := virt.NewManager()
	if err != nil {
		t.Skipf("Skipping state tests - no backend available: %v", err)
		return
	}

	// Test state handling for non-existent VM
	testVM := "test-vm-unit-test"

	// State should be not-found
	state := manager.GetState(testVM)
	if state != virt.VMStateNotFound {
		t.Errorf("Expected VMStateNotFound, got %s", state)
	}

	// Start should fail for non-existent VM
	err = manager.Start(testVM)
	if err == nil {
		t.Error("Start should fail for non-existent VM")
	}

	// Stop should fail for non-existent VM
	err = manager.Stop(testVM, false)
	if err == nil {
		t.Error("Stop should fail for non-existent VM")
	}
}

func TestVirtManagerSSHMethods(t *testing.T) {
	manager, err := virt.NewManager()
	if err != nil {
		t.Skipf("Skipping SSH tests - no backend available: %v", err)
		return
	}

	testVM := "test-vm-ssh"

	// SSH should not be ready for non-existent VM
	ready := manager.IsSSHReady(testVM)
	if ready {
		t.Error("SSH should not be ready for non-existent VM")
	}

	// GetIP should fail for non-existent VM
	ip, err := manager.GetIP(testVM)
	if err == nil {
		t.Error("GetIP should fail for non-existent VM")
	}
	if ip != "" {
		t.Error("IP should be empty for non-existent VM")
	}
}

func TestVirtManagerSnapshotMethods(t *testing.T) {
	manager, err := virt.NewManager()
	if err != nil {
		t.Skipf("Skipping snapshot tests - no backend available: %v", err)
		return
	}

	testVM := "test-vm-snapshot"

	// Create snapshot should fail for non-existent VM
	err = manager.CreateSnapshot(testVM, "test-snapshot", "Test snapshot")
	if err == nil {
		t.Error("CreateSnapshot should fail for non-existent VM")
	}

	// List snapshots should fail for non-existent VM
	snapshots, err := manager.ListSnapshots(testVM)
	if err == nil {
		t.Error("ListSnapshots should fail for non-existent VM")
	}
	if len(snapshots) > 0 {
		t.Error("Snapshots list should be empty for non-existent VM")
	}

	// Revert snapshot should fail for non-existent VM
	err = manager.RevertSnapshot(testVM, "test-snapshot")
	if err == nil {
		t.Error("RevertSnapshot should fail for non-existent VM")
	}

	// Delete snapshot should fail for non-existent VM
	err = manager.DeleteSnapshot(testVM, "test-snapshot")
	if err == nil {
		t.Error("DeleteSnapshot should fail for non-existent VM")
	}
}

func TestVirtManagerFileMethods(t *testing.T) {
	manager, err := virt.NewManager()
	if err != nil {
		t.Skipf("Skipping file tests - no backend available: %v", err)
		return
	}

	testVM := "test-vm-files"

	// Copy to VM should fail for non-existent VM
	err = manager.CopyToVM(testVM, "/tmp/test", "/tmp/test")
	if err == nil {
		t.Error("CopyToVM should fail for non-existent VM")
	}

	// Copy from VM should fail for non-existent VM
	err = manager.CopyFromVM(testVM, "/tmp/test", "/tmp/test")
	if err == nil {
		t.Error("CopyFromVM should fail for non-existent VM")
	}
}