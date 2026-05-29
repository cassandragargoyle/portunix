/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package drivers

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// sqliteDriver implements Driver for SQLite. SQLite is embedded — there is no
// service lifecycle; lifecycle methods are no-ops that report status based on
// file existence. The "instance" identifier is treated as the path to a .db
// file. List/tables/schema shell out to the sqlite3 CLI.
type sqliteDriver struct{}

func newSqliteDriver() Driver { return sqliteDriver{} }

func init() { Register(newSqliteDriver()) }

func (sqliteDriver) Engine() string         { return "sqlite" }
func (sqliteDriver) SupportedModes() []Mode { return []Mode{ModeEmbedded} }

func (d sqliteDriver) Status(opts LifecycleOpts) (*Status, error) {
	st := &Status{Engine: d.Engine(), Instance: opts.Instance, Mode: ModeEmbedded}
	if opts.Instance == "" {
		st.Message = "embedded engine — no service lifecycle"
		return st, nil
	}
	if _, err := os.Stat(opts.Instance); err == nil {
		st.Running = true
		st.DataDir = opts.Instance
		if v, err := d.cliVersion(); err == nil {
			st.Version = v
		}
	} else {
		st.Message = fmt.Sprintf("file not found: %s", opts.Instance)
	}
	return st, nil
}

func (sqliteDriver) Start(opts LifecycleOpts) error {
	return fmt.Errorf("sqlite is embedded — no start operation")
}

func (sqliteDriver) Stop(opts LifecycleOpts) error {
	return fmt.Errorf("sqlite is embedded — no stop operation")
}

func (sqliteDriver) Restart(opts LifecycleOpts) error {
	return fmt.Errorf("sqlite is embedded — no restart operation")
}

func (d sqliteDriver) Health(opts LifecycleOpts) (*HealthReport, error) {
	rep := &HealthReport{Engine: d.Engine()}
	v, err := d.cliVersion()
	if err != nil {
		rep.OK = false
		rep.Message = err.Error()
		return rep, nil
	}
	rep.OK = true
	rep.Message = "sqlite3 cli " + v
	return rep, nil
}

func (sqliteDriver) ListDatabases(opts LifecycleOpts) ([]Database, error) {
	if opts.Instance == "" {
		return nil, fmt.Errorf("sqlite: instance must be a path to a .db file")
	}
	info, err := os.Stat(opts.Instance)
	if err != nil {
		return nil, err
	}
	return []Database{{Name: filepath.Base(opts.Instance), Size: humanBytes(info.Size())}}, nil
}

func (d sqliteDriver) ListTables(opts LifecycleOpts, _ string) ([]Table, error) {
	if opts.Instance == "" {
		return nil, fmt.Errorf("sqlite: instance must be a path to a .db file")
	}
	out, err := d.cli(opts.Instance,
		`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name;`)
	if err != nil {
		return nil, err
	}
	var tables []Table
	for _, name := range nonEmptyLines(out) {
		tables = append(tables, Table{Name: name})
	}
	return tables, nil
}

func (d sqliteDriver) DumpSchema(opts LifecycleOpts, _ string, format string) (*SchemaDef, error) {
	if opts.Instance == "" {
		return nil, fmt.Errorf("sqlite: instance must be a path to a .db file")
	}
	if format == "" {
		format = "sql"
	}
	switch format {
	case "sql":
		out, err := d.cli(opts.Instance, ".schema")
		if err != nil {
			return nil, err
		}
		return &SchemaDef{Format: "sql", Body: out}, nil
	case "json":
		// Build a simple JSON schema using PRAGMA table_info per table.
		tables, err := d.ListTables(opts, "")
		if err != nil {
			return nil, err
		}
		var b strings.Builder
		b.WriteString("[")
		for i, t := range tables {
			if i > 0 {
				b.WriteString(",")
			}
			cols, _ := d.cli(opts.Instance, fmt.Sprintf("PRAGMA table_info(%s);", quoteIdent(t.Name)))
			fmt.Fprintf(&b, "{\"table\":%q,\"columns\":[", t.Name)
			for j, line := range nonEmptyLines(cols) {
				if j > 0 {
					b.WriteString(",")
				}
				parts := splitPipe(line)
				name := safeIndex(parts, 1)
				typ := safeIndex(parts, 2)
				fmt.Fprintf(&b, "{\"name\":%q,\"type\":%q}", name, typ)
			}
			b.WriteString("]}")
		}
		b.WriteString("]")
		return &SchemaDef{Format: "json", Body: b.String()}, nil
	default:
		return nil, fmt.Errorf("unsupported format %q (want sql|json)", format)
	}
}

func (d sqliteDriver) Backup(opts LifecycleOpts, b BackupOpts) (*BackupResult, error) {
	if opts.Instance == "" {
		return nil, fmt.Errorf("sqlite: instance must be a path to a .db file")
	}
	dest, err := resolveBackupDir(b.Destination, "sqlite", filepath.Base(opts.Instance))
	if err != nil {
		return nil, err
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	base := strings.TrimSuffix(filepath.Base(opts.Instance), filepath.Ext(opts.Instance))
	out := filepath.Join(dest, fmt.Sprintf("%s-%s.db", base, stamp))

	// Use sqlite3 .backup (online backup) for consistency.
	cmd := exec.Command("sqlite3", opts.Instance, fmt.Sprintf(".backup '%s'", out))
	if combined, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("sqlite3 .backup failed: %w (output: %s)", err, string(combined))
	}
	info, _ := os.Stat(out)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}
	return &BackupResult{
		Path: out, SizeBytes: size, Engine: "sqlite",
		Database: filepath.Base(opts.Instance), Timestamp: stamp,
	}, nil
}

func (d sqliteDriver) Restore(opts LifecycleOpts, r RestoreOpts) error {
	if opts.Instance == "" {
		return fmt.Errorf("sqlite: --instance must be a path to the target .db file")
	}
	if r.File == "" {
		return fmt.Errorf("backup file required")
	}
	src, err := os.Open(r.File)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(opts.Instance)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func (sqliteDriver) cli(dbPath, query string) (string, error) {
	cmd := exec.Command("sqlite3", "-bail", "-noheader", dbPath, query)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (sqliteDriver) cliVersion() (string, error) {
	out, err := exec.Command("sqlite3", "--version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sqlite3 cli not found: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected sqlite3 --version output")
	}
	return parts[0], nil
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
