package github

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.client == nil {
		t.Fatal("GitHub client not initialized")
	}
}

func TestFindAsset(t *testing.T) {
	release := &Release{
		Assets: []*Asset{
			{Name: "app-linux-amd64.tar.gz", Size: 1024},
			{Name: "app-windows-amd64.zip", Size: 2048},
			{Name: "app-darwin-amd64.tar.gz", Size: 1536},
		},
	}

	asset := release.FindAsset("app-linux-amd64.tar.gz")
	if asset == nil {
		t.Fatal("Expected to find exact match asset")
	}
	if asset.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", asset.Size)
	}

	notFound := release.FindAsset("nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent asset")
	}
}

func TestFindAssetByPlatform(t *testing.T) {
	release := &Release{
		Assets: []*Asset{
			{Name: "app-linux-amd64.tar.gz", Size: 1024},
			{Name: "app-windows-amd64.zip", Size: 2048},
			{Name: "app-darwin-amd64.tar.gz", Size: 1536},
		},
	}

	tests := []struct {
		platform     string
		expectedName string
	}{
		{"linux-amd64", "app-linux-amd64.tar.gz"},
		{"windows-amd64", "app-windows-amd64.zip"},
		{"darwin-amd64", "app-darwin-amd64.tar.gz"},
	}

	for _, tt := range tests {
		asset := release.FindAssetByPlatform(tt.platform)
		if asset == nil {
			t.Errorf("Expected to find asset for platform %s", tt.platform)
			continue
		}
		if asset.Name != tt.expectedName {
			t.Errorf("Platform %s: expected %s, got %s", tt.platform, tt.expectedName, asset.Name)
		}
	}

	unknownAsset := release.FindAssetByPlatform("unsupported")
	if unknownAsset != nil {
		t.Error("Expected nil for unsupported platform")
	}
}