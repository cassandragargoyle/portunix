package sandbox

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDirectory(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")

	if err := os.MkdirAll(filepath.Join(src, "sub"), 0750); err != nil {
		t.Fatalf("create source tree: %v", err)
	}
	files := map[string]string{"a.txt": "alpha", filepath.Join("sub", "b.txt"): "beta"}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(src, name), []byte(content), 0600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	if err := copyDirectory(src, dst); err != nil {
		t.Fatalf("copyDirectory failed: %v", err)
	}

	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s: expected %q, got %q", name, want, got)
		}
	}
}
