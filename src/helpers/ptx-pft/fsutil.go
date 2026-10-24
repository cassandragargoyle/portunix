/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeSharedFile writes a file other users may read: project documents shared
// through a git working tree and deploy files mounted into containers
func writeSharedFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644) // #nosec G306 -- see function comment, no credentials are written here
}

// ensureSharedDir creates a directory other users may traverse: project trees
// shared through git and deploy directories mounted into containers
func ensureSharedDir(path string) error {
	return os.MkdirAll(path, 0755) // #nosec G301 -- see function comment, container users must traverse mounted directories
}

// writePrivateFile writes a file that may hold credentials (API tokens, passwords)
// and narrows the mode of a file created earlier with wider permissions
func writePrivateFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

// rewriteExistingFile replaces the content of an existing regular file and keeps
// its mode; the write is confined to the file's own directory through os.Root,
// so it never creates a new path and never follows a symlink out of that directory
func rewriteExistingFile(path string, data []byte) error {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer root.Close()

	name := filepath.Base(path)
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to write %s: not a regular file", name)
	}

	file, err := root.OpenFile(name, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
