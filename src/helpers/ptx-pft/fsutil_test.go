/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRewriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "item.md")
	if err := os.WriteFile(path, []byte("old content, longer"), 0640); err != nil {
		t.Fatal(err)
	}

	if err := rewriteExistingFile(path, []byte("new")); err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("unexpected content: %q", got)
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(path)
		if info.Mode().Perm() != 0640 {
			t.Errorf("mode changed to %v", info.Mode().Perm())
		}
	}
}

func TestRewriteExistingFile_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.md")

	if err := rewriteExistingFile(path, []byte("x")); err == nil {
		t.Error("expected an error for a missing file")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file must not be created")
	}
}

func TestRewriteExistingFile_Symlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need elevated rights on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(target, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := rewriteExistingFile(link, []byte("x")); err == nil {
		t.Error("expected an error for a symlink")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "outside" {
		t.Errorf("symlink target was modified: %q", got)
	}
}

func TestWritePrivateFile_NarrowsExistingMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not available on Windows")
	}
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := writePrivateFile(path, []byte(`{"api_token":"x"}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %v", info.Mode().Perm())
	}
}

func TestHashItem_NonASCIIStable(t *testing.T) {
	// Reference value of the rune-based hash, cached entries must stay valid
	item := &FeedbackItem{Title: "Žluťoučký kůň", Description: "úpěl", Status: "open"}

	var want uint32
	for _, c := range item.Title + "|" + item.Description + "|" + item.Status {
		want = want*31 + uint32(c)
	}

	if got := hashItem(item); got != fmt.Sprintf("%08x", want) {
		t.Errorf("hash changed: got %s", got)
	}
}
