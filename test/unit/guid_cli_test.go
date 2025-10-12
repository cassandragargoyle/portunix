package unit

import (
	"testing"

	"portunix.ai/app/guid"
)

// TestGUIDModuleBasic tests basic functionality of the GUID module
func TestGUIDModuleBasic(t *testing.T) {
	// Test random generation
	uuid1, err := guid.GenerateRandom()
	if err != nil {
		t.Fatalf("GenerateRandom failed: %v", err)
	}

	if !guid.Validate(uuid1) {
		t.Errorf("Generated random UUID is invalid: %s", uuid1)
	}

	// Test deterministic generation
	uuid2, err := guid.GenerateFromStrings("test", "value")
	if err != nil {
		t.Fatalf("GenerateFromStrings failed: %v", err)
	}

	if !guid.Validate(uuid2) {
		t.Errorf("Generated deterministic UUID is invalid: %s", uuid2)
	}

	// Test consistency
	uuid3, err := guid.GenerateFromStrings("test", "value")
	if err != nil {
		t.Fatalf("Second GenerateFromStrings failed: %v", err)
	}

	if uuid2 != uuid3 {
		t.Errorf("Deterministic generation not consistent: %s != %s", uuid2, uuid3)
	}

	// Test validation
	if !guid.Validate("550e8400-e29b-41d4-a716-446655440000") {
		t.Error("Valid UUID rejected by validation")
	}

	if guid.Validate("invalid-uuid") {
		t.Error("Invalid UUID accepted by validation")
	}
}