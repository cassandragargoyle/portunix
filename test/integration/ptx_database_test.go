/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue013_PtxDatabase_Phase1 smoke-tests the end-to-end Phase 1 surface
// of ptx-database. It avoids touching real database services so it can run
// in any environment — only the dispatcher routing, command tree, JSON output
// shape, and the embedded SQLite driver are exercised. PostgreSQL lifecycle
// tests run in containers per ISSUE-DEVELOPMENT-METHODOLOGY.md and are scoped
// to a separate (manual) test run, not this hermetic suite.
//
// Run with:
//
//	go test ./test/integration/ptx_database_test.go ./test/integration/ptx_specpm_test.go -v -timeout 2m
func TestIssue013_PtxDatabase_Phase1(t *testing.T) {
	tf := testframework.NewTestFramework("Issue013_PtxDatabase_Phase1")
	tf.Start(t, "ptx-database Phase 1: dispatcher routing, command tree, JSON shape")

	success := true
	defer func() { tf.Finish(t, success) }()

	root := repoRoot(t, tf)
	binary := filepath.Join(root, "portunix")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if _, err := os.Stat(binary); err != nil {
		tf.Error(t, "portunix binary missing — run 'make build' first", err.Error())
		t.Skipf("portunix binary not built at %s", binary)
		return
	}

	// 1. dispatcher routes "database --help" to ptx-database
	tf.Step(t, "portunix database --help (dispatcher form)")
	out, err := runCmd(binary, "database", "--help")
	if err != nil {
		tf.Error(t, "database --help failed", err.Error()+"\n"+out)
		success = false
		return
	}
	for _, want := range []string{"engines", "install", "status", "backup", "schema"} {
		if !strings.Contains(out, want) {
			tf.Error(t, "database --help missing subcommand", want)
			success = false
		}
	}
	tf.Success(t, "database --help lists Phase 1 subcommands")

	// 2. alias 'db' also routes correctly
	tf.Step(t, "portunix db engines (alias form)")
	out, err = runCmd(binary, "db", "engines", "--format", "json")
	if err != nil {
		tf.Error(t, "db engines failed", err.Error()+"\n"+out)
		success = false
		return
	}
	var engines []string
	if jsonStart := strings.Index(out, "["); jsonStart >= 0 {
		_ = json.Unmarshal([]byte(out[jsonStart:]), &engines)
	}
	hasPostgres, hasSqlite := false, false
	for _, e := range engines {
		if e == "postgresql" {
			hasPostgres = true
		}
		if e == "sqlite" {
			hasSqlite = true
		}
	}
	if !hasPostgres || !hasSqlite {
		tf.Error(t, "engines list missing postgresql or sqlite", out)
		success = false
	} else {
		tf.Success(t, "alias 'db' routes correctly; engines list contains postgresql + sqlite")
	}

	// 3. SQLite end-to-end (no service required) — embedded driver only
	tf.Step(t, "SQLite end-to-end: create file, list tables, dump schema, backup")
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "smoke.db")

	// Create a tiny SQLite database via the sqlite3 CLI directly so the test
	// stays valid even if the helper's create flow is not in Phase 1.
	if !haveSqlite3(tf) {
		tf.Info(t, "sqlite3 CLI not installed — skipping SQLite end-to-end tests")
		tf.Success(t, "ptx-database Phase 1 smoke (SQLite skipped, dispatcher OK)")
		return
	}
	if out, err := runCmd("sqlite3", dbPath, "CREATE TABLE t1 (id INTEGER PRIMARY KEY, name TEXT); INSERT INTO t1(name) VALUES('alice'),('bob');"); err != nil {
		tf.Error(t, "sqlite3 fixture creation failed", err.Error()+"\n"+out)
		success = false
		return
	}

	// status
	out, err = runCmd(binary, "database", "status", "sqlite", "--instance", dbPath, "--format", "json")
	if err != nil {
		tf.Error(t, "database status sqlite failed", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, `"running": true`) {
		tf.Error(t, "expected status.running=true for existing sqlite file", out)
		success = false
	}

	// tables
	out, err = runCmd(binary, "database", "tables", "ignored", "--engine", "sqlite", "--instance", dbPath, "--format", "json")
	if err != nil {
		tf.Error(t, "database tables sqlite failed", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, `"name": "t1"`) {
		tf.Error(t, "tables output missing t1", out)
		success = false
	}

	// schema (sql)
	out, err = runCmd(binary, "database", "schema", "ignored", "--engine", "sqlite", "--instance", dbPath, "--schema-format", "sql")
	if err != nil {
		tf.Error(t, "database schema sqlite failed", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(strings.ToLower(out), "create table") {
		tf.Error(t, "schema output missing CREATE TABLE", out)
		success = false
	}

	// backup
	backupDir := filepath.Join(tmp, "backups")
	out, err = runCmd(binary, "database", "backup", "smoke", "--engine", "sqlite", "--instance", dbPath, "--destination", backupDir, "--format", "json")
	if err != nil {
		tf.Error(t, "database backup sqlite failed", err.Error()+"\n"+out)
		success = false
		return
	}
	var br struct {
		Path      string `json:"path"`
		SizeBytes int64  `json:"size_bytes"`
	}
	if jsonStart := strings.Index(out, "{"); jsonStart >= 0 {
		_ = json.Unmarshal([]byte(out[jsonStart:]), &br)
	}
	if br.Path == "" || br.SizeBytes == 0 {
		tf.Error(t, "backup result did not include path/size_bytes", out)
		success = false
	} else if _, statErr := os.Stat(br.Path); statErr != nil {
		tf.Error(t, "backup file missing on disk", br.Path)
		success = false
	} else {
		tf.Success(t, "SQLite backup wrote a non-empty file at "+br.Path)
	}

	// 4. unknown engine surfaces a clear error
	tf.Step(t, "Unknown engine error path")
	out, _ = runCmd(binary, "database", "status", "no-such-engine")
	if !strings.Contains(out, "unknown engine") {
		tf.Error(t, "expected 'unknown engine' message in error path", out)
		success = false
	} else {
		tf.Success(t, "unknown engine returns a clear error")
	}
}

// haveSqlite3 reports whether the sqlite3 CLI is on PATH for use as a fixture.
func haveSqlite3(tf *testframework.TestFramework) bool {
	_, err := runCmd("sqlite3", "--version")
	return err == nil
}
